package inav

// What the font layout says about a code, apart from its name: which codes are
// digits, graphics or icons. Nothing outside this package should repeat these
// ranges.

// MaxGlyph is the last code of the OSD font, which has 512 characters.
const MaxGlyph = 0x1ff

// Glyph ranges that the header only defines by their first code.
const (
	TrailingDotFirst = 0xA1 // SYM_ZERO_HALF_TRAILING_DOT: "0." ... "9."
	TrailingDotLast  = 0xAA
	LeadingDotFirst  = 0xB1 // SYM_ZERO_HALF_LEADING_DOT: ".0" ... ".9"
	LeadingDotLast   = 0xBA
	decorationFirst  = 0x17 // SYM_DECORATION: small directional arrows
	decorationLast   = 0x1E
	LogoFirst        = 0x101 // SYM_LOGO_START: INAV logo (4x10)
	LogoLast         = LogoFirst + 40 - 1
)

// IsHalfDot reports whether the glyph is a digit with half a decimal point on
// its right ("3.") or on its left (".6"): INAV writes 3.69 as [3.][.6]9.
func IsHalfDot(c uint16) bool {
	return c >= TrailingDotFirst && c <= TrailingDotLast || c >= LeadingDotFirst && c <= LeadingDotLast
}

// IsLogo reports whether the glyph is a piece of the INAV logo, which the OSD
// draws on the splash screen shown while arming.
func IsLogo(c uint16) bool { return c >= LogoFirst && c <= LogoLast }

// IsGraphic reports whether the glyph belongs to a graphic element (artificial
// horizon, crosshair, variometer, HUD arrows), which carries no value.
func IsGraphic(c uint16) bool {
	switch {
	case c >= 0x12c && c <= 0x133, // horizon arrows and scale
		c >= 0x14c && c <= 0x164, // horizontal/vertical horizon, variometer, HUD
		c == 0x166,               // crosshair center
		c >= 0x190 && c <= 0x1a6, // crosshair types
		c >= 0x1ae && c <= 0x1b9: // HUD arrows
		return true
	}
	return false
}

// IsGlyph reports whether a cell holds an icon or a unit of the font, as
// opposed to a blank, a digit, a letter or punctuation, a digit with half a
// decimal point, a graphic or a piece of the logo.
//
// Besides the codes below the space and above ASCII, INAV puts glyphs on the
// lowercase codes (its text is all uppercase), such as the 'v' that stands for
// metres. The glyphs that sit on other ASCII codes (the quote of Ah/km, for
// one) cannot be told from punctuation; the rules that know them look for them
// by code.
func IsGlyph(c uint16) bool {
	switch {
	case c == 0 || c == ' ':
		return false
	case c < ' ':
		return true
	case c < 0x60:
		return false // digits, uppercase letters and punctuation
	case c > MaxGlyph:
		return false
	}
	return !IsHalfDot(c) && !IsGraphic(c) && !IsLogo(c)
}
