package analyzer_test

import (
	"path/filepath"
	"testing"

	"github.com/naoyafurudono/dirty/analyzer"
	"golang.org/x/tools/go/analysis/analysistest"
)

// TestFactCrossPackage tests cross-package analysis with Facts
func TestFactCrossPackage(t *testing.T) {
	// Create a custom test runner that supports multiple packages
	testdata := analysistest.TestData()

	// Run analysis on all packages in the test directory
	results := analysistest.Run(t, testdata, analyzer.Analyzer, "facttest/...")

	// Check that we got results for both packages
	if len(results) < 2 {
		t.Errorf("Expected at least 2 packages, got %d", len(results))
	}
}

// TestSimpleCrossPackage tests a simple cross-package scenario
func TestSimpleCrossPackage(t *testing.T) {
	testdata := analysistest.TestData()

	// Define our test packages
	pkgs := []string{
		filepath.Join("simpletest", "pkg1"),
		filepath.Join("simpletest", "pkg2"),
	}

	// Run analyzer on each package
	for _, pkg := range pkgs {
		t.Run(pkg, func(t *testing.T) {
			analysistest.Run(t, testdata, analyzer.Analyzer, pkg)
		})
	}
}
