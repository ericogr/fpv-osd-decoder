// Package inav interprets the glyph codes of the INAV OSD font.
package inav

import (
	"fmt"
	"strings"
)

// Name returns a readable name for a glyph code, or "" for plain ASCII
// characters (or unknown codes).
func Name(c uint16) string {
	switch {
	case c >= TrailingDotFirst && c <= TrailingDotLast:
		return fmt.Sprintf("%d.", c-TrailingDotFirst)
	case c >= LeadingDotFirst && c <= LeadingDotLast:
		return fmt.Sprintf(".%d", c-LeadingDotFirst)
	case c >= decorationFirst && c <= decorationLast:
		return fmt.Sprintf("DECORATION_%d", c-decorationFirst)
	case c >= LogoFirst && c <= LogoLast:
		return fmt.Sprintf("LOGO_%d", c-LogoFirst)
	}
	return symbolNames[c]
}

// Text converts a sequence of numeric/ASCII glyphs into text. Digits with a
// half point are combined: [3.][.6]9 becomes "3.69". Non-text glyphs become
// '?'.
func Text(codes []uint16) string {
	var sb strings.Builder
	for _, c := range codes {
		switch {
		case c >= TrailingDotFirst && c <= TrailingDotLast:
			sb.WriteByte(byte('0' + c - TrailingDotFirst))
			sb.WriteByte('.')
		case c >= LeadingDotFirst && c <= LeadingDotLast:
			// the point was already written if the previous glyph had a half point on its right
			if s := sb.String(); !strings.HasSuffix(s, ".") {
				sb.WriteByte('.')
			}
			sb.WriteByte(byte('0' + c - LeadingDotFirst))
		case c >= 0x20 && c < 0x7f:
			sb.WriteByte(byte(c))
		default:
			sb.WriteByte('?')
		}
	}
	return sb.String()
}
