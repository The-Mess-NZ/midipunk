package main

import (
	"flag"
	"log"

	"github.com/The-Mess-NZ/midipunk/api"
	"github.com/The-Mess-NZ/midipunk/config"
	"github.com/The-Mess-NZ/midipunk/logger"
	"github.com/The-Mess-NZ/midipunk/midiport"
	"github.com/The-Mess-NZ/midipunk/midirouter"

	midi "gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

func main() {
	defer midi.CloseDriver()

	// Parse command-line arguments
	var logLevel = flag.String("log-level", "", "Set log level (none, error, info, debug, all) - overrides config file")
	flag.Parse()

	// Load configuration from YAML
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		logger.Error("Failed to load config: %v", err)
		log.Fatalf("Failed to load config: %v", err)
	}

	// Determine log level (command line overrides config)
	var finalLogLevel string
	if *logLevel != "" {
		finalLogLevel = *logLevel
	} else {
		finalLogLevel = cfg.LogLevel
	}

	// Initialize logger
	logger.Init(logger.LogLevelFromString(finalLogLevel))
	defer logger.Shutdown()

	logger.Info("MidiPunk starting with log level: %s", finalLogLevel)

	// Print out ports
	if logger.GetLevel() >= logger.INFO {
		logger.Info("Available MIDI input ports:")
		for _, in := range midi.GetInPorts() {
			logger.Info("- %s", in)
		}

		logger.Info("Available MIDI output ports:")
		for _, out := range midi.GetOutPorts() {
			logger.Info("- %s", out)
		}
	}

	// Create port configurations from config
	portMap := make(map[string]*midiport.PortConfig)
	for _, p := range cfg.Ports {
		var pc *midiport.PortConfig
		label := p.Label
		if label == "" && p.Device != "" {
			label = p.Device
		}
		switch p.Type {
		case "USB":
			pc = midiport.NewUSBPortConfig(p.ID, "", midiport.PortDirectionFromString(p.Direction), label)
		case "DIN":
			pc = midiport.NewDINPortConfig(p.ID, p.Device, midiport.PortDirectionFromString(p.Direction), label)
		}
		portMap[p.ID] = pc
	}

	// Create route inputs and outputs from config
	var routes []*midirouter.Route
	for _, r := range cfg.Routes {
		var inputs []*midirouter.RouteInput
		for _, in := range r.Inputs {
			pc, ok := portMap[in.PortID]
			if !ok {
				continue
			}
			ch := in.Channel
			if ch == 0 {
				ch = 1
			}
			inputs = append(inputs, midirouter.NewRouteInput(uint8(ch), pc, in.Label))
		}
		var outputs []*midirouter.RouteOutput
		for _, out := range r.Outputs {
			pc, ok := portMap[out.PortID]
			if !ok {
				continue
			}
			ch := out.Channel
			if ch == 0 {
				ch = 1 // default to channel 1 if not set
			}
			outputs = append(outputs, midirouter.NewRouteOutput(uint8(ch), pc, out.Label))
		}
		route := midirouter.NewRoute(inputs, outputs, r.Label)
		routes = append(routes, route)
	}

	// Create and start the router
	router := midirouter.NewRouter(routes)

	err = router.Start()
	if err != nil {
		logger.Error("Failed to start router: %v", err)
		log.Fatalf("Failed to start router: %v", err)
	}

	defer router.Stop()

	// Start USB device watcher in a goroutine
	deviceEventChan := make(chan struct{})
	go midiport.UsbDeviceDetect(2, deviceEventChan)

	logger.Info("MidiPunk router started. Press Ctrl+C to exit...")

	// Listen for device events and print updated MIDI ports
	go func() {
		for range deviceEventChan {
			logger.Info("MIDI device change detected!")
			logger.Info("Available MIDI ports: %v", midi.GetInPorts())
		}
	}()

	go api.StartAPIServer()

	// Keep main thread running
	select {}
}
