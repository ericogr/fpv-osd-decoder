package main

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeOSD writes a tiny 6x2 .osd file with two frames and returns its path.
// Frame 0 reads "AB" followed by the RSSI icon; frame 1 reads "AC".
func writeOSD(t *testing.T) string {
	t.Helper()
	return writeFrames(t, 6, 2, [][]uint16{{'A', 'B', 0x01}, {'A', 'C'}})
}

// writeFrames writes an .osd file with a cols x rows grid and one frame per
// entry of frames, 151 ms apart. Each frame lists its first cells; the rest
// are blank.
func writeFrames(t *testing.T, cols, rows int, frames [][]uint16) string {
	t.Helper()
	b := make([]byte, 40)
	copy(b, "INAV")
	binary.LittleEndian.PutUint16(b[0x24:], uint16(cols))
	binary.LittleEndian.PutUint16(b[0x26:], uint16(rows))
	for i, text := range frames {
		b = binary.LittleEndian.AppendUint32(b, uint32(i)*151)
		for j := 0; j < cols*rows; j++ {
			c := uint16(' ')
			if j < len(text) {
				c = text[j]
			}
			b = binary.LittleEndian.AppendUint16(b, c)
		}
	}
	p := filepath.Join(t.TempDir(), "x.osd")
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// writeHeaderOnlyOSD writes an .osd file that has a valid header and no frames.
func writeHeaderOnlyOSD(t *testing.T) string {
	t.Helper()
	b := make([]byte, 40)
	copy(b, "INAV")
	binary.LittleEndian.PutUint16(b[0x24:], 6)
	binary.LittleEndian.PutUint16(b[0x26:], 2)
	p := filepath.Join(t.TempDir(), "empty.osd")
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRunMain(t *testing.T) {
	osdFile := writeOSD(t)
	headerOnly := writeHeaderOnlyOSD(t)
	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{"no arguments", nil, 2, "", "usage:"},
		{"help", []string{"help"}, 0, "fpv-osd csv", ""},
		{"-h", []string{"-h"}, 0, "fpv-osd dump", ""},
		{"unknown command", []string{"frobnicate"}, 2, "", `unknown command "frobnicate"`},
		{"csv without a file", []string{"csv"}, 2, "", "usage: fpv-osd csv"},
		{"csv -h", []string{"csv", "-h"}, 0, "", "-decimal"},
		{"csv bad flag", []string{"csv", "-nope", osdFile}, 2, "", "flag provided but not defined"},
		{"csv equal separators", []string{"csv", "-sep", ",", "-decimal", ",", osdFile}, 2, "", "cannot be the same"},
		{"csv missing file", []string{"csv", "/nonexistent/x.osd"}, 1, "", "error:"},
		{"dump without a file", []string{"dump"}, 2, "", "usage: fpv-osd dump"},
		{"dump -h", []string{"dump", "-h"}, 0, "", "-census"},
		{"dump missing file", []string{"dump", "/nonexistent/x.osd"}, 1, "", "error:"},
		{"dump frame 1", []string{"dump", "-frame", "1", osdFile}, 0, "=== frame 1  t=151 ms ===\nAC", ""},
		{"dump hex names glyphs", []string{"dump", "-hex", osdFile}, 0, "AB{RSSI}", ""},
		{"dump region", []string{"dump", "-region", "0,0,3", osdFile}, 0, `1  "AB{RSSI}"`, ""},
		{"dump region malformed", []string{"dump", "-region", "oops", osdFile}, 2, "", "-region must be x,y,width"},
		{"dump region outside grid", []string{"dump", "-region", "0,9,3", osdFile}, 2, "", "outside the 6x2 grid"},
		{"dump census", []string{"dump", "-census", osdFile}, 0, "{RSSI}", ""},
		{"csv bad -mode", []string{"csv", "-mode", "1,2", osdFile}, 2, "", "-mode must be x,y,width"},
		{"csv bad -msg-row", []string{"csv", "-msg-row", "top", osdFile}, 2, "", "-msg-row must be a row number or auto"},
		{"csv says what the layout is", []string{"csv", "-o", "-", osdFile}, 0, "", "layout: flight mode not found (give -mode x,y,width); status messages not found (give -msg-row)"},
		{"csv bad -sep", []string{"csv", "-sep", ";;", osdFile}, 2, "", "invalid -sep/-decimal"},
		{"csv unknown profile", []string{"csv", "-profile", "frobnicate", osdFile}, 2, "", `unknown profile "frobnicate"`},
		{"csv maps profile has no separator", []string{"csv", "-profile", "maps", "-sep", ";", osdFile}, 2, "", "do not apply to -profile maps"},
		{"csv maps profile has no decimal", []string{"csv", "-profile", "maps", "-decimal", ",", osdFile}, 2, "", "do not apply to -profile maps"},
		{"csv maps profile without GPS", []string{"csv", "-profile", "maps", "-o", "-", osdFile}, 1, "", "profile maps: the recording has no valid GPS position"},
		{"csv explicit default profile", []string{"csv", "-q", "-profile", "default", "-o", "-", osdFile}, 0, "t_ms\n0\n151\n", ""},
		{"csv to stdout", []string{"csv", "-q", "-o", "-", osdFile}, 0, "t_ms\n0\n151\n", ""},
		{"csv prints the summary", []string{"csv", "-o", "-", osdFile}, 0, "", "OSD: 6x2 grid, 2 frames"},
		{"csv -unmapped", []string{"csv", "-q", "-unmapped", "-o", "-", osdFile}, 0, "", "AB"},
		{"dump file without frames", []string{"dump", headerOnly}, 1, "", "no frames after the header"},
		{"csv file without frames", []string{"csv", headerOnly}, 1, "", "no frames after the header"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if got := runMain(tt.args, &stdout, &stderr); got != tt.wantCode {
				t.Errorf("exit code = %d, want %d\nstdout: %s\nstderr: %s", got, tt.wantCode, &stdout, &stderr)
			}
			if !strings.Contains(stdout.String(), tt.wantStdout) {
				t.Errorf("stdout = %q, want it to contain %q", &stdout, tt.wantStdout)
			}
			if !strings.Contains(stderr.String(), tt.wantStderr) {
				t.Errorf("stderr = %q, want it to contain %q", &stderr, tt.wantStderr)
			}
		})
	}
}

// "csv" on a file without a sibling .srt must still produce a CSV.
func TestCSVWithoutSRT(t *testing.T) {
	osdFile := writeOSD(t)
	out := filepath.Join(t.TempDir(), "out.csv")
	var stderr bytes.Buffer
	if code := runMain([]string{"csv", "-q", "-o", out, osdFile}, nil, &stderr); code != 0 {
		t.Fatalf("exit code = %d: %s", code, &stderr)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	// the tiny grid holds no telemetry, so only the time column remains
	if want := "t_ms\n0\n151\n"; string(got) != want {
		t.Errorf("csv = %q, want %q", got, want)
	}
}

// A profile other than the default must not overwrite the full CSV.
func TestCSVDefaultOutputNameHasTheProfile(t *testing.T) {
	data, err := os.ReadFile(sampleFlight)
	if err != nil {
		t.Skip("AvatarG0031.osd is not in the project")
	}
	dir := t.TempDir()
	osdFile := filepath.Join(dir, "flight.osd")
	if err := os.WriteFile(osdFile, data, 0o644); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		profile, want string
	}{
		{"maps", "flight.maps.kml"},
		{"default", "flight.csv"},
	} {
		var stderr bytes.Buffer
		if code := runMain([]string{"csv", "-q", "-profile", tt.profile, osdFile}, nil, &stderr); code != 0 {
			t.Fatalf("%s: exit code = %d: %s", tt.profile, code, &stderr)
		}
		if _, err := os.Stat(filepath.Join(dir, tt.want)); err != nil {
			t.Errorf("-profile %s: %v", tt.profile, err)
		}
	}
}

// An element no rule knows still gets a column, named after its glyphs, with
// the number as it is on screen.
func TestCSVGenericColumn(t *testing.T) {
	const telemetry0, scale = 0x06, 0x0D
	// frame 0: an unknown icon, 42 and an unknown unit; frame 1: the same with
	// a decimal number (3.69 is [3.][.6]9); frame 2: nothing
	osdFile := writeFrames(t, 12, 2, [][]uint16{
		{telemetry0, '4', '2', scale},
		{telemetry0, 0xA1 + 3, 0xB1 + 6, '9', scale},
		{},
	})
	out := filepath.Join(t.TempDir(), "out.csv")
	var stderr bytes.Buffer
	if code := runMain([]string{"csv", "-o", out, osdFile}, nil, &stderr); code != 0 {
		t.Fatalf("exit code = %d: %s", code, &stderr)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if want := "t_ms,osd_raw_telemetry_0__scale\n0,42\n151,3.69\n302,\n"; string(got) != want {
		t.Errorf("csv = %q, want %q", got, want)
	}
	if !strings.Contains(stderr.String(), "osd_raw_telemetry_0__scale") || !strings.Contains(stderr.String(), "no rule knows") {
		t.Errorf("the summary should list the generic column: %q", &stderr)
	}
}
