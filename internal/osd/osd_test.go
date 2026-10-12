package osd

import (
	"encoding/binary"
	"strings"
	"testing"
)

// build assembles an .osd file with the given grid and one timestamp per frame.
func build(cols, rows int, times []uint32, fill uint16) []byte {
	b := make([]byte, headerSize)
	copy(b, magic)
	binary.LittleEndian.PutUint16(b[colsOffset:], uint16(cols))
	binary.LittleEndian.PutUint16(b[rowsOffset:], uint16(rows))
	for _, ts := range times {
		b = binary.LittleEndian.AppendUint32(b, ts)
		for i := 0; i < cols*rows; i++ {
			b = binary.LittleEndian.AppendUint16(b, fill+uint16(i))
		}
	}
	return b
}

func TestParse(t *testing.T) {
	f, err := Parse(build(53, 20, []uint32{0, 151, 302}, 0x20))
	if err != nil {
		t.Fatal(err)
	}
	if f.Cols != 53 || f.Rows != 20 || len(f.Frames) != 3 {
		t.Fatalf("grid %dx%d with %d frames", f.Cols, f.Rows, len(f.Frames))
	}
	if f.Frames[2].TimeMs != 302 {
		t.Errorf("TimeMs = %d, want 302", f.Frames[2].TimeMs)
	}
	// cell (x=2, y=1) has index 1*53+2 = 55
	if got := f.Row(&f.Frames[1], 1)[2]; got != 0x20+55 {
		t.Errorf("cell (2,1) = %#x, want %#x", got, 0x20+55)
	}
}

func TestRowAndDuration(t *testing.T) {
	f, err := Parse(build(4, 3, []uint32{0, 151, 302}, 0x20))
	if err != nil {
		t.Fatal(err)
	}
	row := f.Row(&f.Frames[1], 2) // cells 8..11 of the frame
	if len(row) != 4 || row[0] != 0x20+8 || row[3] != 0x20+11 {
		t.Errorf("Row(1, 2) = %v", row)
	}
	if got := f.Duration(); got != 302 {
		t.Errorf("Duration = %d, want 302", got)
	}
}

func TestParseErrors(t *testing.T) {
	bad := build(53, 20, []uint32{0}, 0x20)
	copy(bad, "XXXX")
	tests := map[string]struct {
		data []byte
		want string
	}{
		"short":           {[]byte("INAV"), "shorter than the header"},
		"wrong magic":     {bad, "magic"},
		"truncated frame": {build(53, 20, []uint32{0, 1}, 0x20)[:headerSize+2124+10], "multiple"},
		"zero grid":       {build(0, 20, nil, 0), "invalid grid"},
		"no frames":       {build(53, 20, nil, 0), "no frames"},
	}
	for name, tt := range tests {
		_, err := Parse(tt.data)
		if err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("%s: error = %v, want it to contain %q", name, err, tt.want)
		}
	}
}
