package inav

import "testing"

func TestText(t *testing.T) {
	tests := []struct {
		name  string
		codes []uint16
		want  string
	}{
		{"integer", []uint16{'-', '7', '2'}, "-72"},
		{"point embedded in the glyphs", []uint16{0xA4, 0xB7, '9'}, "3.69"},
		{"coordinate", []uint16{'-', '2', 0xA4, 0xB4, '1', '1', '5', '8', '3'}, "-23.311583"},
		{"leading point only", []uint16{0xBA}, ".9"},
		{"time", []uint16{'0', '7', ':', '2', '3'}, "07:23"},
		{"non-text glyph", []uint16{'1', SymVolt}, "1?"},
	}
	for _, tt := range tests {
		if got := Text(tt.codes); got != tt.want {
			t.Errorf("%s: Text = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestName(t *testing.T) {
	tests := map[uint16]string{
		SymRssi:   "RSSI",
		0xA1:      "0.",
		0xAA:      "9.",
		0xB1:      ".0",
		0xBA:      ".9",
		0x17:      "DECORATION_0",
		0x1E:      "DECORATION_7",
		0x101:     "LOGO_0",
		'A':       "",
		SymDistKm: "DIST_KM",
	}
	for c, want := range tests {
		if got := Name(c); got != want {
			t.Errorf("Name(%#x) = %q, want %q", c, got, want)
		}
	}
}
