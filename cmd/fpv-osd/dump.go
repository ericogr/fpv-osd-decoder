package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/ericogr/fpv-osd-decoder/internal/inav"
	"github.com/ericogr/fpv-osd-decoder/internal/osd"
	"github.com/ericogr/fpv-osd-decoder/internal/telemetry"
)

// dumpCommand implements "fpv-osd dump": a diagnostic view of the raw
// OSD screens. It returns the exit code.
func dumpCommand(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("dump", flag.ContinueOnError)
	fs.SetOutput(stderr)
	frame := fs.Int("frame", 0, "frame index")
	count := fs.Int("n", 1, "number of frames")
	hex := fs.Bool("hex", false, "show non-ASCII glyphs by name/code")
	stats := fs.Bool("stats", false, "show a summary of the file")
	census := fs.Bool("census", false, "list non-numeric glyphs, their frequency and positions")
	region := fs.String("region", "", "x,y,width: histogram of the contents of this region across all frames")
	fs.Usage = func() {
		outln(stderr, "usage: fpv-osd dump [options] file.osd")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return 2
	}

	var r telemetry.Region
	if *region != "" {
		var err error
		if r, err = parseRegion("-region", *region); err != nil {
			outln(stderr, "error:", err)
			return 2
		}
	}
	f, err := osd.Open(fs.Arg(0))
	if err != nil {
		outln(stderr, "error:", err)
		return 1
	}

	switch {
	case *stats:
		outf(stdout, "%dx%d grid, %d frames, %.1f s\n", f.Cols, f.Rows, len(f.Frames), float64(f.Duration())/1000)
	case *census:
		printCensus(stdout, f)
	case *region != "":
		if err := printRegion(stdout, f, r); err != nil {
			outln(stderr, "error:", err)
			return 2
		}
	default:
		printFrames(stdout, f, *frame, *count, *hex)
	}
	return 0
}

// glyph renders a cell for the text grid. Non-ASCII glyphs show as "·", or
// as {NAME}/{0xCODE} when verbose.
func glyph(c uint16, verbose bool) string {
	switch {
	case c == ' ' || (c > 0x20 && c < 0x7f && (!verbose || inav.Name(c) == "")):
		return string(rune(c))
	case verbose:
		if n := inav.Name(c); n != "" {
			return fmt.Sprintf("{%s}", n)
		}
		return fmt.Sprintf("{%#x}", c)
	}
	return "·"
}

func printFrames(w io.Writer, f *osd.File, from, n int, verbose bool) {
	for i := from; i < from+n && i < len(f.Frames); i++ {
		fr := &f.Frames[i]
		outf(w, "=== frame %d  t=%d ms ===\n", i, fr.TimeMs)
		for y := 0; y < f.Rows; y++ {
			var sb strings.Builder
			for _, c := range f.Row(fr, y) {
				sb.WriteString(glyph(c, verbose))
			}
			outln(w, strings.TrimRight(sb.String(), " "))
		}
	}
}

// printRegion shows the distinct values (and how many frames each) of a strip
// of cells.
func printRegion(w io.Writer, f *osd.File, r telemetry.Region) error {
	if r.X < 0 || r.Y < 0 || r.Y >= f.Rows {
		return fmt.Errorf("-region %s is outside the %dx%d grid", r, f.Cols, f.Rows)
	}
	hist := map[string]int{}
	for i := range f.Frames {
		row := f.Row(&f.Frames[i], r.Y)
		var sb strings.Builder
		for x := r.X; x < r.X+r.Width && x < f.Cols; x++ {
			sb.WriteString(glyph(row[x], true))
		}
		hist[sb.String()]++
	}
	keys := rankedKeys(hist)
	for i, k := range keys {
		if i == maxListed {
			outf(w, "... and %d more distinct values\n", len(keys)-i)
			break
		}
		outf(w, "%6d  %q\n", hist[k], k)
	}
	return nil
}

type spot struct{ x, y int }

// glyphStat is how often a glyph shows up and where.
type glyphStat struct {
	frames int // frames in which the glyph appears at least once
	spots  map[spot]int
}

// topSpots formats the n most frequent positions of the glyph.
func (g *glyphStat) topSpots(n int) string {
	type counted struct {
		spot
		n int
	}
	list := make([]counted, 0, len(g.spots))
	for p, c := range g.spots {
		list = append(list, counted{p, c})
	}
	sort.Slice(list, func(i, j int) bool {
		a, b := list[i], list[j]
		switch {
		case a.n != b.n:
			return a.n > b.n
		case a.y != b.y:
			return a.y < b.y
		}
		return a.x < b.x
	})
	var top []string
	for i := 0; i < len(list) && i < n; i++ {
		top = append(top, fmt.Sprintf("(%d,%d)x%d", list[i].x, list[i].y, list[i].n))
	}
	return strings.Join(top, " ")
}

// collectCensus counts every glyph that is not a digit or a space.
func collectCensus(f *osd.File) map[uint16]*glyphStat {
	stats := map[uint16]*glyphStat{}
	for i := range f.Frames {
		seen := map[uint16]bool{}
		for y := 0; y < f.Rows; y++ {
			for x, c := range f.Row(&f.Frames[i], y) {
				if c == ' ' || c == 0 || (c >= '0' && c <= '9') {
					continue
				}
				s := stats[c]
				if s == nil {
					s = &glyphStat{spots: map[spot]int{}}
					stats[c] = s
				}
				s.spots[spot{x, y}]++
				if !seen[c] {
					seen[c] = true
					s.frames++
				}
			}
		}
	}
	return stats
}

// printCensus lists, for each glyph that is not a digit or space, in how many
// frames it appears and its most frequent positions.
func printCensus(w io.Writer, f *osd.File) {
	stats := collectCensus(f)
	codes := make([]uint16, 0, len(stats))
	for c := range stats {
		codes = append(codes, c)
	}
	sort.Slice(codes, func(i, j int) bool {
		a, b := stats[codes[i]], stats[codes[j]]
		if a.frames != b.frames {
			return a.frames > b.frames
		}
		return codes[i] < codes[j]
	})
	for _, c := range codes {
		s := stats[c]
		outf(w, "%#05x %-22s %5d frames  %d positions  %s\n", c, glyph(c, true), s.frames, len(s.spots), s.topSpots(3))
	}
}
