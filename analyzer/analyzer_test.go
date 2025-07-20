package analyzer_test

import (
	"testing"

	"github.com/naoyafurudono/dirty/analyzer"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, analyzer.Analyzer, "basic")
}

func TestParseEffects(t *testing.T) {
	tests := []struct {
		name    string
		comment string
		want    []analyzer.Effect
	}{
		{
			name:    "single effect",
			comment: "//dirty: select[user]",
			want: []analyzer.Effect{
				{Action: "select", Target: "user"},
			},
		},
		{
			name:    "multiple effects",
			comment: "//dirty: select[user], insert[log]",
			want: []analyzer.Effect{
				{Action: "select", Target: "user"},
				{Action: "insert", Target: "log"},
			},
		},
		{
			name:    "effects with spaces",
			comment: "//dirty: select[user] , update[member] , delete[session]",
			want: []analyzer.Effect{
				{Action: "select", Target: "user"},
				{Action: "update", Target: "member"},
				{Action: "delete", Target: "session"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: Implement when parseEffects is exposed
		})
	}
}

func TestCheckFunctionEffects(t *testing.T) {
	// TODO: Add tests for function effect checking
}