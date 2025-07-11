package midiport

import (
	"log"
	"os"
	"strings"
	"time"
)

/*
Detects changes to connected USB MIDI devices and returns
channels that will each report either connect or disconnect
events, with pertinent device data, such as name, and port number.
*/
// USBDeviceEvent represents a USB MIDI device connection event
// For now, just log details
func UsbDeviceDetect(pollIntervalSeconds int, eventChan chan<- struct{}) {
	watchPath := "/dev/snd"
	prevDevices := map[string]struct{}{}

	for {
		currentDevices := map[string]struct{}{}

		files, err := os.ReadDir(watchPath)
		if err != nil {
			log.Printf("Error reading %s: %v", watchPath, err)
			time.Sleep(time.Duration(pollIntervalSeconds) * time.Second)
			continue
		}

		for _, f := range files {
			if f.IsDir() || !strings.Contains(f.Name(), "midi") {
				continue
			}

			currentDevices[f.Name()] = struct{}{}
		}

		changed := false
		// Detect new devices
		for fname := range currentDevices {
			if _, found := prevDevices[fname]; !found {
				log.Printf("USB MIDI device connected: %s", fname)
				changed = true
			}
		}
		// Detect removed devices
		for fname := range prevDevices {
			if _, found := currentDevices[fname]; !found {
				log.Printf("USB MIDI device disconnected: %s", fname)
				changed = true
			}
		}

		if changed {
			eventChan <- struct{}{}
		}

		prevDevices = currentDevices

		time.Sleep(time.Duration(pollIntervalSeconds) * time.Second)
	}
}
