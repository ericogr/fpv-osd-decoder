package telemetry

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/ericogr/fpv-osd-decoder/internal/inav"
)

// emitFunc stores what a matched rule read into the sample. g holds the
// matched text (g[0]) and the capture groups of the rule's expression.
type emitFunc func(g []string, s Sample)

// rule recognizes one field on a row: a regular expression over the glyph
// codes and the function that turns a match into sample values.
type rule struct {
	re   *regexp.Regexp
	emit emitFunc
	// needs lists the glyphs of which a row must hold one for the rule to have
	// anything to read (see prefilter.go); nil means the rule always runs.
	needs []rune
}

func newRule(re *regexp.Regexp, emit emitFunc) rule {
	return rule{re: re, emit: emit, needs: neededGlyphs(re)}
}

// apply runs the rule on every place of line where it matches. A match emits
// its values and is erased from the line that comes back; one that does not
// match leaves the line unchanged.
//
// A match that would store a column the sample already has is left alone: the
// first reading in screen order wins and the other one stays on the line for
// the generic pass. Some INAV elements look exactly alike on screen (the
// capacity used and the capacity left, for example), so glyphs alone cannot
// tell them apart; this way neither reading is lost nor overwritten.
func (r rule) apply(line string, s Sample) string {
	var spans [][2]int
	for _, loc := range r.re.FindAllStringSubmatchIndex(line, -1) {
		g := make([]string, len(loc)/2)
		for i := range g {
			if loc[2*i] >= 0 {
				g[i] = line[loc[2*i]:loc[2*i+1]]
			}
		}
		got := Sample{}
		r.emit(g, got)
		if s.hasAny(got) {
			continue
		}
		for key, v := range got {
			s[key] = v
		}
		spans = append(spans, [2]int{loc[0], loc[1]})
	}
	return erase(line, spans)
}

// erase replaces the spans (byte offsets, in order) of line with one blank per
// glyph.
func erase(line string, spans [][2]int) string {
	// from the end, so the offsets of the spans still to do stay valid
	for i := len(spans) - 1; i >= 0; i-- {
		from, to := spans[i][0], spans[i][1]
		line = line[:from] + strings.Repeat(" ", utf8.RuneCountInString(line[from:to])) + line[to:]
	}
	return line
}

// Regular expression fragments. Each OSD cell becomes a rune holding the
// glyph code, so \x{NN} matches glyph NN.
const spaceRE = ` *`

// numRE is an INAV number: ASCII digits, sign, point and the digits with half a
// decimal point ("3." and ".6"). It is one capture group.
var numRE = fmt.Sprintf(`(-?[0-9.\x{%x}-\x{%x}\x{%x}-\x{%x}]+)`,
	inav.TrailingDotFirst, inav.TrailingDotLast, inav.LeadingDotFirst, inav.LeadingDotLast)

func sym(c uint16) string { return fmt.Sprintf(`\x{%x}`, c) }

// syms is a character class of glyphs.
func syms(cs ...uint16) string {
	var sb strings.Builder
	sb.WriteByte('[')
	for _, c := range cs {
		sb.WriteString(sym(c))
	}
	sb.WriteByte(']')
	return sb.String()
}

func symRange(lo, hi uint16) string { return fmt.Sprintf(`[\x{%x}-\x{%x}]`, lo, hi) }

// unitsRE is a capture group matching any of the unit glyphs of the table.
func unitsRE(units map[uint16]float64) string {
	cs := make([]uint16, 0, len(units))
	for c := range units {
		cs = append(cs, c)
	}
	sort.Slice(cs, func(i, j int) bool { return cs[i] < cs[j] })
	return "(" + syms(cs...) + ")"
}

func rx(parts ...string) *regexp.Regexp { return regexp.MustCompile(strings.Join(parts, "")) }

// scaledReading is the rule for a number followed by a unit glyph, with an
// optional icon before it: the icon, blanks, the number and the unit. The value
// is the number times the factor the table gives for the unit glyph. prefix is
// the expression of the icon and has no capture group of its own. With no
// prefix the reading has no icon, which makes it the generic one for its units.
func scaledReading(key string, units map[uint16]float64, prefix ...string) rule {
	parts := prefix
	if len(prefix) > 0 {
		parts = append(parts[:len(parts):len(parts)], spaceRE)
	}
	parts = append(parts, numRE, unitsRE(units))
	return newRule(rx(parts...), scaled(key, 1, 2, units))
}

// codes returns the glyph codes of a matched span.
func codes(g string) []uint16 {
	cs := make([]uint16, 0, len(g))
	for _, r := range g {
		cs = append(cs, uint16(r))
	}
	return cs
}

// firstGlyph returns the code of the first glyph of a matched span.
func firstGlyph(g string) uint16 { return codes(g)[0] }

func parseNum(g string) (float64, bool) {
	v, err := strconv.ParseFloat(inav.Text(codes(g)), 64)
	return v, err == nil
}

// Emitters shared by several rules. idx is the capture group they read.

// numField emits the INAV number in group idx.
func numField(key string, idx int) emitFunc {
	return func(g []string, s Sample) {
		if v, ok := parseNum(g[idx]); ok {
			s.num(key, v)
		}
	}
}

// intField emits the plain integer in group idx.
func intField(key string, idx int) emitFunc {
	return func(g []string, s Sample) {
		if v, err := strconv.Atoi(g[idx]); err == nil {
			s.num(key, float64(v))
		}
	}
}

// textField emits the text in group idx, without the padding blanks.
func textField(key string, idx int) emitFunc {
	return func(g []string, s Sample) {
		if t := strings.Join(strings.Fields(g[idx]), " "); t != "" {
			s[key] = Value{Text: t}
		}
	}
}

// noReading is for the rules that recognize an element only to take it off the
// screen: it is there, but shows nothing to read.
func noReading([]string, Sample) {}

// each runs several emitters on the same match.
func each(fns ...emitFunc) emitFunc {
	return func(g []string, s Sample) {
		for _, fn := range fns {
			fn(g, s)
		}
	}
}

// scaled emits key = number * unit factor, where g[ni] is the number and
// g[ui] the unit glyph.
func scaled(key string, ni, ui int, units map[uint16]float64) emitFunc {
	return func(g []string, s Sample) {
		if v, ok := parseNum(g[ni]); ok {
			s.num(key, v*units[firstGlyph(g[ui])])
		}
	}
}

// signedByIcon emits the number in group 2, negated when the icon in group 1
// is the given one.
func signedByIcon(key string, negative uint16) emitFunc {
	return func(g []string, s Sample) {
		v, ok := parseNum(g[2])
		if !ok {
			return
		}
		if firstGlyph(g[1]) == negative {
			v = -v
		}
		s.num(key, v)
	}
}
