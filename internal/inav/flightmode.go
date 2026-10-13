package inav

// FlightModeWidth is the width of the flight mode element (OSD_FLYMODE): INAV
// always writes four characters, padding the short names with blanks.
const FlightModeWidth = 4

// flightModes are the texts of the flight mode element, as INAV 9.1.0 writes
// them (osdDrawSingleElement in osd.c), blanks included.
var flightModes = map[string]bool{
	"ACRO": true, "ANGL": true, "ANGH": true, "HOR ": true, "MANU": true,
	"RTH ": true, "WRTH": true, "HOLD": true, "LOTR": true, "CRUZ": true,
	"CRSH": true, " WP ": true, " AH ": true, "GEO ": true, "TURT": true,
	"LAND": true, "!FS!": true,
}

// IsFlightMode reports whether text is exactly what the flight mode element
// can show: four characters, with the padding blanks.
func IsFlightMode(text string) bool { return flightModes[text] }
