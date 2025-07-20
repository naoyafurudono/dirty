package analyzer_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/naoyafurudono/dirty/analyzer"
)

func TestLoadSQLCEffects(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "test-operations.json")
	
	testJSON := `{
		"GetUser": [
			{"operation": "select", "table": "users"}
		],
		"CreateUserWithAudit": [
			{"operation": "insert", "table": "users"},
			{"operation": "insert", "table": "audit_logs"}
		]
	}`
	
	if err := os.WriteFile(jsonPath, []byte(testJSON), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	
	tests := []struct {
		name     string
		jsonPath string
		wantErr  bool
		wantLen  int
	}{
		{
			name:     "valid JSON",
			jsonPath: jsonPath,
			wantErr:  false,
			wantLen:  2,
		},
		{
			name:     "empty path",
			jsonPath: "",
			wantErr:  false,
			wantLen:  0,
		},
		{
			name:     "non-existent file",
			jsonPath: "/does/not/exist.json",
			wantErr:  true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queryMap, err := analyzer.LoadSQLCEffects(tt.jsonPath)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadSQLCEffects() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && len(queryMap) != tt.wantLen {
				t.Errorf("LoadSQLCEffects() returned %d queries, want %d", len(queryMap), tt.wantLen)
			}
		})
	}
}

func TestConvertToEffects(t *testing.T) {
	tests := []struct {
		name       string
		operations []analyzer.SQLCOperation
		want       []string
	}{
		{
			name: "single operation",
			operations: []analyzer.SQLCOperation{
				{Operation: "select", Table: "users"},
			},
			want: []string{"select[users]"},
		},
		{
			name: "multiple operations",
			operations: []analyzer.SQLCOperation{
				{Operation: "insert", Table: "users"},
				{Operation: "insert", Table: "audit_logs"},
			},
			want: []string{"insert[audit_logs]", "insert[users]"},
		},
		{
			name: "duplicate operations",
			operations: []analyzer.SQLCOperation{
				{Operation: "select", Table: "users"},
				{Operation: "select", Table: "users"},
			},
			want: []string{"select[users]"},
		},
		{
			name: "all operation types",
			operations: []analyzer.SQLCOperation{
				{Operation: "select", Table: "users"},
				{Operation: "insert", Table: "posts"},
				{Operation: "update", Table: "comments"},
				{Operation: "delete", Table: "sessions"},
			},
			want: []string{"delete[sessions]", "insert[posts]", "select[users]", "update[comments]"},
		},
		{
			name:       "empty operations",
			operations: []analyzer.SQLCOperation{},
			want:       []string{},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzer.ConvertToEffects(tt.operations)
			
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertToEffects() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLCIntegration(t *testing.T) {
	// Test the full flow: load JSON and convert to effects
	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "query-table-operations.json")
	
	testJSON := `{
		"ComplexQuery": [
			{"operation": "select", "table": "users"},
			{"operation": "select", "table": "organizations"},
			{"operation": "update", "table": "member_counts"},
			{"operation": "insert", "table": "activity_logs"}
		]
	}`
	
	if err := os.WriteFile(jsonPath, []byte(testJSON), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	
	// Load JSON
	queryMap, err := analyzer.LoadSQLCEffects(jsonPath)
	if err != nil {
		t.Fatalf("LoadSQLCEffects() failed: %v", err)
	}
	
	// Convert ComplexQuery operations
	operations, ok := queryMap["ComplexQuery"]
	if !ok {
		t.Fatal("ComplexQuery not found in query map")
	}
	
	effects := analyzer.ConvertToEffects(operations)
	
	// Verify all effects are present and sorted
	want := []string{
		"insert[activity_logs]",
		"select[organizations]",
		"select[users]",
		"update[member_counts]",
	}
	
	if !reflect.DeepEqual(effects, want) {
		t.Errorf("ConvertToEffects() = %v, want %v", effects, want)
	}
}