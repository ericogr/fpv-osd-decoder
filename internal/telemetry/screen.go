package telemetry

import (
	"github.com/ericogr/fpv-osd-decoder/internal/inav"
	"github.com/ericogr/fpv-osd-decoder/internal/osd"
)

// screen says what an OSD frame is showing.
type screen int

const (
	flightScreen screen = iota // the pilot's own layout
	splashScreen               // the INAV logo, shown while arming
	statsScreen                // the statistics, shown after disarming
)

// statsTitle starts the title line of the statistics screen.
const statsTitle = "*** STATS"

// screenOf tells which screen a frame shows.
func screenOf(f *osd.File, fr *osd.Frame) screen {
	kind := flightScreen
	for y := 0; y < f.Rows; y++ {
		row := f.Row(fr, y)
		if hasText(row, statsTitle) {
			return statsScreen
		}
		for _, c := range row {
			if inav.IsLogo(c) {
				kind = splashScreen
			}
		}
	}
	return kind
}

// hasText reports whether the ASCII text appears in the cells of a row.
func hasText(row []uint16, text string) bool {
	for i := 0; i+len(text) <= len(row); i++ {
		j := 0
		for j < len(text) && row[i+j] == uint16(text[j]) {
			j++
		}
		if j == len(text) {
			return true
		}
	}
	return false
}
