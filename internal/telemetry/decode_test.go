package telemetry

import (
	"math"
	"testing"

	"github.com/ericogr/fpv-osd-decoder/internal/inav"
	"github.com/ericogr/fpv-osd-decoder/internal/osd"
)

const cols, rows = 53, 20

// td/ld are the digits with a half point on the right ("3.") and on the left (".6").
func td(n int) uint16 { return 0xA1 + uint16(n) }
func ld(n int) uint16 { return 0xB1 + uint16(n) }

func newFile() *osd.File {
	cells := make([]uint16, cols*rows)
	for i := range cells {
		cells[i] = ' '
	}
	return &osd.File{Cols: cols, Rows: rows, Frames: []osd.Frame{{Cells: cells}}}
}

// put writes into the grid starting at (x, y): strings become ASCII, uint16
// values become the glyph with the same code.
func put(f *osd.File, x, y int, parts ...any) {
	cells := f.Frames[0].Cells
	for _, p := range parts {
		switch v := p.(type) {
		case string:
			for _, r := range v {
				cells[y*cols+x] = uint16(r)
				x++
			}
		case uint16:
			cells[y*cols+x] = v
			x++
		default:
			panic("unexpected type")
		}
	}
}

func decode(t *testing.T, f *osd.File) Sample {
	t.Helper()
	s, _ := New(DefaultOptions()).Decode(f, &f.Frames[0])
	known := map[string]bool{}
	for _, fld := range Fields() {
		known[fld.Key] = true
	}
	for k := range s {
		if !known[k] {
			t.Errorf("key %q is not in the catalog", k)
		}
	}
	return s
}

func checkNum(t *testing.T, s Sample, want map[string]float64) {
	t.Helper()
	for k, w := range want {
		v, ok := s[k]
		if !ok {
			t.Errorf("%s missing", k)
			continue
		}
		if math.Abs(v.Num-w) > 1e-9 {
			t.Errorf("%s = %v, want %v", k, v.Num, w)
		}
	}
}

// frame3000 reproduces, glyph by glyph, frame 3000 of AvatarG0031.osd.
func frame3000() *osd.File {
	f := newFile()
	put(f, 3, 1, inav.SymRssi, " -72", inav.SymDbm)
	put(f, 3, 2, inav.SymLq, "2:100")
	put(f, 3, 3, inav.SymFlyM, "07:23")
	put(f, 3, 4, inav.SymBatt4, td(3), ld(6), "9", inav.SymVolt)
	put(f, 3, 5, " 969", inav.SymMah) // mAh without an icon
	put(f, 3, 6, "208", inav.SymMahKm0, inav.SymMahKm1)
	put(f, 3, 7, inav.SymHeading, "254", inav.SymDegrees)
	put(f, 3, 8, inav.SymRollRight, " ", td(3), ld(1))
	put(f, 3, 9, inav.SymPitchDown, " ", td(0), ld(9))
	put(f, 3, 10, inav.SymGforce, td(1), ld(0), "3")
	put(f, 3, 15, td(6), ld(1), "9", inav.SymAmp)
	put(f, 3, 17, inav.SymLat, "-2", td(3), ld(3), "11583", " ", inav.SymLon, "-4", td(7), ld(3), "78889", " M0")

	put(f, 45, 4, "213", inav.SymAltM)
	put(f, 44, 5, inav.SymHome, td(1), ld(4), "1", inav.SymDistKm)
	put(f, 44, 6, inav.SymTripDist, td(3), ld(9), "3", inav.SymDistKm)
	put(f, 45, 7, "29", inav.SymKmh)
	put(f, 43, 8, inav.SymWindHorizontal, inav.SymDecoration+4, td(7), ld(2), "0", inav.SymKmh)
	put(f, 44, 9, inav.SymAir, " 36", inav.SymKmh)
	put(f, 44, 17, inav.SymAutoThr0, inav.SymAutoThr1, " 50")
	put(f, 27, 1, inav.SymArrow7)
	put(f, 45, 15, inav.SymSatL, inav.SymSatR, "18")
	put(f, 45, 16, "CRUZ")
	put(f, 18, 13, "EN ROUTE TO", inav.SymAhHStart, "HOME")
	return f
}

func TestDecodeFlightFrame(t *testing.T) {
	s := decode(t, frame3000())
	checkNum(t, s, map[string]float64{
		"osd_rssi_dbm":       -72,
		"osd_rssi_antenna":   1,
		"osd_rf_mode":        2,
		"osd_lq_pct":         100,
		"osd_flight_time_s":  443,
		"osd_batt_cell_v":    3.69,
		"osd_mah_used":       969,
		"osd_mah_per_km":     208,
		"osd_heading_deg":    254,
		"osd_roll_deg":       3.1,
		"osd_pitch_deg":      -0.9,
		"osd_gforce":         1.03,
		"osd_current_a":      6.19,
		"osd_gps_lat":        -23.311583,
		"osd_gps_lon":        -47.378889,
		"osd_altitude_m":     213,
		"osd_home_dist_m":    1410,
		"osd_trip_dist_m":    3930,
		"osd_gps_speed_kmh":  29,
		"osd_wind_dir_deg":   180,
		"osd_wind_speed_kmh": 7.2,
		"osd_air_speed_kmh":  36,
		"osd_throttle_pct":   50,
		"osd_throttle_auto":  1,
		"osd_home_dir_deg":   135,
		"osd_gps_sats":       18,
	})
	if got := s["osd_flight_mode"].Text; got != "CRUZ" {
		t.Errorf("mode = %q, want CRUZ", got)
	}
	if got := s["osd_message"].Text; got != "EN ROUTE TO HOME" {
		t.Errorf("message = %q, want %q", got, "EN ROUTE TO HOME")
	}
	if _, ok := s["osd_batt_v"]; ok {
		t.Error("osd_batt_v should not exist: 3.69 V is a cell voltage")
	}
}

// Wind/air/ground speeds share the same unit; each reading must go to the
// right column and the speed without an icon must not steal the other two.
func TestSpeedsDoNotCollide(t *testing.T) {
	f := newFile()
	put(f, 44, 9, inav.SymAir, " 36", inav.SymKmh)
	s := decode(t, f)
	if _, ok := s["osd_gps_speed_kmh"]; ok {
		t.Error("airspeed was also read as GPS speed")
	}
	checkNum(t, s, map[string]float64{"osd_air_speed_kmh": 36})
}

func TestUnitConversions(t *testing.T) {
	tests := []struct {
		name  string
		parts []any
		key   string
		want  float64
	}{
		{"home in meters", []any{inav.SymHome, " 850", inav.SymDistM}, "osd_home_dist_m", 850},
		{"home in feet", []any{inav.SymHome, "100", inav.SymDistFt}, "osd_home_dist_m", 30.48},
		{"home in miles", []any{inav.SymHome, td(1), ld(0), inav.SymDistMi}, "osd_home_dist_m", 1609.344},
		{"altitude in feet", []any{"100", inav.SymAltFt}, "osd_altitude_m", 30.48},
		{"negative altitude", []any{"-1", inav.SymAltM}, "osd_altitude_m", -1},
		{"speed in mph", []any{"10", inav.SymMph}, "osd_gps_speed_kmh", 16.09344},
		{"speed in knots", []any{"10", inav.SymKt}, "osd_gps_speed_kmh", 18.52},
		{"consumption in Ah", []any{td(1), ld(2), "3", inav.SymAh}, "osd_mah_used", 1230},
		{"mAh/km as Ah/km", []any{td(1), ld(5), inav.SymAhKm}, "osd_mah_per_km", 1500},
		{"flight time in hours", []any{inav.SymFlyH, "01:40"}, "osd_flight_time_s", 6000},
		{"second antenna", []any{inav.Sym2rss, "-80", inav.SymDbm}, "osd_rssi_antenna", 2},
		{"pack voltage", []any{inav.SymBatt5, "1", td(6), ld(0), inav.SymVolt}, "osd_batt_v", 16.0},
		{"level roll", []any{inav.SymRollLevel, td(0), ld(0)}, "osd_roll_deg", 0},
		{"roll to the left", []any{inav.SymRollLeft, td(4), ld(2)}, "osd_roll_deg", -4.2},
		{"pitch up", []any{inav.SymPitchUp, td(2), ld(5)}, "osd_pitch_deg", 2.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFile()
			put(f, 10, 5, tt.parts...)
			checkNum(t, decode(t, f), map[string]float64{tt.key: tt.want})
		})
	}
}

// Blinking or invalid elements are absent, never zero.
func TestMissingIsNotZero(t *testing.T) {
	f := newFile()
	put(f, 3, 7, inav.SymHeading, "---", inav.SymDegrees)
	put(f, 45, 15, inav.SymSatL, inav.SymSatR, inav.SymAlert)
	put(f, 3, 6, "---", inav.SymMahKm0, inav.SymMahKm1)
	s := decode(t, f)
	if len(s) != 0 {
		t.Errorf("expected no readings, got %v", s)
	}
}

func TestLeftoversIgnoreGraphics(t *testing.T) {
	f := newFile()
	put(f, 21, 12, inav.SymAhHStart, uint16(0x14d), uint16(0x14e)) // horizon
	put(f, 25, 11, inav.SymAhChType7, inav.SymAhChType7+1)         // crosshair
	put(f, 25, 17, "M0")
	_, left := New(DefaultOptions()).Decode(f, &f.Frames[0])
	if len(left) != 1 || left[0] != (Leftover{X: 25, Y: 17, Text: "M0"}) {
		t.Errorf("leftovers = %+v, want only M0", left)
	}
}

// A flight mode region outside the grid must be ignored, not crash.
func TestModeRegionOutsideGrid(t *testing.T) {
	f := newFile()
	for _, region := range []Region{
		{X: cols, Y: 16, Width: 4},
		{X: cols + 10, Y: 16, Width: 4},
		{X: -1, Y: 16, Width: 4},
	} {
		opts := Options{FlightMode: region, MessageRow: -1}
		s, _ := New(opts).Decode(f, &f.Frames[0])
		if _, ok := s[keyFlightMode]; ok {
			t.Errorf("%+v: read a flight mode outside the grid", region)
		}
	}
}

// Every catalog key is unique.
func TestCatalogKeysAreUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, f := range Fields() {
		if seen[f.Key] {
			t.Errorf("duplicate catalog key %q", f.Key)
		}
		seen[f.Key] = true
	}
}

// Fields returns a copy: changing it must not change the catalog.
func TestFieldsReturnsACopy(t *testing.T) {
	a := Fields()
	a[0].Key = "changed"
	if b := Fields(); b[0].Key == "changed" {
		t.Error("Fields exposes the internal catalog")
	}
}
