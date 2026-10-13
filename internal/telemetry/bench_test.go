package telemetry

import (
	"os"
	"testing"

	"github.com/ericogr/fpv-osd-decoder/internal/osd"
)

// BenchmarkDecodeSampleFlight decodes every frame of the sample flight, if it
// is in the project.
func BenchmarkDecodeSampleFlight(b *testing.B) {
	data, err := os.ReadFile("../../AvatarG0031.osd")
	if err != nil {
		b.Skip("AvatarG0031.osd is not in the project")
	}
	f, err := osd.Parse(data)
	if err != nil {
		b.Fatal(err)
	}
	d := New(sampleLayout())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := range f.Frames {
			d.Decode(f, &f.Frames[j])
		}
	}
}
