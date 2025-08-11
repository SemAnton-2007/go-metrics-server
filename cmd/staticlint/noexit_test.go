package main

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestNoExitAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()

	analysistest.Run(t, testdata, NoExitAnalyzer, "exit")

	analysistest.Run(t, testdata, NoExitAnalyzer, "noexit")
}
