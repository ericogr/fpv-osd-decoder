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

func TestIsLogo(t *testing.T) {
	tests := map[uint16]bool{
		SymLogoStart - 1:  false,
		SymLogoStart:      true,
		SymLogoStart + 39: true, // the last of the 4x10 pieces
		SymLogoStart + 40: false,
		'A':               false,
		SymRssi:           false,
	}
	for c, want := range tests {
		if got := IsLogo(c); got != want {
			t.Errorf("IsLogo(%#x) = %v, want %v", c, got, want)
		}
	}
}

// The glyphs the rules built from the firmware source rely on keep the codes
// that osd_symbols.h of INAV 9.1.0 gives them.
func TestKeyGlyphCodes(t *testing.T) {
	tests := map[string]uint16{
		"SYM_ODOMETER": SymOdometer, "SYM_RX_BAND": SymRxBand, "SYM_FLIGHT_DIST_REMAINING": SymFlightDistRemaining,
		"SYM_TEMP_C": SymTempC, "SYM_WH": SymWh, "SYM_MAX": SymMax, "SYM_AH_DECORATION_DOWN": SymAhDecorationDown,
	}
	want := map[string]uint16{
		"SYM_ODOMETER": 0x168, "SYM_RX_BAND": 0x169, "SYM_FLIGHT_DIST_REMAINING": 0x167,
		"SYM_TEMP_C": 0x97, "SYM_WH": 0x6D, "SYM_MAX": 0xCE, "SYM_AH_DECORATION_DOWN": 0x16,
	}
	for name, got := range tests {
		if got != want[name] {
			t.Errorf("%s = %#x, want %#x", name, got, want[name])
		}
	}
}

func TestIsGlyph(t *testing.T) {
	tests := []struct {
		name string
		c    uint16
		want bool
	}{
		{"nothing", 0, false},
		{"blank", ' ', false},
		{"a low icon", 0x01, true},
		{"the last low icon", 0x1f, true},
		{"a digit", '5', false},
		{"an uppercase letter", 'A', false},
		{"punctuation", '-', false},
		{"the quote that stands for Ah/km", '"', false},
		{"a lowercase code, where units sit", 'v', true},
		{"the tilde", '~', true},
		{"a unit above ASCII", 0x90, true},
		{"a digit with a point on its right", 0xA1, false},
		{"a digit with a point on its left", 0xB1, false},
		{"the code after those digits", 0xBC, true},
		{"the artificial horizon", 0x14d, false},
		{"the crosshair", 0x190, false},
		{"a piece of the logo", 0x105, false},
		{"the last code of the font", MaxGlyph, true},
		{"above the font", MaxGlyph + 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsGlyph(tt.c); got != tt.want {
				t.Errorf("IsGlyph(%#x) = %v, want %v", tt.c, got, tt.want)
			}
		})
	}
}

func TestIsHalfDot(t *testing.T) {
	tests := map[uint16]bool{
		'0': false, 0xA0: false, 0xA1: true, 0xAA: true, 0xAB: false,
		0xB0: false, 0xB1: true, 0xBA: true, 0xBB: false,
	}
	for c, want := range tests {
		if got := IsHalfDot(c); got != want {
			t.Errorf("IsHalfDot(%#x) = %v, want %v", c, got, want)
		}
	}
}

func TestIsGraphic(t *testing.T) {
	tests := map[uint16]bool{
		SymRssi: false, 'A': false, 0x12b: false, 0x12c: true, 0x133: true, 0x134: false,
		0x14c: true, 0x164: true, 0x165: false, 0x166: true, 0x190: true, 0x1a6: true,
		0x1ae: true, 0x1b9: true, 0x1ba: false,
	}
	for c, want := range tests {
		if got := IsGraphic(c); got != want {
			t.Errorf("IsGraphic(%#x) = %v, want %v", c, got, want)
		}
	}
}

func TestIsFlightMode(t *testing.T) {
	tests := map[string]bool{
		"ACRO": true, "ANGL": true, "HOR ": true, " WP ": true, " AH ": true, "RTH ": true, "!FS!": true,
		"RTH":  false, // without the padding blank
		"WP":   false,
		"ACR ": false,
		"acro": false,
		"":     false,
	}
	for text, want := range tests {
		if got := IsFlightMode(text); got != want {
			t.Errorf("IsFlightMode(%q) = %v, want %v", text, got, want)
		}
	}
	for text := range flightModes {
		if len(text) != FlightModeWidth {
			t.Errorf("flight mode %q is not %d characters wide", text, FlightModeWidth)
		}
	}
}
