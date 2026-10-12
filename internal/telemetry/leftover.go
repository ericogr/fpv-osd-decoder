package telemetry

import (
	"fmt"
	"strings"

	"github.com/ericogr/fpv-osd-decoder/internal/inav"
)

// Leftover is a piece of the screen that no rule recognized.
type Leftover struct {
	X, Y int
	Text string // named glyphs appear as {NAME}
}

// isGraphic reports whether the glyph belongs to a graphic element
// (artificial horizon, crosshair, variometer, HUD arrows), which carries no
// value.
func isGraphic(c rune) bool {
	switch {
	case c >= 0x12c && c <= 0x133, // horizon arrows and scale
		c >= 0x14c && c <= 0x164, // horizontal/vertical horizon, variometer, HUD
		c == 0x166,               // crosshair center
		c >= 0x190 && c <= 0x1a6, // crosshair types
		c >= 0x1ae && c <= 0x1b9: // HUD arrows
		return true
	}
	return false
}

func glyphText(c rune) string {
	if c >= 0x20 && c < 0x7f && inav.Name(uint16(c)) == "" {
		return string(c)
	}
	if n := inav.Name(uint16(c)); n != "" {
		return "{" + n + "}"
	}
	return fmt.Sprintf("{%#x}", c)
}

// appendLeftovers splits the row into contiguous runs of occupied cells and
// appends the ones that are not purely graphic.
func appendLeftovers(out []Leftover, y int, line string) []Leftover {
	var sb strings.Builder
	start, graphic, n := 0, true, 0
	flush := func() {
		if n > 0 && !graphic {
			out = append(out, Leftover{X: start, Y: y, Text: sb.String()})
		}
		sb.Reset()
		graphic, n = true, 0
	}
	x := 0
	for _, c := range line {
		if c == ' ' {
			flush()
		} else {
			if n == 0 {
				start = x
			}
			n++
			sb.WriteString(glyphText(c))
			graphic = graphic && isGraphic(c)
		}
		x++
	}
	flush()
	return out
}
