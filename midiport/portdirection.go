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

// PortDirectionFromString converts a string to a PortDirection.
func PortDirectionFromString(s string) PortDirection {
	switch s {
	case "INPUT":
		return INPUT
	case "OUTPUT":
		return OUTPUT
	default:
		return INPUT // default to INPUT if unknown
	}
}
