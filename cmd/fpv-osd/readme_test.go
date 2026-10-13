package main

import (
	"os"
	"strings"
	"testing"

	"github.com/ericogr/fpv-osd-decoder/internal/telemetry"
)

// The README documents every column the program can write.
func TestReadmeListsEveryColumn(t *testing.T) {
	readme, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Skip("README.md is not in the project")
	}
	for _, f := range telemetry.Fields() {
		if !strings.Contains(string(readme), "`"+f.Key+"`") {
			t.Errorf("README.md does not mention %s", f.Key)
		}
	}
}
