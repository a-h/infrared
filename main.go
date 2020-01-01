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
		var last time.Time
		freq := time.Microsecond * 50
		for {
			time.Sleep(freq - time.Now().Sub(last))
			ed.Read(d)
			last = time.Now()
		}
	}()

	var op edge.Edges
	for e := range d {
		op = append(op, e)
		if e.Tail {
			// Write it out.
			// fmt.Println(op)
			// Now decode it.
			var bits []int
			d := decoder.New(op)
			for {
				v, ok, err := d.Next()
				if err != nil {
					fmt.Printf("Decoding error: %v\n", err)
					break
				}
				bits = append(bits, v)
				if !ok {
					break
				}
			}

			// Create an integer.
			var n uint64
			for i := len(bits) - 1; i >= 0; i-- {
				if bits[i] == 0 {
					continue
				}
				n |= (1 << i)
			}
			// Look up the key.
			fmt.Println(n, keys[n])

			// Clear the array to make space.
			op = make(edge.Edges, 0)
		}
	}
}

var keys = map[uint64]string{
	502790354903042: "OK",
	439873378983938: "1",
	440977185579010: "2",
	417754448404482: "NetF",
}
