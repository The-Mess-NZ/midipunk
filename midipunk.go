package main

import (
	"fmt"
	"log"

	midi "gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
	"go.bug.st/serial"
)

// func readSerial(ch chan midi.Message) {
func readSerial() {
	mode := &serial.Mode{
		BaudRate: 38400,
	}

	port, err := serial.Open("/dev/ttyAMA5", mode)

	if err != nil {
		log.Fatal(err)
	}

	defer port.Close()

	buf := make([]byte, 8)

	for {
		n, err := port.Read(buf)

		if err != nil {
			log.Fatal(err)
		}

		if n > 0 {
			fmt.Printf("Received: %x\n", buf[:n])
			// msg := midi.Message(buf[:n])
			// ch <- msg
		}
	}
}

func handleMessage(msg midi.Message, timestamps int32) {
	// Handle the MIDI message
	fmt.Printf("MIDI Message: %s\n", msg)
}

func main() {
	defer midi.CloseDriver()

	fmt.Println(midi.GetInPorts())

	var in, err = midi.InPort(2)

	if err != nil {
		log.Fatal(err)
	}

	stop, _ := midi.ListenTo(in, handleMessage)
	defer stop()

	// Create a communication channel
	// ch := make(chan midi.Message)

	// Start reading from serial port
	go readSerial()

	// Keep main thread running
	select {}
}
