package analyzer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/naoyafurudono/dirty/analyzer"
	"golang.org/x/tools/go/analysis/analysistest"
)

// These tests require Facts to work properly.
// They test cross-package functionality that depends on Facts.
// We skip them if running in an environment that doesn't support Facts.

func skipIfNoFactsSupport(t *testing.T) {
	// Skip these tests when running with go test ./...
	// They should be run explicitly with: go test -run TestCrossPackageWithFacts
	if os.Getenv("ENABLE_FACTS_TESTS") != "1" {
		t.Skip("Skipping Facts-based tests. Set ENABLE_FACTS_TESTS=1 to run them.")
	}
}

func TestCrossPackageAnalysisWithFacts(t *testing.T) {
	skipIfNoFactsSupport(t)

	// Enable Facts explicitly
	os.Unsetenv("DIRTY_DISABLE_FACTS")

	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, analyzer.Analyzer, "crosspackage/...")
}

func TestCrossPackageDebugWithFacts(t *testing.T) {
	skipIfNoFactsSupport(t)

	// Enable Facts explicitly
	os.Unsetenv("DIRTY_DISABLE_FACTS")

	// Enable verbose mode for debugging
	os.Setenv("DIRTY_VERBOSE", "1")
	defer os.Unsetenv("DIRTY_VERBOSE")

	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, analyzer.Analyzer, "crosspackage/pkg2")
}

func TestFactCrossPackageWithFacts(t *testing.T) {
	skipIfNoFactsSupport(t)

	// Enable Facts explicitly
	os.Unsetenv("DIRTY_DISABLE_FACTS")

	testdata := analysistest.TestData()

	// Create a temporary module for cross-package testing
	tmpDir := t.TempDir()

	// Copy test files
	srcDir := filepath.Join(testdata, "src", "crosspackage")
	if err := copyDir(srcDir, filepath.Join(tmpDir, "crosspackage")); err != nil {
		t.Fatalf("Failed to copy test files: %v", err)
	}

	// Run analysis on the packages
	pkgs := []string{
		filepath.Join(tmpDir, "crosspackage", "pkg1"),
		filepath.Join(tmpDir, "crosspackage", "pkg2"),
		filepath.Join(tmpDir, "crosspackage", "pkg3"),
	}

	for _, pkg := range pkgs {
		result := analysistest.Run(t, tmpDir, analyzer.Analyzer, pkg)
		if result == nil {
			t.Errorf("Analysis failed for package %s", pkg)
		}
	}
}

func TestSimpleCrossPackageWithFacts(t *testing.T) {
	skipIfNoFactsSupport(t)

	// Enable Facts explicitly
	os.Unsetenv("DIRTY_DISABLE_FACTS")

	testdata := analysistest.TestData()

	tests := []struct {
		name     string
		patterns []string
	}{
		{"simpletest/pkg1", []string{"simpletest/pkg1"}},
		{"simpletest/pkg2", []string{"simpletest/pkg2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysistest.Run(t, testdata, analyzer.Analyzer, tt.patterns...)
		})
	}
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(dstPath, data, info.Mode())
	})
}
