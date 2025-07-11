package midiport

import (
	"fmt"
	"log"

	midi "gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
	"go.bug.st/serial"
)

/*
PortConfig is a struct that holds the configuration for a MIDI port.

A MIDI port, is a physical device port, either input or output. So, a
USB MIDI device, might have both an input and an output port configured.
*/
type PortConfig struct {
	portType  PortType
	id        string
	path      string
	direction PortDirection
}

// NewUSBPortConfig creates a new PortConfig for a USB MIDI device
func NewUSBPortConfig(id string, direction PortDirection) *PortConfig {
	return &PortConfig{
		portType:  USB,
		id:        id,
		path:      "", // USB ports don't use path
		direction: direction,
	}
}

// NewDINPortConfig creates a new PortConfig for a DIN MIDI device
func NewDINPortConfig(id string, path string, direction PortDirection) *PortConfig {
	return &PortConfig{
		portType:  DIN,
		id:        id,
		path:      path,
		direction: direction,
	}
}

// GetPortType returns the port type
func (pc *PortConfig) GetPortType() PortType {
	return pc.portType
}

// GetID returns the port ID
func (pc *PortConfig) GetID() string {
	return pc.id
}

// GetPath returns the port path
func (pc *PortConfig) GetPath() string {
	return pc.path
}

// GetDirection returns the port direction
func (pc *PortConfig) GetDirection() PortDirection {
	return pc.direction
}

// String returns a string representation of the port config
func (pc *PortConfig) String() string {
	var typeStr, dirStr string

	if pc.portType == USB {
		typeStr = "USB"
	} else {
		typeStr = "DIN"
	}

	if pc.direction == INPUT {
		dirStr = "INPUT"
	} else {
		dirStr = "OUTPUT"
	}

	return fmt.Sprintf("Port[%s %s %s %s]", typeStr, dirStr, pc.id, pc.path)
}

// StartListening starts listening for MIDI messages on this port
func (pc *PortConfig) StartListening(msgChannel chan<- midi.Message) (func(), error) {
	if pc.direction != INPUT {
		return nil, fmt.Errorf("cannot listen on OUTPUT port")
	}

	switch pc.portType {
	case USB:
		return pc.startUSBListening(msgChannel)
	case DIN:
		return pc.startDINListening(msgChannel)
	default:
		return nil, fmt.Errorf("unknown port type")
	}
}

// startUSBListening handles USB MIDI input
func (pc *PortConfig) startUSBListening(msgChannel chan<- midi.Message) (func(), error) {
	// Get the USB MIDI port by ID (converting string to int)
	var portNum int
	fmt.Sscanf(pc.id, "%d", &portNum)

	in, err := midi.InPort(portNum)
	if err != nil {
		return nil, fmt.Errorf("failed to open USB MIDI port %s: %v", pc.id, err)
	}

	stop, err := midi.ListenTo(in, func(msg midi.Message, timestamp int32) {
		fmt.Printf("USB MIDI Message from %s: %s\n", pc.String(), msg)
		msgChannel <- msg
	})

	if err != nil {
		return nil, fmt.Errorf("failed to start listening on USB port %s: %v", pc.id, err)
	}

	return stop, nil
}

// startDINListening handles DIN MIDI input via serial
func (pc *PortConfig) startDINListening(msgChannel chan<- midi.Message) (func(), error) {
	mode := &serial.Mode{
		BaudRate: 38400,
	}

	port, err := serial.Open(pc.path, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to open serial port %s: %v", pc.path, err)
	}

	// Start reading in a goroutine
	go func() {
		defer port.Close()
		buf := make([]byte, 3)

		for {
			n, err := port.Read(buf)
			if err != nil {
				log.Printf("Error reading from serial port %s: %v", pc.path, err)
				return
			}

			if n > 0 {
				msg := midi.Message(buf[:n])
				fmt.Printf("DIN MIDI Message from %s: %s\n", pc.String(), msg)
				msgChannel <- msg
			}
		}
	}()

	// Return a stop function that closes the port
	return func() {
		port.Close()
	}, nil
}
