package midirouter

import (
	"fmt"

	midi "gitlab.com/gomidi/midi/v2"
)

/*
Router should be spawned as a "singleton" goroutine.
*/
type Router struct {
	routes     []*Route
	msgChannel chan midi.Message
	stopFuncs  []func()
}

// NewRouter creates a new Router instance
func NewRouter(routes []*Route) *Router {
	return &Router{
		routes:     routes,
		msgChannel: make(chan midi.Message, 100), // Buffered channel
		stopFuncs:  make([]func(), 0),
	}
}

// Start begins the router and starts listening on all input ports
func (r *Router) Start() error {
	// Start listening on all input ports from all routes
	for _, route := range r.routes {
		for _, input := range route.GetInputs() {
			stopFunc, err := input.StartListening(r.msgChannel)
			if err != nil {
				return fmt.Errorf("failed to start listening on input: %v", err)
			}
			r.stopFuncs = append(r.stopFuncs, stopFunc)
		}
	}

	// Start the message handling goroutine
	go r.handleMessages()

	fmt.Printf("Router started with %d routes\n", len(r.routes))
	return nil
}

// Stop stops the router and all listening ports
func (r *Router) Stop() {
	for _, stopFunc := range r.stopFuncs {
		stopFunc()
	}
	close(r.msgChannel)
	fmt.Println("Router stopped")
}

// handleMessages processes incoming MIDI messages and routes them appropriately
func (r *Router) handleMessages() {
	for msg := range r.msgChannel {
		fmt.Printf("Router received message: %s\n", msg)

		// For now, just log the message
		// In a full implementation, this would route messages to appropriate outputs
		// based on the routing rules defined in the routes
	}
}
