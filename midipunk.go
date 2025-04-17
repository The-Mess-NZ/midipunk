package main

import (
	"fmt"
	"log"

	midi "gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
	"go.bug.st/serial"
)

func readSerial(ch chan midi.Message) {
	mode := &serial.Mode{
		BaudRate: 38400,
	}

	port, err := serial.Open("/dev/ttyAMA2", mode)

	if err != nil {
		log.Fatal(err)
	}

	defer port.Close()

	buf := make([]byte, 3)

	for {
		n, err := port.Read(buf)

		if err != nil {
			log.Fatal(err)
		}

		if n > 0 {
			// fmt.Printf("Received: %x\n", buf[:n])
			msg := midi.Message(buf[:n])
			fmt.Printf("Received Serial Message: %s\n", msg)
			ch <- msg
		}
	}
}

func handleUsbMessage(msg midi.Message, timestamps int32, ch chan midi.Message) {
	// Handle the MIDI message
	fmt.Printf("MIDI Message: %s\n", msg)

	ch <- msg
}

func handleChannelMessage(ch chan midi.Message) {
	for msg := range ch {
		// Process the message received from the channel
		fmt.Printf("Channel Message: %s\n", msg)
	}
}

func main() {
	defer midi.CloseDriver()

	fmt.Println(midi.GetInPorts())

	var in, err = midi.InPort(2)

	if err != nil {
		log.Fatal(err)
	}

	// Create a communication channel
	ch := make(chan midi.Message)

	stop, _ := midi.ListenTo(
		in,
		func(msg midi.Message, timestamps int32) {
			handleUsbMessage(msg, timestamps, ch)
		},
	)
	defer stop()

	// Start reading from serial port
	go readSerial(ch)

	// Start a goroutine to handle messages from the channel
	go handleChannelMessage(ch)

	// Keep main thread running
	select {}
}
