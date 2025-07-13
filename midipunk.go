package main

import (
	"fmt"
	"log"

	"github.com/The-Mess-NZ/midipunk/api"
	"github.com/The-Mess-NZ/midipunk/config"
	"github.com/The-Mess-NZ/midipunk/midiport"
	"github.com/The-Mess-NZ/midipunk/midirouter"

	midi "gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

func main() {
	defer midi.CloseDriver()

	// Print out ins
	fmt.Println("Available MIDI input ports:")
	for _, in := range midi.GetInPorts() {
		fmt.Printf("- %s\n", in)
	}

	// Load configuration from YAML
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
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
			if pc, ok := portMap[in.PortID]; ok {
				ch := in.Channel
				if ch == 0 {
					ch = 1 // default to channel 1 if not set
				}
				inputs = append(inputs, midirouter.NewRouteInput(uint8(ch), pc, in.Label))
			}
		}
		var outputs []*midirouter.RouteOutput
		for _, out := range r.Outputs {
			if pc, ok := portMap[out.PortID]; ok {
				ch := out.Channel
				if ch == 0 {
					ch = 1 // default to channel 1 if not set
				}
				outputs = append(outputs, midirouter.NewRouteOutput(uint8(ch), pc, out.Label))
			}
		}
		route := midirouter.NewRoute(inputs, outputs, r.Label)
		routes = append(routes, route)
	}

	// Create and start the router
	router := midirouter.NewRouter(routes)

	err = router.Start()
	if err != nil {
		log.Fatalf("Failed to start router: %v", err)
	}

	defer router.Stop()

	// Start USB device watcher in a goroutine
	deviceEventChan := make(chan struct{})
	go midiport.UsbDeviceDetect(2, deviceEventChan)

	fmt.Println("MidiPunk router started. Press Ctrl+C to exit...")

	// Listen for device events and print updated MIDI ports
	go func() {
		for range deviceEventChan {
			fmt.Println("MIDI device change detected!")
			fmt.Println("Available MIDI ports:", midi.GetInPorts())
		}
	}()

	go api.StartAPIServer()

	// Keep main thread running
	select {}
}
