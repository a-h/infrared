package main

import (
	"context"
	"fmt"
	"log"

	"github.com/a-h/infrared/decoder"
	"github.com/a-h/infrared/edge"
	"github.com/stianeikeland/go-rpio/v4"
)

func main() {
	err := rpio.Open()
	defer rpio.Close()
	if err != nil {
		log.Fatal("GPIO open error")
	}

	irIn := rpio.Pin(4)
	irIn.Input()

	ed := edge.NewDetector(irIn)

	// Create a channel to receive codes from the IR.
	// Use a buffered channel to avoid failing to handle IR if the processing code is slow to process the results.
	codes := make(chan uint64, 64)

	// Start a routine to receive IR codes.
	go func() {
		// It's important to close the channel if the context is cancelled or this routine will exit abruptly.
		for code := range codes {
			// This program expects Panasonic codes so we can look up the names and print them.
			fmt.Println(code, decoder.PanasonicCodeToKey[code])
		}
	}()

	// Start decoding.
	ed.Decode(context.Background(), decoder.Panasonic, codes)
	close(codes)
}
