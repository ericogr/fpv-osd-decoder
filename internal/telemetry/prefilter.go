package telemetry

import (
	"regexp"
	"regexp/syntax"

	"github.com/ericogr/fpv-osd-decoder/internal/inav"
)

// Running seventy regular expressions on every row of every frame is slow, and
// almost all of them cannot match: a rule that reads the altitude needs one of
// four unit glyphs, and most rows have none. The decoder therefore asks of each
// rule which glyphs it needs, and skips the rule on a row that has none of
// them. This is only a shortcut: it must never change what a rule reads.

// maxNeeded is the largest character set worth checking for: a larger one
// (all the digits and half-point digits, say) is on almost every row and would
// not skip anything.
const maxNeeded = 24

// rowGlyphs says which glyph codes a row holds. A code above the font is not
// recorded: no rule needs one.
type rowGlyphs [inav.MaxGlyph + 1]bool

func glyphsOf(line string) (p rowGlyphs) {
	for _, r := range line {
		if r >= 0 && r <= inav.MaxGlyph {
			p[r] = true
		}
	}
	return p
}

// hasAny reports whether the row holds any of the codes. A nil list means the
// rule has no requirement, so it always runs.
func (p *rowGlyphs) hasAny(codes []rune) bool {
	if codes == nil {
		return true
	}
	for _, c := range codes {
		if p[c] {
			return true
		}
	}
	return false
}

// neededGlyphs returns the codes of which a row must hold at least one for re
// to match it, or nil when the expression gives no such short list. It looks
// at the parts of the expression that every match must contain and takes the
// smallest set of characters among them.
func neededGlyphs(re *regexp.Regexp) []rune {
	tree, err := syntax.Parse(re.String(), syntax.Perl)
	if err != nil {
		return nil
	}
	set := needed(tree)
	for _, c := range set {
		if c < 0 || c > inav.MaxGlyph {
			return nil
		}
	}
	return set
}

// needed returns a short list of characters of which any match of re contains
// at least one, or nil if it cannot tell.
func needed(re *syntax.Regexp) []rune {
	switch re.Op {
	case syntax.OpLiteral:
		// every match holds the whole literal, so its first character will do
		if len(re.Rune) > 0 && re.Flags&syntax.FoldCase == 0 {
			return re.Rune[:1]
		}
	case syntax.OpCharClass:
		var set []rune
		for i := 0; i+1 < len(re.Rune); i += 2 {
			for c := re.Rune[i]; c <= re.Rune[i+1]; c++ {
				if len(set) == maxNeeded {
					return nil
				}
				set = append(set, c)
			}
		}
		return set
	case syntax.OpCapture, syntax.OpPlus:
		return needed(re.Sub[0])
	case syntax.OpRepeat:
		if re.Min >= 1 {
			return needed(re.Sub[0])
		}
	case syntax.OpConcat:
		var best []rune
		for _, sub := range re.Sub {
			if set := needed(sub); set != nil && (best == nil || len(set) < len(best)) {
				best = set
			}
		}
		return best
	case syntax.OpAlternate:
		var all []rune
		for _, sub := range re.Sub {
			set := needed(sub)
			if set == nil {
				return nil
			}
			all = append(all, set...)
		}
		if len(all) <= maxNeeded {
			return all
		}
	}
	return nil
}
