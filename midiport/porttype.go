package midiport

/*
PortType is a type that represents the type of MIDI port.

A MIDI port is a physical device port, either USB or DIN.
*/
type PortType int

const (
	// A 5-pin DIN or 3.5mm TRS MIDI port running at 31.25kbaud.
	DIN PortType = iota
	// A USB device port with a USB MIDI complaint device attached.
	USB PortType = iota
)
