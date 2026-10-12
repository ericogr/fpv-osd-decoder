package export

import (
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/ericogr/fpv-osd-decoder/internal/telemetry"
)

// Profile says what the exported CSV contains. The zero value is the default.
type Profile int

const (
	// ProfileDefault exports every column and every frame.
	ProfileDefault Profile = iota
	// ProfileMaps exports the flight path and the GPS positions, with a few
	// flight attributes on each one, as a KML file for Google My Maps.
	ProfileMaps
)

var profileNames = [...]string{ProfileDefault: "default", ProfileMaps: "maps"}

// Google My Maps imports at most 2000 rows from a spreadsheet; the KML keeps
// the same limit for its points. The two other constants name the position
// columns of the maps table.
const (
	mapsMaxRows   = 2000
	mapsLatColumn = "latitude"
	mapsLonColumn = "longitude"
)

// mapsAttributes are the columns that travel with each position, shown in the
// description of the point. My Maps does not use them to place anything. Only
// the ones the recording has are written, and the names are the same as in the
// default profile.
var mapsAttributes = []string{
	telemetry.KeyFlightMode,
	telemetry.KeyAltitude,
	"osd_gps_speed_kmh",
	"osd_air_speed_kmh",
	"osd_heading_deg",
	"osd_home_dist_m",
	"osd_gps_sats",
	"osd_rssi_dbm",
	"osd_lq_pct",
	"srt_signal",
	"osd_batt_cell_v",
	"osd_batt_v",
	"osd_current_a",
}

var errNoGPS = errors.New("the recording has no valid GPS position")

// String returns the name used on the command line.
func (p Profile) String() string { return profileNames[p] }

// ParseProfile finds a profile by its name.
func ParseProfile(name string) (Profile, error) {
	for p, n := range profileNames {
		if n == name {
			return Profile(p), nil
		}
	}
	return 0, fmt.Errorf("unknown profile %q (use %s)", name, strings.Join(profileNames[:], " or "))
}

// UsesCSV reports whether the profile writes a CSV, and so whether the
// column and decimal separators matter.
func (p Profile) UsesCSV() bool { return p != ProfileMaps }

// Ext is the extension of the file the profile writes.
func (p Profile) Ext() string {
	if p == ProfileMaps {
		return ".kml"
	}
	return ".csv"
}

// Write writes the table, already reshaped by Apply, in the file format of the
// profile. format only matters for CSV; title names the flight in a KML.
func (p Profile) Write(w io.Writer, t *Table, format Format, title string) error {
	if p == ProfileMaps {
		return t.WriteKML(w, title)
	}
	return t.WriteCSV(w, format)
}

// Reduction counts the rows a profile left out of the table, by reason.
type Reduction struct {
	NoGPS    int // no valid position
	Repeated int // same position as the row before it
	Thinned  int // dropped, evenly spaced, to fit the row limit
}

// Total is the number of rows left out.
func (r Reduction) Total() int { return r.NoGPS + r.Repeated + r.Thinned }

// Apply reshapes the table for the profile. It returns t itself when the
// profile keeps everything.
func (p Profile) Apply(t *Table) (*Table, Reduction, error) {
	if p == ProfileMaps {
		return mapsTable(t)
	}
	return t, Reduction{}, nil
}

// mapsTable keeps the time, the position and the attributes of the rows with
// a valid GPS fix. A position equal to the one before it adds nothing to a
// map, so it is dropped; if more than mapsMaxRows rows are left, they are
// thinned out evenly (the first and the last always stay).
func mapsTable(t *Table) (*Table, Reduction, error) {
	var red Reduction
	lat, lon := t.columnIndex(telemetry.KeyLat), t.columnIndex(telemetry.KeyLon)
	if lat < 0 || lon < 0 {
		return nil, red, errNoGPS
	}

	var kept []int // indexes of the rows to write
	for i, row := range t.Rows {
		switch {
		case !validPosition(row[lat], row[lon]):
			red.NoGPS++
		case len(kept) > 0 && samePosition(t.Rows[kept[len(kept)-1]], row, lat, lon):
			red.Repeated++
		default:
			kept = append(kept, i)
		}
	}
	if len(kept) == 0 {
		return nil, red, errNoGPS
	}
	if len(kept) > mapsMaxRows {
		picked := make([]int, mapsMaxRows)
		for k := range picked {
			picked[k] = kept[k*(len(kept)-1)/(mapsMaxRows-1)]
		}
		red.Thinned = len(kept) - len(picked)
		kept = picked
	}

	// the position first, then the attributes that have a value in some kept row
	src := []int{lat, lon}
	for _, key := range mapsAttributes {
		if i := t.columnIndex(key); i >= 0 && countValues(t, i, kept) > 0 {
			src = append(src, i)
		}
	}
	out := &Table{
		Columns: make([]Column, len(src)),
		Times:   make([]uint32, len(kept)),
		Rows:    make([][]string, len(kept)),
	}
	for c, j := range src {
		out.Columns[c] = t.Columns[j]
		out.Columns[c].Frames = countValues(t, j, kept)
	}
	out.Columns[0].Key, out.Columns[1].Key = mapsLatColumn, mapsLonColumn
	for k, i := range kept {
		out.Times[k] = t.Times[i]
		out.Rows[k] = make([]string, len(src))
		for c, j := range src {
			out.Rows[k][c] = t.Rows[i][j]
		}
	}
	return out, red, nil
}

// countValues is the number of the given rows in which column col has a value.
func countValues(t *Table, col int, rows []int) int {
	n := 0
	for _, i := range rows {
		if t.Rows[i][col] != "" {
			n++
		}
	}
	return n
}

// columnIndex returns the index of the column with this key, or -1.
func (t *Table) columnIndex(key string) int {
	for i, c := range t.Columns {
		if c.Key == key {
			return i
		}
	}
	return -1
}

// validPosition reports whether lat and lon (as written in the table) are a
// real position: both present, in range, and not the 0,0 a receiver without a
// fix reports.
func validPosition(lat, lon string) bool {
	la, errLat := strconv.ParseFloat(lat, 64)
	lo, errLon := strconv.ParseFloat(lon, 64)
	if errLat != nil || errLon != nil {
		return false
	}
	return math.Abs(la) <= 90 && math.Abs(lo) <= 180 && (la != 0 || lo != 0)
}

func samePosition(a, b []string, lat, lon int) bool {
	return a[lat] == b[lat] && a[lon] == b[lon]
}
