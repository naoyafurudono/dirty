package analyzer_test

import (
	"testing"

	"github.com/naoyafurudono/dirty/analyzer"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
)

// analyzerWithoutFacts is a wrapper that disables fact export/import for testing
var analyzerWithoutFacts = &analysis.Analyzer{
	Name:       analyzer.Analyzer.Name,
	Doc:        analyzer.Analyzer.Doc,
	URL:        analyzer.Analyzer.URL,
	Requires:   analyzer.Analyzer.Requires,
	ResultType: analyzer.Analyzer.ResultType,
	FactTypes:  nil, // Disable facts for these tests
	Run:        analyzer.Analyzer.Run,
}

func TestAnalyzerWithoutFacts(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, analyzerWithoutFacts, "basic", "complex", "implicit")
}

func TestAnalyzerWithJSONEffectsWithoutFacts(t *testing.T) {
	// Set JSON effects for this test - use absolute path
	testdata := analysistest.TestData()
	jsonPath := testdata + "/src/jsoneffects/effect-registry.json"
	t.Setenv("DIRTY_EFFECTS_JSON", jsonPath)
	analysistest.Run(t, testdata, analyzerWithoutFacts, "jsoneffects")
}

func TestCrossPackageAnalysisWithoutFacts(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, analyzerWithoutFacts, "crosspackage/...")
}

func TestCrossPackageDebugWithoutFacts(t *testing.T) {
	// Enable verbose mode for debugging
	t.Setenv("DIRTY_VERBOSE", "1")
	defer t.Setenv("DIRTY_VERBOSE", "")

	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, analyzerWithoutFacts, "crosspackage/pkg2")
}
