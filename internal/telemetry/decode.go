package telemetry

import (
	"fmt"
	"strings"

	"github.com/ericogr/fpv-osd-decoder/internal/inav"
	"github.com/ericogr/fpv-osd-decoder/internal/osd"
)

// Value is a telemetry value read from a frame.
type Value struct {
	Num  float64
	Text string
}

// Sample holds what was read from a frame, by column key. A field that was
// not readable (blinking, no GPS, splash screen) is absent: it never becomes
// zero.
type Sample map[string]Value

func (s Sample) num(key string, v float64) { s[key] = Value{Num: v} }

// Region is a strip of cells on one row of the OSD.
type Region struct{ X, Y, Width int }

// String formats the region as "x,y,width", the form the command line takes.
func (r Region) String() string { return fmt.Sprintf("%d,%d,%d", r.X, r.Y, r.Width) }

// Options says where to find the fields that have no anchor glyph and
// therefore depend on the layout configured in INAV. The defaults match the
// layout of this project's AvatarG00xx videos.
type Options struct {
	FlightMode Region // where the flight mode text is shown
	MessageRow int    // row of the status messages; <0 disables
}

// DefaultOptions returns the positions of the known layout.
func DefaultOptions() Options {
	return Options{FlightMode: Region{X: 45, Y: 16, Width: 4}, MessageRow: 13}
}

// Decoder turns OSD frames into telemetry samples.
type Decoder struct {
	opts  Options
	rules []rule
}

// New creates a decoder.
func New(opts Options) *Decoder {
	return &Decoder{opts: opts, rules: buildRules()}
}

// Decode reads every recognized field in a frame and also returns what was
// left on the screen unrecognized (not counting the artificial horizon, the
// crosshair and the variometer, which are graphics).
func (d *Decoder) Decode(f *osd.File, fr *osd.Frame) (Sample, []Leftover) {
	s := Sample{}
	var left []Leftover
	for y := 0; y < f.Rows; y++ {
		row := f.Row(fr, y)
		line := d.decodeRow(row, s)
		line = d.readTextFields(y, row, line, s)
		left = appendLeftovers(left, y, line)
	}
	return s, left
}

// decodeRow applies the rules in order to one row and returns the row with
// everything recognized erased. Each recognized span disappears before the
// next rule runs, so generic rules (a speed without an icon, for example)
// cannot reuse what a more specific field already consumed.
func (d *Decoder) decodeRow(row []uint16, s Sample) string {
	var sb strings.Builder
	blank := true
	for _, c := range row {
		if c == 0 {
			c = ' '
		}
		if c != ' ' {
			blank = false
		}
		sb.WriteRune(rune(c))
	}
	line := sb.String()
	if blank {
		return line
	}
	for _, r := range d.rules {
		line = r.apply(line, s)
	}
	return line
}

// readTextFields reads the fields that have no anchor glyph (flight mode and
// status messages) by position, and erases them from line.
func (d *Decoder) readTextFields(y int, row []uint16, line string, s Sample) string {
	if r := d.opts.FlightMode; y == r.Y && r.Width > 0 && r.X >= 0 && r.X < len(row) {
		end := min(r.X+r.Width, len(row))
		if t := plainText(row[r.X:end]); t != "" {
			s[keyFlightMode] = Value{Text: t}
		}
		line = blankRunes(line, r.X, end)
	}
	if y == d.opts.MessageRow {
		if t := plainText(row); t != "" {
			s[keyMessage] = Value{Text: t}
		}
		line = blankRunes(line, 0, len(row))
	}
	return line
}

func blankRunes(line string, from, to int) string {
	rs := []rune(line)
	for i := from; i < to && i < len(rs); i++ {
		rs[i] = ' '
	}
	return string(rs)
}

// plainText returns only the "real" ASCII text of the cells: uppercase
// letters, digits and punctuation. Font glyphs (including those INAV puts in
// ASCII codes, such as 'k', 'v' or '~') become spaces.
func plainText(cells []uint16) string {
	b := make([]byte, len(cells))
	for i, c := range cells {
		if c >= 0x20 && c < 0x7f && (c == ' ' || inav.Name(c) == "") {
			b[i] = byte(c)
		} else {
			b[i] = ' '
		}
	}
	return strings.Join(strings.Fields(string(b)), " ")
}
