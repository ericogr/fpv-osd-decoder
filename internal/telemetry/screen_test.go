package telemetry

import (
	"testing"

	"github.com/ericogr/fpv-osd-decoder/internal/inav"
	"github.com/ericogr/fpv-osd-decoder/internal/osd"
)

func TestScreenKinds(t *testing.T) {
	splash := newFile()
	put(splash, 21, 1, inav.SymLogoStart, inav.SymLogoStart+1)
	stats := newFile()
	put(stats, 16, 1, "*** STATS ", inav.SymBlackbox, "00000 ***")
	statsPage := newFile()
	put(statsPage, 16, 1, "*** STATS   1/2 -> ***")
	notStats := newFile()
	put(notStats, 16, 1, "*** STAT ", "STATS")

	for _, tt := range []struct {
		name string
		f    *osd.File
		want screen
	}{
		{"flight", frame3000(), flightScreen},
		{"empty", newFile(), flightScreen},
		{"splash", splash, splashScreen},
		{"stats", stats, statsScreen},
		{"stats on two pages", statsPage, statsScreen},
		{"a title that only looks like it", notStats, flightScreen},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := screenOf(tt.f, &tt.f.Frames[0]); got != tt.want {
				t.Errorf("screen = %d, want %d", got, tt.want)
			}
		})
	}
}

// The statistics are summaries of the flight, not the state of the aircraft:
// nothing is read from them, but they still show up among the leftovers.
func TestStatsScreenIsNotRead(t *testing.T) {
	f := newFile()
	put(f, 16, 1, "*** STATS ", inav.SymBlackbox, "00000 ***")
	put(f, 6, 3, "MAX ALTITUDE", " ", "213", inav.SymAltM)
	put(f, 6, 4, "MAX SPEED", " ", "90", inav.SymKmh)
	put(f, 6, 5, inav.SymHome, " 850", inav.SymDistM)
	put(f, 6, 13, "SOMETHING")
	s, left := New(sampleLayout()).Decode(f, &f.Frames[0])
	if len(s) != 0 {
		t.Errorf("read %v from the statistics screen", s)
	}
	if len(left) == 0 {
		t.Error("the statistics screen should still be listed as leftovers")
	}
}

// The splash screen shown while arming does carry the position, so the rules
// read it; the generic pass stays away, since all it would find is totals.
func TestSplashScreenIsReadByTheRulesOnly(t *testing.T) {
	f := newFile()
	put(f, 21, 1, inav.SymLogoStart, inav.SymLogoStart+1)
	put(f, 16, 6, inav.SymLat, "-2", td(3), ld(2), "99915", " ", inav.SymLon, "-4", td(7), ld(3), "73494")
	put(f, 15, 14, "ODOMETER:", " 0", inav.SymKm)
	s, _ := New(sampleLayout()).Decode(f, &f.Frames[0])
	checkNum(t, s, map[string]float64{"osd_gps_lat": -23.299915, "osd_gps_lon": -47.373494})
	if len(s) != 2 {
		t.Errorf("read %v, want only the position", s)
	}
}

// There is no flight mode or status message on the splash screen: what it shows
// where the pilot's layout has them is something else.
func TestSplashScreenHasNoFlightModeOrMessage(t *testing.T) {
	f := newFile()
	put(f, 21, 1, inav.SymLogoStart, inav.SymLogoStart+1)
	put(f, 17, 13, "INAV VERSION: 9.1.0") // the row of the messages
	put(f, 45, 16, "ACRO")                // the flight mode
	s, left := New(sampleLayout()).Decode(f, &f.Frames[0])
	if len(s) != 0 {
		t.Errorf("read %v, want nothing", s)
	}
	texts := map[string]bool{}
	for _, l := range left {
		texts[l.Text] = true
	}
	if !texts["ACRO"] || !texts["INAV"] || !texts["VERSION:"] {
		t.Errorf("leftovers = %+v, want both texts left on the screen", left)
	}
}
