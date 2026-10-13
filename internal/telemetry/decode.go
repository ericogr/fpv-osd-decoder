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

// hasAny reports whether s has any of the columns of other.
func (s Sample) hasAny(other Sample) bool {
	for key := range other {
		if _, ok := s[key]; ok {
			return true
		}
	}
	return false
}

// Region is a strip of cells on one row of the OSD.
type Region struct{ X, Y, Width int }

// String formats the region as "x,y,width", the form the command line takes.
func (r Region) String() string { return fmt.Sprintf("%d,%d,%d", r.X, r.Y, r.Width) }

// Options says where to find the fields that have no anchor glyph and
// therefore depend on the layout configured in INAV. DetectFlightModes and
// DetectMessageRow find them in a recording.
type Options struct {
	// FlightModes are the strips where the flight mode is shown, one for each
	// OSD layout that shows it. With more than one, a strip only counts when it
	// holds a text the flight mode element can show.
	FlightModes []Region
	// MessageRow is the row of the status messages; <0 disables them. The zero
	// value is the first row, so a decoder with no messages has to say so.
	MessageRow int
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
//
// The post-flight statistics screen is not read at all: its numbers are
// summaries of the flight, not what the aircraft was doing at that moment. The
// splash screen shown while arming is read by the rules, since it does show
// the position, but not by the generic pass, which would only find the
// version and the totals it prints.
func (d *Decoder) Decode(f *osd.File, fr *osd.Frame) (Sample, []Leftover) {
	kind := screenOf(f, fr)
	s := Sample{}
	var left []Leftover
	for y := 0; y < f.Rows; y++ {
		row := f.Row(fr, y)
		line, blank := rowText(row)
		if blank {
			continue
		}
		if kind != statsScreen {
			line = d.readRow(y, row, line, s, kind == flightScreen)
		}
		left = appendLeftovers(left, y, line)
	}
	return s, left
}

// rowText turns the cells of a row into a string with one rune per cell
// holding the glyph code (an empty cell becomes a blank), and says whether
// the row is empty.
func rowText(row []uint16) (line string, blank bool) {
	var sb strings.Builder
	blank = true
	for _, c := range row {
		if c == 0 {
			c = ' '
		}
		if c != ' ' {
			blank = false
		}
		sb.WriteRune(rune(c))
	}
	return sb.String(), blank
}

// readRow reads what a row of the screen holds, in order: the fields the rules
// know; then, on the pilot's own screen (flight), the numbers next to glyphs no
// rule knows and last the text, which is what is left. The flight mode and the
// messages are not on the other screens: what a splash screen shows in those
// places is not them. cells are the glyph codes of the row and line is the same
// row as text; it returns the line without what was read.
func (d *Decoder) readRow(y int, cells []uint16, line string, s Sample, flight bool) string {
	line = d.decodeRow(line, s)
	if !flight {
		return line
	}
	return d.readTextFields(y, cells, readGeneric(line, s), s)
}

// decodeRow applies the rules in order to one row and returns the row with
// everything recognized erased. Each recognized span disappears before the
// next rule runs, so generic rules (a speed without an icon, for example)
// cannot reuse what a more specific field already consumed.
func (d *Decoder) decodeRow(line string, s Sample) string {
	// The glyphs of the row as it is now. A rule that erases something leaves
	// its glyphs here, which only means a later rule may be tried for nothing.
	present := glyphsOf(line)
	for _, r := range d.rules {
		if !present.hasAny(r.needs) {
			continue
		}
		line = r.apply(line, s)
	}
	return line
}

// readTextFields reads the fields that have no anchor glyph (flight mode and
// status messages) by position, and erases them from line.
func (d *Decoder) readTextFields(y int, row []uint16, line string, s Sample) string {
	for _, r := range d.opts.FlightModes {
		if y != r.Y || r.Width <= 0 || r.X < 0 || r.X >= len(row) {
			continue
		}
		end := min(r.X+r.Width, len(row))
		line = blankRunes(line, r.X, end)
		if _, read := s[keyFlightMode]; read {
			continue
		}
		strip, _ := rowText(row[r.X:end])
		if len(d.opts.FlightModes) > 1 && !inav.IsFlightMode(strip) {
			continue
		}
		if t := plainText(row[r.X:end]); t != "" {
			s[keyFlightMode] = Value{Text: t}
		}
	}
	if y == d.opts.MessageRow {
		// From what the rules left, so that another element on the row does not
		// end up in the message.
		var t string
		if t, line = takeText(line); t != "" {
			s[keyMessage] = Value{Text: t}
		}
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

// isText reports whether a cell holds a real ASCII character: a letter, digit
// or punctuation. Font glyphs, including those INAV puts in ASCII codes such as
// 'k', 'v' or '~', are not text.
func isText(c uint16) bool { return c > ' ' && c < 0x7f && inav.Name(c) == "" }

// plainText returns the ASCII text of the cells, with the blanks squeezed.
func plainText(cells []uint16) string {
	b := make([]byte, len(cells))
	for i, c := range cells {
		if isText(c) {
			b[i] = byte(c)
		} else {
			b[i] = ' '
		}
	}
	return strings.Join(strings.Fields(string(b)), " ")
}

// takeText returns the ASCII text on the line, with the blanks squeezed, and
// the line without that text. Whatever is not text stays on the line.
func takeText(line string) (text, rest string) {
	cells := []rune(line)
	found := make([]rune, len(cells))
	for i, c := range cells {
		found[i] = ' '
		if c <= 0xffff && isText(uint16(c)) {
			found[i], cells[i] = c, ' '
		}
	}
	return strings.Join(strings.Fields(string(found)), " "), string(cells)
}
