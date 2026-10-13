// Package export builds a table from decoded OSD telemetry and the matching
// SRT data, and writes it as CSV.
package export

import (
	"math"
	"strconv"
	"strings"

	"github.com/ericogr/fpv-osd-decoder/internal/srt"
	"github.com/ericogr/fpv-osd-decoder/internal/telemetry"
)

// timeColumn is the name of the first CSV column: the OSD frame timestamp.
const timeColumn = "t_ms"

// Source says where a column's values come from.
type Source int

const (
	FromOSD Source = iota // decoded from the OSD screens
	FromSRT               // read from the .srt subtitles
)

// Column describes one telemetry column of a Table.
type Column struct {
	Key      string
	Desc     string
	Decimals int // decimal places to keep; srt.AllDigits keeps them all
	Text     bool
	Source   Source
	Frames   int // rows in which the column has a value
}

// Frame is one decoded OSD frame: its timestamp and what was read from it.
type Frame struct {
	TimeMs uint32
	Sample telemetry.Sample
}

// Table holds the rows of a flight. Only columns that have a value in at
// least one row are kept.
type Table struct {
	Columns []Column
	Times   []uint32   // Times[i] is the timestamp of row i, in ms
	Rows    [][]string // Rows[i][j] is Columns[j] in row i; "" means absent
}

// Build lays the frames out as a table. sf may be nil; when present, each
// frame is joined to the SRT entry that covers its timestamp.
func Build(frames []Frame, sf *srt.File) *Table {
	samples := make([]telemetry.Sample, len(frames))
	for i, fr := range frames {
		samples[i] = fr.Sample
	}
	cols := candidateColumns(samples, sf)

	rows := make([][]string, len(frames))
	for i, fr := range frames {
		var entry *srt.Entry
		if sf != nil {
			entry = sf.At(int64(fr.TimeMs))
		}
		row := make([]string, len(cols))
		for j := range cols {
			row[j] = cell(&cols[j], fr.Sample, entry)
			if row[j] != "" {
				cols[j].Frames++
			}
		}
		rows[i] = row
	}

	t := &Table{Times: make([]uint32, len(frames)), Rows: make([][]string, len(frames))}
	for i, fr := range frames {
		t.Times[i] = fr.TimeMs
	}
	var used []int
	for j, c := range cols {
		if c.Frames > 0 {
			used = append(used, j)
			t.Columns = append(t.Columns, c)
		}
	}
	for i, row := range rows {
		t.Rows[i] = make([]string, len(used))
		for k, j := range used {
			t.Rows[i][k] = row[j]
		}
	}
	return t
}

// candidateColumns lists every column that could appear: the OSD ones (the
// catalog and any reading the samples hold that no rule knows), then the ones
// present in the SRT (if any).
func candidateColumns(samples []telemetry.Sample, sf *srt.File) []Column {
	var cols []Column
	for _, f := range telemetry.FieldsFor(samples) {
		cols = append(cols, Column{Key: f.Key, Desc: f.Desc, Decimals: f.Decimals, Text: f.Text, Source: FromOSD})
	}
	if sf != nil {
		for _, f := range sf.Columns() {
			cols = append(cols, Column{Key: f.Key, Desc: f.Desc, Decimals: f.Decimals, Source: FromSRT})
		}
	}
	return cols
}

// cell formats the value of column c for this sample/entry ("" = absent).
func cell(c *Column, s telemetry.Sample, e *srt.Entry) string {
	if c.Source == FromSRT {
		if e == nil {
			return ""
		}
		v, ok := e.Values[c.Key]
		if !ok {
			return ""
		}
		return formatNum(v, c.Decimals)
	}
	v, ok := s[c.Key]
	if !ok {
		return ""
	}
	if c.Text {
		return v.Text
	}
	return formatNum(v.Num, c.Decimals)
}

// formatNum rounds to the given decimal places and strips trailing zeros; a
// negative number of decimals does not round.
func formatNum(v float64, decimals int) string {
	if decimals < 0 {
		return strconv.FormatFloat(v, 'f', -1, 64)
	}
	s := strconv.FormatFloat(v, 'f', decimals, 64)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	}
	if s == "-0" || math.IsNaN(v) {
		return "0"
	}
	return s
}
