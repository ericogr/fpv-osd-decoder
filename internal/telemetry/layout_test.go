package telemetry

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/ericogr/fpv-osd-decoder/internal/inav"
	"github.com/ericogr/fpv-osd-decoder/internal/osd"
)

// recording joins one-frame files into a recording.
func recording(frames ...*osd.File) *osd.File {
	rec := newFile()
	rec.Frames = nil
	for _, f := range frames {
		rec.Frames = append(rec.Frames, f.Frames[0])
	}
	return rec
}

// drawn is a frame with the texts put at the given places: a text per cell
// position, as x, y, text triples.
func drawn(parts ...any) *osd.File {
	f := newFile()
	for i := 0; i < len(parts); i += 3 {
		put(f, parts[i].(int), parts[i+1].(int), parts[i+2])
	}
	return f
}

// times is n copies of a frame.
func times(n int, f *osd.File) []*osd.File {
	frames := make([]*osd.File, n)
	for i := range frames {
		frames[i] = f
	}
	return frames
}

func TestDetectFlightModes(t *testing.T) {
	at := func(x, y int) Region { return Region{X: x, Y: y, Width: inav.FlightModeWidth} }
	stats := drawn(16, 1, "*** STATS", 20, 7, "ACRO")
	splash := newFile()
	put(splash, 21, 1, inav.SymLogoStart, inav.SymLogoStart+1)
	put(splash, 20, 7, "ACRO")

	tests := []struct {
		name   string
		frames []*osd.File
		want   []Region
	}{
		{"nothing on screen", times(3, newFile()), []Region{}},
		{"wherever the pilot put it", times(3, drawn(9, 3, "ACRO")), []Region{at(9, 3)}},
		{"in the corners", append(times(2, drawn(0, 0, "ACRO")), times(2, drawn(cols-4, rows-1, "ACRO"))...),
			[]Region{at(0, 0), at(cols-4, rows-1)}},
		{"names that are padded with blanks", times(2, drawn(20, 7, "RTH ")), []Region{at(20, 7)}},
		{"a name that starts with a blank", times(2, drawn(20, 7, " WP ")), []Region{at(20, 7)}},
		{"the most frequent position comes first",
			append(times(2, drawn(1, 1, "ANGL")), times(5, drawn(30, 9, "ANGL"))...),
			[]Region{at(30, 9), at(1, 1)}},
		{"a position seen far less than the best is not a layout",
			append(times(1, drawn(1, 1, "(HOLD)")), times(20, drawn(30, 9, "ACRO"))...), []Region{at(30, 9)}},
		{"inside a longer word", times(3, drawn(10, 5, "ADJUSTING RTH ALTITUDE", 10, 6, "HOLDING", 10, 7, "1ACRO")), []Region{}},
		{"not on the stats screen", times(3, stats), []Region{}},
		{"not on the splash screen", times(3, splash), []Region{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectFlightModes(recording(tt.frames...))
			if got == nil {
				got = []Region{}
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DetectFlightModes = %v, want %v", got, tt.want)
			}
		})
	}
}

// Every name the flight mode element can show is found, wherever it is.
func TestDetectEveryFlightMode(t *testing.T) {
	for name := range map[string]bool{"ACRO": true, "ANGL": true, "ANGH": true, "HOR ": true, "MANU": true,
		"RTH ": true, "WRTH": true, "HOLD": true, "LOTR": true, "CRUZ": true, "CRSH": true, " WP ": true,
		" AH ": true, "GEO ": true, "TURT": true, "LAND": true, "!FS!": true} {
		got := DetectFlightModes(recording(drawn(12, 8, name)))
		if want := []Region{{X: 12, Y: 8, Width: 4}}; !reflect.DeepEqual(got, want) {
			t.Errorf("%q: DetectFlightModes = %v, want %v", name, got, want)
		}
	}
}

func TestDetectMessageRow(t *testing.T) {
	messages := func(y int, texts ...string) []*osd.File {
		var frames []*osd.File
		for _, text := range texts {
			frames = append(frames, drawn(14, y, text))
		}
		return frames
	}
	join := func(groups ...[]*osd.File) *osd.File {
		var all []*osd.File
		for _, g := range groups {
			all = append(all, g...)
		}
		return recording(all...)
	}
	mode := []Region{{X: 45, Y: 16, Width: 4}}
	tests := []struct {
		name    string
		rec     *osd.File
		modes   []Region
		wantRow int
		wantOK  bool
	}{
		{"the row whose text changes", join(messages(9, "EN ROUTE TO HOME", "ADJUSTING RTH ALTITUDE", "LANDING")), mode, 9, true},
		{"a name that never changes is not a message",
			join(messages(2, "MINI X8", "MINI X8", "MINI X8"), messages(9, "EN ROUTE TO HOME", "LANDING")), mode, 9, true},
		{"only text that does not change", join(messages(2, "MINI X8", "MINI X8", "MINI X8")), mode, 0, false},
		{"a single message is not enough to tell", join(messages(9, "EN ROUTE TO HOME")), mode, 0, false},
		{"two rows that change alike", join(messages(9, "ONE TWO", "TWO ONE"), messages(11, "RED BLUE", "BLUE RED")), mode, 0, false},
		{"a row that changes more wins",
			join(messages(9, "ONE TWO", "TWO ONE", "TWO TWO"), messages(11, "RED BLUE", "BLUE RED")), mode, 9, true},
		{"a plus code changes all the time and is not a message",
			join(messages(4, "8FRP4JQ7+M5", "8FRP4JQ7+M6", "8FRP4JQ7+M7", "8FRP4JQ7+M8"), messages(9, "EN ROUTE TO HOME", "LANDING")), mode, 9, true},
		{"words are needed", join(messages(9, "12", "13", "14")), mode, 0, false},
		{"no text at all", recording(times(3, newFile())...), mode, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row, ok := DetectMessageRow(tt.rec, tt.modes)
			if ok != tt.wantOK || row != tt.wantRow {
				t.Errorf("DetectMessageRow = %d, %v; want %d, %v", row, ok, tt.wantRow, tt.wantOK)
			}
		})
	}

	// The flight mode changes too, which makes its row look like messages:
	// that is why the messages are looked for after the mode.
	rec := recording(drawn(45, 16, "ACRO"), drawn(45, 16, "ANGL"), drawn(45, 16, "CRUZ"))
	if row, ok := DetectMessageRow(rec, nil); !ok || row != 16 {
		t.Errorf("without the mode: %d, %v; want row 16", row, ok)
	}
	if row, ok := DetectMessageRow(rec, mode); ok {
		t.Errorf("with the mode: row %d, want none", row)
	}
}

// The text of a row is what the rules left: another element on the row is
// neither part of the message nor a reason to take the row for another.
func TestDetectMessageRowIgnoresOtherElements(t *testing.T) {
	var frames []*osd.File
	for i, text := range []string{"EN ROUTE TO HOME", "LANDING", "HOLDING POSITION"} {
		f := drawn(8, 13, text)
		put(f, 40, 13, fmt.Sprint(100+i), inav.SymAltM)
		frames = append(frames, f)
	}
	if row, ok := DetectMessageRow(recording(frames...), nil); !ok || row != 13 {
		t.Errorf("DetectMessageRow = %d, %v; want row 13", row, ok)
	}
}

func TestDetectMessageRowIgnoresTheStatsScreen(t *testing.T) {
	var frames []*osd.File
	for _, text := range []string{"MAX ALTITUDE", "MAX SPEED", "FLIGHT TIME"} {
		frames = append(frames, drawn(16, 1, "*** STATS", 6, 5, text))
	}
	if row, ok := DetectMessageRow(recording(frames...), nil); ok {
		t.Errorf("DetectMessageRow = row %d on the stats screen, want none", row)
	}
}
