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
			graphic = graphic && inav.IsGraphic(uint16(c))
		}
		x++
	}
	flush()
	return out
}
