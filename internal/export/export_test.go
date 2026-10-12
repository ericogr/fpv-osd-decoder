package export

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/ericogr/fpv-osd-decoder/internal/srt"
	"github.com/ericogr/fpv-osd-decoder/internal/telemetry"
)

func TestFormatNum(t *testing.T) {
	tests := []struct {
		v        float64
		decimals int
		want     string
	}{
		{29, 0, "29"},
		{29, 1, "29"},
		{3.69, 2, "3.69"},
		{3.6, 2, "3.6"},
		{7.2, 1, "7.2"},
		{16.09344, 1, "16.1"},
		{-23.311583, 7, "-23.311583"},
		{-0.0001, 1, "0"},
		{25.04, srt.AllDigits, "25.04"},
		{1410, 1, "1410"},
	}
	for _, tt := range tests {
		if got := formatNum(tt.v, tt.decimals); got != tt.want {
			t.Errorf("formatNum(%v, %d) = %q, want %q", tt.v, tt.decimals, got, tt.want)
		}
	}
}

func TestNewFormat(t *testing.T) {
	ok := []struct {
		sep, decimal string
		want         Format
	}{
		{",", ".", Format{',', "."}},
		{";", ",", Format{';', ","}},
		{";", ".", Format{';', "."}},
	}
	for _, tt := range ok {
		got, err := NewFormat(tt.sep, tt.decimal)
		if err != nil || got != tt.want {
			t.Errorf("NewFormat(%q, %q) = %v, %v; want %v", tt.sep, tt.decimal, got, err, tt.want)
		}
	}
	bad := []struct{ sep, decimal, want string }{
		{",", ",", "cannot be the same"},
		{"", ".", "single character"},
		{";;", ".", "single character"},
		{"\"", ".", "single character"},
		{"\n", ".", "single character"},
		{";", "x", "decimal separator"},
		{";", "", "decimal separator"},
	}
	for _, tt := range bad {
		_, err := NewFormat(tt.sep, tt.decimal)
		if err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("NewFormat(%q, %q) error = %v, want it to contain %q", tt.sep, tt.decimal, err, tt.want)
		}
	}
}

const subtitles = `1
00:00:00,000 --> 00:00:00,150
Signal:4 Delay:32ms Foo:1.5dB

2
00:00:00,150 --> 00:00:00,300
Signal:3 Delay:33ms Foo:2.5dB
`

func flight(t *testing.T) *Table {
	t.Helper()
	sf, err := srt.Parse(strings.NewReader(subtitles))
	if err != nil {
		t.Fatal(err)
	}
	frames := []Frame{
		{0, telemetry.Sample{"osd_batt_cell_v": {Num: 3.6}, "osd_flight_mode": {Text: "ACRO"}}},
		{150, telemetry.Sample{"osd_batt_cell_v": {Num: 3.69}}},
		{400, telemetry.Sample{}}, // past the last subtitle, nothing readable
	}
	return Build(frames, sf)
}

func TestBuildKeepsOnlyColumnsWithValues(t *testing.T) {
	tab := flight(t)
	var keys []string
	for _, c := range tab.Columns {
		keys = append(keys, c.Key)
	}
	want := []string{"osd_batt_cell_v", "osd_flight_mode", "srt_signal", "srt_latency_ms", "srt_foo_db"}
	if strings.Join(keys, " ") != strings.Join(want, " ") {
		t.Fatalf("columns = %v, want %v", keys, want)
	}
	// 2 of the 3 frames have a battery reading; only the first has a mode
	wantFrames := map[string]int{"osd_batt_cell_v": 2, "osd_flight_mode": 1, "srt_signal": 2}
	for _, c := range tab.Columns {
		if w, ok := wantFrames[c.Key]; ok && c.Frames != w {
			t.Errorf("%s.Frames = %d, want %d", c.Key, c.Frames, w)
		}
	}
	if tab.Columns[0].Source != FromOSD || tab.Columns[2].Source != FromSRT {
		t.Errorf("sources = %v, %v", tab.Columns[0].Source, tab.Columns[2].Source)
	}
}

func TestBuildJoinsByTimeAndLeavesGapsEmpty(t *testing.T) {
	tab := flight(t)
	want := [][]string{
		{"3.6", "ACRO", "4", "32", "1.5"},
		{"3.69", "", "3", "33", "2.5"},
		{"", "", "", "", ""}, // t=400 ms: no subtitle covers it and nothing was read
	}
	for i, w := range want {
		if strings.Join(tab.Rows[i], "|") != strings.Join(w, "|") {
			t.Errorf("row %d = %q, want %q", i, tab.Rows[i], w)
		}
	}
	if tab.Times[1] != 150 || tab.Times[2] != 400 {
		t.Errorf("times = %v", tab.Times)
	}
}

func TestBuildWithoutSRT(t *testing.T) {
	tab := Build([]Frame{{0, telemetry.Sample{"osd_gps_sats": {Num: 12}}}}, nil)
	if len(tab.Columns) != 1 || tab.Columns[0].Key != "osd_gps_sats" || tab.Rows[0][0] != "12" {
		t.Errorf("table = %+v", tab)
	}
}

func TestWriteCSV(t *testing.T) {
	tab := flight(t)
	tests := []struct {
		name   string
		format Format
		want   string
	}{
		{"zero value is the default", Format{}, "t_ms,osd_batt_cell_v,osd_flight_mode,srt_signal,srt_latency_ms,srt_foo_db\n" +
			"0,3.6,ACRO,4,32,1.5\n150,3.69,,3,33,2.5\n400,,,,,\n"},
		{"pt-BR", Format{Sep: ';', Decimal: ","}, "t_ms;osd_batt_cell_v;osd_flight_mode;srt_signal;srt_latency_ms;srt_foo_db\n" +
			"0;3,6;ACRO;4;32;1,5\n150;3,69;;3;33;2,5\n400;;;;;\n"},
	}
	for _, tt := range tests {
		var buf bytes.Buffer
		if err := tab.WriteCSV(&buf, tt.format); err != nil {
			t.Fatalf("%s: %v", tt.name, err)
		}
		if buf.String() != tt.want {
			t.Errorf("%s:\ngot  %q\nwant %q", tt.name, buf.String(), tt.want)
		}
	}
}

// The decimal separator must not touch text columns.
func TestWriteCSVLeavesTextAlone(t *testing.T) {
	tab := Build([]Frame{{0, telemetry.Sample{"osd_message": {Text: "RTH 1.5 KM"}}}}, nil)
	var buf bytes.Buffer
	if err := tab.WriteCSV(&buf, Format{Sep: ';', Decimal: ","}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "RTH 1.5 KM") {
		t.Errorf("text was altered: %q", buf.String())
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("disk full") }

func TestWriteCSVReportsWriteErrors(t *testing.T) {
	// the csv writer buffers, so a large table is needed to hit the writer
	frames := make([]Frame, 5000)
	for i := range frames {
		frames[i] = Frame{uint32(i), telemetry.Sample{"osd_gps_sats": {Num: 12}}}
	}
	if err := Build(frames, nil).WriteCSV(failingWriter{}, Format{}); err == nil {
		t.Error("expected the write error to be reported")
	}
}

func TestParseProfile(t *testing.T) {
	for name, want := range map[string]Profile{"default": ProfileDefault, "maps": ProfileMaps} {
		got, err := ParseProfile(name)
		if err != nil || got != want || got.String() != name {
			t.Errorf("ParseProfile(%q) = %v, %v; want %v", name, got, err, want)
		}
	}
	for _, name := range []string{"", "Maps", "google"} {
		if _, err := ParseProfile(name); err == nil || !strings.Contains(err.Error(), "default or maps") {
			t.Errorf("ParseProfile(%q) error = %v", name, err)
		}
	}
	if !ProfileDefault.UsesCSV() || ProfileDefault.Ext() != ".csv" {
		t.Error("the default profile writes a CSV")
	}
	if ProfileMaps.UsesCSV() || ProfileMaps.Ext() != ".kml" {
		t.Error("the maps profile writes a KML")
	}
}

func TestProfileWrite(t *testing.T) {
	tab := flight(t)
	var csvOut, kmlOut bytes.Buffer
	if err := ProfileDefault.Write(&csvOut, tab, Format{Sep: ';', Decimal: ","}, "x"); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(csvOut.String(), "t_ms;osd_batt_cell_v;") {
		t.Errorf("default profile wrote %q", csvOut.String())
	}
	maps, _, err := ProfileMaps.Apply(Build(gpsFrames("-23.1,-47.1"), nil))
	if err != nil {
		t.Fatal(err)
	}
	if err := ProfileMaps.Write(&kmlOut, maps, Format{}, "x"); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(kmlOut.String(), "<?xml") {
		t.Errorf("maps profile wrote %q", kmlOut.String())
	}
}

func TestDefaultProfileKeepsTheTable(t *testing.T) {
	tab := flight(t)
	got, red, err := ProfileDefault.Apply(tab)
	if err != nil || got != tab || red.Total() != 0 {
		t.Errorf("Apply = %p, %+v, %v; want the same table untouched", got, red, err)
	}
}

// gpsFrames builds frames from "lat,lon" strings; "" is a frame without GPS.
func gpsFrames(positions ...string) []Frame {
	frames := make([]Frame, len(positions))
	for i, p := range positions {
		frames[i] = Frame{TimeMs: uint32(i * 100), Sample: telemetry.Sample{"osd_gps_sats": {Num: 10}}}
		if p == "" {
			continue
		}
		la, lo, _ := strings.Cut(p, ",")
		frames[i].Sample[telemetry.KeyLat] = telemetry.Value{Num: mustFloat(la)}
		frames[i].Sample[telemetry.KeyLon] = telemetry.Value{Num: mustFloat(lo)}
	}
	return frames
}

func mustFloat(s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(err)
	}
	return v
}

func TestMapsProfileKeepsValidDistinctPositions(t *testing.T) {
	tab := Build(gpsFrames(
		"",            // no fix
		"-23.1,-47.1", // kept
		"-23.1,-47.1", // repeats the one before
		"",            // lost the fix
		"-23.1,-47.1", // repeats the last kept position
		"0,0",         // a receiver without a fix reports 0,0
		"95,-47.1",    // latitude out of range
		"-23.1,181",   // longitude out of range
		"-23.2,-47.2", // kept
		"-23.1,-47.1", // kept: it is not the position right before
	), nil)
	got, red, err := ProfileMaps.Apply(tab)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := got.WriteCSV(&buf, Format{}); err != nil {
		t.Fatal(err)
	}
	want := "t_ms,latitude,longitude,osd_gps_sats\n100,-23.1,-47.1,10\n800,-23.2,-47.2,10\n900,-23.1,-47.1,10\n"
	if buf.String() != want {
		t.Errorf("csv = %q, want %q", buf.String(), want)
	}
	if want := (Reduction{NoGPS: 5, Repeated: 2}); red != want {
		t.Errorf("reduction = %+v, want %+v", red, want)
	}
	for _, c := range got.Columns {
		if c.Frames != 3 {
			t.Errorf("%s.Frames = %d, want 3", c.Key, c.Frames)
		}
	}
}

func TestMapsProfileThinsToTheRowLimit(t *testing.T) {
	const n = 5001
	positions := make([]string, n)
	for i := range positions {
		positions[i] = fmt.Sprintf("-23.%07d,-47.1", i+1) // every position is distinct
	}
	got, red, err := ProfileMaps.Apply(Build(gpsFrames(positions...), nil))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Rows) != mapsMaxRows || len(got.Times) != mapsMaxRows {
		t.Fatalf("%d rows, %d times; want %d", len(got.Rows), len(got.Times), mapsMaxRows)
	}
	if red.Thinned != n-mapsMaxRows || red.Total() != n-mapsMaxRows {
		t.Errorf("reduction = %+v", red)
	}
	if got.Times[0] != 0 || got.Times[mapsMaxRows-1] != uint32((n-1)*100) {
		t.Errorf("first and last times = %d, %d; want the ends of the flight", got.Times[0], got.Times[mapsMaxRows-1])
	}
	for i := 1; i < len(got.Times); i++ {
		if got.Times[i] <= got.Times[i-1] {
			t.Fatalf("times not increasing at %d: %d after %d", i, got.Times[i], got.Times[i-1])
		}
	}
}

func TestMapsProfileKeepsExactlyTheLimit(t *testing.T) {
	positions := make([]string, mapsMaxRows)
	for i := range positions {
		positions[i] = fmt.Sprintf("-23.%07d,-47.1", i+1)
	}
	got, red, err := ProfileMaps.Apply(Build(gpsFrames(positions...), nil))
	if err != nil || len(got.Rows) != mapsMaxRows || red.Total() != 0 {
		t.Errorf("rows = %d, reduction = %+v, err = %v", len(got.Rows), red, err)
	}
}

func TestMapsProfileWithoutPositions(t *testing.T) {
	for name, frames := range map[string][]Frame{
		"no GPS column":      gpsFrames("", ""),
		"only invalid fixes": gpsFrames("0,0", "91,0"),
	} {
		if _, _, err := ProfileMaps.Apply(Build(frames, nil)); !errors.Is(err, errNoGPS) {
			t.Errorf("%s: error = %v, want %v", name, err, errNoGPS)
		}
	}
}

func TestMapsProfileCarriesTheAttributes(t *testing.T) {
	pos := func(la, lo float64) telemetry.Sample {
		return telemetry.Sample{telemetry.KeyLat: {Num: la}, telemetry.KeyLon: {Num: lo}}
	}
	with := func(s telemetry.Sample, extra telemetry.Sample) telemetry.Sample {
		for k, v := range extra {
			s[k] = v
		}
		return s
	}
	frames := []Frame{
		// before the fix: the battery is readable, but the row has no position
		{0, telemetry.Sample{"osd_batt_v": {Num: 16.5}}},
		{100, with(pos(-23.1, -47.1), telemetry.Sample{"osd_altitude_m": {Num: 12.5}, "osd_flight_mode": {Text: "ACRO"}, "osd_wind_speed_kmh": {Num: 7}})},
		{200, with(pos(-23.2, -47.2), telemetry.Sample{"osd_altitude_m": {Num: 30}, "osd_gps_speed_kmh": {Num: 55.5}})},
	}
	got, _, err := ProfileMaps.Apply(Build(frames, nil))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := got.WriteCSV(&buf, Format{Sep: ';', Decimal: ","}); err != nil {
		t.Fatal(err)
	}
	// the attributes follow the order of mapsAttributes; the battery (only on a
	// row without a position) and the wind (not an attribute) are left out
	want := "t_ms;latitude;longitude;osd_flight_mode;osd_altitude_m;osd_gps_speed_kmh\n" +
		"100;-23,1;-47,1;ACRO;12,5;\n200;-23,2;-47,2;;30;55,5\n"
	if buf.String() != want {
		t.Errorf("csv = %q, want %q", buf.String(), want)
	}
	frames2 := map[string]int{}
	for _, c := range got.Columns {
		frames2[c.Key] = c.Frames
	}
	if frames2["osd_flight_mode"] != 1 || frames2["osd_altitude_m"] != 2 || frames2["latitude"] != 2 {
		t.Errorf("Frames = %v", frames2)
	}
}

// Every attribute must be a column the program can produce; a typo would
// silently drop it from the maps CSV.
func TestMapsAttributesExist(t *testing.T) {
	sf, err := srt.Parse(strings.NewReader(subtitles))
	if err != nil {
		t.Fatal(err)
	}
	known := map[string]bool{}
	for _, c := range candidateColumns(sf) {
		known[c.Key] = true
	}
	seen := map[string]bool{}
	for _, key := range mapsAttributes {
		if !known[key] {
			t.Errorf("attribute %q is not a column of the program", key)
		}
		if seen[key] {
			t.Errorf("attribute %q is listed twice", key)
		}
		seen[key] = true
		if key == mapsLatColumn || key == mapsLonColumn || key == telemetry.KeyLat || key == telemetry.KeyLon {
			t.Errorf("attribute %q is the position", key)
		}
	}
}

// kmlDoc is the part of a KML the tests look at.
type kmlDoc struct {
	Name   string `xml:"Document>name"`
	Styles []struct {
		ID        string  `xml:"id,attr"`
		Icon      *string `xml:"IconStyle>Icon>href"`
		IconColor string  `xml:"IconStyle>color"`
		IconScale string  `xml:"IconStyle>scale"`
		LabelSize string  `xml:"LabelStyle>scale"`
		LineColor string  `xml:"LineStyle>color"`
	} `xml:"Document>Style"`
	Placemarks []kmlPlacemark `xml:"Document>Placemark"`
}

type kmlPlacemark struct {
	Name        string `xml:"name"`
	Description string `xml:"description"`
	Style       string `xml:"styleUrl"`
	Point       string `xml:"Point>coordinates"`
	PointAlt    string `xml:"Point>altitudeMode"`
	Line        string `xml:"LineString>coordinates"`
	LineAlt     string `xml:"LineString>altitudeMode"`
	Tessellate  string `xml:"LineString>tessellate"`
}

// lines and points split the placemarks of a KML by what they hold.
func (d kmlDoc) lines() (out []kmlPlacemark) {
	for _, p := range d.Placemarks {
		if p.Line != "" {
			out = append(out, p)
		}
	}
	return out
}

func (d kmlDoc) points() (out []kmlPlacemark) {
	for _, p := range d.Placemarks {
		if p.Point != "" {
			out = append(out, p)
		}
	}
	return out
}

// colorOf returns the colour of the style a placemark points to (#line3, #dot3).
func (d kmlDoc) colorOf(t *testing.T, p kmlPlacemark) string {
	t.Helper()
	for _, s := range d.Styles {
		if "#"+s.ID == p.Style {
			if s.LineColor != "" {
				return s.LineColor
			}
			return s.IconColor
		}
	}
	t.Fatalf("placemark %q uses the style %q, which is not defined", p.Name, p.Style)
	return ""
}

func writeKML(t *testing.T, frames []Frame, title string) kmlDoc {
	t.Helper()
	tab, _, err := ProfileMaps.Apply(Build(frames, nil))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := tab.WriteKML(&buf, title); err != nil {
		t.Fatal(err)
	}
	var doc kmlDoc
	if err := xml.Unmarshal(buf.Bytes(), &doc); err != nil {
		t.Fatalf("not well-formed XML: %v\n%s", err, buf.Bytes())
	}
	return doc
}

// kmlFrames builds a frame per spec, one second apart and at distinct
// positions. A spec is "MODE@ALTITUDE"; leave out the mode (or "@ALTITUDE") to
// have a frame without it.
func kmlFrames(specs ...string) []Frame {
	frames := make([]Frame, len(specs))
	for i, spec := range specs {
		mode, alt, hasAlt := strings.Cut(spec, "@")
		s := telemetry.Sample{
			telemetry.KeyLat: {Num: -23 - float64(i)/10},
			telemetry.KeyLon: {Num: -47},
		}
		if mode != "" {
			s[telemetry.KeyFlightMode] = telemetry.Value{Text: mode}
		}
		if hasAlt {
			s[telemetry.KeyAltitude] = telemetry.Value{Num: mustFloat(alt)}
		}
		frames[i] = Frame{TimeMs: uint32(i) * 1000, Sample: s}
	}
	return frames
}

func TestWriteKML(t *testing.T) {
	frames := []Frame{
		{0, telemetry.Sample{telemetry.KeyLat: {Num: -23.1}, telemetry.KeyLon: {Num: -47.1}}},
		{4027, telemetry.Sample{telemetry.KeyLat: {Num: -23.2}, telemetry.KeyLon: {Num: -47.2},
			telemetry.KeyAltitude: {Num: 12.5}, telemetry.KeyFlightMode: {Text: "A&B <C>"}}},
		{65000, telemetry.Sample{telemetry.KeyLat: {Num: -23.3}, telemetry.KeyLon: {Num: -47.3}}},
	}
	doc := writeKML(t, frames, "Flight <1>")

	if doc.Name != "Flight <1>" {
		t.Errorf("document name = %q", doc.Name)
	}
	lines, pts := doc.lines(), doc.points()
	if len(lines) != 1 || len(pts) != 3 || len(doc.Placemarks) != 4 {
		t.Fatalf("%d lines and %d points in %d placemarks, want 1 line and 3 points", len(lines), len(pts), len(doc.Placemarks))
	}
	// longitude comes first; the altitude that blinked out keeps the last one read,
	// and 0 before the first reading
	wantLine := "-47.1,-23.1,0 -47.2,-23.2,12.5 -47.3,-23.3,12.5"
	if lines[0].Line != wantLine || lines[0].Name != "Flight <1> A&B <C>" {
		t.Errorf("line = %+v, want %q named after the mode", lines[0], wantLine)
	}
	if lines[0].LineAlt != "relativeToGround" || lines[0].Tessellate != "" {
		t.Errorf("line altitude mode = %q, tessellate = %q", lines[0].LineAlt, lines[0].Tessellate)
	}
	for i, want := range []struct{ name, coord string }{
		{"0:00", "-47.1,-23.1,0"}, {"0:04", "-47.2,-23.2,12.5"}, {"1:05", "-47.3,-23.3,12.5"},
	} {
		if pts[i].Name != want.name || pts[i].Point != want.coord || pts[i].PointAlt != "relativeToGround" {
			t.Errorf("point %d = %+v, want %q at %q, relative to the ground", i, pts[i], want.name, want.coord)
		}
	}
	// the description lists the time and the columns the row has, as plain text
	if want := "t_ms: 4027\nosd_flight_mode: A&B <C>\nosd_altitude_m: 12.5"; pts[1].Description != want {
		t.Errorf("description = %q, want %q", pts[1].Description, want)
	}
	if want := "t_ms: 0"; pts[0].Description != want {
		t.Errorf("description = %q, want %q", pts[0].Description, want)
	}
}

func TestWriteKMLSplitsThePathByFlightMode(t *testing.T) {
	doc := writeKML(t, kmlFrames("ACRO@1", "ACRO@2", "CRUZ@3", "CRUZ@4", "ACRO@5"), "x")
	lines := doc.lines()
	want := []struct {
		name  string
		coord []string
		color string
	}{
		{"x ACRO", []string{"-47,-23,1", "-47,-23.1,2"}, "ff0080ff"},
		// a segment starts where the one before it ended: no gap in the path
		{"x CRUZ", []string{"-47,-23.1,2", "-47,-23.2,3", "-47,-23.3,4"}, "ff00c800"},
		{"x ACRO", []string{"-47,-23.3,4", "-47,-23.4,5"}, "ff0080ff"},
	}
	if len(lines) != len(want) {
		t.Fatalf("%d lines, want %d: %+v", len(lines), len(want), lines)
	}
	for i, w := range want {
		if lines[i].Name != w.name || lines[i].Line != strings.Join(w.coord, " ") {
			t.Errorf("line %d = %q %q, want %q %q", i, lines[i].Name, lines[i].Line, w.name, strings.Join(w.coord, " "))
		}
		if got := doc.colorOf(t, lines[i]); got != w.color {
			t.Errorf("line %d colour = %s, want %s", i, got, w.color)
		}
	}
	// each dot has the colour of the mode of its own row
	for i, wantColor := range []string{"ff0080ff", "ff0080ff", "ff00c800", "ff00c800", "ff0080ff"} {
		if got := doc.colorOf(t, doc.points()[i]); got != wantColor {
			t.Errorf("point %d colour = %s, want %s", i, got, wantColor)
		}
	}
}

func TestWriteKMLColorsOfTheModes(t *testing.T) {
	// the first rows come before any mode is read; unknown modes take the spare colours in turn
	doc := writeKML(t, kmlFrames("", "", "XYZ", "XYZ", "ABC", "ABC", "RTH", "RTH", "MANU", "MANU"), "x")
	got := map[string]string{}
	for _, l := range doc.lines() {
		got[l.Name] = doc.colorOf(t, l)
	}
	want := map[string]string{
		"x no mode": "ff808080", "x XYZ": "ff00ffff", "x ABC": "ffffff00", "x RTH": "ff0000ff", "x MANU": "ffff00ff",
	}
	for name, color := range want {
		if got[name] != color {
			t.Errorf("%s colour = %q, want %q (all: %v)", name, got[name], color, got)
		}
	}
}

// Without a flight mode or an altitude, the path is one line that hugs the
// terrain.
func TestWriteKMLWithoutFlightModeOrAltitude(t *testing.T) {
	doc := writeKML(t, kmlFrames("", "", ""), "x")
	lines := doc.lines()
	if len(lines) != 1 {
		t.Fatalf("%d lines, want 1", len(lines))
	}
	if l := lines[0]; l.Name != "x path" || l.Line != "-47,-23 -47,-23.1 -47,-23.2" || l.Tessellate != "1" || l.LineAlt != "" {
		t.Errorf("line = %+v", l)
	}
	if got := doc.colorOf(t, lines[0]); got != "ff0080ff" {
		t.Errorf("colour = %s, want the default", got)
	}
	for i, p := range doc.points() {
		if p.PointAlt != "" || strings.Count(p.Point, ",") != 1 {
			t.Errorf("point %d = %+v, want a position without altitude", i, p)
		}
	}
}

func TestWriteKMLWithOnePositionHasNoLine(t *testing.T) {
	doc := writeKML(t, kmlFrames("ACRO@1"), "x")
	if len(doc.lines()) != 0 || len(doc.points()) != 1 {
		t.Errorf("%d lines and %d points, want only the point", len(doc.lines()), len(doc.points()))
	}
}

// Without a style Google Earth puts a pushpin on every point: the points must
// use the small dot and the lines must have no icon.
func TestWriteKMLStyles(t *testing.T) {
	doc := writeKML(t, kmlFrames("ACRO@1", "ACRO@2", "CRUZ@3", "CRUZ@4"), "x")
	kinds := map[string]string{}
	for _, s := range doc.Styles {
		switch {
		case strings.HasPrefix(s.ID, "dot"):
			kinds[s.ID] = "dot"
			if s.Icon == nil || !strings.HasPrefix(*s.Icon, "https://") || s.IconScale != "0.5" || s.LabelSize != "0" {
				t.Errorf("style %s = %+v, want a small dot without a label", s.ID, s)
			}
		case strings.HasPrefix(s.ID, "line"):
			kinds[s.ID] = "line"
			if s.IconScale != "0" || s.LineColor == "" {
				t.Errorf("style %s = %+v, want a line without an icon", s.ID, s)
			}
		default:
			t.Errorf("unexpected style %q", s.ID)
		}
	}
	if len(kinds) != 4 {
		t.Errorf("styles = %v, want a line and a dot for each of the 2 modes", kinds)
	}
	for _, p := range doc.Placemarks {
		want := "dot"
		if p.Line != "" {
			want = "line"
		}
		if kinds[strings.TrimPrefix(p.Style, "#")] != want {
			t.Errorf("%q uses the style %q, want a %s style", p.Name, p.Style, want)
		}
	}
}

// My Maps drops folders, MultiGeometry, time elements and HTML balloons.
func TestWriteKMLAvoidsWhatMyMapsDrops(t *testing.T) {
	tab, _, err := ProfileMaps.Apply(Build(gpsFrames("-23.1,-47.1", "-23.2,-47.2"), nil))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := tab.WriteKML(&buf, "x"); err != nil {
		t.Fatal(err)
	}
	for _, tag := range []string{"<Folder", "<MultiGeometry", "<TimeStamp", "<TimeSpan", "<gx:", "<![CDATA["} {
		if strings.Contains(buf.String(), tag) {
			t.Errorf("KML contains %s", tag)
		}
	}
}

func TestWriteKMLNeedsThePositionColumns(t *testing.T) {
	if err := flight(t).WriteKML(&bytes.Buffer{}, "x"); err == nil || !strings.Contains(err.Error(), "latitude") {
		t.Errorf("error = %v", err)
	}
}

func TestWriteKMLReportsWriteErrors(t *testing.T) {
	tab, _, err := ProfileMaps.Apply(Build(gpsFrames("-23.1,-47.1"), nil))
	if err != nil {
		t.Fatal(err)
	}
	if err := tab.WriteKML(failingWriter{}, "x"); err == nil {
		t.Error("expected the write error to be reported")
	}
}

func TestClock(t *testing.T) {
	for ms, want := range map[uint32]string{0: "0:00", 999: "0:00", 4027: "0:04", 65000: "1:05", 989749: "16:29", 4500000: "75:00"} {
		if got := clock(ms); got != want {
			t.Errorf("clock(%d) = %q, want %q", ms, got, want)
		}
	}
}
