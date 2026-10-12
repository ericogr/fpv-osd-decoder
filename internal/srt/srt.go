// Package srt reads the .srt subtitles recorded by Walksnail goggles, which
// carry video link data (signal, channel, latency, bitrate...) every ~150 ms.
package srt

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// AllDigits is the Field.Decimals value that means "do not round".
const AllDigits = -1

// Field describes a column that comes out of the SRT.
type Field struct {
	Key      string // CSV column name
	Desc     string
	Decimals int // decimal places shown in the file; AllDigits keeps them all
}

// knownField ties a field name from the .srt to the column it becomes.
type knownField struct {
	name string // lowercase name as it appears in the .srt
	Field
}

// known lists the fields the goggles are known to write, in the order their
// columns appear in the CSV. Any other field becomes a generic
// "srt_<name>[_<unit>]" column.
var known = []knownField{
	{"signal", Field{Key: "srt_signal", Desc: "video signal strength (0 to 4 bars)"}},
	{"ch", Field{Key: "srt_channel", Desc: "video channel"}},
	{"flighttime", Field{Key: "srt_flight_time_s", Desc: "flight time counted by the video system (s)"}},
	{"sbat", Field{Key: "srt_sky_batt_v", Desc: "voltage of the aircraft battery (SBat, the \"sky\" side) (V)", Decimals: 1}},
	{"gbat", Field{Key: "srt_ground_batt_v", Desc: "voltage of the battery that powers the ground video receiver (GBat) (V)", Decimals: 1}},
	{"delay", Field{Key: "srt_latency_ms", Desc: "video latency (ms)"}},
	{"bitrate", Field{Key: "srt_bitrate_mbps", Desc: "video bitrate (Mbps)", Decimals: 1}},
	{"distance", Field{Key: "srt_link_dist_m", Desc: "distance between the video receiver and the aircraft (m)"}},
}

var knownByName = func() map[string]Field {
	m := make(map[string]Field, len(known))
	for _, k := range known {
		m[k.name] = k.Field
	}
	return m
}()

// Entry is one subtitle: a time interval and its values.
type Entry struct {
	StartMs, EndMs int64
	Values         map[string]float64 // column key -> value
}

// File is a loaded .srt, sorted by time.
type File struct {
	Entries []Entry
}

var (
	timeLine = regexp.MustCompile(`^(\d+):(\d+):(\d+)[,.](\d+)\s*-->\s*(\d+):(\d+):(\d+)[,.](\d+)`)
	pair     = regexp.MustCompile(`([A-Za-z][A-Za-z0-9]*):\s*(-?[0-9]+(?:\.[0-9]+)?)([A-Za-z]*)`)
)

// parseMillis converts the digit groups of a timestamp to milliseconds. The
// groups are guaranteed to be digits by the timeLine expression, so conversion
// errors cannot happen; the fraction is normalized to exactly three digits.
func parseMillis(h, m, s, frac string) int64 {
	hh, _ := strconv.Atoi(h)
	mm, _ := strconv.Atoi(m)
	ss, _ := strconv.Atoi(s)
	ff, _ := strconv.Atoi((frac + "000")[:3])
	return int64(((hh*60+mm)*60+ss)*1000 + ff)
}

// Open reads an .srt file.
func Open(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(bytes.NewReader(data))
}

// Parse reads .srt subtitles. Lines it does not understand are ignored.
func Parse(r io.Reader) (*File, error) {
	f := &File{}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	var cur *Entry
	for sc.Scan() {
		line := strings.TrimSpace(strings.TrimPrefix(sc.Text(), "\xef\xbb\xbf"))
		if m := timeLine.FindStringSubmatch(line); m != nil {
			f.Entries = append(f.Entries, Entry{
				StartMs: parseMillis(m[1], m[2], m[3], m[4]),
				EndMs:   parseMillis(m[5], m[6], m[7], m[8]),
				Values:  map[string]float64{},
			})
			cur = &f.Entries[len(f.Entries)-1]
			continue
		}
		if line == "" {
			cur = nil
			continue
		}
		if cur == nil {
			continue // sequence number
		}
		for _, p := range pair.FindAllStringSubmatch(line, -1) {
			v, err := strconv.ParseFloat(p[2], 64)
			if err != nil {
				continue
			}
			cur.Values[columnKey(p[1], p[3])] = v
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("srt: %w", err)
	}
	sort.SliceStable(f.Entries, func(i, j int) bool { return f.Entries[i].StartMs < f.Entries[j].StartMs })
	return f, nil
}

// columnKey gives the column name for a "Name:value<unit>" field.
func columnKey(name, unit string) string {
	lower := strings.ToLower(name)
	if f, ok := knownByName[lower]; ok {
		return f.Key
	}
	if unit != "" {
		return "srt_" + lower + "_" + strings.ToLower(unit)
	}
	return "srt_" + lower
}

// Columns lists the columns present in the file: the known ones in canonical
// order, then the unknown ones alphabetically (with Decimals = AllDigits).
func (f *File) Columns() []Field {
	seen := map[string]bool{}
	for i := range f.Entries {
		for k := range f.Entries[i].Values {
			seen[k] = true
		}
	}
	var cols []Field
	for _, k := range known {
		if seen[k.Key] {
			cols = append(cols, k.Field)
			delete(seen, k.Key)
		}
	}
	var rest []string
	for k := range seen {
		rest = append(rest, k)
	}
	sort.Strings(rest)
	for _, k := range rest {
		cols = append(cols, Field{Key: k, Desc: "SRT field with no known description", Decimals: AllDigits})
	}
	return cols
}

// At returns the entry covering time t (ms), or nil if there is none.
func (f *File) At(t int64) *Entry {
	i := sort.Search(len(f.Entries), func(i int) bool { return f.Entries[i].StartMs > t }) - 1
	if i < 0 || t > f.Entries[i].EndMs {
		return nil
	}
	return &f.Entries[i]
}
