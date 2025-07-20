package analyzer

import (
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/analysis"
)

// Analyzer is the dirty effect analyzer
var Analyzer = &analysis.Analyzer{
	Name: "dirty",
	Doc:  "checks that function effect declarations are consistent",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	// TODO: Implement effect checking logic
	return nil, nil
}

// Effect represents a single effect declaration
type Effect struct {
	Action string
	Target string
}

// parseEffects extracts effects from a //dirty: comment
func parseEffects(comment string) []Effect {
	// TODO: Implement parsing logic
	return nil
}

// checkFunctionEffects verifies that a function's declared effects
// include all effects from functions it calls
func checkFunctionEffects(fset *token.FileSet, fn *ast.FuncDecl) error {
	// TODO: Implement checking logic
	return nil
}