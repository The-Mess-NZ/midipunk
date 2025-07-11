package midiport

import (
	"log"
	"os"
	"time"
)

/*
Detects changes to connected USB MIDI devices and returns
channels that will each report either connect or disconnect
events, with pertinent device data, such as name, and port number.
*/
// USBDeviceEvent represents a USB MIDI device connection event
// For now, just log details
func UsbDeviceDetect() {
	watchPath := "/dev/snd"
	prevDevices := map[string]os.FileInfo{}

	for {
		currentDevices := map[string]os.FileInfo{}
		// List all files in /dev/snd
		files, err := os.ReadDir(watchPath)
		if err != nil {
			log.Printf("Error reading %s: %v", watchPath, err)
			time.Sleep(2 * time.Second)
			continue
		}
		for _, f := range files {
			if !f.IsDir() {
				info, err := f.Info()
				if err == nil {
					currentDevices[f.Name()] = info
				}
			}
		}
		// Detect new devices
		for name, info := range currentDevices {
			if _, found := prevDevices[name]; !found {
				log.Printf("USB MIDI device connected: %s (size: %d bytes)", name, info.Size())
			}
		}
		// Detect removed devices
		for name := range prevDevices {
			if _, found := currentDevices[name]; !found {
				log.Printf("USB MIDI device disconnected: %s", name)
			}
		}
		prevDevices = currentDevices
		// Poll every 2 seconds
		time.Sleep(2 * time.Second)
	}
}
