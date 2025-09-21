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
	PortType  PortType      // Type of the MIDI port (USB or DIN)
	ID        string        // Unique identifier for the port (USB client number or DIN name)
	Direction PortDirection // Specifies if the port is an input or output
	Serial    string        // Serial number for USB devices, used to uniquely identify devices of the same make
	Label     string        // Label for UI display, set from configuration
	dinPort   DinPort       // The DIN port (only relevant if PortType is DIN), set from ID
}

type DinPort int

const (
	DIN1 DinPort = 1
	DIN2 DinPort = 2
	DIN3 DinPort = 3
	DIN4 DinPort = 4
)

/*
String returns a string representation of the DIN port, e.g. "din1"
*/
func (d DinPort) String() string {
	return fmt.Sprintf("DIN%d", d)
}

/*
ParseDinPort parses a string like "din1" into a DinPort constant.
*/
func ParseDinPort(s string) (DinPort, error) {
	switch strings.ToLower(s) {
	case "din1":
		return DIN1, nil
	case "din2":
		return DIN2, nil
	case "din3":
		return DIN3, nil
	case "din4":
		return DIN4, nil
	default:
		return 0, fmt.Errorf("invalid DIN port: %s", s)
	}
}

/*
GetSelectPin returns the GPIO pin used to select INPUT/OUTPUT mode for the DIN port.
*/
func (d DinPort) GetSelectPin() DinSelectPin {
	switch d {
	case DIN1:
		return DIN1_SELECT_PIN
	case DIN2:
		return DIN2_SELECT_PIN
	case DIN3:
		return DIN3_SELECT_PIN
	case DIN4:
		return DIN4_SELECT_PIN
	default:
		return -1 // Invalid
	}
}

/*
GetPath returns the device path for the DIN port:
  - /dev/ttyAMA0 for DIN1
  - /dev/ttyAMA2 for DIN2
  - /dev/ttyAMA3 for DIN3
  - /dev/ttyAMA5 for DIN4
*/
func (d DinPort) GetPath() string {
	switch d {
	case DIN1:
		return "/dev/ttyAMA0"
	case DIN2:
		return "/dev/ttyAMA2"
	case DIN3:
		return "/dev/ttyAMA3"
	case DIN4:
		return "/dev/ttyAMA5"
	default:
		return "" // Invalid
	}
}

/*
DinSelectPin represents the GPIO pin used to select whether the DIN is INPUT (LOW) or OUTPUT (HIGH).
*/
type DinSelectPin int

const (
	DIN1_SELECT_PIN DinSelectPin = 6
	DIN2_SELECT_PIN DinSelectPin = 19
	DIN3_SELECT_PIN DinSelectPin = 16
	DIN4_SELECT_PIN DinSelectPin = 26
)

/*
Whether the port is an input or output. Contains the drive mode for the select GPIO pin for raspi-gpio.
Output mode sets the pin LOW, to turn off the optocoupler, and enable the output buffer.
*/
type DinMode string

const (
	DIN_MODE_INPUT  DinMode = "op dh"
	DIN_MODE_OUTPUT DinMode = "op dl"
)

// NewUSBPortConfig creates a new PortConfig for a USB MIDI device
func NewUSBPortConfig(id string, serial string, direction PortDirection, label string) *PortConfig {
	return &PortConfig{
		PortType:  USB,
		ID:        id,
		Direction: direction,
		Serial:    serial, // set serial number for USB
		Label:     label,
		dinPort:   DIN1, // not used for USB
	}
}

// NewDINPortConfig creates a new PortConfig for a DIN MIDI device
func NewDINPortConfig(id string, direction PortDirection, label string) (*PortConfig, error) {
	dinPort, err := ParseDinPort(id)
	if err != nil {
		return nil, err
	}

	return &PortConfig{
		PortType:  DIN,
		ID:        id,
		Direction: direction,
		Serial:    "", // DIN ports do not have serial numbers
		Label:     label,
		dinPort:   dinPort,
	}, nil
}

// GetPortType returns the port type
func (pc *PortConfig) GetPortType() PortType {
	return pc.PortType
}

// GetID returns the port ID
func (pc *PortConfig) GetID() string {
	return pc.ID
}

// GetPath returns the port path
func (pc *PortConfig) GetPath() (string, error) {
	if pc.PortType == USB {
		return "", nil // USB ports do not have a path
	}

	// For DIN, parse the ID to get the path
	path, err := ParseDinPort(pc.ID)

	if err != nil {
		return "", err
	}

	return path.GetPath(), nil
}

// GetDirection returns the port direction
func (pc *PortConfig) GetDirection() PortDirection {
	return pc.Direction
}

// String returns a string representation of the port config
func (pc *PortConfig) String() string {
	var typeStr, dirStr string

	if pc.PortType == USB {
		typeStr = "USB"
	} else {
		typeStr = "DIN"
	}

	if pc.Direction == INPUT {
		dirStr = "INPUT"
	} else {
		dirStr = "OUTPUT"
	}

	return fmt.Sprintf("Port[%s %s %s %s]", typeStr, dirStr, pc.ID, pc.Label)
}

// GetPortFullName returns the full name of the port, e.g. 'Oxygen Pro Mini USB MIDI'
func (pc *PortConfig) GetPortFullName() string {
	if pc.PortType == USB {
		// Try to get full name from aconnect -i
		aconnectOut, err := exec.Command("aconnect", "-i").Output()
		if err == nil {
			lines := strings.Split(string(aconnectOut), "\n")
			clientNum := pc.ID
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
	if pc.Direction != INPUT {
		return nil, fmt.Errorf("cannot listen on OUTPUT port")
	}

	switch pc.PortType {
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
	fmt.Sscanf(pc.ID, "%d", &portNum)

	in, err := midi.InPort(portNum)
	if err != nil {
		return nil, fmt.Errorf("failed to open USB MIDI port %s: %v", pc.ID, err)
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
		return nil, fmt.Errorf("failed to start listening on USB port %s: %v", pc.ID, err)
	}

	return stop, nil
}

// startDINListening handles DIN MIDI input via serial
func (pc *PortConfig) startDINListening(msgChannel chan<- midi.Message) (func(), error) {
	mode := &serial.Mode{
		BaudRate: 38400,
	}

	// Set GPIO pin to INPUT mode for listening
	selectPin := pc.dinPort.GetSelectPin()
	if selectPin == -1 {
		return nil, fmt.Errorf("invalid DIN port ID: %s", pc.ID)
	}

	args := append([]string{"set", fmt.Sprintf("%d", selectPin)}, strings.Fields(string(DIN_MODE_OUTPUT))...)
	cmd := exec.Command("pinctrl", args...)
	if err := cmd.Run(); err != nil {
		logger.Error("Failed to set GPIO pin %d to OUTPUT HIGH for DIN %s: %v", selectPin, pc.ID, err)
	} else {
		logger.Info("Set GPIO pin %d to OUTPUT HIGH for DIN %s", selectPin, pc.ID)
	}

	path := pc.dinPort.GetPath()

	port, err := serial.Open(path, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to open serial port %s: %v", path, err)
	}

	// Start reading in a goroutine
	go func() {
		defer port.Close()
		buf := make([]byte, 3)

		for {
			n, err := port.Read(buf)
			if err != nil {
				logger.Error("Error reading from serial port %s: %v", path, err)
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
	if pc.PortType != DIN {
		return nil, fmt.Errorf("cannot create DIN sender for non-DIN port")
	}
	if pc.Direction != OUTPUT {
		return nil, fmt.Errorf("cannot create sender for INPUT port")
	}

	mode := &serial.Mode{
		BaudRate: 38400,
	}

	selectPin := pc.dinPort.GetSelectPin()
	if selectPin == -1 {
		return nil, fmt.Errorf("invalid DIN port ID: %s", pc.ID)
	}

	// Set GPIO pin to OUTPUT HIGH for sending
	args := append([]string{"set", fmt.Sprintf("%d", selectPin)}, strings.Fields(string(DIN_MODE_OUTPUT))...)
	cmd := exec.Command("/usr/bin/pinctrl", args...)
	if err := cmd.Run(); err != nil {
		logger.Error("Failed to set GPIO pin %d to OUTPUT LOW for DIN %s: %v", selectPin, pc.ID, err)
	} else {
		logger.Info("Set GPIO pin %d to OUTPUT LOW for DIN %s", selectPin, pc.ID)
	}

	// Open the serial port
	port, err := serial.Open(pc.dinPort.GetPath(), mode)
	if err != nil {
		return nil, fmt.Errorf("failed to open serial port %s for output: %v", pc.dinPort.GetPath(), err)
	}

	// Create the sender function
	sender := func(msg midi.Message) error {
		if len(msg) == 0 {
			return nil // Nothing to send
		}

		// Write MIDI message bytes to serial port
		n, err := port.Write([]byte(msg))
		if err != nil {
			return fmt.Errorf("failed to write to serial port %s: %v", pc.dinPort.GetPath(), err)
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
