package midirouter

import "github.com/The-Mess-NZ/midipunk/midiport"

/*
A MIDI port output, which is a PortConfig broadcasting messages on the given channel.

This can be used in a Route's outputs.
*/
type RouteOutput struct {
	channel    uint8
	portConfig *midiport.PortConfig
}

// NewRouteOutput creates a new RouteOutput
func NewRouteOutput(channel uint8, portConfig *midiport.PortConfig) *RouteOutput {
	return &RouteOutput{
		channel:    channel,
		portConfig: portConfig,
	}
}

// GetChannel returns the MIDI channel
func (ro *RouteOutput) GetChannel() uint8 {
	return ro.channel
}

// GetPortConfig returns the port configuration
func (ro *RouteOutput) GetPortConfig() *midiport.PortConfig {
	return ro.portConfig
}
