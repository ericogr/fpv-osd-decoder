package telemetry

import "github.com/ericogr/fpv-osd-decoder/internal/inav"

// linkRules read the radio link (RSSI, link quality, power, receiver) and the
// video transmitter.
func linkRules() []rule {
	return []rule{
		newRule(rx(`(`+syms(inav.SymRssi, inav.Sym2rss)+`)`, spaceRE, numRE, sym(inav.SymDbm)), rssiDbm),
		// The downlink quality ends with a down arrow; it must go first or the
		// uplink rule below would take it.
		newRule(rx(sym(inav.SymLq), spaceRE, `([0-9]+)`, sym(inav.SymAhDecorationDown)), intField(keyLQDown, 1)),
		// "n:LQ" (RF mode and quality) or just the quality
		newRule(rx(sym(inav.SymLq), spaceRE, `(?:([0-9]+):)?`, spaceRE, `([0-9]+)`),
			each(intField(keyRFMode, 1), intField(keyLQ, 2))),
		// RSSI as a percentage. The dBm rule has already taken the elements
		// that show dBm, which start with the same icon.
		newRule(rx(sym(inav.SymRssi), spaceRE, `(?:([0-9]+)|(`+sym(inav.SymMax)+`))`), rssiPercent),
		newRule(rx(sym(inav.SymSnr), spaceRE, `(-?[0-9]+)`, sym(inav.SymDb)), intField(keySNR, 1)),
		// The downlink power ends with a down arrow and must go before the
		// uplink one.
		newRule(rx(numRE, sym(inav.SymMw), sym(inav.SymAhDecorationDown)), numField(keyRXPower, 1)),
		newRule(rx(numRE, sym(inav.SymMw)), numField(keyTXPower, 1)),
		newRule(rx(sym(inav.SymRxBand), `([ -~]{1,4})`), textField(keyRXBand, 1)),
		newRule(rx(sym(inav.SymRxMode), `([ -~]{1,6})`), textField(keyRXMode, 1)),
		// "CH:", the band letter, the channel number, ":" and the power letter.
		// It goes before the VTX power element, which shows the same letter: if
		// both are on screen the channel must not be lost.
		newRule(rx(`CH:([A-Z][0-9A-Z]*):([0-9A-Z]?)`), each(textField(keyVTXChannel, 1), textField(keyVTXPower, 2))),
		// The glyph sits on the apostrophe, so it only counts at the start of a
		// word: "DON'T" in a text is not a transmitter power.
		newRule(rx(`\B`, sym(inav.SymVtxPower), `([0-9A-Z])`), textField(keyVTXPower, 1)),
	}
}

// rssiDbm reads the RSSI in dBm (group 2). The icon (group 1) says which
// antenna is active.
func rssiDbm(g []string, s Sample) {
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
}

// rssiPercent reads the RSSI in percent: the number in group 1, or the MAX
// glyph in group 2, which stands for 100.
func rssiPercent(g []string, s Sample) {
	if g[2] != "" {
		s.num(keyRSSIPct, 100)
		return
	}
	intField(keyRSSIPct, 1)(g, s)
}
