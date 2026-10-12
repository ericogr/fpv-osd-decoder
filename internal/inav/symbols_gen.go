// Code generated from INAV src/main/drivers/osd_symbols.h; DO NOT EDIT.

package inav

// Constants for the glyphs of the INAV OSD font (the firmware's SYM_*).
const (
	SymRssi                 uint16 = 0x01
	SymLq                   uint16 = 0x02
	SymLat                  uint16 = 0x03
	SymLon                  uint16 = 0x04
	SymAzimuth              uint16 = 0x05
	SymTelemetry0           uint16 = 0x06
	SymTelemetry1           uint16 = 0x07
	SymSatL                 uint16 = 0x08
	SymSatR                 uint16 = 0x09
	SymHomeNear             uint16 = 0x0A
	SymDegrees              uint16 = 0x0B
	SymHeading              uint16 = 0x0C
	SymScale                uint16 = 0x0D
	SymHdpL                 uint16 = 0x0E
	SymHdpR                 uint16 = 0x0F
	SymHome                 uint16 = 0x10
	Sym2rss                 uint16 = 0x11
	SymDb                   uint16 = 0x12
	SymDbm                  uint16 = 0x13
	SymSnr                  uint16 = 0x14
	SymAhDecorationUp       uint16 = 0x15
	SymAhDecorationDown     uint16 = 0x16
	SymDecoration           uint16 = 0x17
	SymVolt                 uint16 = 0x1F
	SymMah                  uint16 = 0x99
	SymAhKm                 uint16 = 0x22
	SymAhMi                 uint16 = 0x24
	SymVtxPower             uint16 = 0x27
	SymAhNm                 uint16 = 0x3F
	SymMahNm0               uint16 = 0x60
	SymMahNm1               uint16 = 0x61
	SymMahKm0               uint16 = 0x6B
	SymMahKm1               uint16 = 0x6C
	SymMilliohm             uint16 = 0x62
	SymBattFull             uint16 = 0x63
	SymBatt5                uint16 = 0x64
	SymBatt4                uint16 = 0x65
	SymBatt3                uint16 = 0x66
	SymBatt2                uint16 = 0x67
	SymBatt1                uint16 = 0x68
	SymBattEmpty            uint16 = 0x69
	SymAmp                  uint16 = 0x6A
	SymWh                   uint16 = 0x6D
	SymWhKm                 uint16 = 0x6E
	SymWhMi                 uint16 = 0x6F
	SymWhNm                 uint16 = 0x70
	SymWatt                 uint16 = 0x71
	SymMw                   uint16 = 0x72
	SymKilowatt             uint16 = 0x73
	SymFt                   uint16 = 0x74
	SymTripDist             uint16 = 0x75
	SymTotal                uint16 = 0x75
	SymAltM                 uint16 = 0x76
	SymAltKm                uint16 = 0x77
	SymAltFt                uint16 = 0x78
	SymAltKft               uint16 = 0x79
	SymDistM                uint16 = 0x7A
	SymDistKm               uint16 = 0x7E
	SymDistFt               uint16 = 0x7F
	SymDistMi               uint16 = 0x80
	SymDistNm               uint16 = 0x81
	SymM                    uint16 = 0x82
	SymKm                   uint16 = 0x83
	SymMi                   uint16 = 0x84
	SymNm                   uint16 = 0x85
	SymWindHorizontal       uint16 = 0x86
	SymWindVertical         uint16 = 0x87
	Sym3dKmh                uint16 = 0x88
	Sym3dMph                uint16 = 0x89
	Sym3dKt                 uint16 = 0x8A
	SymRpm                  uint16 = 0x8B
	SymAir                  uint16 = 0x8C
	SymFts                  uint16 = 0x8D
	Sym100ftm               uint16 = 0x8E
	SymMs                   uint16 = 0x8F
	SymKmh                  uint16 = 0x90
	SymMph                  uint16 = 0x91
	SymKt                   uint16 = 0x92
	SymMahMi0               uint16 = 0x93
	SymMahMi1               uint16 = 0x94
	SymThr                  uint16 = 0x95
	SymTempF                uint16 = 0x96
	SymTempC                uint16 = 0x97
	SymBlank                uint16 = 0x20
	SymOnH                  uint16 = 0x9A
	SymFlyH                 uint16 = 0x9B
	SymOnM                  uint16 = 0x9E
	SymFlyM                 uint16 = 0x9F
	SymGlideslope           uint16 = 0x9C
	SymWaypoint             uint16 = 0x9D
	SymClock                uint16 = 0xA0
	SymZeroHalfTrailingDot  uint16 = 0xA1
	SymZeroHalfLeadingDot   uint16 = 0xB1
	SymAutoThr0             uint16 = 0xAB
	SymAutoThr1             uint16 = 0xAC
	SymRollLeft             uint16 = 0xAD
	SymRollLevel            uint16 = 0xAE
	SymRollRight            uint16 = 0xAF
	SymPitchUp              uint16 = 0xB0
	SymPitchDown            uint16 = 0xBB
	SymGforce               uint16 = 0xBC
	SymGforceX              uint16 = 0xBD
	SymGforceY              uint16 = 0xBE
	SymGforceZ              uint16 = 0xBF
	SymBaroTemp             uint16 = 0xC0
	SymImuTemp              uint16 = 0xC1
	SymTemp                 uint16 = 0xC2
	SymTempSensorFirst      uint16 = 0xC2
	SymEscTemp              uint16 = 0xC3
	SymTempSensorLast       uint16 = 0xC7
	SymHeadingN             uint16 = 0xC8
	SymHeadingS             uint16 = 0xC9
	SymHeadingE             uint16 = 0xCA
	SymHeadingW             uint16 = 0xCB
	SymHeadingDividedLine   uint16 = 0xCC
	SymHeadingLine          uint16 = 0xCD
	SymMax                  uint16 = 0xCE
	SymProfile              uint16 = 0xCF
	SymSwitchIndicatorLow   uint16 = 0xD0
	SymSwitchIndicatorMid   uint16 = 0xD1
	SymSwitchIndicatorHigh  uint16 = 0xD2
	SymAh                   uint16 = 0xD3
	SymGlideDist            uint16 = 0xD4
	SymGlideMins            uint16 = 0xD5
	SymAhVFt0               uint16 = 0xD6
	SymAhVFt1               uint16 = 0xD7
	SymAhVM0                uint16 = 0xD8
	SymAhVM1                uint16 = 0xD9
	SymFlightMinsRemaining  uint16 = 0xDA
	SymFlightHoursRemaining uint16 = 0xDB
	SymGroundCourse         uint16 = 0xDC
	SymAlert                uint16 = 0xDD
	SymMinGroundSpeed       uint16 = 0xDE
	SymTerrainFollowing     uint16 = 0xFB
	SymCrossTrackError      uint16 = 0xFC
	SymAdsb                 uint16 = 0xFD
	SymBlackbox             uint16 = 0xFE
	SymLogoStart            uint16 = 0x101
	SymAhLeft               uint16 = 0x12C
	SymAhRight              uint16 = 0x12D
	SymAhDecorationMin      uint16 = 0x12E
	SymAhDecoration         uint16 = 0x131
	SymAhDecorationMax      uint16 = 0x133
	SymAhChLeft             uint16 = 0x13A
	SymAhChRight            uint16 = 0x13B
	SymArrowUp              uint16 = 0x13C
	SymArrow2               uint16 = 0x13D
	SymArrow3               uint16 = 0x13E
	SymArrow4               uint16 = 0x13F
	SymArrowRight           uint16 = 0x140
	SymArrow6               uint16 = 0x141
	SymArrow7               uint16 = 0x142
	SymArrow8               uint16 = 0x143
	SymArrowDown            uint16 = 0x144
	SymArrow10              uint16 = 0x145
	SymArrow11              uint16 = 0x146
	SymArrow12              uint16 = 0x147
	SymArrowLeft            uint16 = 0x148
	SymArrow14              uint16 = 0x149
	SymArrow15              uint16 = 0x14A
	SymArrow16              uint16 = 0x14B
	SymAhHStart             uint16 = 0x14C
	SymAhVStart             uint16 = 0x15A
	SymVarioUp2a            uint16 = 0x155
	SymVarioUp1a            uint16 = 0x156
	SymVarioDown1a          uint16 = 0x157
	SymVarioDown2a          uint16 = 0x158
	SymAlt                  uint16 = 0x159
	SymHudSignal0           uint16 = 0x160
	SymHudSignal1           uint16 = 0x161
	SymHudSignal2           uint16 = 0x162
	SymHudSignal3           uint16 = 0x163
	SymHudSignal4           uint16 = 0x164
	SymHomeDist             uint16 = 0x165
	SymAhChCenter           uint16 = 0x166
	SymFlightDistRemaining  uint16 = 0x167
	SymOdometer             uint16 = 0x168
	SymRxBand               uint16 = 0x169
	SymRxMode               uint16 = 0x16A
	SymThrGaugeEmpty        uint16 = 0x16B
	SymThrGaugeHalf         uint16 = 0x16C
	SymThrGaugeFull         uint16 = 0x16D
	SymAhChType3            uint16 = 0x190
	SymAhChType4            uint16 = 0x193
	SymAhChType5            uint16 = 0x196
	SymAhChType6            uint16 = 0x199
	SymAhChType7            uint16 = 0x19C
	SymAhChType8            uint16 = 0x19F
	SymAhChAircraft0        uint16 = 0x1A2
	SymAhChAircraft1        uint16 = 0x1A3
	SymAhChAircraft2        uint16 = 0x1A4
	SymAhChAircraft3        uint16 = 0x1A5
	SymAhChAircraft4        uint16 = 0x1A6
	SymHudArrowsL1          uint16 = 0x1AE
	SymHudArrowsL2          uint16 = 0x1AF
	SymHudArrowsL3          uint16 = 0x1B0
	SymHudArrowsR1          uint16 = 0x1B1
	SymHudArrowsR2          uint16 = 0x1B2
	SymHudArrowsR3          uint16 = 0x1B3
	SymHudArrowsU1          uint16 = 0x1B4
	SymHudArrowsU2          uint16 = 0x1B5
	SymHudArrowsU3          uint16 = 0x1B6
	SymHudArrowsD1          uint16 = 0x1B7
	SymHudArrowsD2          uint16 = 0x1B8
	SymHudArrowsD3          uint16 = 0x1B9
	SymHudCardinal          uint16 = 0x1BA
	SymServoPanIsCentred    uint16 = 0x1C6
	SymServoPanIsOffsetL    uint16 = 0x1C7
	SymServoPanIsOffsetR    uint16 = 0x1C8
	SymPilotLogoSmlL        uint16 = 0x1D5
	SymPilotLogoSmlC        uint16 = 0x1D6
	SymPilotLogoSmlR        uint16 = 0x1D7
	SymPilotLogoLrgStart    uint16 = 0x1D8
)

// symbolNames maps code -> SYM_* name (without the prefix). When several names
// share the same code, the first one in the header wins.
var symbolNames = map[uint16]string{
	0x01:  "RSSI",
	0x02:  "LQ",
	0x03:  "LAT",
	0x04:  "LON",
	0x05:  "AZIMUTH",
	0x06:  "TELEMETRY_0",
	0x07:  "TELEMETRY_1",
	0x08:  "SAT_L",
	0x09:  "SAT_R",
	0x0A:  "HOME_NEAR",
	0x0B:  "DEGREES",
	0x0C:  "HEADING",
	0x0D:  "SCALE",
	0x0E:  "HDP_L",
	0x0F:  "HDP_R",
	0x101: "LOGO_START",
	0x10:  "HOME",
	0x11:  "2RSS",
	0x12C: "AH_LEFT",
	0x12D: "AH_RIGHT",
	0x12:  "DB",
	0x12E: "AH_DECORATION_MIN",
	0x131: "AH_DECORATION",
	0x133: "AH_DECORATION_MAX",
	0x13A: "AH_CH_LEFT",
	0x13B: "AH_CH_RIGHT",
	0x13C: "ARROW_UP",
	0x13D: "ARROW_2",
	0x13:  "DBM",
	0x13E: "ARROW_3",
	0x13F: "ARROW_4",
	0x140: "ARROW_RIGHT",
	0x141: "ARROW_6",
	0x142: "ARROW_7",
	0x143: "ARROW_8",
	0x144: "ARROW_DOWN",
	0x145: "ARROW_10",
	0x146: "ARROW_11",
	0x147: "ARROW_12",
	0x148: "ARROW_LEFT",
	0x149: "ARROW_14",
	0x14A: "ARROW_15",
	0x14B: "ARROW_16",
	0x14C: "AH_H_START",
	0x14:  "SNR",
	0x155: "VARIO_UP_2A",
	0x156: "VARIO_UP_1A",
	0x157: "VARIO_DOWN_1A",
	0x158: "VARIO_DOWN_2A",
	0x159: "ALT",
	0x15A: "AH_V_START",
	0x15:  "AH_DECORATION_UP",
	0x160: "HUD_SIGNAL_0",
	0x161: "HUD_SIGNAL_1",
	0x162: "HUD_SIGNAL_2",
	0x163: "HUD_SIGNAL_3",
	0x164: "HUD_SIGNAL_4",
	0x165: "HOME_DIST",
	0x166: "AH_CH_CENTER",
	0x167: "FLIGHT_DIST_REMAINING",
	0x168: "ODOMETER",
	0x169: "RX_BAND",
	0x16:  "AH_DECORATION_DOWN",
	0x16A: "RX_MODE",
	0x16B: "THR_GAUGE_EMPTY",
	0x16C: "THR_GAUGE_HALF",
	0x16D: "THR_GAUGE_FULL",
	0x17:  "DECORATION",
	0x190: "AH_CH_TYPE3",
	0x193: "AH_CH_TYPE4",
	0x196: "AH_CH_TYPE5",
	0x199: "AH_CH_TYPE6",
	0x19C: "AH_CH_TYPE7",
	0x19F: "AH_CH_TYPE8",
	0x1A2: "AH_CH_AIRCRAFT0",
	0x1A3: "AH_CH_AIRCRAFT1",
	0x1A4: "AH_CH_AIRCRAFT2",
	0x1A5: "AH_CH_AIRCRAFT3",
	0x1A6: "AH_CH_AIRCRAFT4",
	0x1AE: "HUD_ARROWS_L1",
	0x1AF: "HUD_ARROWS_L2",
	0x1B0: "HUD_ARROWS_L3",
	0x1B1: "HUD_ARROWS_R1",
	0x1B2: "HUD_ARROWS_R2",
	0x1B3: "HUD_ARROWS_R3",
	0x1B4: "HUD_ARROWS_U1",
	0x1B5: "HUD_ARROWS_U2",
	0x1B6: "HUD_ARROWS_U3",
	0x1B7: "HUD_ARROWS_D1",
	0x1B8: "HUD_ARROWS_D2",
	0x1B9: "HUD_ARROWS_D3",
	0x1BA: "HUD_CARDINAL",
	0x1C6: "SERVO_PAN_IS_CENTRED",
	0x1C7: "SERVO_PAN_IS_OFFSET_L",
	0x1C8: "SERVO_PAN_IS_OFFSET_R",
	0x1D5: "PILOT_LOGO_SML_L",
	0x1D6: "PILOT_LOGO_SML_C",
	0x1D7: "PILOT_LOGO_SML_R",
	0x1D8: "PILOT_LOGO_LRG_START",
	0x1F:  "VOLT",
	0x20:  "BLANK",
	0x22:  "AH_KM",
	0x24:  "AH_MI",
	0x27:  "VTX_POWER",
	0x3F:  "AH_NM",
	0x60:  "MAH_NM_0",
	0x61:  "MAH_NM_1",
	0x62:  "MILLIOHM",
	0x63:  "BATT_FULL",
	0x64:  "BATT_5",
	0x65:  "BATT_4",
	0x66:  "BATT_3",
	0x67:  "BATT_2",
	0x68:  "BATT_1",
	0x69:  "BATT_EMPTY",
	0x6A:  "AMP",
	0x6B:  "MAH_KM_0",
	0x6C:  "MAH_KM_1",
	0x6D:  "WH",
	0x6E:  "WH_KM",
	0x6F:  "WH_MI",
	0x70:  "WH_NM",
	0x71:  "WATT",
	0x72:  "MW",
	0x73:  "KILOWATT",
	0x74:  "FT",
	0x75:  "TRIP_DIST",
	0x76:  "ALT_M",
	0x77:  "ALT_KM",
	0x78:  "ALT_FT",
	0x79:  "ALT_KFT",
	0x7A:  "DIST_M",
	0x7E:  "DIST_KM",
	0x7F:  "DIST_FT",
	0x80:  "DIST_MI",
	0x81:  "DIST_NM",
	0x82:  "M",
	0x83:  "KM",
	0x84:  "MI",
	0x85:  "NM",
	0x86:  "WIND_HORIZONTAL",
	0x87:  "WIND_VERTICAL",
	0x88:  "3D_KMH",
	0x89:  "3D_MPH",
	0x8A:  "3D_KT",
	0x8B:  "RPM",
	0x8C:  "AIR",
	0x8D:  "FTS",
	0x8E:  "100FTM",
	0x8F:  "MS",
	0x90:  "KMH",
	0x91:  "MPH",
	0x92:  "KT",
	0x93:  "MAH_MI_0",
	0x94:  "MAH_MI_1",
	0x95:  "THR",
	0x96:  "TEMP_F",
	0x97:  "TEMP_C",
	0x99:  "MAH",
	0x9A:  "ON_H",
	0x9B:  "FLY_H",
	0x9C:  "GLIDESLOPE",
	0x9D:  "WAYPOINT",
	0x9E:  "ON_M",
	0x9F:  "FLY_M",
	0xA0:  "CLOCK",
	0xA1:  "ZERO_HALF_TRAILING_DOT",
	0xAB:  "AUTO_THR0",
	0xAC:  "AUTO_THR1",
	0xAD:  "ROLL_LEFT",
	0xAE:  "ROLL_LEVEL",
	0xAF:  "ROLL_RIGHT",
	0xB0:  "PITCH_UP",
	0xB1:  "ZERO_HALF_LEADING_DOT",
	0xBB:  "PITCH_DOWN",
	0xBC:  "GFORCE",
	0xBD:  "GFORCE_X",
	0xBE:  "GFORCE_Y",
	0xBF:  "GFORCE_Z",
	0xC0:  "BARO_TEMP",
	0xC1:  "IMU_TEMP",
	0xC2:  "TEMP",
	0xC3:  "ESC_TEMP",
	0xC7:  "TEMP_SENSOR_LAST",
	0xC8:  "HEADING_N",
	0xC9:  "HEADING_S",
	0xCA:  "HEADING_E",
	0xCB:  "HEADING_W",
	0xCC:  "HEADING_DIVIDED_LINE",
	0xCD:  "HEADING_LINE",
	0xCE:  "MAX",
	0xCF:  "PROFILE",
	0xD0:  "SWITCH_INDICATOR_LOW",
	0xD1:  "SWITCH_INDICATOR_MID",
	0xD2:  "SWITCH_INDICATOR_HIGH",
	0xD3:  "AH",
	0xD4:  "GLIDE_DIST",
	0xD5:  "GLIDE_MINS",
	0xD6:  "AH_V_FT_0",
	0xD7:  "AH_V_FT_1",
	0xD8:  "AH_V_M_0",
	0xD9:  "AH_V_M_1",
	0xDA:  "FLIGHT_MINS_REMAINING",
	0xDB:  "FLIGHT_HOURS_REMAINING",
	0xDC:  "GROUND_COURSE",
	0xDD:  "ALERT",
	0xDE:  "MIN_GROUND_SPEED",
	0xFB:  "TERRAIN_FOLLOWING",
	0xFC:  "CROSS_TRACK_ERROR",
	0xFD:  "ADSB",
	0xFE:  "BLACKBOX",
}
