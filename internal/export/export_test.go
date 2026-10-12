package export

import (
	"bytes"
	"errors"
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
