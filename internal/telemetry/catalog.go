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
)

// Field describes a telemetry column that can come out of the OSD.
type Field struct {
	Key      string // CSV column name (unit as suffix)
	Desc     string // what the value represents
	Decimals int    // decimal places shown on screen, used when writing
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
}

// Fields returns every column the decoder can produce, in CSV order.
func Fields() []Field {
	return append([]Field(nil), catalog...)
}
