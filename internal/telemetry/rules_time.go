package telemetry

import (
	"strconv"

	"github.com/ericogr/fpv-osd-decoder/internal/inav"
)

// timeRules read the elements that show a time: the flight time, the time
// since power-on, the time left before the return to home and for gliding, and
// the real-time clock.
func timeRules() []rule {
	return []rule{
		timeRule(keyFlightTime, inav.SymFlyM, inav.SymFlyH),
		timeRule(keyOnTime, inav.SymOnM, inav.SymOnH),
		timeRule(keyRTHTimeLeft, inav.SymFlightMinsRemaining, inav.SymFlightHoursRemaining),
		timeRule(keyGlideTime, inav.SymGlideMins, 0),
		newRule(rx(sym(inav.SymClock), spaceRE, `([0-9]{2}:[0-9]{2}:[0-9]{2})`), textField(keyClock, 1)),
	}
}

// timeRule reads icon + MM:SS into key, in seconds. hours is the icon the OSD
// uses when the time passes 99:59 and shows HH:MM instead; 0 if the element
// never switches.
func timeRule(key string, minutes, hours uint16) rule {
	icons := []uint16{minutes}
	if hours != 0 {
		icons = append(icons, hours)
	}
	return newRule(rx(`(`+syms(icons...)+`)`, spaceRE, `([0-9]+):([0-9]+)`), func(g []string, s Sample) {
		a, errA := strconv.Atoi(g[2])
		b, errB := strconv.Atoi(g[3])
		if errA != nil || errB != nil {
			return
		}
		seconds := a*secondsPerMinute + b
		if hours != 0 && firstGlyph(g[1]) == hours {
			seconds = a*3600 + b*secondsPerMinute
		}
		s.num(key, float64(seconds))
	})
}
