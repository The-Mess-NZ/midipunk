package midiport

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

/*
Detects changes to connected USB MIDI devices and returns
channels that will each report either connect or disconnect
events, with pertinent device data, such as serial, and port number.
*/
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

			devPath := watchPath + "/" + f.Name()
			currentDevices[devPath] = struct{}{}
		}

		changed := false
		// Detect new devices
		for devPath := range currentDevices {
			if _, found := prevDevices[devPath]; !found {
				idVendor, idProduct, iSerial, iManufacturer, iProduct, busNum, devNum, clientNum, cardNum, cardDevNum := getSysfsInfo(devPath)

				dev := &USBMIDIDevice{
					IDVendor:      idVendor,
					IDProduct:     idProduct,
					ISerial:       iSerial,
					IManufacturer: iManufacturer,
					IProduct:      iProduct,
					SystemPath:    devPath,
					ClientNum:     clientNum,
					CardNum:       cardNum,
					CardDevNum:    cardDevNum,
					BusNum:        busNum,
					DevNum:        devNum,
				}
				fmt.Printf("%+v\n", dev)
				USBMIDIDeviceRepo[devPath] = dev
				changed = true
			}
		}
		// Detect removed devices
		for devPath := range prevDevices {
			if _, found := currentDevices[devPath]; !found {
				log.Printf("USB MIDI device disconnected: %s", devPath)
				delete(USBMIDIDeviceRepo, devPath)
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

// USBMIDIDevice holds all relevant info for a USB MIDI device
// Fixed info from sysfs, variable info from /dev/snd
type USBMIDIDevice struct {
	// Hardware vendor ID
	IDVendor string
	// Hardware product ID
	IDProduct string
	// Serial number, if available
	ISerial string
	// Manufacturer name, if available
	IManufacturer string
	// Product name, if available
	IProduct string

	// SystemPath is the absolute path to the MIDI device in /dev/snd
	SystemPath string // e.g. /dev/snd/midiC1D0
	// ClientNum is the ALSA client number, e.g. 20
	ClientNum string // e.g. 20
	// CardNum is the ALSA card number, e.g. 1
	CardNum string // e.g. 1
	// CardDevNum is the ALSA device number on the card, e.g. 0
	CardDevNum string // e.g. 0
	// BusNum is the USB bus number, e.g. 001
	BusNum string // e.g. 001
	// DevNum is the USB device number, e.g. 003
	DevNum string // e.g. 003
}

// USBMIDIDeviceRepo is an in-memory repository of connected devices
var USBMIDIDeviceRepo = map[string]*USBMIDIDevice{}

/*
getSysfsInfo tries to find fixed info for a given midi device path, for example the Vendor ID, Product ID, Serial number, etc.
*/
func getSysfsInfo(devPath string) (idVendor, idProduct, iSerial, iManufacturer, iProduct, busNum, devNum string, clientNum string, cardNum string, cardDevNum string) {
	// Example devPath: /dev/snd/midiC1D0
	parts := strings.Split(devPath, "midiC")
	if len(parts) < 2 {
		return
	}
	cardDev := parts[1]
	cardDevParts := strings.Split(cardDev, "D")
	if len(cardDevParts) < 2 {
		return
	}
	cardNum = cardDevParts[0]
	cardDevNum = cardDevParts[1]

	// Find sysfs path for this card/device
	sndPath := "/sys/class/sound/midiC" + cardNum + "D" + cardDevNum

	// Run aconnect -l to find the client number from the card number
	aconnectOut, err := exec.Command("aconnect", "-l").Output()
	if err != nil {
		log.Printf("Error running aconnect -l: %v", err)
		return
	}
	lines := strings.SplitSeq(string(aconnectOut), "\n")
	for line := range lines {
		if strings.Contains(line, "card="+cardNum) {
			// Extract client number from the line
			parts := strings.Split(line, " ")
			for i := range parts {
				if parts[i] == "client" && i+1 < len(parts) {
					clientNum = strings.Trim(parts[i+1], ":")
					break
				}
			}
		}
	}

	/*
		Use readlink to resolve the real device path
		An example of a raw path:
		 - "../../devices/platform/scb/fd500000.pcie/pci0000:00/0000:00:00.0/0000:01:00.0/usb1/1-1/1-1.4/1-1.4:1.0/sound/card1/midiC1D0"
		Where:
			- usb1 is the USB controller (bus 1).
			- 1-1 is the USB hub - "plugged in" to the controller at port 1.
			- 1-1.4 is the USB device plugged in to the hub at port 4
	*/
	realPath, err := os.Readlink(sndPath)
	if err != nil {
		return
	}

	pathParts := strings.Split(realPath, "/")

	// Traverse up to USB device folder
	var usbDevicePath string = ""

	// Retain path to the usb folder, and two more levels up to the device e.g. 1-1.4
	for i := len(pathParts) - 1; i >= 0; i-- {
		if strings.Contains(pathParts[i], "usb") {
			usbDevicePath = strings.Join(pathParts[:i+3], "/")
			break
		}
	}

	// Absolute
	usbDevicePath = strings.Replace(usbDevicePath, "../..", "/sys", 1)

	log.Printf("USB device path: %s", usbDevicePath)

	idVendor = readSysfsFile(usbDevicePath + "/idVendor")
	idProduct = readSysfsFile(usbDevicePath + "/idProduct")
	iSerial = readSysfsFile(usbDevicePath + "/serial")
	iManufacturer = readSysfsFile(usbDevicePath + "/manufacturer")
	iProduct = readSysfsFile(usbDevicePath + "/product")

	busNum = readSysfsFile(usbDevicePath + "/busnum")
	devNum = readSysfsFile(usbDevicePath + "/devnum")

	return
}

func readSysfsFile(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}
