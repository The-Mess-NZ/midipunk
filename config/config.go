package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type PortConfigYAML struct {
	ID        string `yaml:"id"`
	Type      string `yaml:"type"`
	Device    string `yaml:"device,omitempty"`
	Direction string `yaml:"direction"`
}

type RouteInputYAML struct {
	PortID  string `yaml:"port_id"`
	Channel int    `yaml:"channel"`
}

type RouteOutputYAML struct {
	PortID  string `yaml:"port_id"`
	Channel int    `yaml:"channel"`
}

type RouteYAML struct {
	Inputs  []RouteInputYAML  `yaml:"inputs"`
	Outputs []RouteOutputYAML `yaml:"outputs"`
}

type MidiPunkConfig struct {
	Ports  []PortConfigYAML `yaml:"ports"`
	Routes []RouteYAML      `yaml:"routes"`
}

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
	return &cfg, nil
}
