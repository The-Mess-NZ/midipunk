package midiport

/*
PortDirection is a type that represents the direction of a MIDI port.
*/
type PortDirection int

const (
	// This port receives MIDI messages from the connected MIDI device.
	INPUT PortDirection = iota
	// This port sends MIDI messages to the connected MIDI device.
	OUTPUT PortDirection = iota
)
