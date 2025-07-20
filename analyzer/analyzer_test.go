package analyzer_test

import (
	"testing"

	"github.com/naoyafurudono/dirty/analyzer"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, analyzer.Analyzer, "basic", "complex", "implicit")
}

func TestAnalyzerWithSQLC(t *testing.T) {
	// Set SQLC JSON for this test - use absolute path
	testdata := analysistest.TestData()
	sqlcJSONPath := testdata + "/src/sqlc/query-table-operations.json"
	t.Setenv("DIRTY_SQLC_JSON", sqlcJSONPath)
	analysistest.Run(t, testdata, analyzer.Analyzer, "sqlc")
}

func TestParseEffects(t *testing.T) {
	tests := []struct {
		name    string
		comment string
		want    []string
	}{
		{
			name:    "single effect",
			comment: "// dirty: { select[user] }",
			want:    []string{"select[user]"},
		},
		{
			name:    "multiple effects",
			comment: "// dirty: { select[user] | insert[log] }",
			want:    []string{"insert[log]", "select[user]"},
		},
		{
			name:    "effects with spaces",
			comment: "// dirty: { select[user] | update[member] | delete[session] }",
			want:    []string{"delete[session]", "select[user]", "update[member]"},
		},
		{
			name:    "empty effects",
			comment: "// dirty: { }",
			want:    []string{},
		},
		{
			name:    "not a dirty comment",
			comment: "// regular comment",
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzer.ParseEffects(tt.comment)
			if !equalStringSlices(got, tt.want) {
				t.Errorf("ParseEffects() = %v, want %v", got, tt.want)
			}
		})
	}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestCheckFunctionEffects(t *testing.T) {
	// TODO: Add tests for function effect checking
}
