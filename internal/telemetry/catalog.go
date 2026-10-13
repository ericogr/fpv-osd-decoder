// Package telemetry extracts numeric values and texts from INAV OSD screens.
package telemetry

// Column keys. The catalog and the decoding rules share these constants, so a
// typo cannot split one field in two.
const (
	keyRSSIdBm      = "osd_rssi_dbm"
	keyRSSIAntenna  = "osd_rssi_antenna"
	keyRFMode       = "osd_rf_mode"
	keyLQ           = "osd_lq_pct"
	keyFlightTime   = "osd_flight_time_s"
	keyBattCellV    = "osd_batt_cell_v"
	keyBattV        = "osd_batt_v"
	keyCurrent      = "osd_current_a"
	keyMAhUsed      = "osd_mah_used"
	keyMAhPerKm     = "osd_mah_per_km"
	keyThrottle     = "osd_throttle_pct"
	keyThrottleAuto = "osd_throttle_auto"
	keyFlightMode   = "osd_flight_mode"
	keyMessage      = "osd_message"
	keyHeading      = "osd_heading_deg"
	keyRoll         = "osd_roll_deg"
	keyPitch        = "osd_pitch_deg"
	keyGForce       = "osd_gforce"
	keySats         = "osd_gps_sats"
	keyLat          = "osd_gps_lat"
	keyLon          = "osd_gps_lon"
	keyGPSSpeed     = "osd_gps_speed_kmh"
	keyAirSpeed     = "osd_air_speed_kmh"
	keyAltitude     = "osd_altitude_m"
	keyHomeDist     = "osd_home_dist_m"
	keyHomeDir      = "osd_home_dir_deg"
	keyTripDist     = "osd_trip_dist_m"
	keyWindSpeed    = "osd_wind_speed_kmh"
	keyWindDir      = "osd_wind_dir_deg"

	// Columns the README lists under "Read from the INAV 9.1.0 source".
	keyRSSIPct          = "osd_rssi_pct"
	keyLQDown           = "osd_lq_down_pct"
	keySNR              = "osd_snr_db"
	keyTXPower          = "osd_tx_power_mw"
	keyRXPower          = "osd_rx_power_mw"
	keyRXBand           = "osd_rx_band"
	keyRXMode           = "osd_rx_mode"
	keyVTXChannel       = "osd_vtx_channel"
	keyVTXPower         = "osd_vtx_power"
	keyOnTime           = "osd_on_time_s"
	keyRTHTimeLeft      = "osd_rth_time_left_s"
	keyGlideTime        = "osd_glide_time_s"
	keyClock            = "osd_clock"
	keyBattPct          = "osd_batt_pct"
	keyWhUsed           = "osd_wh_used"
	keyWhPerKm          = "osd_wh_per_km"
	keyPower            = "osd_power_w"
	keyImpedance        = "osd_impedance_mohm"
	keyHDOP             = "osd_gps_hdop"
	keyHomeHeadingError = "osd_home_heading_error_deg"
	keyGroundCourse     = "osd_ground_course_deg"
	keyAzimuth          = "osd_azimuth_deg"
	keyRTHDistLeft      = "osd_rth_dist_left_m"
	keyGlideRange       = "osd_glide_range_m"
	keyCrossTrack       = "osd_cross_track_error_m"
	keyFenceDistH       = "osd_fence_dist_h_m"
	keyFenceDistV       = "osd_fence_dist_v_m"
	keyOdometer         = "osd_odometer_km"
	keyRangefinder      = "osd_rangefinder_m"
	keyGlideslope       = "osd_glideslope"
	keyWindVSpeed       = "osd_wind_vspeed_kmh"
	keyAirMaxSpeed      = "osd_air_max_speed_kmh"
	keyGPSMaxSpeed      = "osd_gps_max_speed_kmh"
	key3DSpeed          = "osd_3d_speed_kmh"
	key3DMaxSpeed       = "osd_3d_max_speed_kmh"
	keyMinGroundSpeed   = "osd_min_ground_speed_kmh"
	keyVSpeed           = "osd_vspeed_ms"
	keyIMUTemp          = "osd_imu_temp_c"
	keyBaroTemp         = "osd_baro_temp_c"
	keyESCTemp          = "osd_esc_temp_c"
	keySensorTemp       = "osd_temp_sensor_c"
	keyESCRPM           = "osd_esc_rpm"
	keyGForceX          = "osd_gforce_x"
	keyGForceY          = "osd_gforce_y"
	keyGForceZ          = "osd_gforce_z"
	keyProfile          = "osd_profile"
	keyGVar0            = "osd_gvar_0"
	keyGVar1            = "osd_gvar_1"
	keyGVar2            = "osd_gvar_2"
	keyGVar3            = "osd_gvar_3"
	keyBlackboxLog      = "osd_blackbox_log"
	keyClimbEfficiency  = "osd_climb_eff_mah_per_m"
)

// Keys of the columns that other packages pick out of a table.
const (
	KeyLat        = keyLat
	KeyLon        = keyLon
	KeyAltitude   = keyAltitude
	KeyFlightMode = keyFlightMode
)

// Field describes a telemetry column that can come out of the OSD.
type Field struct {
	Key      string // CSV column name (unit as suffix)
	Desc     string // what the value represents
	Decimals int    // decimal places shown on screen, used when writing; negative keeps all digits
	Text     bool   // text value instead of numeric
}

// catalog lists every column the decoder can produce, in the order they
// appear in the CSV.
var catalog = []Field{
	{Key: keyRSSIdBm, Desc: "radio link RSSI (dBm)"},
	{Key: keyRSSIAntenna, Desc: "active receiver antenna: 1 (RSSI icon) or 2 (2RSS icon)"},
	{Key: keyRFMode, Desc: "CRSF link RF mode (the n in n:LQ)"},
	{Key: keyLQ, Desc: "CRSF link quality, uplink (%)"},
	{Key: keyFlightTime, Desc: "armed flight time, counted by the flight controller (s)"},
	{Key: keyBattCellV, Desc: "average cell voltage of the main battery (V)", Decimals: 2},
	{Key: keyBattV, Desc: "total voltage of the main battery (V)", Decimals: 2},
	{Key: keyCurrent, Desc: "current draw (A)", Decimals: 2},
	{Key: keyMAhUsed, Desc: "capacity used (mAh)"},
	{Key: keyMAhPerKm, Desc: "energy efficiency (mAh/km)"},
	{Key: keyThrottle, Desc: "throttle position (%)"},
	{Key: keyThrottleAuto, Desc: "1 if the throttle is controlled by navigation (auto-throttle), 0 if manual"},
	{Key: keyFlightMode, Desc: "flight mode (ACRO, ANGL, RTH, CRUZ...)", Text: true},
	{Key: keyMessage, Desc: "INAV status message (EN ROUTE TO HOME...)", Text: true},
	{Key: keyHeading, Desc: "aircraft heading (degrees, 0 = north)"},
	{Key: keyRoll, Desc: "roll (degrees; positive = right wing down)", Decimals: 1},
	{Key: keyPitch, Desc: "pitch (degrees; positive = nose up)", Decimals: 1},
	{Key: keyGForce, Desc: "total G-force", Decimals: 2},
	{Key: keySats, Desc: "GPS satellites in use"},
	{Key: keyLat, Desc: "GPS latitude (decimal degrees)", Decimals: 7},
	{Key: keyLon, Desc: "GPS longitude (decimal degrees)", Decimals: 7},
	{Key: keyGPSSpeed, Desc: "ground speed from GPS (km/h)", Decimals: 1},
	{Key: keyAirSpeed, Desc: "airspeed from the pitot tube (km/h)", Decimals: 1},
	{Key: keyAltitude, Desc: "altitude (m)", Decimals: 1},
	{Key: keyHomeDist, Desc: "distance to the home point (m)", Decimals: 1},
	{Key: keyHomeDir, Desc: "direction of home relative to the direction of flight (degrees, 0 = ahead; 22.5 steps)", Decimals: 1},
	{Key: keyTripDist, Desc: "total distance travelled in the flight (m)", Decimals: 1},
	{Key: keyWindSpeed, Desc: "estimated horizontal wind speed (km/h)", Decimals: 1},
	{Key: keyWindDir, Desc: "wind direction relative to the nose (degrees; 45 steps)"},

	// Read from the format INAV 9.1.0 draws; see the README.
	{Key: keyRSSIPct, Desc: "RSSI as a percentage of the receiver's range (%)"},
	{Key: keyLQDown, Desc: "CRSF link quality, downlink (%)"},
	{Key: keySNR, Desc: "link signal-to-noise ratio (dB)"},
	{Key: keyTXPower, Desc: "radio transmitter power, uplink (mW)"},
	{Key: keyRXPower, Desc: "power of the downlink transmitter on the aircraft (mW)"},
	{Key: keyRXBand, Desc: "receiver band (text)", Text: true},
	{Key: keyRXMode, Desc: "receiver mode (text)", Text: true},
	{Key: keyVTXChannel, Desc: "video transmitter band and channel (text, such as R1)", Text: true},
	{Key: keyVTXPower, Desc: "video transmitter power level, as the letter or digit INAV shows (text)", Text: true},
	{Key: keyOnTime, Desc: "time since the flight controller was powered (s)"},
	{Key: keyRTHTimeLeft, Desc: "estimated flight time left before the return to home must start (s)"},
	{Key: keyGlideTime, Desc: "estimated time the aircraft can still glide (s)"},
	{Key: keyClock, Desc: "real-time clock (HH:MM:SS text)", Text: true},
	{Key: keyBattPct, Desc: "remaining battery charge (%)"},
	{Key: keyWhUsed, Desc: "energy used (Wh)", Decimals: 2},
	{Key: keyWhPerKm, Desc: "energy efficiency (Wh/km)", Decimals: 2},
	{Key: keyPower, Desc: "power draw (W)", Decimals: 2},
	{Key: keyImpedance, Desc: "power supply impedance (mOhm)"},
	{Key: keyHDOP, Desc: "GPS horizontal dilution of precision", Decimals: 1},
	{Key: keyHomeHeadingError, Desc: "angle between the direction of flight and the direction of home (degrees)"},
	{Key: keyGroundCourse, Desc: "GPS course over the ground (degrees)"},
	{Key: keyAzimuth, Desc: "azimuth of the aircraft seen from home (degrees)"},
	{Key: keyRTHDistLeft, Desc: "distance that can still be flown before the return to home must start (m)", Decimals: 1},
	{Key: keyGlideRange, Desc: "distance the aircraft can still glide (m)", Decimals: 1},
	{Key: keyCrossTrack, Desc: "cross-track error from the waypoint leg (m)", Decimals: 1},
	{Key: keyFenceDistH, Desc: "horizontal distance to the nearest geofence zone (m)", Decimals: 1},
	{Key: keyFenceDistV, Desc: "vertical distance to the nearest geofence zone (m; negative = below it)", Decimals: 1},
	{Key: keyOdometer, Desc: "distance travelled in all flights, as kept by the flight controller (km)", Decimals: 2},
	{Key: keyRangefinder, Desc: "rangefinder distance (m)", Decimals: 1},
	{Key: keyGlideslope, Desc: "glide ratio (horizontal distance per unit of height lost)", Decimals: 2},
	{Key: keyWindVSpeed, Desc: "estimated vertical wind speed (km/h; positive = up)", Decimals: 1},
	{Key: keyAirMaxSpeed, Desc: "maximum airspeed of the flight (km/h)", Decimals: 1},
	{Key: keyGPSMaxSpeed, Desc: "maximum ground speed of the flight (km/h)", Decimals: 1},
	{Key: key3DSpeed, Desc: "3D speed, including the vertical component (km/h)", Decimals: 1},
	{Key: key3DMaxSpeed, Desc: "maximum 3D speed of the flight (km/h)", Decimals: 1},
	{Key: keyMinGroundSpeed, Desc: "minimum ground speed set for navigation (km/h)", Decimals: 1},
	{Key: keyVSpeed, Desc: "vertical speed (m/s; positive = climbing)", Decimals: 2},
	{Key: keyIMUTemp, Desc: "flight controller IMU temperature (Celsius)", Decimals: 1},
	{Key: keyBaroTemp, Desc: "barometer temperature (Celsius)", Decimals: 1},
	{Key: keyESCTemp, Desc: "ESC temperature (Celsius)", Decimals: 1},
	{Key: keySensorTemp, Desc: "external temperature sensor (Celsius)", Decimals: 1},
	{Key: keyESCRPM, Desc: "motor speed reported by the ESC (rpm)"},
	{Key: keyGForceX, Desc: "G-force on the X axis", Decimals: 2},
	{Key: keyGForceY, Desc: "G-force on the Y axis", Decimals: 2},
	{Key: keyGForceZ, Desc: "G-force on the Z axis", Decimals: 2},
	{Key: keyProfile, Desc: "active configuration profile"},
	{Key: keyGVar0, Desc: "global variable 0 of the INAV programming framework"},
	{Key: keyGVar1, Desc: "global variable 1 of the INAV programming framework"},
	{Key: keyGVar2, Desc: "global variable 2 of the INAV programming framework"},
	{Key: keyGVar3, Desc: "global variable 3 of the INAV programming framework"},
	{Key: keyBlackboxLog, Desc: "number of the blackbox log being recorded"},
	{Key: keyClimbEfficiency, Desc: "climb efficiency (mAh per vertical metre, as the OSD shows it)", Decimals: 2},
}

// Fields returns every column the decoder can produce, in CSV order.
func Fields() []Field {
	return append([]Field(nil), catalog...)
}

// FieldsFor returns the columns that samples can fill: the catalog, then the
// readings no rule knows that the generic pass found in them, sorted by name.
func FieldsFor(samples []Sample) []Field {
	return append(Fields(), extraFields(samples)...)
}
