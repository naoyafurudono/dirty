package analyzer

import (
	"reflect"
	"testing"
)

func TestParseEffectDecl(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    EffectExpr
		wantErr bool
	}{
		{
			name:  "empty declaration",
			input: "//dirty:",
			want:  &LiteralSet{Elements: []EffectExpr{}},
		},
		{
			name:  "empty set",
			input: "//dirty: { }",
			want:  &LiteralSet{Elements: []EffectExpr{}},
		},
		{
			name:  "single effect",
			input: "//dirty: { select[users] }",
			want: &LiteralSet{
				Elements: []EffectExpr{
					&EffectLabel{Operation: "select", Target: "users"},
				},
			},
		},
		{
			name:  "two effects",
			input: "//dirty: { select[users] | insert[logs] }",
			want: &LiteralSet{
				Elements: []EffectExpr{
					&EffectLabel{Operation: "select", Target: "users"},
					&EffectLabel{Operation: "insert", Target: "logs"},
				},
			},
		},
		{
			name:  "three effects",
			input: "//dirty: { select[users] | insert[logs] | update[users] }",
			want: &LiteralSet{
				Elements: []EffectExpr{
					&EffectLabel{Operation: "select", Target: "users"},
					&EffectLabel{Operation: "insert", Target: "logs"},
					&EffectLabel{Operation: "update", Target: "users"},
				},
			},
		},
		{
			name:  "with extra spaces",
			input: "//dirty:  {  select[users]  |  insert[logs]  }  ",
			want: &LiteralSet{
				Elements: []EffectExpr{
					&EffectLabel{Operation: "select", Target: "users"},
					&EffectLabel{Operation: "insert", Target: "logs"},
				},
			},
		},
		{
			name:  "with parentheses",
			input: "//dirty: { (select[users] | select[posts]) | insert[logs] }",
			want: &LiteralSet{
				Elements: []EffectExpr{
					&LiteralSet{
						Elements: []EffectExpr{
							&EffectLabel{Operation: "select", Target: "users"},
							&EffectLabel{Operation: "select", Target: "posts"},
						},
					},
					&EffectLabel{Operation: "insert", Target: "logs"},
				},
			},
		},
		// Error cases
		{
			name:    "missing opening brace",
			input:   "//dirty: select[users] }",
			wantErr: true,
		},
		{
			name:    "missing closing brace",
			input:   "//dirty: { select[users]",
			wantErr: true,
		},
		{
			name:    "missing bracket",
			input:   "//dirty: { select users] }",
			wantErr: true,
		},
		{
			name:    "missing closing bracket",
			input:   "//dirty: { select[users }",
			wantErr: true,
		},
		{
			name:    "invalid token",
			input:   "//dirty: { select[users] & insert[logs] }",
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseEffectDecl(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseEffectDecl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseEffectDecl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEval(t *testing.T) {
	tests := []struct {
		name string
		expr EffectExpr
		want []string
	}{
		{
			name: "empty set",
			expr: &LiteralSet{Elements: []EffectExpr{}},
			want: []string{},
		},
		{
			name: "single effect",
			expr: &LiteralSet{
				Elements: []EffectExpr{
					&EffectLabel{Operation: "select", Target: "users"},
				},
			},
			want: []string{"select[users]"},
		},
		{
			name: "multiple effects",
			expr: &LiteralSet{
				Elements: []EffectExpr{
					&EffectLabel{Operation: "select", Target: "users"},
					&EffectLabel{Operation: "insert", Target: "logs"},
					&EffectLabel{Operation: "update", Target: "users"},
				},
			},
			want: []string{"select[users]", "insert[logs]", "update[users]"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.expr.Eval(nil)
			if err != nil {
				t.Errorf("Eval() error = %v", err)
				return
			}
			
			// Convert StringSet to sorted slice for comparison
			gotSlice := got.ToSlice()
			if !equalStringSlices(gotSlice, tt.want) {
				t.Errorf("Eval() = %v, want %v", gotSlice, tt.want)
			}
		})
	}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	// Create maps for comparison (order doesn't matter for sets)
	aMap := make(map[string]bool)
	bMap := make(map[string]bool)
	for _, s := range a {
		aMap[s] = true
	}
	for _, s := range b {
		bMap[s] = true
	}
	for k := range aMap {
		if !bMap[k] {
			return false
		}
	}
	return true
}