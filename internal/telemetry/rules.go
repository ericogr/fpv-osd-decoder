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

const (
	// maxCellVolts: no cell goes above ~4.4 V, so a larger battery reading is
	// the whole pack.
	maxCellVolts = 4.5
	// arrowStepDeg is the angle between two home-direction arrow glyphs (16 of
	// them); windStepDeg between two wind-direction arrow glyphs (8 of them).
	arrowStepDeg = 22.5
	windStepDeg  = 45
	// ahToMah converts the Ah the OSD shows for large values back to mAh.
	ahToMah = 1000
)

// emitFunc stores what a matched rule read into the sample. g holds the
// matched text (g[0]) and the capture groups of the rule's expression.
type emitFunc func(g []string, s Sample)

// rule recognizes one field on a row: a regular expression over the glyph
// codes and the function that turns a match into sample values.
type rule struct {
	re   *regexp.Regexp
	emit emitFunc
}

// apply runs the rule on line. On a match it emits the values and returns the
// line with the matched span erased; otherwise the line comes back unchanged.
func (r rule) apply(line string, s Sample) string {
	loc := r.re.FindStringSubmatchIndex(line)
	if loc == nil {
		return line
	}
	g := make([]string, len(loc)/2)
	for i := range g {
		if loc[2*i] >= 0 {
			g[i] = line[loc[2*i]:loc[2*i+1]]
		}
	}
	r.emit(g, s)
	blanks := strings.Repeat(" ", utf8.RuneCountInString(line[loc[0]:loc[1]]))
	return line[:loc[0]] + blanks + line[loc[1]:]
}

// Regular expression fragments. Each OSD cell becomes a rune holding the
// glyph code, so \x{NN} matches glyph NN.
const (
	// INAV number: ASCII digits, sign, point and the digits with a half decimal
	// point (0xA1-0xAA = "n.", 0xB1-0xBA = ".n").
	numRE   = `(-?[0-9.\x{a1}-\x{aa}\x{b1}-\x{ba}]+)`
	spaceRE = ` *`
)

func sym(c uint16) string { return fmt.Sprintf(`\x{%x}`, c) }

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

func unitsRE(m map[uint16]float64) string {
	cs := make([]uint16, 0, len(m))
	for c := range m {
		cs = append(cs, c)
	}
	sort.Slice(cs, func(i, j int) bool { return cs[i] < cs[j] })
	return "(" + syms(cs...) + ")"
}

func rx(parts ...string) *regexp.Regexp { return regexp.MustCompile(strings.Join(parts, "")) }

// Conversion factors from the unit glyph to the column's unit.
var (
	distToMeters = map[uint16]float64{
		inav.SymDistM: 1, inav.SymDistKm: 1000, inav.SymDistFt: 0.3048,
		inav.SymDistMi: 1609.344, inav.SymDistNm: 1852,
	}
	altToMeters = map[uint16]float64{
		inav.SymAltM: 1, inav.SymAltKm: 1000, inav.SymAltFt: 0.3048, inav.SymAltKft: 304.8,
	}
	// Horizontal speed units only: m/s and ft/s belong to the vertical speed
	// indicator.
	speedToKmh = map[uint16]float64{
		inav.SymKmh: 1, inav.SymMph: 1.609344, inav.SymKt: 1.852,
	}
)

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

// thousandsIf emits the number in group 1, times ahToMah when the unit glyph
// in group 2 is the given one (the OSD switches from mAh to Ah on large values).
func thousandsIf(key string, ahUnit uint16) emitFunc {
	return func(g []string, s Sample) {
		v, ok := parseNum(g[1])
		if !ok {
			return
		}
		if firstGlyph(g[2]) == ahUnit {
			v *= ahToMah
		}
		s.num(key, v)
	}
}

// buildRules returns every rule, grouped by subject. The order matters: a
// recognized span is erased before the next rule runs, so the more specific
// rules must come before the generic ones they could collide with.
func buildRules() []rule {
	var rules []rule
	for _, group := range [][]rule{
		linkRules(),
		flightTimeRules(),
		batteryRules(),
		throttleRules(),
		attitudeRules(),
		gpsRules(),
		distanceRules(),
		speedRules(),
		homeArrowRules(),
	} {
		rules = append(rules, group...)
	}
	return rules
}

func linkRules() []rule {
	return []rule{
		{rx(`(`+syms(inav.SymRssi, inav.Sym2rss)+`)`, spaceRE, numRE, sym(inav.SymDbm)), func(g []string, s Sample) {
			v, ok := parseNum(g[2])
			if !ok {
				return
			}
			s.num(keyRSSIdBm, v)
			antenna := 1.0
			if firstGlyph(g[1]) == inav.Sym2rss {
				antenna = 2
			}
			s.num(keyRSSIAntenna, antenna)
		}},
		// "n:LQ" (RF mode and quality) or just the quality
		{rx(sym(inav.SymLq), spaceRE, `(?:([0-9]+):)?`, spaceRE, `([0-9]+)`),
			each(intField(keyRFMode, 1), intField(keyLQ, 2))},
	}
}

// flightTimeRules reads MM:SS, or HH:MM once the flight passes 99:59.
func flightTimeRules() []rule {
	return []rule{
		{rx(`(`+syms(inav.SymFlyM, inav.SymFlyH)+`)`, spaceRE, `([0-9]+):([0-9]+)`), func(g []string, s Sample) {
			a, errA := strconv.Atoi(g[2])
			b, errB := strconv.Atoi(g[3])
			if errA != nil || errB != nil {
				return
			}
			seconds := a*60 + b
			if firstGlyph(g[1]) == inav.SymFlyH {
				seconds = a*3600 + b*60
			}
			s.num(keyFlightTime, float64(seconds))
		}},
	}
}

func batteryRules() []rule {
	return []rule{
		{rx(symRange(inav.SymBattFull, inav.SymBattEmpty), spaceRE, numRE, sym(inav.SymVolt)), func(g []string, s Sample) {
			v, ok := parseNum(g[1])
			if !ok {
				return
			}
			if v > maxCellVolts {
				s.num(keyBattV, v)
			} else {
				s.num(keyBattCellV, v)
			}
		}},
		{rx(numRE, sym(inav.SymAmp)), numField(keyCurrent, 1)},
		{rx(numRE, `(`+syms(inav.SymMah, inav.SymAh)+`)`), thousandsIf(keyMAhUsed, inav.SymAh)},
		{rx(numRE, `(`+sym(inav.SymMahKm0)+sym(inav.SymMahKm1)+`|`+sym(inav.SymAhKm)+`)`),
			thousandsIf(keyMAhPerKm, inav.SymAhKm)},
	}
}

// throttleRules: auto-throttle uses two glyphs, manual uses THR.
func throttleRules() []rule {
	return []rule{
		{rx(`(`+sym(inav.SymAutoThr0)+sym(inav.SymAutoThr1)+`|`+sym(inav.SymThr)+`)`, spaceRE, `([0-9]+)`), func(g []string, s Sample) {
			intField(keyThrottle, 2)(g, s)
			auto := 0.0
			if firstGlyph(g[1]) == inav.SymAutoThr0 {
				auto = 1
			}
			s.num(keyThrottleAuto, auto)
		}},
	}
}

func attitudeRules() []rule {
	return []rule{
		{rx(sym(inav.SymHeading), spaceRE, `([0-9]+)`, sym(inav.SymDegrees)), intField(keyHeading, 1)},
		{rx(`(`+syms(inav.SymRollLeft, inav.SymRollLevel, inav.SymRollRight)+`)`, spaceRE, numRE),
			signedByIcon(keyRoll, inav.SymRollLeft)},
		// INAV shows 'P' in place of the glyph when |pitch| < 0.1°. The glyph
		// shows what the pilot sees: down arrow = nose down.
		{rx(`(`+syms(inav.SymPitchUp, inav.SymPitchDown, 'P')+`)`, spaceRE, numRE),
			signedByIcon(keyPitch, inav.SymPitchDown)},
		{rx(sym(inav.SymGforce), spaceRE, numRE), numField(keyGForce, 1)},
	}
}

func gpsRules() []rule {
	return []rule{
		{rx(sym(inav.SymSatL), sym(inav.SymSatR), spaceRE, `([0-9]+)`), intField(keySats, 1)},
		{rx(sym(inav.SymLat), numRE), numField(keyLat, 1)},
		{rx(sym(inav.SymLon), numRE), numField(keyLon, 1)},
	}
}

// distanceRules read icon + number + a unit glyph that changes with the
// magnitude (m/km, ft/mi...).
func distanceRules() []rule {
	return []rule{
		{rx(sym(inav.SymHome), spaceRE, numRE, unitsRE(distToMeters)), scaled(keyHomeDist, 1, 2, distToMeters)},
		{rx(sym(inav.SymTripDist), spaceRE, numRE, unitsRE(distToMeters)), scaled(keyTripDist, 1, 2, distToMeters)},
		{rx(numRE, unitsRE(altToMeters)), scaled(keyAltitude, 1, 2, altToMeters)},
	}
}

// speedRules go from the most specific to the generic one: the speed without
// an icon is the GPS ground speed, so it must run after wind and airspeed
// have erased theirs.
func speedRules() []rule {
	return []rule{
		{rx(sym(inav.SymWindHorizontal), `(`+symRange(inav.SymDecoration, inav.SymDecoration+7)+`)`, spaceRE, numRE, unitsRE(speedToKmh)),
			each(windDirection, scaled(keyWindSpeed, 2, 3, speedToKmh))},
		{rx(sym(inav.SymAir), spaceRE, numRE, unitsRE(speedToKmh)), scaled(keyAirSpeed, 1, 2, speedToKmh)},
		{rx(numRE, unitsRE(speedToKmh)), scaled(keyGPSSpeed, 1, 2, speedToKmh)},
	}
}

// windDirection reads the arrow glyph in group 1 (8 glyphs, 45° apart).
func windDirection(g []string, s Sample) {
	s.num(keyWindDir, float64(firstGlyph(g[1])-inav.SymDecoration)*windStepDeg)
}

// homeArrowRules read the direction-of-home arrow: 16 glyphs, 22.5° apart.
func homeArrowRules() []rule {
	return []rule{
		{rx(`(` + symRange(inav.SymArrowUp, inav.SymArrow16) + `)`), func(g []string, s Sample) {
			s.num(keyHomeDir, float64(firstGlyph(g[1])-inav.SymArrowUp)*arrowStepDeg)
		}},
	}
}
