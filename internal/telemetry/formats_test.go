package telemetry

import (
	"math"
	"testing"

	"github.com/ericogr/fpv-osd-decoder/internal/inav"
	"github.com/ericogr/fpv-osd-decoder/internal/osd"
)

// The frames below are built from the format strings in INAV 9.1.0's osd.c.

func TestElementFormats(t *testing.T) {
	tests := []struct {
		name  string
		parts []any
		want  map[string]float64
		// absent lists keys that must not be read: the rule that came before
		// or the one that comes after must not take this element
		absent []string
	}{
		// attitude
		{"level pitch is shown with a P", []any{"P", td(0), ld(0)}, map[string]float64{"osd_pitch_deg": 0}, nil},
		{"a P that ends a word is not the pitch icon", []any{"TO WP 3/5"}, map[string]float64{}, []string{"osd_pitch_deg"}},
		// link
		{"rssi in percent", []any{inav.SymRssi, " 87"}, map[string]float64{"osd_rssi_pct": 87}, []string{"osd_rssi_dbm"}},
		{"rssi at 100 shows the MAX glyph", []any{inav.SymRssi, inav.SymMax, " "}, map[string]float64{"osd_rssi_pct": 100}, nil},
		{"rssi in dBm is not a percentage", []any{inav.SymRssi, " -72", inav.SymDbm}, map[string]float64{"osd_rssi_dbm": -72}, []string{"osd_rssi_pct"}},
		{"downlink link quality", []any{inav.SymLq, " 98", inav.SymAhDecorationDown}, map[string]float64{"osd_lq_down_pct": 98}, []string{"osd_lq_pct"}},
		{"uplink link quality is not the downlink one", []any{inav.SymLq, "2:100"}, map[string]float64{"osd_lq_pct": 100, "osd_rf_mode": 2}, []string{"osd_lq_down_pct"}},
		{"snr", []any{inav.SymSnr, "-12", inav.SymDb}, map[string]float64{"osd_snr_db": -12}, nil},
		{"snr of one digit", []any{inav.SymSnr, " 7", inav.SymDb}, map[string]float64{"osd_snr_db": 7}, nil},
		{"uplink transmitter power", []any{" 250", inav.SymMw}, map[string]float64{"osd_tx_power_mw": 250}, []string{"osd_rx_power_mw"}},
		{"downlink transmitter power", []any{" 100", inav.SymMw, inav.SymAhDecorationDown}, map[string]float64{"osd_rx_power_mw": 100}, []string{"osd_tx_power_mw"}},

		// time
		{"time since power on", []any{inav.SymOnM, "20:59"}, map[string]float64{"osd_on_time_s": 1259}, []string{"osd_flight_time_s"}},
		{"time since power on in hours", []any{inav.SymOnH, "01:40"}, map[string]float64{"osd_on_time_s": 6000}, nil},
		{"flight time left before RTH", []any{inav.SymFlightMinsRemaining, "12:30"}, map[string]float64{"osd_rth_time_left_s": 750}, nil},
		{"flight time left in hours", []any{inav.SymFlightHoursRemaining, "02:15"}, map[string]float64{"osd_rth_time_left_s": 8100}, nil},
		{"glide time", []any{inav.SymGlideMins, "02:05"}, map[string]float64{"osd_glide_time_s": 125}, nil},

		// battery
		{"battery charge", []any{inav.SymBatt3, " 85%"}, map[string]float64{"osd_batt_pct": 85}, []string{"osd_batt_cell_v"}},
		{"battery charge at 100", []any{inav.SymBattFull, "100%"}, map[string]float64{"osd_batt_pct": 100}, nil},
		{"energy used", []any{inavNum("1.23"), inav.SymWh}, map[string]float64{"osd_wh_used": 1.23}, nil},
		{"energy per km", []any{inavNum("15.6"), inav.SymWhKm}, map[string]float64{"osd_wh_per_km": 15.6}, nil},
		{"energy per mile is converted to km", []any{inavNum("15.6"), inav.SymWhMi}, map[string]float64{"osd_wh_per_km": 15.6 / 1.609344}, nil},
		{"energy per nautical mile is converted to km", []any{inavNum("15.6"), inav.SymWhNm}, map[string]float64{"osd_wh_per_km": 15.6 / 1.852}, nil},
		{"power in watts", []any{inavNum("5.12"), inav.SymWatt}, map[string]float64{"osd_power_w": 5.12}, nil},
		{"power in kilowatts", []any{inavNum("1.23"), inav.SymKilowatt}, map[string]float64{"osd_power_w": 1230}, nil},
		{"supply impedance", []any{"125", inav.SymMilliohm}, map[string]float64{"osd_impedance_mohm": 125}, nil},
		{"mAh per mile is converted to km", []any{"208", inav.SymMahMi0, inav.SymMahMi1}, map[string]float64{"osd_mah_per_km": 208 / 1.609344}, nil},
		{"mAh per nautical mile is converted to km", []any{"208", inav.SymMahNm0, inav.SymMahNm1}, map[string]float64{"osd_mah_per_km": 208 / 1.852}, nil},

		// angles
		{"home heading error", []any{inav.SymHome, inav.SymHeading, "  45", inav.SymDegrees},
			map[string]float64{"osd_home_heading_error_deg": 45}, []string{"osd_heading_deg", "osd_home_dist_m"}},
		{"negative home heading error", []any{inav.SymHome, inav.SymHeading, " -45", inav.SymDegrees},
			map[string]float64{"osd_home_heading_error_deg": -45}, []string{"osd_heading_deg"}},
		{"ground course", []any{inav.SymGroundCourse, "254", inav.SymDegrees}, map[string]float64{"osd_ground_course_deg": 254}, nil},
		{"azimuth", []any{inav.SymAzimuth, " 90", inav.SymDegrees}, map[string]float64{"osd_azimuth_deg": 90}, nil},
		{"g-force on x", []any{inav.SymGforceX, inavNum("-0.45")}, map[string]float64{"osd_gforce_x": -0.45}, []string{"osd_gforce"}},
		{"g-force on y", []any{inav.SymGforceY, inavNum("1.02")}, map[string]float64{"osd_gforce_y": 1.02}, nil},
		{"g-force on z", []any{inav.SymGforceZ, inavNum("0.98")}, map[string]float64{"osd_gforce_z": 0.98}, nil},

		// gps
		{"hdop", []any{inav.SymHdpL, inav.SymHdpR, inavNum("1.2")}, map[string]float64{"osd_gps_hdop": 1.2}, nil},

		// distances
		{"distance left before RTH", []any{inav.SymFlightDistRemaining, "850", inav.SymDistM}, map[string]float64{"osd_rth_dist_left_m": 850}, []string{"osd_home_dist_m"}},
		{"distance left before RTH in km", []any{inav.SymFlightDistRemaining, inavNum("1.2"), inav.SymDistKm}, map[string]float64{"osd_rth_dist_left_m": 1200}, nil},
		{"glide range", []any{inav.SymGlideDist, "123", inav.SymDistM}, map[string]float64{"osd_glide_range_m": 123}, nil},
		{"cross track error", []any{inav.SymCrossTrackError, " 45", inav.SymDistM}, map[string]float64{"osd_cross_track_error_m": 45}, nil},
		{"horizontal distance to the fence", []any{"FD ", " 45", inav.SymDistM}, map[string]float64{"osd_fence_dist_h_m": 45}, nil},
		{"vertical distance to the fence, above", []any{"FD", " ", " 12", inav.SymAltM, " ", inav.SymDecoration},
			map[string]float64{"osd_fence_dist_v_m": 12}, []string{"osd_altitude_m"}},
		{"vertical distance to the fence, below", []any{"FD", " ", " 12", inav.SymAltM, " ", inav.SymDecoration + 4},
			map[string]float64{"osd_fence_dist_v_m": -12}, []string{"osd_altitude_m"}},
		{"odometer", []any{inav.SymOdometer, "0001", td(0), ld(3), inav.SymKm}, map[string]float64{"osd_odometer_km": 10.3}, nil},
		{"odometer in miles", []any{inav.SymOdometer, "0001", td(0), ld(3), inav.SymMi}, map[string]float64{"osd_odometer_km": 10.3 * 1.609344}, nil},
		{"rangefinder", []any{inavNum("1.5"), inav.SymDistM}, map[string]float64{"osd_rangefinder_m": 1.5}, []string{"osd_home_dist_m"}},

		// speeds
		{"3d speed", []any{" 50", inav.Sym3dKmh}, map[string]float64{"osd_3d_speed_kmh": 50}, []string{"osd_gps_speed_kmh"}},
		{"3d speed in mph", []any{" 50", inav.Sym3dMph}, map[string]float64{"osd_3d_speed_kmh": 50 * 1.609344}, nil},
		{"maximum ground speed", []any{inav.SymMax, "120", inav.SymKmh}, map[string]float64{"osd_gps_max_speed_kmh": 120}, []string{"osd_gps_speed_kmh"}},
		{"maximum airspeed", []any{inav.SymMax, inav.SymAir, "100", inav.SymKmh}, map[string]float64{"osd_air_max_speed_kmh": 100}, []string{"osd_air_speed_kmh", "osd_gps_speed_kmh"}},
		{"maximum 3d speed", []any{inav.SymMax, "130", inav.Sym3dKmh}, map[string]float64{"osd_3d_max_speed_kmh": 130}, []string{"osd_gps_max_speed_kmh", "osd_gps_speed_kmh"}},
		{"minimum ground speed", []any{inav.SymMinGroundSpeed, " 30", inav.SymKmh}, map[string]float64{"osd_min_ground_speed_kmh": 30}, []string{"osd_gps_speed_kmh"}},
		{"vertical wind going up", []any{inav.SymWindVertical, inav.SymAhDecorationUp, td(1), ld(2), "0", inav.SymKmh},
			map[string]float64{"osd_wind_vspeed_kmh": 1.2}, []string{"osd_wind_speed_kmh", "osd_gps_speed_kmh"}},
		{"vertical wind going down", []any{inav.SymWindVertical, inav.SymAhDecorationDown, td(1), ld(5), "0", inav.SymKmh},
			map[string]float64{"osd_wind_vspeed_kmh": -1.5}, nil},
		{"horizontal wind in m/s", []any{inav.SymWindHorizontal, inav.SymDecoration + 2, td(2), ld(5), "0", inav.SymMs},
			map[string]float64{"osd_wind_speed_kmh": 9, "osd_wind_dir_deg": 90}, []string{"osd_vspeed_ms"}},
		{"climbing", []any{inavNum("2.5"), inav.SymMs}, map[string]float64{"osd_vspeed_ms": 2.5}, []string{"osd_gps_speed_kmh"}},
		{"sinking", []any{inavNum("-0.3"), inav.SymMs}, map[string]float64{"osd_vspeed_ms": -0.3}, nil},
		{"climbing in ft/s", []any{inavNum("1.0"), inav.SymFts}, map[string]float64{"osd_vspeed_ms": 0.3048}, nil},
		{"climbing in 100ft/min", []any{inavNum("5.0"), inav.Sym100ftm}, map[string]float64{"osd_vspeed_ms": 5 * 100 * 0.3048 / 60}, nil},

		// sensors
		{"imu temperature", []any{inav.SymImuTemp, " 40", inav.SymTempC}, map[string]float64{"osd_imu_temp_c": 40}, nil},
		{"baro temperature in Fahrenheit", []any{inav.SymBaroTemp, "104", inav.SymTempF}, map[string]float64{"osd_baro_temp_c": 40}, nil},
		{"esc temperature", []any{inav.SymEscTemp, " 55", inav.SymTempC}, map[string]float64{"osd_esc_temp_c": 55}, []string{"osd_temp_sensor_c"}},
		{"sensor temperature with an icon", []any{inav.SymTempSensorFirst + 2, "-10", inav.SymTempC}, map[string]float64{"osd_temp_sensor_c": -10}, nil},
		{"sensor temperature with a label", []any{" 30", inav.SymTempC}, map[string]float64{"osd_temp_sensor_c": 30}, nil},
		{"esc rpm", []any{inav.SymRpm, "1234"}, map[string]float64{"osd_esc_rpm": 1234}, nil},
		{"esc rpm in thousands", []any{inav.SymRpm, "1", td(2), ld(3), "K"}, map[string]float64{"osd_esc_rpm": 12300}, nil},
		{"glide slope", []any{inav.SymGlideslope, inavNum("1.50")}, map[string]float64{"osd_glideslope": 1.5}, nil},
		{"climb efficiency per metre", []any{inavNum("0.45"), inav.SymAhVM0, inav.SymAhVM1}, map[string]float64{"osd_climb_eff_mah_per_m": 0.45}, nil},
		{"climb efficiency per foot is converted to metres", []any{inavNum("0.45"), inav.SymAhVFt0, inav.SymAhVFt1}, map[string]float64{"osd_climb_eff_mah_per_m": 0.45 / 0.3048}, nil},
		{"profile", []any{inav.SymProfile, "2"}, map[string]float64{"osd_profile": 2}, nil},
		{"blackbox log", []any{inav.SymBlackbox, "00012"}, map[string]float64{"osd_blackbox_log": 12}, nil},
		{"global variable", []any{"G1:", "   42"}, map[string]float64{"osd_gvar_1": 42}, nil},
		{"negative global variable", []any{"G3:", "   -5"}, map[string]float64{"osd_gvar_3": -5}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFile()
			put(f, 10, 5, tt.parts...)
			s := decode(t, f)
			checkNum(t, s, tt.want)
			for _, key := range tt.absent {
				if v, ok := s[key]; ok {
					t.Errorf("%s was read as %v", key, v.Num)
				}
			}
			for key := range s {
				if _, expected := tt.want[key]; !expected && isExtraKey(key) {
					t.Errorf("part of the element went to the generic pass: %s", key)
				}
			}
		})
	}
}

func TestTextElementFormats(t *testing.T) {
	tests := []struct {
		name  string
		parts []any
		want  map[string]string
	}{
		{"receiver band", []any{inav.SymRxBand, "ELRS"}, map[string]string{"osd_rx_band": "ELRS"}},
		{"receiver band shorter than its cells", []any{inav.SymRxBand, "AU  "}, map[string]string{"osd_rx_band": "AU"}},
		{"receiver mode", []any{inav.SymRxMode, "F1000 "}, map[string]string{"osd_rx_mode": "F1000"}},
		{"video transmitter channel", []any{"CH:R1:", "3"}, map[string]string{"osd_vtx_channel": "R1", "osd_vtx_power": "3"}},
		{"video transmitter power", []any{inav.SymVtxPower, "M"}, map[string]string{"osd_vtx_power": "M"}},
		{"real-time clock", []any{inav.SymClock, "11:27:51"}, map[string]string{"osd_clock": "11:27:51"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFile()
			put(f, 10, 5, tt.parts...)
			s := decode(t, f)
			for key, want := range tt.want {
				if got := s[key].Text; got != want {
					t.Errorf("%s = %q, want %q", key, got, want)
				}
			}
			if len(s) != len(tt.want) {
				t.Errorf("read %v, want only %v", s, tt.want)
			}
		})
	}
}

// With both channel and power on screen, the channel must not be lost.
func TestVTXChannelSurvivesTheVTXPowerElement(t *testing.T) {
	f := newFile()
	put(f, 3, 5, "CH:A2:", "4")
	put(f, 20, 5, inav.SymVtxPower, "4")
	s := decode(t, f)
	if got := s["osd_vtx_channel"].Text; got != "A2" {
		t.Errorf("channel = %q, want A2", got)
	}
	if got := s["osd_vtx_power"].Text; got != "4" {
		t.Errorf("power = %q, want 4", got)
	}
}

// Everything the wind-speed element can show: the estimate flagged as not
// valid yet has a * in place of the unit and is not a reading.
func TestInvalidWindIsNotAReading(t *testing.T) {
	f := newFile()
	put(f, 43, 8, inav.SymWindHorizontal, inav.SymDecoration+3, td(0), ld(0), "0", "*")
	put(f, 43, 9, inav.SymWindVertical, inav.SymAhDecorationUp, td(0), ld(0), "0", "*")
	s, left := New(sampleLayout()).Decode(f, &f.Frames[0])
	if len(s) != 0 {
		t.Errorf("read %v, want nothing", s)
	}
	for _, l := range left {
		t.Errorf("left on screen: %+v", l)
	}
}

// Two elements the firmware draws alike cannot be told apart: the first one on
// the screen keeps the column and the other one is not lost.
func TestFirstOfTwoAlikeReadingsWins(t *testing.T) {
	f := newFile()
	put(f, 3, 4, " 969", inav.SymMah)  // capacity used
	put(f, 3, 8, " 750", inav.SymMah)  // capacity left: same glyphs
	put(f, 30, 8, " 500", inav.SymMah) // and on the same row, to the right
	s := decode(t, f)
	checkNum(t, s, map[string]float64{"osd_mah_used": 969})
	if got := s["osd_raw_mah"]; got.Num != 750 {
		t.Errorf("osd_raw_mah = %v, want the second reading (750)", got.Num)
	}
}

// A rule's match must not overwrite a column a previous row filled.
func TestLaterRowsDoNotOverwrite(t *testing.T) {
	f := newFile()
	put(f, 45, 4, "213", inav.SymAltM)
	put(f, 45, 12, "99", inav.SymAltM)
	s := decode(t, f)
	checkNum(t, s, map[string]float64{"osd_altitude_m": 213})
}

// The unit glyph of the elements decides the unit conversion: the same number
// in a different unit means a different value.
func TestUnitFactors(t *testing.T) {
	approx := func(a, b float64) bool { return math.Abs(a-b) < 1e-9 }
	if !approx(vSpeedToMS[inav.Sym100ftm], 0.508) {
		t.Errorf("100ft/min = %v m/s, want 0.508", vSpeedToMS[inav.Sym100ftm])
	}
	if !approx(windToKmh[inav.SymMs], 3.6) {
		t.Errorf("1 m/s = %v km/h, want 3.6", windToKmh[inav.SymMs])
	}
}

// A rule that covers several elements (the three G-force axes, the four
// global variables) reads all of them when they share a row.
func TestSeveralElementsOfOneRuleOnARow(t *testing.T) {
	f := newFile()
	put(f, 2, 3, inav.SymGforceX, inavNum("0.10"), " ", inav.SymGforceY, inavNum("0.20"), " ", inav.SymGforceZ, inavNum("0.30"))
	put(f, 2, 6, "G0:", "    1", " G1:", "    2", " G2:", "    3", " G3:", "    4")
	s := decode(t, f)
	checkNum(t, s, map[string]float64{
		"osd_gforce_x": 0.1, "osd_gforce_y": 0.2, "osd_gforce_z": 0.3,
		"osd_gvar_0": 1, "osd_gvar_1": 2, "osd_gvar_2": 3, "osd_gvar_3": 4,
	})
}

// Some glyphs of the font sit on the codes of ASCII punctuation, so the same
// characters appear in text. The craft name is free text, and it must not be
// taken for a reading, nor take the column from the real one below it.
func TestTextIsNotMistakenForGlyphs(t *testing.T) {
	for _, text := range []string{
		"MY X8 2$", // Ah per mile
		"IS IT 5?", // Ah per nautical mile
	} {
		t.Run(text, func(t *testing.T) {
			f := newFile()
			put(f, 3, 1, text)
			put(f, 3, 5, "208", inav.SymMahKm0, inav.SymMahKm1) // the real reading, drawn below
			checkNum(t, decode(t, f), map[string]float64{"osd_mah_per_km": 208})
		})
	}
	t.Run("an apostrophe inside a word is not a VTX power", func(t *testing.T) {
		f := newFile()
		put(f, 3, 1, "DON'T 5")
		if s := decode(t, f); len(s) != 0 {
			t.Errorf("read %v from a text", s)
		}
	})
}

// Every column of the catalog is produced by some rule: the frames below put
// every element on screen at once, away from each other.
func TestEveryCatalogColumnIsProduced(t *testing.T) {
	got := map[string]bool{}
	for _, f := range []*osd.File{frame3000(), everyElement()} {
		s, _ := New(sampleLayout()).Decode(f, &f.Frames[0])
		for key := range s {
			got[key] = true
		}
	}
	for _, f := range Fields() {
		if !got[f.Key] {
			t.Errorf("no frame produces %s", f.Key)
		}
	}
}

// everyElement puts the elements that frame3000 lacks on the screen, three or
// four to a row, with room between them.
func everyElement() *osd.File {
	f := newFile()
	rows := [][][]any{
		{{inav.SymRssi, " 87"}, {inav.SymLq, " 98", inav.SymAhDecorationDown}, {inav.SymSnr, "-12", inav.SymDb}, {" 250", inav.SymMw}},
		{{" 100", inav.SymMw, inav.SymAhDecorationDown}, {inav.SymRxBand, "ELRS"}, {inav.SymRxMode, "F1000 "}, {"CH:R1:", "3"}},
		{{inav.SymOnM, "20:59"}, {inav.SymFlightMinsRemaining, "12:30"}, {inav.SymGlideMins, "02:05"}, {inav.SymClock, "11:27:51"}},
		{{inav.SymBatt3, " 85%"}, {inavNum("1.23"), inav.SymWh}, {inavNum("15.6"), inav.SymWhKm}, {inavNum("5.12"), inav.SymWatt}},
		{{"125", inav.SymMilliohm}, {inav.SymHdpL, inav.SymHdpR, inavNum("1.2")}, {inav.SymHome, inav.SymHeading, " -45", inav.SymDegrees},
			{inav.SymBatt5, "1", td(6), ld(0), inav.SymVolt}},
		{{inav.SymGroundCourse, "254", inav.SymDegrees}, {inav.SymAzimuth, " 90", inav.SymDegrees}, {inav.SymGforceX, inavNum("-0.45")}},
		{{inav.SymGforceY, inavNum("1.02")}, {inav.SymGforceZ, inavNum("0.98")}, {inav.SymFlightDistRemaining, "850", inav.SymDistM}},
		{{inav.SymGlideDist, "123", inav.SymDistM}, {inav.SymCrossTrackError, " 45", inav.SymDistM}, {"FD ", " 45", inav.SymDistM}},
		{{"FD", " ", " 12", inav.SymAltM, " ", inav.SymDecoration}, {inav.SymOdometer, "0001", td(0), ld(3), inav.SymKm}},
		{{inavNum("1.5"), inav.SymDistM}, {" 50", inav.Sym3dKmh}, {inav.SymMax, "120", inav.SymKmh}, {inav.SymMax, inav.SymAir, "100", inav.SymKmh}},
		{{inav.SymMax, "130", inav.Sym3dKmh}, {inav.SymMinGroundSpeed, " 30", inav.SymKmh}, {inav.SymWindVertical, inav.SymAhDecorationDown, td(1), ld(5), "0", inav.SymKmh}},
		{{inavNum("-0.3"), inav.SymMs}, {inav.SymImuTemp, " 40", inav.SymTempC}, {inav.SymBaroTemp, " 42", inav.SymTempC}, {inav.SymEscTemp, " 55", inav.SymTempC}},
		{{inav.SymTempSensorFirst + 2, " 30", inav.SymTempC}, {inav.SymRpm, "1234"}, {inav.SymGlideslope, inavNum("1.50")}},
		{{inavNum("0.45"), inav.SymAhVM0, inav.SymAhVM1}, {inav.SymProfile, "2"}, {inav.SymBlackbox, "00012"}, {"G0:", "   42"}},
		{{"G1:", "   42"}, {"G2:", "   42"}, {"G3:", "   42"}},
	}
	for y, row := range rows {
		for i, element := range row {
			put(f, 1+i*13, y+1, element...)
		}
	}
	return f
}
