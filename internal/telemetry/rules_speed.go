package telemetry

import "github.com/ericogr/fpv-osd-decoder/internal/inav"

// windStepDeg is the angle between two wind-direction arrow glyphs (8 of them).
const windStepDeg = 45

// speedRules go from the most specific to the generic one: the speed without
// an icon is the GPS ground speed, so it must run after every other speed has
// erased its own (they share the unit glyphs).
func speedRules() []rule {
	windArrow := symRange(inav.SymDecoration, inav.SymDecoration+7)
	windVerticalArrow := syms(inav.SymAhDecorationUp, inav.SymAhDecorationDown)
	return []rule{
		newRule(rx(sym(inav.SymWindHorizontal), `(`+windArrow+`)`, spaceRE, numRE, unitsRE(windToKmh)), windHorizontal),
		newRule(rx(sym(inav.SymWindVertical), `(`+windVerticalArrow+`)`, spaceRE, numRE, unitsRE(windToKmh)), windVertical),
		// A wind estimate that is not valid yet shows a * in place of the
		// unit every other second. It is not a reading, but it must not reach
		// the generic pass either.
		newRule(rx(sym(inav.SymWindHorizontal), windArrow, spaceRE, numRE, `\*`), noReading),
		newRule(rx(sym(inav.SymWindVertical), windVerticalArrow, spaceRE, numRE, `\*`), noReading),
		// The maximum speeds start with the MAX glyph and must go before the
		// plain ones, which would read them as the current speed.
		scaledReading(keyAirMaxSpeed, speedToKmh, sym(inav.SymMax), sym(inav.SymAir)),
		scaledReading(keyAirSpeed, speedToKmh, sym(inav.SymAir)),
		scaledReading(key3DMaxSpeed, speed3DToKmh, sym(inav.SymMax)),
		scaledReading(key3DSpeed, speed3DToKmh),
		scaledReading(keyMinGroundSpeed, speedToKmh, sym(inav.SymMinGroundSpeed)),
		scaledReading(keyGPSMaxSpeed, speedToKmh, sym(inav.SymMax)),
		scaledReading(keyVSpeed, vSpeedToMS),
		scaledReading(keyGPSSpeed, speedToKmh),
	}
}

// performanceRules read how efficiently the aircraft is flying.
func performanceRules() []rule {
	return []rule{
		newRule(rx(sym(inav.SymGlideslope), spaceRE, numRE), numField(keyGlideslope, 1)),
		newRule(rx(numRE, `(`+sym(inav.SymAhVM0)+sym(inav.SymAhVM1)+`|`+sym(inav.SymAhVFt0)+sym(inav.SymAhVFt1)+`)`), climbEfficiency),
	}
}

// windHorizontal reads the direction arrow in group 1 (8 glyphs, 45° apart)
// and the speed in groups 2 (number) and 3 (unit).
func windHorizontal(g []string, s Sample) {
	s.num(keyWindDir, float64(firstGlyph(g[1])-inav.SymDecoration)*windStepDeg)
	scaled(keyWindSpeed, 2, 3, windToKmh)(g, s)
}

// windVertical reads the arrow in group 1 (up is positive) and the speed in
// groups 2 (number) and 3 (unit).
func windVertical(g []string, s Sample) {
	v, ok := parseNum(g[2])
	if !ok {
		return
	}
	v *= windToKmh[firstGlyph(g[3])]
	if firstGlyph(g[1]) == inav.SymAhDecorationDown {
		v = -v
	}
	s.num(keyWindVSpeed, v)
}

// climbEfficiency reads the number in group 1 and converts the per-foot
// variant (group 2) to the per-metre one.
func climbEfficiency(g []string, s Sample) {
	v, ok := parseNum(g[1])
	if !ok {
		return
	}
	if firstGlyph(g[2]) == inav.SymAhVFt0 {
		v /= metersPerFoot
	}
	s.num(keyClimbEfficiency, v)
}
