package edge

import (
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
	for _, ee := range e {
		if ee.Value {
			sb.WriteString(strings.Repeat("_", int(ee.Duration.Milliseconds()+1)))
		} else {
			sb.WriteString(strings.Repeat("─", int(ee.Duration.Milliseconds()+1)))
		}
	}
	sb.WriteRune('\n')
	return sb.String()
}

// Reader of a pin. For example, an rpio.Pin.
type Reader interface {
	Read() rpio.State
}

// NewDetector creates a new edge detector.
// Can take in an rpio.Pin.
func NewDetector(pin Reader, timeout time.Duration) *Detector {
	now := time.Now()
	return &Detector{
		Timeout: timeout,
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

	// Previous Value of the pin.
	pv bool
	// Time of the last pin sample.
	t time.Time
	// First sample taken?
	first bool
}

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
	if !r.first && timeSinceLastChange > r.Timeout {
		d <- Edge{
			Value:    r.pv,
			Duration: timeSinceLastChange,
			Tail:     true,
		}
		r.pv = cv
		r.t = now
		// We've now hit the steady state.
		r.first = true
	}
}
