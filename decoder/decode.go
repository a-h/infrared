package decoder

import (
	"errors"
	"fmt"
	"time"

	"github.com/a-h/infrared/edge"
)

// New creates a new Decoder.
func New(edges edge.Edges) *Decoder {
	return &Decoder{
		edges: edges,
	}
}

// Decoder decdes edges to values.
type Decoder struct {
	edges edge.Edges
	index int
}

/*
https://github.com/z3t0/Arduino-IRremote/blob/master/ir_Panasonic.cpp
#define PANASONIC_BITS          48
#define PANASONIC_HDR_MARK    3502
#define PANASONIC_HDR_SPACE   1750
#define PANASONIC_BIT_MARK     502
#define PANASONIC_ONE_SPACE   1244
#define PANASONIC_ZERO_SPACE   400
*/

// Next continues decoding the edges.
func (d *Decoder) Next() (value int, ok bool, err error) {
	if len(d.edges)%2 != 0 {
		//TODO: Error on that because it should be even. But not just yet.
	}
	if d.index >= len(d.edges) {
		return
	}
	if d.index == 0 {
		// Read the low part of the header.
		if d.edges[d.index].Value == false {
			err = errors.New("header mark does not match Panasonic header value")
			return
		}
		if !between(time.Microsecond*1751, time.Microsecond*5253, d.edges[d.index].Duration) {
			err = fmt.Errorf("expected header mark at %d to last 3502µs, but was %v", d.index, d.edges[d.index].Duration)
			return
		}
		d.index++
		// Read the high part of the header.
		if d.edges[d.index].Value == true {
			err = errors.New("header space does not match Panasonic header value")
			return
		}
		if !between(time.Microsecond*875, time.Microsecond*2625, d.edges[d.index].Duration) {
			err = fmt.Errorf("expected header space at %d to last 1750µs, but was %v", d.index, d.edges[d.index].Duration)
			return
		}
		d.index++
		return d.Next()
	}
	// Read the bit mark.
	if d.edges[d.index].Value != true {
		err = fmt.Errorf("expected bit mark at %d to be low", d.index)
		return
	}
	if !between(time.Microsecond*251, time.Microsecond*753, d.edges[d.index].Duration) {
		err = fmt.Errorf("expected bit mark at %d to last 502µs, but was %v", d.index, d.edges[d.index].Duration)
		return
	}
	// Read the space mark to work out whether it's a zero or one.
	d.index++
	if d.edges[d.index].Value != false {
		err = fmt.Errorf("expected bit space at %d to be high", d.index)
		return
	}
	// If it's 1244µs, then it's a one.
	if between(time.Microsecond*622, time.Microsecond*1866, d.edges[d.index].Duration) {
		value = 1
		ok = true
		d.index++
		return
	}
	// If it's 400µs, then it's a zero.
	if between(time.Microsecond*200, time.Microsecond*600, d.edges[d.index].Duration) {
		value = 0
		ok = true
		d.index++
		return
	}
	if d.edges[d.index].Tail && d.edges[d.index].Value == false {
		value = 0
		ok = true
		d.index++
		return
	}
	err = fmt.Errorf("expected bit space mark at %d to be 400µs for a zero, or 1244µs for a one, but was %v", d.index, d.edges[d.index].Duration)
	return
}

func between(min, max, value time.Duration) bool {
	if value > max {
		return false
	}
	if value < min {
		return false
	}
	return true
}
