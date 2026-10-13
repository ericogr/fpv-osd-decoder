package telemetry

import "github.com/ericogr/fpv-osd-decoder/internal/inav"

// The unit glyph of an element says what its number is measured in, and it
// changes with the magnitude (m and km, ft and mi) and with INAV's unit
// setting. These tables give the factor that takes the number to the unit of
// the column, which is always the metric one.

// Exact conversion factors (constant arithmetic is exact, so the values below
// are the decimals written here, rounded once). The nautical mile is written
// with a decimal point so that dividing it by 1000 is not an integer division.
const (
	metersPerFoot         = 0.3048
	metersPerMile         = 1609.344
	metersPerNauticalMile = 1852.0
	kmPerMile             = metersPerMile / 1000
	kmPerNauticalMile     = metersPerNauticalMile / 1000
	kmhPerMetersPerSecond = 3.6
	secondsPerMinute      = 60
)

var (
	distToMeters = map[uint16]float64{
		inav.SymDistM: 1, inav.SymDistKm: 1000, inav.SymDistFt: metersPerFoot,
		inav.SymDistMi: metersPerMile, inav.SymDistNm: metersPerNauticalMile,
	}
	altToMeters = map[uint16]float64{
		inav.SymAltM: 1, inav.SymAltKm: 1000, inav.SymAltFt: metersPerFoot, inav.SymAltKft: 1000 * metersPerFoot,
	}
	// The odometer has its own unit glyphs, different from the distance ones.
	odometerToKm = map[uint16]float64{
		inav.SymKm: 1, inav.SymMi: kmPerMile, inav.SymNm: kmPerNauticalMile,
	}

	// Horizontal speed units only: m/s and ft/s belong to the vertical speed
	// indicator.
	speedToKmh = map[uint16]float64{
		inav.SymKmh: 1, inav.SymMph: kmPerMile, inav.SymKt: kmPerNauticalMile,
	}
	speed3DToKmh = map[uint16]float64{
		inav.Sym3dKmh: 1, inav.Sym3dMph: kmPerMile, inav.Sym3dKt: kmPerNauticalMile,
	}
	// Wind speeds can also be shown in m/s (osd_estimations_wind_mps).
	windToKmh = map[uint16]float64{
		inav.SymKmh: 1, inav.SymMph: kmPerMile, inav.SymKt: kmPerNauticalMile, inav.SymMs: kmhPerMetersPerSecond,
	}
	// The vertical speed indicator; 100ft/min is the aviation unit.
	vSpeedToMS = map[uint16]float64{
		inav.SymMs: 1, inav.SymFts: metersPerFoot, inav.Sym100ftm: 100 * metersPerFoot / secondsPerMinute,
	}

	powerToW = map[uint16]float64{inav.SymWatt: 1, inav.SymKilowatt: 1000}
	// A value per mile or per nautical mile, as a value per km.
	energyPerDistToKm = map[uint16]float64{
		inav.SymWhKm: 1, inav.SymWhMi: 1 / kmPerMile, inav.SymWhNm: 1 / kmPerNauticalMile,
	}
)
