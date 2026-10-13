package telemetry

import "github.com/ericogr/fpv-osd-decoder/internal/inav"

// sensorRules read the temperatures and the motor speed.
func sensorRules() []rule {
	unit := `(` + syms(inav.SymTempC, inav.SymTempF) + `)`
	iconTemperature := func(key string, icon string) rule {
		return newRule(rx(icon, spaceRE, numRE, unit), temperature(key, 1, 2))
	}
	return []rule{
		iconTemperature(keyIMUTemp, sym(inav.SymImuTemp)),
		iconTemperature(keyBaroTemp, sym(inav.SymBaroTemp)),
		iconTemperature(keyESCTemp, sym(inav.SymEscTemp)),
		// The sensors that use one of the other icons. Those with a text label
		// instead of an icon have neither, and are read by the next rule.
		newRule(rx(`(`+symRange(inav.SymTempSensorFirst, inav.SymTempSensorLast)+`)`, spaceRE, numRE, unit),
			temperature(keySensorTemp, 2, 3)),
		newRule(rx(numRE, unit), temperature(keySensorTemp, 1, 2)),
		// "12.3K" is 12300 rpm when the number does not fit
		newRule(rx(sym(inav.SymRpm), spaceRE, numRE, `(K)?`), motorSpeed),
	}
}

// temperature emits the number in group ni, converted to Celsius if the unit
// glyph in group ui says Fahrenheit.
func temperature(key string, ni, ui int) emitFunc {
	return func(g []string, s Sample) {
		v, ok := parseNum(g[ni])
		if !ok {
			return
		}
		if firstGlyph(g[ui]) == inav.SymTempF {
			v = (v - 32) * 5 / 9
		}
		s.num(key, v)
	}
}

// motorSpeed reads the rpm in group 1; the K in group 2 means thousands.
func motorSpeed(g []string, s Sample) {
	v, ok := parseNum(g[1])
	if !ok {
		return
	}
	if g[2] != "" {
		v *= 1000
	}
	s.num(keyESCRPM, v)
}
