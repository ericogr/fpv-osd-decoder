package telemetry

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/ericogr/fpv-osd-decoder/internal/inav"
)

// The generic pass takes the readings no rule recognized: a number anchored by
// glyphs. Each becomes a column named after those glyphs, with the value as
// shown on screen.

// extraPrefix starts the key of every column made by the generic pass.
const extraPrefix = "osd_raw_"

func isExtraKey(key string) bool { return strings.HasPrefix(key, extraPrefix) }

// extraFields describes the columns the generic pass made in the samples,
// sorted by key. The value is exactly what the OSD shows: no unit conversion
// and no rounding.
func extraFields(samples []Sample) []Field {
	seen := map[string]bool{}
	for _, s := range samples {
		for key := range s {
			if isExtraKey(key) {
				seen[key] = true
			}
		}
	}
	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	fields := make([]Field, len(keys))
	for i, key := range keys {
		fields[i] = Field{
			Key:      key,
			Desc:     "a reading no rule knows, as shown on screen (no unit conversion); the glyphs around the number make up the name",
			Decimals: -1,
		}
	}
	return fields
}

// glyphClass builds the regular expression character class of the glyph cells.
func glyphClass() string {
	var sb strings.Builder
	sb.WriteByte('[')
	for c := uint16(1); c <= inav.MaxGlyph; {
		if !inav.IsGlyph(c) {
			c++
			continue
		}
		first := c
		for c <= inav.MaxGlyph && inav.IsGlyph(c) {
			c++
		}
		fmt.Fprintf(&sb, `\x{%x}-\x{%x}`, first, c-1)
	}
	sb.WriteByte(']')
	return sb.String()
}

// genericRE matches a number with glyphs before it (an icon, perhaps with
// blanks after it), after it (a unit), or both. Groups: 1 the glyphs before,
// 2 the number, 3 the glyphs after.
var genericRE = regexp.MustCompile(`(` + glyphClass() + `*) *` + numRE + `(` + glyphClass() + `*)`)

// readGeneric takes the readings no rule recognized and returns the line with
// them erased. A number with no glyph next to it is not taken (it could be
// anything), nor is text, since there is no telling what it means. If the same
// glyphs show a number twice, the first keeps the column and the other stays on
// the line.
func readGeneric(line string, s Sample) string {
	var spans [][2]int
	for _, loc := range genericRE.FindAllStringSubmatchIndex(line, -1) {
		before, number, after := line[loc[2]:loc[3]], line[loc[4]:loc[5]], line[loc[6]:loc[7]]
		if before == "" && after == "" {
			continue
		}
		v, ok := parseNum(number)
		if !ok {
			continue
		}
		key := extraKey(before + after)
		if _, taken := s[key]; taken {
			continue
		}
		s.num(key, v)
		spans = append(spans, [2]int{loc[0], loc[1]})
	}
	return erase(line, spans)
}

// glyphSeparator joins the names of the glyphs in a column name. A glyph name
// has single underscores inside it (hdp_l) and none at its ends, so a double one
// cannot be mistaken for part of a name: without it, the glyphs WH and KM, side
// by side, would give the same column as the single glyph WH_KM.
const glyphSeparator = "__"

// extraKey names the column of a reading after the glyphs around the number:
// osd_raw_ and the names in the font header joined by glyphSeparator, such as
// osd_raw_hdp_l__hdp_r.
func extraKey(glyphs string) string {
	names := make([]string, 0, len(glyphs))
	for _, c := range glyphs {
		names = append(names, glyphID(uint16(c)))
	}
	return extraPrefix + strings.Join(names, glyphSeparator)
}

// glyphID is the name of a glyph in lowercase letters, digits and single
// underscores, or its code when the font header has no name for it.
func glyphID(c uint16) string {
	name := inav.Name(c)
	if name == "" {
		return fmt.Sprintf("x%x", c)
	}
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return '_'
	}, name)
}
