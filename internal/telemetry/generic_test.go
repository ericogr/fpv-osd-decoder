package telemetry

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/ericogr/fpv-osd-decoder/internal/inav"
)

// Glyphs no rule reads, to stand for an element INAV adds in the future.
const (
	unknownIcon uint16 = inav.SymTelemetry0
	unknownUnit uint16 = inav.SymScale
)

func TestGenericPassReadsWhatNoRuleKnows(t *testing.T) {
	tests := []struct {
		name  string
		parts []any
		want  map[string]float64
	}{
		{"icon and number", []any{unknownIcon, "42"}, map[string]float64{"osd_raw_telemetry_0": 42}},
		{"icon, blanks and number", []any{unknownIcon, "   42"}, map[string]float64{"osd_raw_telemetry_0": 42}},
		{"number and unit", []any{"42", unknownUnit}, map[string]float64{"osd_raw_scale": 42}},
		{"icon, number and unit", []any{unknownIcon, " 15", unknownUnit}, map[string]float64{"osd_raw_telemetry_0__scale": 15}},
		{"an icon made of two glyphs", []any{inav.SymTelemetry0, inav.SymTelemetry1, "7"}, map[string]float64{"osd_raw_telemetry_0__telemetry_1": 7}},
		{"decimals are kept as shown", []any{unknownIcon, inavNum("3.69")}, map[string]float64{"osd_raw_telemetry_0": 3.69}},
		{"negative number", []any{unknownIcon, "-12"}, map[string]float64{"osd_raw_telemetry_0": -12}},
		{"a glyph the font header does not name is named by its code", []any{unnamedGlyph(t), "5"}, map[string]float64{fmt.Sprintf("osd_raw_x%x", unnamedGlyph(t)): 5}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFile()
			put(f, 10, 5, tt.parts...)
			s, left := New(sampleLayout()).Decode(f, &f.Frames[0])
			checkNum(t, s, tt.want)
			if len(s) != len(tt.want) {
				t.Errorf("read %v, want only %v", s, tt.want)
			}
			if len(left) != 0 {
				t.Errorf("left on screen after being read: %+v", left)
			}
		})
	}
}

// unnamedGlyph returns a code of the font that osd_symbols.h does not name.
func unnamedGlyph(t *testing.T) uint16 {
	t.Helper()
	for c := uint16(0x01); c <= inav.MaxGlyph; c++ {
		if inav.IsGlyph(c) && inav.Name(c) == "" {
			return c
		}
	}
	t.Skip("every glyph of the font has a name")
	return 0
}

func TestGenericPassIgnoresWhatItCannotTrust(t *testing.T) {
	tests := []struct {
		name  string
		parts []any
	}{
		{"a number with nothing around it", []any{"42"}},
		{"a number inside text", []any{"WP CNT>", " 5"}},
		{"a label", []any{"M0"}},
		{"an icon with no number", []any{unknownIcon, "---"}},
		{"an icon alone", []any{unknownIcon}},
		{"a graphic element", []any{uint16(0x14d), "5"}},
		{"a number that is not one", []any{unknownIcon, "1.2.3"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFile()
			put(f, 10, 5, tt.parts...)
			if s := decode(t, f); len(s) != 0 {
				t.Errorf("read %v, want nothing", s)
			}
		})
	}
}

// What the rules read must not be read twice, and the generic pass must not
// change what the rules took.
func TestGenericPassLeavesTheRulesAlone(t *testing.T) {
	f := frame3000()
	put(f, 3, 12, unknownIcon, "42")
	s := decode(t, f)
	for key := range s {
		if isExtraKey(key) && key != "osd_raw_telemetry_0" {
			t.Errorf("unexpected generic column %s", key)
		}
	}
	checkNum(t, s, map[string]float64{"osd_raw_telemetry_0": 42, "osd_altitude_m": 213, "osd_rssi_dbm": -72})
}

// The same unknown element twice on the screen is one column: the first wins
// and the second stays among the leftovers.
func TestGenericPassKeepsTheFirstOfTwo(t *testing.T) {
	f := newFile()
	put(f, 3, 4, unknownIcon, "42")
	put(f, 20, 9, unknownIcon, "99")
	s, left := New(sampleLayout()).Decode(f, &f.Frames[0])
	checkNum(t, s, map[string]float64{"osd_raw_telemetry_0": 42})
	if len(left) != 1 || left[0].Y != 9 {
		t.Errorf("leftovers = %+v, want only the reading on row 9", left)
	}
}

func TestExtraFields(t *testing.T) {
	samples := []Sample{
		{"osd_raw_b": {Num: 1}, "osd_altitude_m": {Num: 2}},
		{"osd_raw_a": {Num: 3}, "osd_raw_b": {Num: 4}},
		{},
	}
	got := extraFields(samples)
	var keys []string
	for _, f := range got {
		keys = append(keys, f.Key)
		if f.Decimals >= 0 || f.Text || f.Desc == "" {
			t.Errorf("%s: %+v should keep all digits, be numeric and have a description", f.Key, f)
		}
	}
	if want := []string{"osd_raw_a", "osd_raw_b"}; !reflect.DeepEqual(keys, want) {
		t.Errorf("keys = %v, want %v", keys, want)
	}
	if extraFields(nil) != nil && len(extraFields(nil)) != 0 {
		t.Error("no samples, no fields")
	}
}

// A generic key can never be a catalog key, or two columns would collide.
func TestExtraKeysDoNotCollideWithTheCatalog(t *testing.T) {
	for _, f := range Fields() {
		if isExtraKey(f.Key) {
			t.Errorf("catalog key %q starts with the prefix of the generic columns", f.Key)
		}
	}
}

// The separator between glyph names must never be part of a name.
func TestGlyphIDsAreSafeToJoin(t *testing.T) {
	for c := uint16(1); c <= inav.MaxGlyph; c++ {
		if !inav.IsGlyph(c) {
			continue // only glyphs can be in a name
		}
		id := glyphID(c)
		if id == "" || strings.Contains(id, glyphSeparator) || strings.HasPrefix(id, "_") || strings.HasSuffix(id, "_") {
			t.Errorf("glyph %#x is named %q, which cannot be joined unambiguously", c, id)
		}
	}
}

// Two different sequences of glyphs must not give the same column name, or two
// readings would fight for one column. This checks every sequence of one and
// two glyphs; TestGlyphIDsAreSafeToJoin covers the longer ones.
func TestExtraKeysAreUnique(t *testing.T) {
	var glyphs []uint16
	for c := uint16(1); c <= inav.MaxGlyph; c++ {
		if inav.IsGlyph(c) {
			glyphs = append(glyphs, c)
		}
	}
	seen := map[string]string{}
	add := func(codes ...uint16) {
		var sequence []rune
		for _, c := range codes {
			sequence = append(sequence, rune(c))
		}
		key := extraKey(string(sequence))
		if other, dup := seen[key]; dup {
			t.Errorf("%q and %q both give %s", other, string(sequence), key)
		}
		seen[key] = string(sequence)
	}
	for _, a := range glyphs {
		add(a)
		for _, b := range glyphs {
			add(a, b)
		}
	}
}
