package telemetry

import (
	"strings"

	"github.com/ericogr/fpv-osd-decoder/internal/inav"
)

const (
	// maxCellVolts: no cell goes above ~4.4 V, so a larger battery reading is
	// the whole pack.
	maxCellVolts = 4.5
	// ahToMah converts the Ah the OSD shows for large values back to mAh.
	ahToMah = 1000
)

// batteryRules read the voltage, charge, current, capacity, energy, power and
// efficiency of the main battery.
func batteryRules() []rule {
	return []rule{
		newRule(rx(symRange(inav.SymBattFull, inav.SymBattEmpty), spaceRE, numRE, sym(inav.SymVolt)), batteryVoltage),
		// the remaining charge: the battery icon and "nnn%"
		newRule(rx(symRange(inav.SymBattFull, inav.SymBattEmpty), spaceRE, `([0-9]+)%`), intField(keyBattPct, 1)),
		newRule(rx(numRE, sym(inav.SymAmp)), numField(keyCurrent, 1)),
		newRule(rx(numRE, `(`+syms(inav.SymMah, inav.SymAh)+`)`), capacityUsed),
		newRule(rx(numRE, `(`+efficiencyUnitsRE+`)`), efficiencyPerKm),
		newRule(rx(numRE, sym(inav.SymWh)), numField(keyWhUsed, 1)),
		scaledReading(keyWhPerKm, energyPerDistToKm),
		scaledReading(keyPower, powerToW),
		newRule(rx(numRE, sym(inav.SymMilliohm)), numField(keyImpedance, 1)),
	}
}

// batteryVoltage reads the voltage in group 1: a reading above what a cell can
// give is the whole pack.
func batteryVoltage(g []string, s Sample) {
	v, ok := parseNum(g[1])
	if !ok {
		return
	}
	if v > maxCellVolts {
		s.num(keyBattV, v)
	} else {
		s.num(keyBattCellV, v)
	}
}

// capacityUsed reads the capacity in group 1; the OSD switches from mAh to Ah
// (the glyph in group 2) on large values.
func capacityUsed(g []string, s Sample) {
	v, ok := parseNum(g[1])
	if !ok {
		return
	}
	if firstGlyph(g[2]) == inav.SymAh {
		v *= ahToMah
	}
	s.num(keyMAhUsed, v)
}

// efficiencyUnitsRE matches the glyphs that follow an efficiency in capacity
// per distance: two glyphs for mAh, and one for Ah per km. The Ah per mile and
// per nautical mile glyphs sit on the '$' and '?' codes, which text uses too (a
// craft name such as "X8 2$"), so those two are not read.
var efficiencyUnitsRE = strings.Join([]string{
	sym(inav.SymMahKm0) + sym(inav.SymMahKm1),
	sym(inav.SymMahMi0) + sym(inav.SymMahMi1),
	sym(inav.SymMahNm0) + sym(inav.SymMahNm1),
	sym(inav.SymAhKm),
}, "|")

// efficiencyPerKm emits the efficiency in group 1 as mAh per km, whatever the
// units the OSD shows it in (group 2).
func efficiencyPerKm(g []string, s Sample) {
	v, ok := parseNum(g[1])
	if !ok {
		return
	}
	switch firstGlyph(g[2]) {
	case inav.SymMahMi0:
		v /= kmPerMile
	case inav.SymMahNm0:
		v /= kmPerNauticalMile
	case inav.SymAhKm:
		v *= ahToMah
	}
	s.num(keyMAhPerKm, v)
}
