package decoder

import (
	"errors"
	"fmt"
	"time"

	"github.com/a-h/infrared/edge"
)

// PanasonicCodeToKey is a map of Panasonic IR codes to the remote control labels.
var PanasonicCodeToKey = map[uint64]string{
	208069699051522: "Power",
	15973142765570:  "Guide",
	195919270125570: "InputTV",
	146256529727490: "InputAV",
	231249637548034: "Menu",
	142949421686786: "Text",
	145157034876930: "Still",
	99909604745218:  "Aspect",
	203654472671234: "Info",
	92165711601666:  "Exit",
	7211409481730:   "Apps",
	136279471693826: "NetFlix",
	31426435096578:  "Home",
	222419184787458: "Up",
	223522991382530: "Down",
	226834411167746: "Left",
	227938217762818: "Right",
	51294953807874:  "Option",
	93269518196738:  "Back",
	221315378192386: "OK",
}

// Panasonic infrared decoder.
func Panasonic(edges edge.Edges) (v uint64, err error) {
	if len(edges) == 0 {
		fmt.Println("Shouldn't end up with no edges")
	}
	if len(edges)%2 != 0 {
		err = fmt.Errorf("the count of marks and spaces in Panasonic IR signals must be an even number, got %d", len(edges))
		return
	}
	// Read the low part of the header.
	if edges[0].Value == false {
		err = errors.New("header mark does not match Panasonic header value")
		return
	}
	if !between(time.Microsecond*1005, time.Microsecond*5253, edges[0].Duration) {
		err = fmt.Errorf("expected header to last 3502µs, but was %v", edges[0].Duration)
		return
	}
	// Read the high part of the header.
	if edges[1].Value == true {
		err = errors.New("header space does not match Panasonic header value")
		return
	}
	if !between(time.Microsecond*875, time.Microsecond*2625, edges[1].Duration) {
		err = fmt.Errorf("expected header space to last 1750µs, but was %v", edges[1].Duration)
		return
	}
	// Skip the header and read the bits.
	bitOffset := 2
	for i := 0; i < len(edges)-bitOffset; i += 2 {
		// Read the bit mark.
		bit := edges[i+bitOffset]
		if bit.Value != true {
			err = fmt.Errorf("expected bit mark at %d to be low", i+bitOffset)
			return
		}
		// Read the space mark to work out whether it's a zero or one.
		space := edges[i+bitOffset+1]
		if space.Value != false {
			err = fmt.Errorf("expected bit space at %d to be high", i+bitOffset)
			return
		}
		// If the space is 400µs, it's a zero, if it's 1244µs it's a one.
		// The timing can be relatively loose, so if it's greater than 900ms, it's probably a one.
		if space.Duration >= time.Microsecond*800 {
			// There is a mark and a space for each bit, so i is divided by 2.
			v |= (1 << (i / 2))
		}
	}
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
