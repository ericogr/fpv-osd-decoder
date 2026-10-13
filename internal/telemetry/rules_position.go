package telemetry

import "github.com/ericogr/fpv-osd-decoder/internal/inav"

// positionRules read where the aircraft is: the GPS fix, the altitude and the
// distances. Most distances are an icon, a number and a unit glyph that changes
// with the magnitude (m and km, ft and mi).
func positionRules() []rule {
	return []rule{
		newRule(rx(sym(inav.SymSatL), sym(inav.SymSatR), spaceRE, `([0-9]+)`), intField(keySats, 1)),
		newRule(rx(sym(inav.SymLat), numRE), numField(keyLat, 1)),
		newRule(rx(sym(inav.SymLon), numRE), numField(keyLon, 1)),
		newRule(rx(sym(inav.SymHdpL), sym(inav.SymHdpR), spaceRE, numRE), numField(keyHDOP, 1)),

		scaledReading(keyRTHDistLeft, distToMeters, sym(inav.SymFlightDistRemaining)),
		scaledReading(keyGlideRange, distToMeters, sym(inav.SymGlideDist)),
		scaledReading(keyCrossTrack, distToMeters, sym(inav.SymCrossTrackError)),
		scaledReading(keyOdometer, odometerToKm, sym(inav.SymOdometer)),
		// The geofence: the horizontal distance is "FD" and a distance, the
		// vertical one is "FD" and an altitude followed by an arrow, down when
		// the aircraft is below the zone. The vertical one goes before the
		// altitude, which would take it.
		scaledReading(keyFenceDistH, distToMeters, `\bFD`),
		newRule(rx(`\bFD`, spaceRE, numRE, unitsRE(altToMeters), `(?: *(`+symRange(inav.SymDecoration, inav.SymDecoration+7)+`))?`), fenceDistV),

		scaledReading(keyHomeDist, distToMeters, sym(inav.SymHome)),
		scaledReading(keyTripDist, distToMeters, sym(inav.SymTripDist)),
		scaledReading(keyAltitude, altToMeters),
		// a distance with no icon: the rangefinder; every distance with an icon
		// has been taken by now
		scaledReading(keyRangefinder, distToMeters),
	}
}

// fenceDistV reads the altitude in groups 1 (number) and 2 (unit) and the
// optional arrow in group 3, which says the aircraft is below the zone when it
// points down.
func fenceDistV(g []string, s Sample) {
	v, ok := parseNum(g[1])
	if !ok {
		return
	}
	v *= altToMeters[firstGlyph(g[2])]
	if g[3] != "" && firstGlyph(g[3]) == inav.SymDecoration+4 {
		v = -v
	}
	s.num(keyFenceDistV, v)
}
