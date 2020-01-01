package edge

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stianeikeland/go-rpio/v4"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name   string
		pin    *testReader
		values []bool
		isTail []bool
	}{
		{
			name:   "the first value is ignored",
			pin:    readerOf(rpio.High, rpio.High, rpio.High),
			values: []bool{},
		},
		{
			name:   "the first state change is ignored because we missed the start",
			pin:    readerOf(rpio.High, rpio.Low),
			values: []bool{},
		},
		{
			name:   "subsequent edges are detected",
			pin:    readerOf(rpio.High, rpio.Low, rpio.High),
			values: []bool{true},
		},
		{
			name:   "multiple edges are detected",
			pin:    readerOf(rpio.High, rpio.Low, rpio.High, rpio.Low, rpio.High),
			values: []bool{true, false, true},
		},
		{
			name: "the 1ms timeout is respected",
			// The last 3 values are "high" which, because there hasn't been a "low" value for 1ms, is represented as a false.
			pin:    readerOf(rpio.High, rpio.Low, rpio.High, rpio.High, rpio.High),
			values: []bool{true, false},
			isTail: []bool{false, true},
		},
		{
			name:   "tail false values are only counted once",
			pin:    readerOf(rpio.High, rpio.Low, rpio.High, rpio.High, rpio.High, rpio.High, rpio.High),
			values: []bool{true, false},
			isTail: []bool{false, true},
		},
		{
			name:   "a new set of data can be sent after the timeout gap",
			pin:    readerOf(rpio.High, rpio.Low, rpio.High, rpio.High, rpio.High, rpio.High, rpio.High, rpio.Low, rpio.High, rpio.High, rpio.High, rpio.High, rpio.High),
			values: []bool{true, false, true, false},
			isTail: []bool{false, true, false, true},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ed := New(test.pin)
			edges := ed.ReadN(len(test.pin.samples), 500*time.Microsecond)
			if len(test.isTail) > 0 {
				for i, v := range test.isTail {
					if edges[i].Tail != v {
						t.Errorf("expected edge %d to be have Tail %v, but got %v", i, v, edges[i].Tail)
					}
				}
			}

			values := edges.Values()
			if !cmp.Equal(test.values, values) {
				t.Errorf("unexpected values: %v", cmp.Diff(test.values, values))
			}
		})
	}
}

func readerOf(samples ...rpio.State) *testReader {
	return &testReader{
		samples: samples,
	}
}

type testReader struct {
	samples []rpio.State
	index   int
	read    int
}

func (tr *testReader) Read() rpio.State {
	defer func() { tr.read++ }()
	if tr.index >= len(tr.samples) {
		return rpio.High
	}
	defer func() { tr.index++ }()
	return tr.samples[tr.index]
}
