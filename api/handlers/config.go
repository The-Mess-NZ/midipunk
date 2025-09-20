package handlers

import (
	"io"
	"net/http"
	"os"

	"github.com/The-Mess-NZ/midipunk/logger"
	"gopkg.in/yaml.v3"
)

const configPath = "config.yaml"

// A defined MIDI port in the configuration. Must be defined here to be used in routes.
type Port struct {
	Id        string `yaml:"id"`
	Type      string `yaml:"type"`
	Direction string `yaml:"direction"`
	Label     string `yaml:"label,omitempty"`
}

// A defined MIDI route in the configuration.
type Route struct {
	Label   string    `yaml:"label,omitempty"`
	Enabled bool      `yaml:"enabled"`
	Inputs  []RouteIO `yaml:"inputs"`
	Outputs []RouteIO `yaml:"outputs"`
}

// RouteIO represents a MIDI port and optional channel for a route input or output.
type RouteIO struct {
	PortID  string `yaml:"portId"`
	Channel int    `yaml:"channel,omitempty"`
}

// Config represents the root configuration file structure.
type Config struct {
	Label    string          `yaml:"label"`
	LogLevel logger.LogLevel `yaml:"logLevel"`
	Ports    []Port          `yaml:"ports"`
	Routes   []Route         `yaml:"routes"`
}

func readConfig() (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func writeConfig(cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

func ConfigGetHandler(w http.ResponseWriter, r *http.Request) {
	cfg, err := readConfig()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/x-yaml")
	data, _ := yaml.Marshal(cfg)
	w.Write(data)
}

func ConfigPutHandler(w http.ResponseWriter, r *http.Request) {
	var cfg Config
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if err := yaml.Unmarshal(body, &cfg); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if err := writeConfig(&cfg); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
}
