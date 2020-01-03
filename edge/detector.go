package edge

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/stianeikeland/go-rpio/v4"
)

// An Edge is a change in state.
type Edge struct {
	Value    bool
	Duration time.Duration
	Tail     bool
}

func (e Edge) String() string {
	var v int
	if e.Value {
		v = 1
	}
	return fmt.Sprintf("%v, %v, %v", v, e.Duration, e.Tail)
}

// Edges is all of the data.
type Edges []Edge

// Values of the edges.
func (e Edges) Values() []bool {
	op := make([]bool, len(e))
	for i := 0; i < len(e); i++ {
		op[i] = e[i].Value
	}
	return op
}

func (e Edges) String() string {
	var sb strings.Builder
	for i, ee := range e {
		sb.WriteString(fmt.Sprintf("%d, ", i))
		sb.WriteString(ee.String())
		sb.WriteRune('\n')
	}
	sb.WriteString(fmt.Sprintf("%d edges\n", len(e)))
	return sb.String()
}

// Reader of a pin. For example, an rpio.Pin.
type Reader interface {
	Read() rpio.State
}

// DefaultTailTimeout is the default wait time of the end of a signal.
const DefaultTailTimeout = 10 * time.Millisecond

// DefaultBufferSize of edges stored in RAM. Set to 64 bits * 2 (high/low), + 32 extra for start / stop.
const DefaultBufferSize = 128 + 32

// NewDetector creates a new edge detector.
// Can take in an rpio.Pin.
func NewDetector(pin Reader) *Detector {
	now := time.Now()
	return &Detector{
		Timeout: DefaultTailTimeout,
		Buffer:  make(Edges, DefaultBufferSize),
		now:     time.Now,
		pin:     pin,
		t:       now,
		first:   true,
	}
}

// A Detector detects changes in a pin by sampling the pin value.
type Detector struct {
	// Timeout is the maximum the detector will wait for a state change. Used to detect the end
	// of the transmission.
	Timeout time.Duration
	// Now is a function which returns the current time.
	now func() time.Time
	// Pin that we're reading from.
	// Can take in an rpio.Pin.
	pin Reader

	// Buffer of stored edges.
	Buffer Edges
	// Index within the buffer.
	index int

	// Previous Value of the pin.
	pv bool
	// Time of the last pin sample.
	t time.Time
	// First sample taken?
	first bool
}

// Decode infrared signals using the decoder d. The values are output to c.
func (r *Detector) Decode(ctx context.Context, d func(e Edges) (uint64, error), c chan uint64) {
	//TODO: Handle excessive quantities of IR data without a tail.
	for {
		ok := r.Read(&r.Buffer[r.index])
		if ok {
			if r.Buffer[r.index].Tail {
				if r.index > 1 {
					// Decode what we have and put it on the channel.
					// We can leave off the tail, it's not important.
					v, _ := d(r.Buffer[0 : r.index-1])
					c <- v
				}
				r.index = 0
			} else {
				r.index++
			}
		}
		if ctx.Err() != nil {
			break
		}
		if r.index == 0 {
			// Yield to give other routines time to work.
			time.Sleep(time.Microsecond * 20)
		}
	}
}

// Read the pin to see if there have been any changes.
func (r *Detector) Read(edge *Edge) (ok bool) {
	now := r.now()
	cv := r.pin.Read() == rpio.Low
	dur := now.Sub(r.t)
	edge.Value = r.pv
	edge.Duration = dur
	edge.Tail = false
	if cv != r.pv {
		// The first state change is ignored.
		if r.first {
			r.first = false
			r.pv = cv
			r.t = now
			return
		}
		edge.Value = r.pv
		edge.Duration = dur
		edge.Tail = false
		ok = true
		r.pv = cv
		r.t = now
		return
	}
	// Deal with timeouts.
	if !r.first && dur > r.Timeout {
		edge.Value = r.pv
		edge.Duration = dur
		edge.Tail = true
		ok = true
		r.pv = cv
		r.t = now
		// We've now hit the steady state.
		r.first = true
	}
	return
}
