package main

import (
	"bytes"
	"encoding/csv"
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
