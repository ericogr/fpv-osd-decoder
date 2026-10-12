package srt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sample = "\xef\xbb\xbf1\r\n00:00:00,000 --> 00:00:00,150\r\nSignal:4 CH:5 FlightTime:0 SBat:7.8V GBat:11.9V Delay:32ms Bitrate:25.0Mbps Distance:49m\r\n\r\n" +
	"2\r\n00:00:00,150 --> 00:00:00,300\r\nSignal:3 CH:5 FlightTime:1 SBat:7.7V GBat:11.9V Delay:33ms Bitrate:12.5Mbps Distance:50m Foo:2.5dB\r\n\r\n"

func load(t *testing.T) *File {
	t.Helper()
	f, err := Parse(strings.NewReader(sample))
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func TestParse(t *testing.T) {
	f := load(t)
	if len(f.Entries) != 2 {
		t.Fatalf("%d entries, want 2", len(f.Entries))
	}
	e := f.Entries[1]
	if e.StartMs != 150 || e.EndMs != 300 {
		t.Errorf("interval %d-%d, want 150-300", e.StartMs, e.EndMs)
	}
	want := map[string]float64{
		"srt_signal": 3, "srt_channel": 5, "srt_flight_time_s": 1, "srt_sky_batt_v": 7.7,
		"srt_ground_batt_v": 11.9, "srt_latency_ms": 33, "srt_bitrate_mbps": 12.5,
		"srt_link_dist_m": 50, "srt_foo_db": 2.5,
	}
	for k, w := range want {
		if got, ok := e.Values[k]; !ok || got != w {
			t.Errorf("%s = %v (present=%v), want %v", k, got, ok, w)
		}
	}
}

func TestColumns(t *testing.T) {
	cols := load(t).Columns()
	var keys []string
	for _, c := range cols {
		keys = append(keys, c.Key)
	}
	want := []string{"srt_signal", "srt_channel", "srt_flight_time_s", "srt_sky_batt_v", "srt_ground_batt_v",
		"srt_latency_ms", "srt_bitrate_mbps", "srt_link_dist_m", "srt_foo_db"}
	if len(keys) != len(want) {
		t.Fatalf("columns = %v, want %v", keys, want)
	}
	for i := range want {
		if keys[i] != want[i] {
			t.Errorf("column %d = %s, want %s", i, keys[i], want[i])
		}
	}
	if last := cols[len(cols)-1]; last.Decimals != AllDigits {
		t.Errorf("unknown column should keep all digits, has Decimals = %d", last.Decimals)
	}
}

func TestAt(t *testing.T) {
	f := load(t)
	tests := []struct {
		t    int64
		want int64 // expected StartMs; -1 = none
	}{
		{0, 0}, {149, 0}, {150, 150}, {299, 150}, {300, 150}, {301, -1}, {-5, -1},
	}
	for _, tt := range tests {
		e := f.At(tt.t)
		switch {
		case tt.want < 0 && e != nil:
			t.Errorf("At(%d) = %d, want none", tt.t, e.StartMs)
		case tt.want >= 0 && (e == nil || e.StartMs != tt.want):
			t.Errorf("At(%d) = %v, want start %d", tt.t, e, tt.want)
		}
	}
}

func TestOpen(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x.srt")
	if err := os.WriteFile(p, []byte(sample), 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := Open(p)
	if err != nil || len(f.Entries) != 2 {
		t.Fatalf("Open = %v entries, %v", f, err)
	}
	if _, err := Open(filepath.Join(t.TempDir(), "missing.srt")); err == nil {
		t.Error("Open of a missing file should fail")
	}
}

func TestParseIgnoresGarbage(t *testing.T) {
	f, err := Parse(strings.NewReader("not a subtitle\n\n7\nno time line here\nSignal:4\n"))
	if err != nil || len(f.Entries) != 0 {
		t.Errorf("Parse = %+v, %v; want no entries", f, err)
	}
}
