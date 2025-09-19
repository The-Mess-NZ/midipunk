package midirouter

import (
	"fmt"

	"github.com/The-Mess-NZ/midipunk/logger"
	"github.com/The-Mess-NZ/midipunk/midiport"
	midi "gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
)

/*
MessageWithSource wraps a MIDI message with information about where it came from
for routing decisions.
*/
type MessageWithSource struct {
	Message    midi.Message
	SourcePort *midiport.PortConfig
	Channel    uint8 // Channel the message came from (if channel-specific input)
}

/*
Router should be spawned as a "singleton" goroutine.
stopFuncs stores individual stop functions for each input port in routes.
senders caches send functions for each output port in routes, to avoid reopening them repeatedly.
*/
type Router struct {
	routes     []*Route
	msgChannel chan *MessageWithSource
	stopFuncs  []func()
	senders    map[string]func(msg midi.Message) error // Cache of send functions by port ID
}

// NewRouter creates a new Router instance
func NewRouter(routes []*Route) *Router {
	return &Router{
		routes:     routes,
		msgChannel: make(chan *MessageWithSource, 100), // Buffered channel
		stopFuncs:  make([]func(), 0),
		senders:    make(map[string]func(msg midi.Message) error),
	}
}

// Start begins the router and starts listening on all input ports
func (router *Router) Start() error {
	// Start listening on all input ports from all routes
	for _, route := range router.routes {
		for _, input := range route.GetInputs() {
			// Create a separate channel for raw MIDI messages from this input
			rawMsgChannel := make(chan midi.Message, 10)

			stopFunc, err := input.StartListening(rawMsgChannel)
			if err != nil {
				return fmt.Errorf("failed to start listening on input: %v", err)
			}
			router.stopFuncs = append(router.stopFuncs, stopFunc)

			// Start a goroutine to wrap messages with source information
			go func(input *RouteInput, rawCh <-chan midi.Message) {
				for msg := range rawCh {
					wrappedMsg := &MessageWithSource{
						Message:    msg,
						SourcePort: input.GetPortConfig(),
						Channel:    input.GetChannel(),
					}
					router.msgChannel <- wrappedMsg
				}
			}(input, rawMsgChannel)
		}
	}

	// Start the message handling goroutine
	go router.handleMessages()

	logger.Info("Router started with %d routes", len(router.routes))
	return nil
}

// Stops the router, all listening (input) ports, and clears cached senders (outputs)
func (router *Router) Stop() {
	for _, stopFunc := range router.stopFuncs {
		stopFunc()
	}
	close(router.msgChannel)

	// Clear the senders cache
	router.senders = make(map[string]func(msg midi.Message) error)

	logger.Info("Router stopped")
}

// Returns a cached sender function for the given port, creating it if needed
func (router *Router) getSender(portConfig *midiport.PortConfig) (func(msg midi.Message) error, error) {
	portID := portConfig.GetID()

	// Check if we already have a sender for this port
	if sender, exists := router.senders[portID]; exists {
		return sender, nil
	}

	// Create a new sender based on port type
	var out drivers.Out
	var err error

	switch portConfig.GetPortType() {
	case midiport.USB:
		// Get the USB MIDI port by ID (converting string to int)
		var portNum int
		fmt.Sscanf(portID, "%d", &portNum)
		out, err = midi.OutPort(portNum)
		if err != nil {
			return nil, fmt.Errorf("failed to open USB MIDI out port %s: %v", portID, err)
		}
	case midiport.DIN:
		// For DIN ports, we'll need to implement serial output
		// For now, return an error as DIN output is not fully implemented
		return nil, fmt.Errorf("DIN output not yet implemented for port %s", portID)
	default:
		return nil, fmt.Errorf("unknown port type for port %s", portID)
	}

	// Create sender function using midi.SendTo
	sender, err := midi.SendTo(out)
	if err != nil {
		return nil, fmt.Errorf("failed to create sender for port %s: %v", portID, err)
	}

	// Cache the sender
	router.senders[portID] = sender

	return sender, nil
}

// shouldRouteMessage determines if a message should be routed from input to output
func (router *Router) shouldRouteMessage(msg *MessageWithSource, routeInput *RouteInput, routeOutput *RouteOutput) bool {
	// Check if the message source matches the route input port
	if msg.SourcePort.GetID() != routeInput.GetPortConfig().GetID() {
		return false
	}

	// If the route input specifies a channel (non-zero), check if the message matches
	if routeInput.GetChannel() != 0 {
		// Extract channel from MIDI message if it's a channel message
		if len(msg.Message) > 0 {
			// Most MIDI channel messages have channel in the lower 4 bits of the first byte
			msgByte := msg.Message[0]
			if msgByte >= 0x80 && msgByte <= 0xEF { // Channel messages
				msgChannel := (msgByte & 0x0F) + 1 // Convert 0-15 to 1-16
				if msgChannel != routeInput.GetChannel() {
					return false
				}
			}
		}
	}

	return true
}

// applyChannelChange modifies a MIDI message to change its channel if needed
func (router *Router) applyChannelChange(msg midi.Message, targetChannel uint8) midi.Message {
	if targetChannel == 0 || len(msg) == 0 {
		return msg // No channel change needed
	}

	// Make a copy of the message
	newMsg := make(midi.Message, len(msg))
	copy(newMsg, msg)

	// Check if this is a channel message
	msgByte := newMsg[0]
	if msgByte >= 0x80 && msgByte <= 0xEF { // Channel messages
		// Clear the channel bits and set the new channel (convert 1-16 to 0-15)
		newMsg[0] = (msgByte & 0xF0) | ((targetChannel - 1) & 0x0F)
	}

	return newMsg
}

// handleMessages processes incoming MIDI messages and routes them appropriately
func (router *Router) handleMessages() {
	for msg := range router.msgChannel {
		// Check if this is a timing message
		isTiming := len(msg.Message) == 1 && (msg.Message[0] == 0xF8 || msg.Message[0] == 0xFA || msg.Message[0] == 0xFB || msg.Message[0] == 0xFC || msg.Message[0] == 0xFE || msg.Message[0] == 0xFF)
		if isTiming {
			logger.DebugMIDITiming("Router received message from %s: %s", msg.SourcePort.GetID(), msg.Message)
		} else {
			logger.DebugMIDI("Router received message from %s: %s", msg.SourcePort.GetID(), msg.Message)
		}

		// Find all routes that should handle this message
		for _, route := range router.routes {
			// Check if any input in this route matches the incoming message
			var matchingInput *RouteInput
			for _, routeInput := range route.GetInputs() {
				if router.shouldRouteMessage(msg, routeInput, nil) {
					matchingInput = routeInput
					break // Only process the first matching input per route
				}
			}

			// If we found a matching input, route to all outputs of this route
			if matchingInput != nil {
				for _, routeOutput := range route.GetOutputs() {
					// Get sender for the output port
					sender, err := router.getSender(routeOutput.GetPortConfig())
					if err != nil {
						logger.Error("Error getting sender for port %s: %v", routeOutput.GetPortConfig().GetID(), err)
						continue
					}

					// Apply channel change if needed
					outputMsg := router.applyChannelChange(msg.Message, routeOutput.GetChannel())

					// Send the message
					err = sender(outputMsg)
					if err != nil {
						logger.Error("Error sending message to port %s: %v", routeOutput.GetPortConfig().GetID(), err)
					} else {
						logger.DebugMIDI("Routed message from %s to %s (channel %d)",
							msg.SourcePort.GetID(), routeOutput.GetPortConfig().GetID(), routeOutput.GetChannel())
					}
				}
			}
		}
	}
}
