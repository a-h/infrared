package main

import (
	"fmt"
	"log"
	"time"

	"github.com/a-h/infrared/decoder"
	"github.com/a-h/infrared/edge"
	"github.com/stianeikeland/go-rpio/v4"
)

func main() {
	err := rpio.Open()
	if err != nil {
		log.Fatal("GPIO open error")
	}

	irIn := rpio.Pin(4)
	irIn.Input()

	ed := edge.NewDetector(irIn, time.Millisecond*10)

	// Create a channel to receive the data.
	d := make(chan edge.Edge)

	// Start reading in another thread.
	go func() {
		for {
			ed.Read(d)
			time.Sleep(time.Microsecond * 50)
		}
	}()

	var op edge.Edges
	for e := range d {
		op = append(op, e)
		if e.Tail {
			// Write it out.
			fmt.Println(op)
			// Now decode it.
			d := decoder.New(op)
			for {
				v, ok, err := d.Next()
				if err != nil {
					fmt.Println()
					fmt.Printf("Decoding error: %v\n", err)
					break
				}
				if !ok {
					break
				}
				fmt.Printf("%v", v)
			}
			fmt.Println()

			// Clear the array to make space.
			op = make(edge.Edges, 0)
		}
	}
}
