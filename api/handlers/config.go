package handlers

import (
	"io"
	"net/http"
	"os"

	"github.com/The-Mess-NZ/midipunk/config"
	"gopkg.in/yaml.v3"
)

const configPath = "config.yaml"

func readConfig() (*config.MidiPunkConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var cfg config.MidiPunkConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func writeConfig(cfg *config.MidiPunkConfig) error {
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
	var cfg config.MidiPunkConfig
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
