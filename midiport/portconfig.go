package midiport

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/The-Mess-NZ/midipunk/logger"
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
	portType  PortType      // Type of the MIDI port (USB or DIN)
	id        string        // Unique identifier for the port (USB client number or DIN name)
	path      string        // Device path for DIN ports (e.g., serial port path); unused for USB
	direction PortDirection // Specifies if the port is an input or output
	Serial    string        // Serial number for USB devices, used to uniquely identify devices of the same make
	Label     string        // Label for UI display, set from configuration
}

// NewUSBPortConfig creates a new PortConfig for a USB MIDI device
func NewUSBPortConfig(id string, serial string, direction PortDirection, label string) *PortConfig {
	return &PortConfig{
		portType:  USB,
		id:        id,
		path:      "", // USB ports don't use path
		direction: direction,
		Serial:    serial, // set serial number for USB
		Label:     label,
	}
}

// NewDINPortConfig creates a new PortConfig for a DIN MIDI device
func NewDINPortConfig(id string, path string, direction PortDirection, label string) *PortConfig {
	return &PortConfig{
		portType:  DIN,
		id:        id,
		path:      path,
		direction: direction,
		Serial:    "", // DIN ports do not have serial numbers
		Label:     label,
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

// GetPortFullName returns the full name of the port, e.g. 'Oxygen Pro Mini USB MIDI'
func (pc *PortConfig) GetPortFullName() string {
	if pc.portType == USB {
		// Try to get full name from aconnect -i
		aconnectOut, err := exec.Command("aconnect", "-i").Output()
		if err == nil {
			lines := strings.Split(string(aconnectOut), "\n")
			clientNum := pc.id
			for _, line := range lines {
				if !strings.Contains(line, "client "+clientNum+":") {
					continue
				}
				// Next lines are ports for this client
				for i := 1; i <= 4; i++ {
					if len(lines) > i {
						portLine := lines[i]
						if strings.Contains(portLine, "'") {
							name := strings.TrimSpace(strings.Split(portLine, "'")[1])
							return name
						}
					}
				}
			}
		}
		if pc.Serial != "" {
			return pc.Serial
		}
	}
	return pc.Label
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
		// Check if this is a timing message (Clock, Start, Stop, Continue, Active Sensing, Reset)
		if msg.Is(midi.TimingClockMsg) {
			logger.DebugMIDITiming("USB MIDI Message from %s: %s", pc.String(), msg)
		} else {
			logger.DebugMIDI("USB MIDI Message from %s: %s", pc.String(), msg)
		}
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
				logger.Error("Error reading from serial port %s: %v", pc.path, err)
				return
			}

			if n > 0 {
				// Create a copy of the buffer to avoid reuse issues
				msgBytes := make([]byte, n)
				copy(msgBytes, buf[:n])
				msg := midi.Message(msgBytes)
				// Check if this is a timing message
				if msg.Is(midi.TimingClockMsg) {
					logger.DebugMIDITiming("DIN MIDI Message from %s: %s", pc.String(), msg)
				} else {
					logger.DebugMIDI("DIN MIDI Message from %s: %s", pc.String(), msg)
				}
				msgChannel <- msg
			}
		}
	}()

	// Return a stop function that closes the port
	return func() {
		port.Close()
	}, nil
}

// CreateDINSender creates a sender function for DIN MIDI output via serial
func (pc *PortConfig) CreateDINSender() (func(msg midi.Message) error, error) {
	if pc.portType != DIN {
		return nil, fmt.Errorf("cannot create DIN sender for non-DIN port")
	}
	if pc.direction != OUTPUT {
		return nil, fmt.Errorf("cannot create sender for INPUT port")
	}

	mode := &serial.Mode{
		BaudRate: 38400,
	}

	port, err := serial.Open(pc.path, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to open serial port %s for output: %v", pc.path, err)
	}

	// Create the sender function
	sender := func(msg midi.Message) error {
		if len(msg) == 0 {
			return nil // Nothing to send
		}

		// Write MIDI message bytes to serial port
		n, err := port.Write([]byte(msg))
		if err != nil {
			return fmt.Errorf("failed to write to serial port %s: %v", pc.path, err)
		}

		// Check if this is a timing message for appropriate logging
		if msg.Is(midi.TimingClockMsg) {
			logger.DebugMIDITiming("DIN MIDI Message sent to %s (%d bytes): %s", pc.String(), n, msg)
		} else {
			logger.DebugMIDI("DIN MIDI Message sent to %s (%d bytes): %s", pc.String(), n, msg)
		}

		return nil
	}

	return sender, nil
}
