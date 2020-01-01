package edge

import (
	"time"

	"github.com/stianeikeland/go-rpio/v4"
)

// An Edge is a change in state.
type Edge struct {
	Value    bool
	Duration time.Duration
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

// Reader of a pin. For example, an rpio.Pin.
type Reader interface {
	Read() rpio.State
}

// New creates a new edge detector.
// Can take in an rpio.Pin.
func New(pin Reader) *Detector {
	now := time.Now()
	return &Detector{
		now:   time.Now,
		pin:   pin,
		t:     now,
		first: true,
	}
}

// A Detector detects changes in a pin by sampling the pin value.
type Detector struct {
	// Now is a function which returns the current time.
	now func() time.Time
	// Pin that we're reading from.
	// Can take in an rpio.Pin.
	pin Reader

	// Previous Value of the pin.
	pv bool
	// Time of the last pin sample.
	t time.Time
	// First sample taken?
	first bool
}

const timeout = time.Millisecond

func stateToBool(s rpio.State) bool {
	return s == rpio.Low
}

// ReadN samples from the pin.
func (r *Detector) ReadN(n int, every time.Duration) Edges {
	d := make(chan Edge)

	go func() {
		defer close(d)
		for i := 0; i < n; i++ {
			r.Read(d)
			time.Sleep(every)
		}
	}()

	var op []Edge
	for v := range d {
		op = append(op, v)
	}
	return op
}

// Read the pin to see if there have been any changes.
func (r *Detector) Read(d chan Edge) {
	now := r.now()
	cv := stateToBool(r.pin.Read())
	timeSinceLastChange := now.Sub(r.t)
	if cv != r.pv {
		// The first state change is ignored.
		if r.first {
			r.first = false
			r.pv = cv
			r.t = now
			return
		}
		d <- Edge{
			Value:    r.pv,
			Duration: timeSinceLastChange,
		}
		r.pv = cv
		r.t = now
		return
	}
	// Deal with timeouts.
	if !r.first && timeSinceLastChange > timeout {
		d <- Edge{
			Value:    r.pv,
			Duration: timeSinceLastChange,
		}
		r.pv = cv
		r.t = now
		// We've now hit the steady state.
		r.first = true
	}
}
