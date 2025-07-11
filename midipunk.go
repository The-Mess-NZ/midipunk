package main

import (
	"fmt"
	"log"

	"github.com/The-Mess-NZ/midipunk/midiport"
	"github.com/The-Mess-NZ/midipunk/midirouter"

	midi "gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

func main() {
	defer midi.CloseDriver()

	fmt.Println("Available MIDI ports:", midi.GetInPorts())

	// Create port configurations
	usbInputPort := midiport.NewUSBPortConfig("1", midiport.INPUT)
	dinInputPort := midiport.NewDINPortConfig("din1", "/dev/ttyAMA2", midiport.INPUT)

	fmt.Printf("Created USB port: %s\n", usbInputPort.String())
	fmt.Printf("Created DIN port: %s\n", dinInputPort.String())

	// Create route inputs
	usbInput := midirouter.NewRouteInput(1, usbInputPort)
	dinInput := midirouter.NewRouteInput(1, dinInputPort)

	// Create a simple route that takes both inputs (no outputs for now)
	route := midirouter.NewRoute([]*midirouter.RouteInput{usbInput, dinInput}, []*midirouter.RouteOutput{})

	// Create and start the router
	router := midirouter.NewRouter([]*midirouter.Route{route})

	err := router.Start()
	if err != nil {
		log.Fatalf("Failed to start router: %v", err)
	}

	defer router.Stop()

	// Start USB device watcher in a goroutine
	go midiport.UsbDeviceDetect()

	fmt.Println("MidiPunk router started. Press Ctrl+C to exit...")

	// Keep main thread running
	select {}
}
