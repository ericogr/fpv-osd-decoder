package telemetry

import (
	"math/rand"
	"os"
	"reflect"
	"regexp"
	"slices"
	"testing"

	"github.com/ericogr/fpv-osd-decoder/internal/inav"
	"github.com/ericogr/fpv-osd-decoder/internal/osd"
)

func TestNeededGlyphs(t *testing.T) {
	tests := []struct {
		name string
		re   string
		want []rune
	}{
		{"a literal", `CH:([A-Z])`, []rune{'C'}},
		{"an icon before a number", `\x{10} *(-?[0-9.]+)`, []rune{0x10}},
		{"a number before a short list of units", `(-?[0-9.]+)([\x{76}-\x{79}])`, []rune{0x76, 0x77, 0x78, 0x79}},
		{"the smaller of two sets", `[\x{1}\x{2}\x{3}] *[\x{5}\x{6}]`, []rune{5, 6}},
		{"alternatives are joined", `(\x{1}|\x{2}\x{3})`, []rune{1, 2}},
		{"a repeated part that is required", `(\x{7})+`, []rune{7}},
		{"an optional part is not required", `(?:\x{7})? *[0-9]+`, []rune("0123456789")},
		{"a short class", `[0-9]+`, []rune("0123456789")},
		{"only a long class", `[0-9A-Za-z]+`, nil},
		{"only a star", `\x{7}*`, nil},
		{"a code outside the font", `\x{3000}`, nil},
		{"an alternative with no requirement", `(\x{1}|[0-9A-Za-z]+)`, nil},
		{"a word boundary is ignored", `\bG([0-3]):`, []rune{'G'}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := neededGlyphs(regexp.MustCompile(tt.re))
			slices.Sort(got)
			if !slices.Equal(got, tt.want) {
				t.Errorf("neededGlyphs(%s) = %v, want %v", tt.re, got, tt.want)
			}
		})
	}
}

func TestRowGlyphs(t *testing.T) {
	p := glyphsOf("AB\x01 ā")
	for _, c := range []rune{'A', 'B', 1, 0x101} {
		if !p[c] {
			t.Errorf("code %#x should be present", c)
		}
	}
	if p['C'] || p[2] {
		t.Error("absent codes reported as present")
	}
	if !p.hasAny(nil) {
		t.Error("a rule with no requirement must always run")
	}
	if p.hasAny([]rune{'C', 2}) || !p.hasAny([]rune{'C', 'B'}) {
		t.Error("hasAny is wrong")
	}
	// codes the font does not have are ignored, not a crash
	glyphsOf("　")
}

// A frame filled with glyphs, digits and letters at random, from a fixed seed:
// the shortcut must give the same readings as running every rule.
func randomFrame(rng *rand.Rand) *osd.File {
	f := newFile()
	cells := f.Frames[0].Cells
	alphabet := []uint16{' ', ' ', ' ', ' ', '-', '.', ':', '%', '*'}
	for c := '0'; c <= '9'; c++ {
		alphabet = append(alphabet, uint16(c))
	}
	for c := uint16(0x01); c <= inav.MaxGlyph; c += 7 {
		alphabet = append(alphabet, c)
	}
	for _, c := range []uint16{inav.SymHome, inav.SymAltM, inav.SymKmh, inav.SymMax, inav.SymAir, inav.SymMah, inav.SymVolt,
		inav.SymAmp, inav.SymDbm, inav.SymRssi, inav.SymLq, inav.SymSatL, inav.SymSatR, 'G', 'F', 'D', 'C', 'H', 0xA3, 0xB4} {
		alphabet = append(alphabet, c, c, c)
	}
	for i := range cells {
		cells[i] = alphabet[rng.Intn(len(alphabet))]
	}
	return f
}

func TestPrefilterDoesNotChangeResults(t *testing.T) {
	filtered := New(sampleLayout())
	everything := New(sampleLayout())
	for i := range everything.rules {
		everything.rules[i].needs = nil // every rule runs on every row
	}
	withFilter := 0
	for _, r := range filtered.rules {
		if r.needs != nil {
			withFilter++
		}
	}
	if withFilter < len(filtered.rules)/2 {
		t.Errorf("only %d of %d rules have a filter: the shortcut is not working", withFilter, len(filtered.rules))
	}

	check := func(name string, f *osd.File, fr *osd.Frame) {
		t.Helper()
		a, leftA := filtered.Decode(f, fr)
		b, leftB := everything.Decode(f, fr)
		if !reflect.DeepEqual(a, b) || !reflect.DeepEqual(leftA, leftB) {
			t.Errorf("%s: the filter changed the result\n with:    %v\n without: %v", name, a, b)
		}
	}
	for name, f := range map[string]*osd.File{"flight": frame3000(), "every element": everyElement()} {
		check(name, f, &f.Frames[0])
	}
	rng := rand.New(rand.NewSource(1))
	for i := 0; i < 400; i++ {
		f := randomFrame(rng)
		check("random frame", f, &f.Frames[0])
	}
	if data, err := os.ReadFile("../../AvatarG0031.osd"); err == nil {
		f, err := osd.Parse(data)
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < len(f.Frames); i += 11 {
			check("sample flight", f, &f.Frames[i])
		}
	}
}
