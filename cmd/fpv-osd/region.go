package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ericogr/fpv-osd-decoder/internal/telemetry"
)

// parseRegion parses "x,y,width". option names the flag in error messages.
func parseRegion(option, spec string) (telemetry.Region, error) {
	parts := strings.Split(spec, ",")
	if len(parts) != 3 {
		return telemetry.Region{}, fmt.Errorf("%s must be x,y,width (got %q)", option, spec)
	}
	var n [3]int
	for i, p := range parts {
		v, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return telemetry.Region{}, fmt.Errorf("%s must be x,y,width (got %q)", option, spec)
		}
		n[i] = v
	}
	return telemetry.Region{X: n[0], Y: n[1], Width: n[2]}, nil
}
