package midirouter

import (
	"github.com/The-Mess-NZ/midipunk/midiport"
	midi "gitlab.com/gomidi/midi/v2"
)

/*
A MIDI route input, which is a PortConfig listening on a specific MIDI channel.

This can be used in a Route's inputs.
*/
type RouteInput struct {
	channel    uint8
	portConfig *midiport.PortConfig
}

// NewRouteInput creates a new RouteInput
func NewRouteInput(channel uint8, portConfig *midiport.PortConfig) *RouteInput {
	return &RouteInput{
		channel:    channel,
		portConfig: portConfig,
	}
}

// GetChannel returns the MIDI channel
func (ri *RouteInput) GetChannel() uint8 {
	return ri.channel
}

// GetPortConfig returns the port configuration
func (ri *RouteInput) GetPortConfig() *midiport.PortConfig {
	return ri.portConfig
}

// StartListening starts listening for MIDI messages on this input
func (ri *RouteInput) StartListening(msgChannel chan<- midi.Message) (func(), error) {
	return ri.portConfig.StartListening(msgChannel)
}
