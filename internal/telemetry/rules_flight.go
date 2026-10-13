package telemetry

import "github.com/ericogr/fpv-osd-decoder/internal/inav"

// arrowStepDeg is the angle between two home-direction arrow glyphs (16 of
// them).
const arrowStepDeg = 22.5

// gforceAxisKeys is in the order of the glyphs of the three axes.
var gforceAxisKeys = [...]string{keyGForceX, keyGForceY, keyGForceZ}

// flightRules read how the aircraft is flying: the throttle, the heading and
// the other angles, the attitude, the G-force and the direction of home.
func flightRules() []rule {
	return []rule{
		// auto-throttle uses two glyphs, manual uses THR
		newRule(rx(`(`+sym(inav.SymAutoThr0)+sym(inav.SymAutoThr1)+`|`+sym(inav.SymThr)+`)`, spaceRE, `([0-9]+)`), throttle),
		// The angles that end with the degrees glyph. The home heading error
		// goes before the heading, whose rule would take it.
		newRule(rx(sym(inav.SymHome), sym(inav.SymHeading), spaceRE, `(-?[0-9]+)`, sym(inav.SymDegrees)), intField(keyHomeHeadingError, 1)),
		newRule(rx(sym(inav.SymGroundCourse), spaceRE, `([0-9]+)`, sym(inav.SymDegrees)), intField(keyGroundCourse, 1)),
		newRule(rx(sym(inav.SymAzimuth), spaceRE, `([0-9]+)`, sym(inav.SymDegrees)), intField(keyAzimuth, 1)),
		newRule(rx(sym(inav.SymHeading), spaceRE, `([0-9]+)`, sym(inav.SymDegrees)), intField(keyHeading, 1)),
		newRule(rx(`(`+syms(inav.SymRollLeft, inav.SymRollLevel, inav.SymRollRight)+`)`, spaceRE, numRE),
			signedByIcon(keyRoll, inav.SymRollLeft)),
		// INAV shows 'P' in place of the glyph when |pitch| < 0.1°. That 'P' does
		// not end a word, or the "P 3" of the text "TO WP 3/5" would be a pitch.
		// The glyph shows what the pilot sees: down arrow = nose down.
		newRule(rx(`(\bP|`+syms(inav.SymPitchUp, inav.SymPitchDown)+`)`, spaceRE, numRE),
			signedByIcon(keyPitch, inav.SymPitchDown)),
		newRule(rx(sym(inav.SymGforce), spaceRE, numRE), numField(keyGForce, 1)),
		newRule(rx(`(`+symRange(inav.SymGforceX, inav.SymGforceZ)+`)`, spaceRE, numRE), gforceAxis),
		// the direction-of-home arrow: 16 glyphs, 22.5° apart
		newRule(rx(`(`+symRange(inav.SymArrowUp, inav.SymArrow16)+`)`), homeArrow),
	}
}

// throttle reads the throttle in group 2; the glyphs in group 1 say whether
// navigation is controlling it.
func throttle(g []string, s Sample) {
	intField(keyThrottle, 2)(g, s)
	auto := 0.0
	if firstGlyph(g[1]) == inav.SymAutoThr0 {
		auto = 1
	}
	s.num(keyThrottleAuto, auto)
}

// gforceAxis reads the G-force of the axis whose glyph is in group 1.
func gforceAxis(g []string, s Sample) {
	if v, ok := parseNum(g[2]); ok {
		s.num(gforceAxisKeys[firstGlyph(g[1])-inav.SymGforceX], v)
	}
}

func homeArrow(g []string, s Sample) {
	s.num(keyHomeDir, float64(firstGlyph(g[1])-inav.SymArrowUp)*arrowStepDeg)
}
