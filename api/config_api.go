package api

import (
	"io"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
)

const configPath = "config.yaml"

type Config struct {
	Ports  []map[string]any `yaml:"ports"`
	Routes []map[string]any `yaml:"routes"`
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
	w.Header().Set("Content-Type", "application/json")
	jsonData, _ := yaml.Marshal(cfg)
	w.Write(jsonData)
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
