package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ericogr/fpv-osd-decoder/internal/osd"
	"github.com/ericogr/fpv-osd-decoder/internal/telemetry"
)

// autoLayout is the value of -mode and -msg-row that finds the position in the
// recording.
const autoLayout = "auto"

// layoutSpec is what -mode and -msg-row ask for. A nil field is to be found in
// the recording.
type layoutSpec struct {
	mode   *telemetry.Region
	msgRow *int // <0 disables the messages
}

func parseLayout(mode, msgRow string) (layoutSpec, error) {
	var spec layoutSpec
	if mode != autoLayout {
		r, err := parseRegion("-mode", mode)
		if err != nil {
			return spec, err
		}
		spec.mode = &r
	}
	if msgRow != autoLayout {
		row, err := strconv.Atoi(strings.TrimSpace(msgRow))
		if err != nil {
			return spec, fmt.Errorf("-msg-row must be a row number or %s (got %q)", autoLayout, msgRow)
		}
		spec.msgRow = &row
	}
	return spec, nil
}

// resolve finds in the recording what the spec leaves open. It returns the
// decoder options and a line that says what was used, so that a position that
// could not be found does not go unnoticed.
func (l layoutSpec) resolve(f *osd.File) (telemetry.Options, string) {
	var opts telemetry.Options
	var mode string
	if l.mode != nil {
		opts.FlightModes = []telemetry.Region{*l.mode}
		mode = fmt.Sprintf("flight mode at %s", l.mode)
	} else {
		opts.FlightModes = telemetry.DetectFlightModes(f)
		mode = describeModes(opts.FlightModes)
	}

	var msg string
	switch {
	case l.msgRow != nil && *l.msgRow < 0:
		opts.MessageRow = -1
		msg = "status messages off"
	case l.msgRow != nil:
		opts.MessageRow = *l.msgRow
		msg = fmt.Sprintf("status messages on row %d", *l.msgRow)
	default:
		row, ok := telemetry.DetectMessageRow(f, opts.FlightModes)
		opts.MessageRow = -1
		if ok {
			opts.MessageRow = row
			msg = fmt.Sprintf("status messages on row %d (found)", row)
		} else {
			msg = "status messages not found (give -msg-row)"
		}
	}
	return opts, "layout: " + mode + "; " + msg
}

func describeModes(modes []telemetry.Region) string {
	if len(modes) == 0 {
		return "flight mode not found (give -mode x,y,width)"
	}
	where := make([]string, len(modes))
	for i, r := range modes {
		where[i] = r.String()
	}
	return "flight mode at " + strings.Join(where, " and ") + " (found)"
}
