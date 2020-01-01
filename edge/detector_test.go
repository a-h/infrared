package edge

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stianeikeland/go-rpio/v4"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name     string
		pin      *testReader
		expected []bool
	}{
		{
			name:     "the first value is ignored",
			pin:      readerOf(rpio.High, rpio.High, rpio.High),
			expected: []bool{},
		},
		{
			name:     "the first state change is ignored because we missed the start",
			pin:      readerOf(rpio.High, rpio.Low),
			expected: []bool{},
		},
		{
			name:     "subsequent edges are detected",
			pin:      readerOf(rpio.High, rpio.Low, rpio.High),
			expected: []bool{true},
		},
		{
			name:     "multiple edges are detected",
			pin:      readerOf(rpio.High, rpio.Low, rpio.High, rpio.Low, rpio.High),
			expected: []bool{true, false, true},
		},
		{
			name: "the 1ms timeout is respected",
			// The last 3 values are "high" which, because there hasn't been a "low" value for 1ms, is represented as a false.
			pin:      readerOf(rpio.High, rpio.Low, rpio.High, rpio.High, rpio.High),
			expected: []bool{true, false},
		},
		{
			name:     "tail false values are only counted once",
			pin:      readerOf(rpio.High, rpio.Low, rpio.High, rpio.High, rpio.High, rpio.High, rpio.High),
			expected: []bool{true, false},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ed := New(test.pin)
			values := ed.ReadN(len(test.pin.samples), 500*time.Microsecond).Values()

			if !cmp.Equal(test.expected, values) {
				t.Errorf(cmp.Diff(test.expected, values))
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
