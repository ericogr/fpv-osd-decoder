package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ericogr/fpv-osd-decoder/internal/export"
	"github.com/ericogr/fpv-osd-decoder/internal/osd"
	"github.com/ericogr/fpv-osd-decoder/internal/srt"
	"github.com/ericogr/fpv-osd-decoder/internal/telemetry"
)

const (
	stdoutPath = "-"    // -o value that writes the CSV to stdout
	noSRT      = "none" // -srt value that disables the join
)

// csvOptions are the settings of the csv subcommand.
type csvOptions struct {
	osdPath, outPath, srtPath string
	decoder                   telemetry.Options
	format                    export.Format
	profile                   export.Profile
	quiet, unmapped           bool
}

// csvCommand implements "fpv-osd csv" and returns the exit code.
func csvCommand(args []string, stdout, stderr io.Writer) int {
	def := telemetry.DefaultOptions()
	fs := flag.NewFlagSet("csv", flag.ContinueOnError)
	fs.SetOutput(stderr)
	out := fs.String("o", "", "output file (default: same name as the .osd, plus .<profile> unless it is the default profile; "+stdoutPath+" for stdout)")
	srtPath := fs.String("srt", "", ".srt file (default: same name as the .osd, if it exists; "+noSRT+" disables)")
	mode := fs.String("mode", def.FlightMode.String(), "x,y,width of the flight mode region on the OSD")
	msgRow := fs.Int("msg-row", def.MessageRow, "OSD row of the status messages (-1 disables)")
	sep := fs.String("sep", ",", "CSV column separator (use ; for Portuguese-locale spreadsheets; not for -profile maps)")
	decimal := fs.String("decimal", ".", "decimal separator of numbers (use , for Portuguese-locale spreadsheets; not for -profile maps)")
	profile := fs.String("profile", export.ProfileDefault.String(), "what to export: default (a CSV with every column and frame) or maps (a KML for Google My Maps: the path and up to 2000 valid GPS points)")
	quiet := fs.Bool("q", false, "do not print the column summary")
	unmapped := fs.Bool("unmapped", false, "list what was left on the screen undecoded (useful to map a new layout)")
	fs.Usage = func() {
		outln(stderr, "usage: fpv-osd csv [options] file.osd")
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

	opts := csvOptions{
		osdPath:  fs.Arg(0),
		outPath:  *out,
		srtPath:  *srtPath,
		decoder:  telemetry.Options{MessageRow: *msgRow},
		quiet:    *quiet,
		unmapped: *unmapped,
	}
	var err error
	if opts.decoder.FlightMode, err = parseRegion("-mode", *mode); err != nil {
		outln(stderr, "error:", err)
		return 2
	}
	if opts.profile, err = export.ParseProfile(*profile); err != nil {
		outln(stderr, "error: invalid -profile:", err)
		return 2
	}
	if opts.profile.UsesCSV() {
		if opts.format, err = export.NewFormat(*sep, *decimal); err != nil {
			outln(stderr, "error: invalid -sep/-decimal:", err)
			return 2
		}
	} else {
		given := false
		fs.Visit(func(f *flag.Flag) { given = given || f.Name == "sep" || f.Name == "decimal" })
		if given {
			outf(stderr, "error: -sep and -decimal do not apply to -profile %s, which does not write a CSV\n", opts.profile)
			return 2
		}
	}
	if err := runCSV(opts, stdout, stderr); err != nil {
		outln(stderr, "error:", err)
		return 1
	}
	return 0
}

// runCSV decodes the recording, joins the SRT, applies the profile and writes
// the output file.
func runCSV(o csvOptions, stdout, stderr io.Writer) error {
	f, err := osd.Open(o.osdPath)
	if err != nil {
		return err
	}
	sf, err := openSRT(o)
	if err != nil {
		return err
	}

	frames, unmapped := decodeFrames(f, telemetry.New(o.decoder))
	table, dropped, err := o.profile.Apply(export.Build(frames, sf))
	if err != nil {
		return fmt.Errorf("profile %s: %w", o.profile, err)
	}

	if err := writeOutput(table, o, stdout); err != nil {
		return err
	}
	if !o.quiet {
		printSummary(stderr, f, sf, table, dropped)
	}
	if o.unmapped {
		printUnmapped(stderr, unmapped, len(f.Frames))
	}
	return nil
}

// openSRT opens the subtitles to join: the -srt file, or by default the one
// next to the .osd. It returns nil when there is none or the join is disabled.
func openSRT(o csvOptions) (*srt.File, error) {
	path := o.srtPath
	if path == noSRT {
		return nil, nil
	}
	if path == "" {
		path = basePath(o.osdPath) + ".srt"
		if _, err := os.Stat(path); err != nil {
			return nil, nil
		}
	}
	return srt.Open(path)
}

// decodeFrames decodes every frame and counts, by position and content, what
// the decoder left on the screen unrecognized.
func decodeFrames(f *osd.File, dec *telemetry.Decoder) ([]export.Frame, map[string]int) {
	frames := make([]export.Frame, len(f.Frames))
	unmapped := map[string]int{}
	for i := range f.Frames {
		sample, left := dec.Decode(f, &f.Frames[i])
		frames[i] = export.Frame{TimeMs: f.Frames[i].TimeMs, Sample: sample}
		for _, l := range left {
			unmapped[fmt.Sprintf("(%2d,%2d) %s", l.X, l.Y, l.Text)]++
		}
	}
	return frames, unmapped
}

// writeOutput writes the table to the -o file (default: next to the .osd, with
// the profile in the name unless it is the default one) or to stdout.
func writeOutput(t *export.Table, o csvOptions, stdout io.Writer) (err error) {
	title := filepath.Base(basePath(o.osdPath))
	if o.outPath == stdoutPath {
		return o.profile.Write(stdout, t, o.format, title)
	}
	path := o.outPath
	if path == "" {
		path = basePath(o.osdPath)
		if o.profile != export.ProfileDefault {
			path += "." + o.profile.String()
		}
		path += o.profile.Ext()
	}
	fh, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := fh.Close(); err == nil {
			err = cerr
		}
	}()
	return o.profile.Write(fh, t, o.format, title)
}

// basePath is path without its extension.
func basePath(path string) string {
	return strings.TrimSuffix(path, filepath.Ext(path))
}

func printSummary(w io.Writer, f *osd.File, sf *srt.File, t *export.Table, dropped export.Reduction) {
	outf(w, "OSD: %dx%d grid, %d frames, %.1f s\n", f.Cols, f.Rows, len(f.Frames), float64(f.Duration())/1000)
	if sf != nil {
		outf(w, "SRT: %d entries\n", len(sf.Entries))
	}
	if dropped.Total() > 0 {
		outf(w, "rows written: %d of %d frames (left out: %d without a valid GPS position, %d repeating the previous position, %d thinned to fit the row limit)\n",
			len(t.Rows), len(f.Frames), dropped.NoGPS, dropped.Repeated, dropped.Thinned)
	}
	outf(w, "columns (%d), with the number of rows in which each value appeared:\n", len(t.Columns))
	for _, c := range t.Columns {
		outf(w, "  %-22s %6d/%d  %s\n", c.Key, c.Frames, len(t.Rows), c.Desc)
	}
}

// printUnmapped shows the most frequent undecoded pieces of the screen.
func printUnmapped(w io.Writer, hist map[string]int, frames int) {
	keys := rankedKeys(hist)
	outf(w, "undecoded (x,y position and content; frames in which it appeared), %d distinct:\n", len(keys))
	for i, k := range keys {
		if i == maxListed {
			outf(w, "  ... and %d more\n", len(keys)-i)
			break
		}
		outf(w, "  %6d/%d  %s\n", hist[k], frames, k)
	}
}
