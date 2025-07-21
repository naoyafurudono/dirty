package analyzer_test

import (
	"os"
	"testing"

	"github.com/naoyafurudono/dirty/analyzer"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestCrossPackageDebug(t *testing.T) {
	// Enable verbose mode for debugging
	os.Setenv("DIRTY_VERBOSE", "1")
	defer os.Unsetenv("DIRTY_VERBOSE")

	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, analyzer.Analyzer, "crosspackage/pkg2")
}
