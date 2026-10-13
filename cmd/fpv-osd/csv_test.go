package main

import (
	"bytes"
	"encoding/csv"
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleFlight = "../../AvatarG0031.osd"

// readCSV runs "fpv-osd csv" on the sample flight with extra options and
// returns the parsed records.
func readCSV(t *testing.T, sep rune, args ...string) [][]string {
	t.Helper()
	if _, err := os.Stat(sampleFlight); err != nil {
		t.Skip("AvatarG0031.osd is not in the project")
	}
	out := filepath.Join(t.TempDir(), "flight.csv")
	cmd := append([]string{"csv", "-q", "-o", out}, args...)
	var stderr bytes.Buffer
	if code := runMain(append(cmd, sampleFlight), nil, &stderr); code != 0 {
		t.Fatalf("exit code = %d: %s", code, &stderr)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	r := csv.NewReader(bytes.NewReader(data))
	r.Comma = sep
	recs, err := r.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 6606+1 {
		t.Fatalf("%d lines, want 6607", len(recs))
	}
	return recs
}

func columns(header []string) map[string]int {
	idx := map[string]int{}
	for i, h := range header {
		idx[h] = i
	}
	return idx
}

// Runs the whole program on the sample flight, if it is in the project.
func TestCSVSampleFlight(t *testing.T) {
	recs := readCSV(t, ',')
	idx := columns(recs[0])
	for _, col := range []string{"t_ms", "osd_rssi_dbm", "osd_gps_lat", "osd_flight_mode", "srt_link_dist_m"} {
		if _, ok := idx[col]; !ok {
			t.Errorf("column %s missing from the header", col)
		}
	}
	// frame 3000, checked by hand against the screen
	row := recs[3001]
	want := map[string]string{
		"t_ms": "449678", "osd_rssi_dbm": "-72", "osd_batt_cell_v": "3.69",
		"osd_gps_lat": "-23.311583", "osd_gps_lon": "-47.378889",
		"osd_altitude_m": "213", "osd_home_dist_m": "1410", "osd_flight_mode": "CRUZ",
	}
	for col, w := range want {
		if got := row[idx[col]]; got != w {
			t.Errorf("%s = %q, want %q", col, got, w)
		}
	}
	// the splash screen has no flight telemetry: empty cells, not zero
	if got := recs[1][idx["osd_rssi_dbm"]]; got != "" {
		t.Errorf("osd_rssi_dbm in frame 0 = %q, want empty", got)
	}
}

// The frame right before the statistics shows the layout INAV uses when
// disarmed, which has elements the flight layout lacks: the rules built from
// the firmware source read them, with the values that are on screen.
func TestCSVSampleFlightDisarmedLayout(t *testing.T) {
	recs := readCSV(t, ',')
	idx := columns(recs[0])
	row := recs[6586] // frame 6585
	want := map[string]string{
		"t_ms": "986745", "osd_flight_time_s": "979", "osd_on_time_s": "1259",
		"osd_imu_temp_c": "40", "osd_baro_temp_c": "42", "osd_odometer_km": "10.3",
	}
	for col, w := range want {
		j, ok := idx[col]
		if !ok {
			t.Errorf("column %s missing from the header", col)
			continue
		}
		if row[j] != w {
			t.Errorf("%s = %q, want %q", col, row[j], w)
		}
	}
	// nothing the generic pass could take is in this flight, and the
	// statistics screen (it has a blackbox log number) is not read
	for col := range idx {
		if strings.HasPrefix(col, "osd_raw_") || col == "osd_blackbox_log" {
			t.Errorf("unexpected column %s", col)
		}
	}
	// the splash screen at the start does show the position
	if got := recs[1][idx["osd_gps_lat"]]; got != "-23.299915" {
		t.Errorf("osd_gps_lat in frame 0 = %q, want the position on the splash screen", got)
	}
}

// With ; and , the "3.6" from the Sheets example must not come out with a
// point, and the texts must not be altered.
func TestCSVBrazilianFormat(t *testing.T) {
	recs := readCSV(t, ';', "-sep", ";", "-decimal", ",")
	idx := columns(recs[0])
	row := recs[3001]
	want := map[string]string{
		"osd_batt_cell_v": "3,69", "osd_gps_lat": "-23,311583", "osd_wind_speed_kmh": "7,2",
		"srt_ground_batt_v": "11,8", "osd_flight_mode": "CRUZ", "osd_rssi_dbm": "-72",
	}
	for col, w := range want {
		if got := row[idx[col]]; got != w {
			t.Errorf("%s = %q, want %q", col, got, w)
		}
	}
	for i, rec := range recs[1:] {
		for j, v := range rec {
			if strings.Contains(v, ".") && !strings.HasPrefix(recs[0][j], "osd_message") {
				t.Fatalf("line %d, %s = %q: decimal point left over", i+1, recs[0][j], v)
			}
		}
	}
}

// The maps profile writes a KML with the flight path and at most 2000 valid,
// distinct points.
func TestCSVMapsProfile(t *testing.T) {
	if _, err := os.Stat(sampleFlight); err != nil {
		t.Skip("AvatarG0031.osd is not in the project")
	}
	var stdout, stderr bytes.Buffer
	if code := runMain([]string{"csv", "-profile", "maps", "-o", "-", sampleFlight}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d: %s", code, &stderr)
	}
	// My Maps takes unzipped KML files of up to 5 MB
	if stdout.Len() > 5<<20 {
		t.Errorf("KML is %d bytes, over the 5 MB limit of My Maps", stdout.Len())
	}
	var doc struct {
		Name       string `xml:"Document>name"`
		Placemarks []struct {
			Name        string `xml:"name"`
			Description string `xml:"description"`
			Point       string `xml:"Point>coordinates"`
			Line        string `xml:"LineString>coordinates"`
			LineAlt     string `xml:"LineString>altitudeMode"`
		} `xml:"Document>Placemark"`
	}
	if err := xml.Unmarshal(stdout.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Name != "AvatarG0031" {
		t.Errorf("document name = %q", doc.Name)
	}
	var pts []string          // coordinates of the points
	var descriptions []string // and what they say
	var names []string        // names of the lines, in order
	var path []string         // positions of the lines, joined
	for _, p := range doc.Placemarks {
		switch {
		case p.Point != "":
			pts = append(pts, p.Point)
			descriptions = append(descriptions, p.Description)
		case p.Line != "":
			names = append(names, p.Name)
			path = append(path, strings.Fields(p.Line)...)
			if p.LineAlt != "relativeToGround" {
				t.Errorf("line %q altitude mode = %q", p.Name, p.LineAlt)
			}
		}
	}
	if len(pts) != 2000 {
		t.Fatalf("%d points, want 2000", len(pts))
	}
	for i := 1; i < len(pts); i++ {
		if pts[i] == pts[i-1] {
			t.Fatalf("point %d repeats the position of the one before it: %q", i, pts[i])
		}
	}
	// the path is one line per flight mode: the first and the last positions of the
	// flight stay, each cut shares its position with the segment before it, and the
	// modes of the flight (ACRO, ANGL, CRUZ, RTH, MANU) all appear
	if len(names) < 5 {
		t.Errorf("%d lines, want one per change of flight mode: %q", len(names), names)
	}
	if want := len(pts) + len(names) - 1; len(path) != want {
		t.Errorf("the lines have %d positions, want %d (the points plus a shared one per cut)", len(path), want)
	}
	if !strings.HasPrefix(path[0], "-47.373494,-23.299915,") || !strings.HasPrefix(pts[0], "-47.373494,-23.299915,") {
		t.Errorf("first position = %q / %q", path[0], pts[0])
	}
	for _, mode := range []string{"ACRO", "ANGL", "CRUZ", "RTH", "MANU"} {
		found := false
		for _, n := range names {
			found = found || n == "AvatarG0031 "+mode
		}
		if !found {
			t.Errorf("no line for the mode %s: %q", mode, names)
		}
	}
	for _, want := range []string{"t_ms: 4027", "osd_flight_mode: ANGL", "osd_altitude_m:", "osd_gps_speed_kmh:", "srt_signal:"} {
		if !strings.Contains(descriptions[1], want) {
			t.Errorf("description of the second point lacks %q: %q", want, descriptions[1])
		}
	}
	for _, left := range []string{"osd_message", "osd_roll_deg", "osd_wind_speed_kmh", "srt_latency_ms"} {
		for _, d := range descriptions {
			if strings.Contains(d, left) {
				t.Fatalf("%s should not be in the descriptions: %q", left, d)
			}
		}
	}
	if !strings.Contains(stderr.String(), "rows written: 2000 of 6606 frames") {
		t.Errorf("summary = %q", &stderr)
	}
}

func TestParseRegion(t *testing.T) {
	r, err := parseRegion("-mode", "45, 16,4")
	if err != nil || r.X != 45 || r.Y != 16 || r.Width != 4 {
		t.Errorf("parseRegion = %+v, %v", r, err)
	}
	if got := r.String(); got != "45,16,4" {
		t.Errorf("String = %q, want 45,16,4", got)
	}
	for _, spec := range []string{"", "1,2", "1,2,3,4", "a,b,c", "1,2,3abc", "1,,3"} {
		if _, err := parseRegion("-mode", spec); err == nil || !strings.Contains(err.Error(), "-mode must be x,y,width") {
			t.Errorf("parseRegion(%q) error = %v", spec, err)
		}
	}
}
