package telemetry

import (
	"regexp"
	"sort"
	"strings"

	"github.com/ericogr/fpv-osd-decoder/internal/inav"
	"github.com/ericogr/fpv-osd-decoder/internal/osd"
)

// The flight mode and the status messages are text with no glyph to anchor
// them, so where they are depends on the layout the pilot set up in INAV. They
// can be found in the recording itself, which is what this file does.

// minShare is how much of the best candidate's support a flight mode position
// needs to count as a layout of its own: a message that happens to hold a mode
// name ("(HOLD)") shows up far less often than the element does.
const minShare = 5 // one in five

// DetectFlightModes finds where the recording shows the flight mode. The
// element is always four characters and always one of a short list of names,
// and it is on screen in nearly every frame, so the positions where those names
// appear most often are it. The best supported position comes first. It returns
// nothing when no frame shows one.
func DetectFlightModes(f *osd.File) []Region {
	hits := map[Region]int{}
	for i := range f.Frames {
		fr := &f.Frames[i]
		if screenOf(f, fr) != flightScreen {
			continue
		}
		for y := 0; y < f.Rows; y++ {
			row := f.Row(fr, y)
			for x := 0; x+inav.FlightModeWidth <= len(row); x++ {
				if modeAt(row, x) {
					hits[Region{X: x, Y: y, Width: inav.FlightModeWidth}]++
				}
			}
		}
	}
	regions := make([]Region, 0, len(hits))
	for r := range hits {
		regions = append(regions, r)
	}
	sort.Slice(regions, func(i, j int) bool {
		a, b := regions[i], regions[j]
		if hits[a] != hits[b] {
			return hits[a] > hits[b]
		}
		return a.Y < b.Y || a.Y == b.Y && a.X < b.X
	})
	for i, r := range regions {
		if hits[r]*minShare < hits[regions[0]] {
			return regions[:i]
		}
	}
	return regions
}

// modeAt reports whether the cells from x on are the name of a flight mode, set
// apart from the text around it: a mode name inside a longer word is not the
// element.
func modeAt(row []uint16, x int) bool {
	end := x + inav.FlightModeWidth
	if x > 0 && isAlphanumeric(row[x-1]) || end < len(row) && isAlphanumeric(row[end]) {
		return false
	}
	var name [inav.FlightModeWidth]byte
	for i, c := range row[x:end] {
		if c == 0 {
			c = ' '
		}
		if c > 0x7e {
			return false
		}
		name[i] = byte(c)
	}
	return inav.IsFlightMode(string(name[:]))
}

// isAlphanumeric reports whether a cell holds a digit or a letter of INAV's
// text, which is uppercase: the lowercase codes are font glyphs.
func isAlphanumeric(c uint16) bool {
	return c >= '0' && c <= '9' || c >= 'A' && c <= 'Z'
}

// A message has a word of at least three letters, which leaves out labels such
// as "M0" and the two-letter codes of the radio.
var wordRE = regexp.MustCompile(`[A-Z]{3}`)

// minMessageTexts is how many different texts a row must show to be taken for
// the messages. Text that never changes is a name, not a message.
const minMessageTexts = 2

// DetectMessageRow finds the row of the status messages, given where the flight
// mode is. Messages are free text, so there is no list of names to look for:
// they are the text that changes. Of the text left on the screen after the
// rules have read their elements, the row that shows the most different texts
// is the messages row. It is only taken when no other row shows as many
// different texts and when the texts really change, since a wrong row is worse
// than none; ok is false when there is no such row.
func DetectMessageRow(f *osd.File, modes []Region) (row int, ok bool) {
	d := New(Options{FlightModes: modes, MessageRow: -1})
	texts := map[int]map[string]bool{}
	for i := range f.Frames {
		fr := &f.Frames[i]
		if screenOf(f, fr) != flightScreen {
			continue
		}
		for y := 0; y < f.Rows; y++ {
			cells := f.Row(fr, y)
			line, blank := rowText(cells)
			if blank || !wordRE.MatchString(line) {
				continue
			}
			text, _ := takeText(d.readRow(y, cells, line, Sample{}, true))
			// a plus code (8FRP4JQ7+M5) changes with every move
			if !wordRE.MatchString(text) || strings.Contains(text, "+") {
				continue
			}
			if texts[y] == nil {
				texts[y] = map[string]bool{}
			}
			texts[y][text] = true
		}
	}
	best, runnerUp := -1, 0
	for y, set := range texts {
		switch n := len(set); {
		case best < 0 || n > len(texts[best]):
			if best >= 0 {
				runnerUp = len(texts[best])
			}
			best = y
		case n > runnerUp:
			runnerUp = n
		}
	}
	if best < 0 || len(texts[best]) < minMessageTexts || len(texts[best]) == runnerUp {
		return 0, false
	}
	return best, true
}
