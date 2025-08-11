package buildinfo_test

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"go-metrics-server/internal/buildinfo"
)

func TestPrint(t *testing.T) {
	originalLogger := log.Default().Writer()
	defer log.SetOutput(originalLogger)

	var buf bytes.Buffer
	log.SetOutput(&buf)

	buildinfo.Print()

	output := buf.String()
	expected := []string{
		"Build version: " + buildinfo.Version,
		"Build date: " + buildinfo.Date,
		"Build commit: " + buildinfo.Commit,
	}
	for _, s := range expected {
		if !strings.Contains(output, s) {
			t.Errorf("Output missing: %s", s)
		}
	}
}
