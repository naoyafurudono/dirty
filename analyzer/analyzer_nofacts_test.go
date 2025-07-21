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
	// Also set env var to ensure Facts are disabled
	t.Setenv("DIRTY_DISABLE_FACTS", "1")
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, analyzerWithoutFacts, "basic", "complex", "implicit")
}

func TestAnalyzerWithJSONEffectsWithoutFacts(t *testing.T) {
	// Also set env var to ensure Facts are disabled
	t.Setenv("DIRTY_DISABLE_FACTS", "1")
	// Set JSON effects for this test - use absolute path
	testdata := analysistest.TestData()
	jsonPath := testdata + "/src/jsoneffects/effect-registry.json"
	t.Setenv("DIRTY_EFFECTS_JSON", jsonPath)
	analysistest.Run(t, testdata, analyzerWithoutFacts, "jsoneffects")
}

func TestCrossPackageAnalysisWithoutFacts(t *testing.T) {
	// Also set env var to ensure Facts are disabled
	t.Setenv("DIRTY_DISABLE_FACTS", "1")
	testdata := analysistest.TestData()
	// Cross-package analysis won't work without Facts, so only test pkg1 in isolation
	analysistest.Run(t, testdata, analyzerWithoutFacts, "crosspackage/pkg1")
}

func TestNoFactsBasic(t *testing.T) {
	// Also set env var to ensure Facts are disabled
	t.Setenv("DIRTY_DISABLE_FACTS", "1")

	testdata := analysistest.TestData()
	// Test basic functionality without cross-package dependencies
	analysistest.Run(t, testdata, analyzerWithoutFacts, "nofacts")
}
