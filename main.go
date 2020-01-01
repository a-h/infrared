package main

import (
	"fmt"
	"log"
	"time"

	"github.com/stianeikeland/go-rpio/v4"
)

var sleepTime = time.Microsecond * 500

func main() {
	err := rpio.Open()
	if err != nil {
		log.Fatal("GPIO open error")
	}

	irIn := rpio.Pin(4)
	irIn.Input()

	for {
		currentValue := irIn.Read()
		if currentValue == rpio.Low {
			// Take the first value and put it into the array.
			record := make([]int, 64)
			record[0] = int(currentValue)
			time.Sleep(sleepTime)
			//TODO: Read the next 31 values and put it into the array.
			for i := 1; i < len(record); i++ {
				record[i] = int(irIn.Read())
				time.Sleep(sleepTime)
			}
			fmt.Println(record)
		}
		time.Sleep(sleepTime)
	}
}

// Record infrared signals and store them against a remote control key.
// Display the right key when a signal is received.
// Play them back later.
