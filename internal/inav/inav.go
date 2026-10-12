// Package inav interprets the glyph codes of the INAV OSD font.
package inav

import (
	"fmt"
	"strings"
)

// Glyph ranges that the header only defines by their first code.
const (
	trailingDotFirst = 0xA1 // SYM_ZERO_HALF_TRAILING_DOT: "0." ... "9."
	trailingDotLast  = 0xAA
	leadingDotFirst  = 0xB1 // SYM_ZERO_HALF_LEADING_DOT: ".0" ... ".9"
	leadingDotLast   = 0xBA
	decorationFirst  = 0x17 // SYM_DECORATION: small directional arrows
	decorationLast   = 0x1E
	logoFirst        = 0x101 // SYM_LOGO_START: INAV logo (4x10)
	logoLast         = 0x101 + 40 - 1
)

// Name returns a readable name for a glyph code, or "" for plain ASCII
// characters (or unknown codes).
func Name(c uint16) string {
	switch {
	case c >= trailingDotFirst && c <= trailingDotLast:
		return fmt.Sprintf("%d.", c-trailingDotFirst)
	case c >= leadingDotFirst && c <= leadingDotLast:
		return fmt.Sprintf(".%d", c-leadingDotFirst)
	case c >= decorationFirst && c <= decorationLast:
		return fmt.Sprintf("DECORATION_%d", c-decorationFirst)
	case c >= logoFirst && c <= logoLast:
		return fmt.Sprintf("LOGO_%d", c-logoFirst)
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
		case c >= trailingDotFirst && c <= trailingDotLast:
			sb.WriteByte(byte('0' + c - trailingDotFirst))
			sb.WriteByte('.')
		case c >= leadingDotFirst && c <= leadingDotLast:
			// the point was already written if the previous glyph had a half point on its right
			if s := sb.String(); !strings.HasSuffix(s, ".") {
				sb.WriteByte('.')
			}
			sb.WriteByte(byte('0' + c - leadingDotFirst))
		case c >= 0x20 && c < 0x7f:
			sb.WriteByte(byte(c))
		default:
			sb.WriteByte('?')
		}
	}
	return sb.String()
}
