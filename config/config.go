package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// MidiPunkConfig represents the main configuration for the MIDI router.
// It is loaded from a YAML file and contains port and routing definitions.
//   - LogLevel: Controls the verbosity of logging output. Default is "NONE".
//   - Label: Required field to identify the configuration in the UI.
//   - ProgramNumber: Optional field to assign a program change number to the configuration.
//   - Ports: Defines all available MIDI ports (USB or DIN) and their properties.
//   - Routes: Defines how MIDI messages are routed between ports and channels.
type MidiPunkConfig struct {
	LogLevel      string           `yaml:"logLevel"`
	Label         string           `yaml:"label"`
	ProgramNumber int              `yaml:"programNumber,omitempty"`
	Ports         []PortConfigYAML `yaml:"ports"`
	Routes        []RouteYAML      `yaml:"routes"`
}

// PortConfigYAML describes a single MIDI port's configuration.
//   - ID: Unique identifier for the port (used in routing).
//   - Type: Port type (e.g., "usb", "din").
//   - Serial: Serial number for USB devices, used to uniquely identify devices of the same make.
//   - Device: Optional device path (for DIN) or name (for USB).
//   - Direction: "input" or "output"; determines if port receives or sends MIDI.
//   - Label: Optional custom label for the port, defaults to device name if available.
type PortConfigYAML struct {
	ID        string `yaml:"id"`
	Type      string `yaml:"type"`
	Serial    string `yaml:"serial,omitempty"`
	Device    string `yaml:"device,omitempty"`
	Direction string `yaml:"direction"`
	Label     string `yaml:"label,omitempty"`
}

// RouteInputYAML specifies a source port and MIDI channel for routing.
// PortID: The ID of the input port.
// Channel: MIDI channel number (1-16) to listen to.
// Label: Optional label for UI display.
type RouteInputYAML struct {
	PortID  string `yaml:"portId"`
	Channel int    `yaml:"channel"`
	Label   string `yaml:"label,omitempty"`
}

// RouteOutputYAML specifies a destination port and MIDI channel for routing.
//   - PortID: The ID of the output port.
//   - Channel: MIDI channel number (1-16) to send to.
//   - Label: Optional label for UI display.
type RouteOutputYAML struct {
	PortID  string `yaml:"portId"`
	Channel int    `yaml:"channel"`
	Label   string `yaml:"label,omitempty"`
}

// RouteYAML defines a routing rule from one or more inputs to one or more outputs.
//   - Inputs: List of input ports/channels.
//   - Outputs: List of output ports/channels.
//   - Label: Optional label to describe the route's purpose.
type RouteYAML struct {
	Inputs  []RouteInputYAML  `yaml:"inputs"`
	Outputs []RouteOutputYAML `yaml:"outputs"`
	Label   string            `yaml:"label,omitempty"`
	Enabled *bool             `yaml:"enabled,omitempty"`
}

// LoadConfig loads the MidiPunk configuration from a YAML file at the given path.
// Returns a MidiPunkConfig struct or an error if loading fails.
func LoadConfig(path string) (*MidiPunkConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg MidiPunkConfig
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	// Set default port labels if not provided
	for i := range cfg.Ports {
		if cfg.Ports[i].Label == "" && cfg.Ports[i].Device != "" {
			cfg.Ports[i].Label = cfg.Ports[i].Device
		}
	}
	for i := range cfg.Routes {
		if cfg.Routes[i].Enabled == nil {
			enabled := true
			cfg.Routes[i].Enabled = &enabled
		}
	}

	return &cfg, nil
}

// IsEnabled returns true when a route is enabled or omitted from the YAML.
func (r RouteYAML) IsEnabled() bool {
	return r.Enabled == nil || *r.Enabled
}

// SetEnabled updates the route enabled state in memory.
// TODO: We should send an "all notes off" message to all outputs when disabling a route to prevent stuck notes. Perhaps optionally.
func (r *RouteYAML) SetEnabled(enabled bool) {
	r.Enabled = &enabled
}
