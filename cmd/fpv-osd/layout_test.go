package main

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ericogr/fpv-osd-decoder/internal/osd"
	"github.com/ericogr/fpv-osd-decoder/internal/telemetry"
)

func TestParseLayout(t *testing.T) {
	spec, err := parseLayout("auto", "auto")
	if err != nil || spec.mode != nil || spec.msgRow != nil {
		t.Errorf("auto, auto = %+v, %v; want both left to the recording", spec, err)
	}
	spec, err = parseLayout("45, 16,4", " 13")
	if err != nil || spec.mode == nil || *spec.mode != (telemetry.Region{X: 45, Y: 16, Width: 4}) || spec.msgRow == nil || *spec.msgRow != 13 {
		t.Errorf("45,16,4 and 13 = %+v, %v", spec, err)
	}
	if spec, err = parseLayout("auto", "-1"); err != nil || spec.msgRow == nil || *spec.msgRow != -1 {
		t.Errorf("-1 = %+v, %v; want the messages off", spec, err)
	}
	for _, tt := range []struct{ mode, row, want string }{
		{"1,2", "auto", "-mode must be x,y,width"},
		{"automatic", "auto", "-mode must be x,y,width"},
		{"auto", "top", "-msg-row must be a row number or auto"},
		{"auto", "", "-msg-row must be a row number or auto"},
	} {
		if _, err := parseLayout(tt.mode, tt.row); err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("parseLayout(%q, %q) error = %v, want %q", tt.mode, tt.row, err, tt.want)
		}
	}
}

// layoutFrames are frames of a 20x6 grid in which the flight mode is at (9,3)
// and the status messages are on row 5.
func layoutFrames() [][]uint16 {
	frame := func(mode, message string) []uint16 {
		cells := make([]uint16, 20*6)
		for i := range cells {
			cells[i] = ' '
		}
		copy(cells[3*20+9:], codes(mode))
		copy(cells[5*20+4:], codes(message))
		return cells
	}
	return [][]uint16{frame("ACRO", "GO HOME"), frame("ACRO", ""), frame("ANGL", "WAIT HERE"), frame("ANGL", "")}
}

func codes(s string) []uint16 {
	cs := make([]uint16, len(s))
	for i, c := range []byte(s) {
		cs[i] = uint16(c)
	}
	return cs
}

// A layout that is not the one of the sample is found without being told.
func TestCSVFindsAMovedLayout(t *testing.T) {
	file := writeFrames(t, 20, 6, layoutFrames())
	var stdout, stderr bytes.Buffer
	if code := runMain([]string{"csv", "-o", "-", file}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d: %s", code, &stderr)
	}
	if want := "t_ms,osd_flight_mode,osd_message\n0,ACRO,GO HOME\n151,ACRO,\n302,ANGL,WAIT HERE\n453,ANGL,\n"; stdout.String() != want {
		t.Errorf("csv = %q, want %q", stdout.String(), want)
	}
	if want := "layout: flight mode at 9,3,4 (found); status messages on row 5 (found)"; !strings.Contains(stderr.String(), want) {
		t.Errorf("summary = %q, want it to say %q", &stderr, want)
	}
}

// What is given on the command line is used as it is, found or not.
func TestCSVGivenLayoutIsUsed(t *testing.T) {
	file := writeFrames(t, 20, 6, layoutFrames())
	tests := []struct {
		name   string
		args   []string
		stdout string
		layout string
	}{
		{"the right places", []string{"-mode", "9,3,4", "-msg-row", "5"},
			"t_ms,osd_flight_mode,osd_message\n0,ACRO,GO HOME\n151,ACRO,\n302,ANGL,WAIT HERE\n453,ANGL,\n",
			"layout: flight mode at 9,3,4; status messages on row 5\n"},
		{"messages off", []string{"-msg-row", "-1"},
			"t_ms,osd_flight_mode\n0,ACRO\n151,ACRO\n302,ANGL\n453,ANGL\n",
			"layout: flight mode at 9,3,4 (found); status messages off\n"},
		{"the mode somewhere else", []string{"-mode", "0,0,4", "-msg-row", "5"},
			"t_ms,osd_message\n0,GO HOME\n151,\n302,WAIT HERE\n453,\n",
			"layout: flight mode at 0,0,4; status messages on row 5\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			args := append([]string{"csv", "-o", "-"}, tt.args...)
			if code := runMain(append(args, file), &stdout, &stderr); code != 0 {
				t.Fatalf("exit code = %d: %s", code, &stderr)
			}
			if stdout.String() != tt.stdout {
				t.Errorf("csv = %q, want %q", stdout.String(), tt.stdout)
			}
			if !strings.Contains(stderr.String(), tt.layout) {
				t.Errorf("summary = %q, want it to say %q", &stderr, tt.layout)
			}
		})
	}
}

// The sample flight is the layout the old defaults were made for.
func TestSampleFlightLayoutIsFound(t *testing.T) {
	f, err := osd.Open(sampleFlight)
	if err != nil {
		t.Skip("AvatarG0031.osd is not in the project")
	}
	opts, line := layoutSpec{}.resolve(f)
	if want := []telemetry.Region{{X: 45, Y: 16, Width: 4}}; len(opts.FlightModes) != 1 || opts.FlightModes[0] != want[0] || opts.MessageRow != 13 {
		t.Errorf("layout = %+v, want the mode at %v and the messages on row 13", opts, want)
	}
	if want := "layout: flight mode at 45,16,4 (found); status messages on row 13 (found)"; line != want {
		t.Errorf("line = %q, want %q", line, want)
	}
}

// shiftedSample is the sample flight with the whole OSD moved dx columns and dy
// rows, what the cells push out of the grid lost and the vacated ones blank.
func shiftedSample(t *testing.T, dx, dy int) string {
	t.Helper()
	data, err := os.ReadFile(sampleFlight)
	if err != nil {
		t.Skip("AvatarG0031.osd is not in the project")
	}
	const header = 40
	cols := int(binary.LittleEndian.Uint16(data[0x24:]))
	rows := int(binary.LittleEndian.Uint16(data[0x26:]))
	size := 4 + cols*rows*2
	out := append([]byte(nil), data[:header]...)
	for at := header; at+size <= len(data); at += size {
		out = append(out, data[at:at+4]...)
		cells := make([]uint16, cols*rows)
		for i := range cells {
			cells[i] = ' '
		}
		for y := 0; y < rows; y++ {
			for x := 0; x < cols; x++ {
				if nx, ny := x+dx, y+dy; nx >= 0 && nx < cols && ny >= 0 && ny < rows {
					cells[ny*cols+nx] = binary.LittleEndian.Uint16(data[at+4+(y*cols+x)*2:])
				}
			}
		}
		for _, c := range cells {
			out = binary.LittleEndian.AppendUint16(out, c)
		}
	}
	path := filepath.Join(t.TempDir(), "shifted.osd")
	if err := os.WriteFile(path, out, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// The pilot can move every element of the OSD: nothing read changes, and the
// flight mode and the messages are found in their new places.
func TestMovedOSDReadsTheSame(t *testing.T) {
	moved := shiftedSample(t, -2, 1)
	run := func(file string) (string, string) {
		var stdout, stderr bytes.Buffer
		if code := runMain([]string{"csv", "-srt", "none", "-o", "-", file}, &stdout, &stderr); code != 0 {
			t.Fatalf("exit code = %d: %s", code, &stderr)
		}
		return stdout.String(), stderr.String()
	}
	want, _ := run(sampleFlight)
	got, summary := run(moved)
	if got != want {
		t.Error("the CSV of the moved OSD is not the one of the original")
	}
	if line := "layout: flight mode at 43,17,4 (found); status messages on row 14 (found)"; !strings.Contains(summary, line) {
		t.Errorf("summary = %q, want it to say %q", summary, line)
	}
}
