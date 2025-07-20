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

// ExtractDirtyComment extracts the //dirty: comment from a function declaration
func ExtractDirtyComment(fn *ast.FuncDecl) string {
	if fn.Doc == nil {
		return ""
	}
	
	for _, comment := range fn.Doc.List {
		text := strings.TrimSpace(comment.Text)
		if strings.HasPrefix(text, "//dirty:") {
			return text
		}
	}
	return ""
}

// AssertEffects checks if two effect slices are equal
func AssertEffects(t *testing.T, got, want []Effect) {
	t.Helper()
	
	if len(got) != len(want) {
		t.Errorf("effect count mismatch: got %d, want %d", len(got), len(want))
		return
	}
	
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("effect[%d] mismatch: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

// Effect represents a parsed effect for testing
type Effect struct {
	Action string
	Target string
}

// ParseEffectsFromComment parses effects from a //dirty: comment
func ParseEffectsFromComment(comment string) ([]Effect, error) {
	// Remove //dirty: prefix
	comment = strings.TrimPrefix(comment, "//dirty:")
	comment = strings.TrimSpace(comment)
	
	if comment == "" {
		return nil, nil
	}
	
	var effects []Effect
	parts := strings.Split(comment, ",")
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		
		// Parse action[target] format
		openBracket := strings.Index(part, "[")
		closeBracket := strings.Index(part, "]")
		
		if openBracket == -1 || closeBracket == -1 || closeBracket <= openBracket {
			return nil, fmt.Errorf("invalid effect format: %s", part)
		}
		
		action := part[:openBracket]
		target := part[openBracket+1 : closeBracket]
		
		effects = append(effects, Effect{
			Action: action,
			Target: target,
		})
	}
	
	return effects, nil
}