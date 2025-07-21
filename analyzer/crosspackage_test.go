package analyzer_test

import (
	"testing"

	"github.com/naoyafurudono/dirty/analyzer"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestCrossPackageAnalysis(t *testing.T) {
	// Enable cross-package analysis test

	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, analyzer.Analyzer, "crosspackage/...")
}
