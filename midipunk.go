package main

import (
	"fmt"

	midi "gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

func main() {
	defer midi.CloseDriver()
	fmt.Println(midi.GetInPorts())
}
