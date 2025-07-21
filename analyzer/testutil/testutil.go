// Package testutil provides testing utilities for the dirty analyzer
package testutil

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// ParseFile parses a Go source file and returns the AST
func ParseFile(t *testing.T, src string) (*ast.File, *token.FileSet) {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}
	return file, fset
}

// ExtractDirtyComment extracts the // dirty: comment from a function declaration
func ExtractDirtyComment(fn *ast.FuncDecl) string {
	if fn.Doc == nil {
		return ""
	}

	for _, comment := range fn.Doc.List {
		text := strings.TrimSpace(comment.Text)
		if strings.HasPrefix(text, "// dirty:") {
			return text
		}
	}
	return ""
}

// AssertEffects checks if two effect slices are equal
func AssertEffects(t *testing.T, got, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Errorf("effect count mismatch: got %d, want %d", len(got), len(want))
		return
	}

	for i := range got {
		if got[i] != want[i] {
			t.Errorf("effect[%d] mismatch: got %q, want %q", i, got[i], want[i])
		}
	}
}

// ParseEffectsFromComment parses effects from a // dirty: comment
func ParseEffectsFromComment(comment string) ([]string, error) {
	// Remove // dirty: prefix
	comment = strings.TrimSpace(comment)
	if !strings.HasPrefix(comment, "// dirty:") {
		return nil, fmt.Errorf("not a dirty comment: %s", comment)
	}

	comment = strings.TrimPrefix(comment, "// dirty:")
	comment = strings.TrimSpace(comment)

	if comment == "" {
		return []string{}, nil
	}

	// Split by comma and trim spaces
	parts := strings.Split(comment, ",")
	effects := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			effects = append(effects, part)
		}
	}

	return effects, nil
}
