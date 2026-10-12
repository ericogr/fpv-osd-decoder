package export

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/ericogr/fpv-osd-decoder/internal/telemetry"
)

const (
	// kmlDot is a plain circle from Google's own icon set, tinted with the
	// colour of the flight mode. Without a style, Google Earth draws a yellow
	// pushpin at every point.
	kmlDot = "https://maps.google.com/mapfiles/kml/shapes/placemark_circle.png"

	kmlPathColor   = "ff0080ff" // when the recording has no flight mode
	kmlNoModeColor = "ff808080" // before the first flight mode is read
)

// kmlModeColors are the colours of the flight modes seen in INAV recordings,
// in KML's aabbggrr order. Any other mode takes one of kmlSpareColors in turn.
var kmlModeColors = map[string]string{
	"ACRO": "ff0080ff", // orange
	"ANGL": "ffff8000", // blue
	"CRUZ": "ff00c800", // green
	"RTH":  "ff0000ff", // red
	"MANU": "ffff00ff", // magenta
}

var kmlSpareColors = []string{"ff00ffff", "ffffff00", "ffff0080", "ffffffff"} // yellow, cyan, violet, white

// kmlSegment is a stretch of the path flown in one flight mode: the rows from
// through to, both included.
type kmlSegment struct {
	mode     string
	from, to int
}

// WriteKML writes a maps table (see ProfileMaps) as a KML file for Google My
// Maps and Google Earth, the way flight log converters do: the path is a line
// per flight mode, each in its own colour, and the altitude, if the table has
// it, lifts the path off the ground (it is relative to the take-off point).
// On top of that, every row is a small dot with the time as its name and the
// other columns as plain text in its description. The dots have no label and
// the lines no icon, so a viewer shows the path instead of a pin on every
// position.
//
// My Maps leaves out parts of a KML (folders, MultiGeometry, time elements and
// HTML balloons), so the file is a flat list of placemarks with none of them.
func (t *Table) WriteKML(w io.Writer, title string) error {
	lat, lon := t.columnIndex(mapsLatColumn), t.columnIndex(mapsLonColumn)
	if lat < 0 || lon < 0 {
		return fmt.Errorf("kml needs the %s and %s columns", mapsLatColumn, mapsLonColumn)
	}
	modeCol, altCol := t.columnIndex(telemetry.KeyFlightMode), t.columnIndex(telemetry.KeyAltitude)

	// a value that blinked out of the OSD keeps the last one read
	modes := carryForward(t, modeCol, "")
	alts := carryForward(t, altCol, "0")

	coords := make([]string, len(t.Rows))
	for i, row := range t.Rows {
		coords[i] = row[lon] + "," + row[lat] // KML wants longitude first
		if altCol >= 0 {
			coords[i] += "," + alts[i]
		}
	}
	// altitude is relative to the take-off point; without it, hug the terrain
	lineMode, pointMode := "<tessellate>1</tessellate>", ""
	if altCol >= 0 {
		lineMode = "<altitudeMode>relativeToGround</altitudeMode>"
		pointMode = lineMode
	}

	order, styleOf := styleIDs(modes)
	colors := modeColors(order, modeCol >= 0)

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<kml xmlns="http://www.opengis.net/kml/2.2">` + "\n<Document>\n")
	fmt.Fprintf(&b, "<name>%s</name>\n", xmlText(title))
	for i, color := range colors {
		fmt.Fprintf(&b, "<Style id=\"line%d\"><IconStyle><scale>0</scale></IconStyle><LineStyle><color>%s</color><width>3</width></LineStyle></Style>\n", i, color)
		fmt.Fprintf(&b, "<Style id=\"dot%d\"><IconStyle><color>%s</color><scale>0.5</scale><Icon><href>%s</href></Icon></IconStyle><LabelStyle><scale>0</scale></LabelStyle></Style>\n", i, color, kmlDot)
	}

	for _, seg := range segments(modes) {
		if seg.to == seg.from {
			continue // a line needs two positions
		}
		name := title + " path"
		if modeCol >= 0 {
			name = title + " " + modeName(seg.mode)
		}
		fmt.Fprintf(&b, "<Placemark><name>%s</name><styleUrl>#line%d</styleUrl><LineString>%s<coordinates>%s</coordinates></LineString></Placemark>\n",
			xmlText(name), styleOf[seg.mode], lineMode, strings.Join(coords[seg.from:seg.to+1], " "))
	}

	for i, row := range t.Rows {
		var desc strings.Builder
		fmt.Fprintf(&desc, "%s: %d", timeColumn, t.Times[i])
		for c, col := range t.Columns {
			if c != lat && c != lon && row[c] != "" {
				fmt.Fprintf(&desc, "\n%s: %s", col.Key, row[c])
			}
		}
		fmt.Fprintf(&b, "<Placemark><name>%s</name><styleUrl>#dot%d</styleUrl><description>%s</description><Point>%s<coordinates>%s</coordinates></Point></Placemark>\n",
			clock(t.Times[i]), styleOf[modes[i]], xmlText(desc.String()), pointMode, coords[i])
	}
	b.WriteString("</Document>\n</kml>\n")

	_, err := io.WriteString(w, b.String())
	return err
}

// carryForward returns the values of column col, with each empty one replaced
// by the last value read before it (initial if there is none). A negative col
// means the table has no such column.
func carryForward(t *Table, col int, initial string) []string {
	out := make([]string, len(t.Rows))
	last := initial
	for i, row := range t.Rows {
		if col >= 0 && row[col] != "" {
			last = row[col]
		}
		out[i] = last
	}
	return out
}

// segments cuts the path where the flight mode changes. A segment starts on
// the last row of the one before it, so the path has no gap between them.
func segments(modes []string) []kmlSegment {
	if len(modes) == 0 {
		return nil
	}
	segs := []kmlSegment{{mode: modes[0]}}
	for i := 1; i < len(modes); i++ {
		if modes[i] != modes[i-1] {
			segs[len(segs)-1].to = i - 1
			segs = append(segs, kmlSegment{mode: modes[i], from: i - 1})
		}
	}
	segs[len(segs)-1].to = len(modes) - 1
	return segs
}

// styleIDs numbers the distinct modes in order of first appearance.
func styleIDs(modes []string) (order []string, id map[string]int) {
	id = map[string]int{}
	for _, m := range modes {
		if _, ok := id[m]; !ok {
			id[m] = len(order)
			order = append(order, m)
		}
	}
	return order, id
}

// modeColors gives each mode its colour. hasModes is false when the recording
// has no flight mode at all, in which case the one path takes the default.
func modeColors(order []string, hasModes bool) []string {
	colors := make([]string, len(order))
	spare := 0
	for i, m := range order {
		fixed, known := kmlModeColors[m]
		switch {
		case known:
			colors[i] = fixed
		case m != "":
			colors[i] = kmlSpareColors[spare%len(kmlSpareColors)]
			spare++
		case hasModes:
			colors[i] = kmlNoModeColor
		default:
			colors[i] = kmlPathColor
		}
	}
	return colors
}

// modeName is how a mode appears in the name of its line.
func modeName(mode string) string {
	if mode == "" {
		return "no mode"
	}
	return mode
}

// clock formats a time in ms as m:ss.
func clock(ms uint32) string {
	s := ms / 1000
	return fmt.Sprintf("%d:%02d", s/60, s%60)
}

// xmlText escapes s for use as XML text.
func xmlText(s string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(s)) // a strings.Builder never fails
	return b.String()
}
