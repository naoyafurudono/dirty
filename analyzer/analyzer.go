// Package analyzer implements the dirty effect checking analyzer
package analyzer

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer is the dirty effect analyzer
var Analyzer = &analysis.Analyzer{
	Name:     "dirty",
	Doc:      "checks that function effect declarations are consistent",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	FactTypes: []analysis.Fact{
		(*PackageEffectsFact)(nil),
		(*FunctionEffectsFact)(nil),
	},
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Create effect analysis
	effectAnalysis := NewEffectAnalysis(pass, insp)

	// Load JSON effects if available
	jsonPath := os.Getenv("DIRTY_EFFECTS_JSON")
	if jsonPath == "" {
		// Try to find in package directory
		if len(pass.Files) > 0 {
			pkgDir := filepath.Dir(pass.Fset.Position(pass.Files[0].Pos()).Filename)
			jsonPath = filepath.Join(pkgDir, "effect-registry.json")
		}
	}

	var jsonEffects ParsedEffects
	if jsonPath != "" && fileExists(jsonPath) {
		decls, err := LoadEffectDeclarations(jsonPath)
		if err == nil {
			jsonEffects, _ = decls.ParseAll()
		}
		// Silently ignore errors loading JSON
	}
	effectAnalysis.JSONEffects = jsonEffects

	// Phase 1: Collect all functions and their declared effects
	effectAnalysis.CollectFunctions()

	// Phase 2: Build call graph
	effectAnalysis.BuildCallGraph()

	// Phase 2.5: Enhance with cross-package support
	EnhanceWithCrossPackageSupportV2(effectAnalysis)

	// Phase 3: Propagate effects
	effectAnalysis.PropagateEffects()

	// Phase 4: Check effect consistency
	effectAnalysis.CheckEffects()

	// Phase 5: Export effects as Facts for dependent packages
	effectAnalysis.ExportPackageEffects()

	return nil, nil
}

// ParseEffects extracts effects from a // dirty: comment
func ParseEffects(comment string) []string {
	comment = strings.TrimSpace(comment)
	if !strings.HasPrefix(comment, "// dirty:") {
		return nil
	}

	// Use the new parser
	expr, err := ParseEffectDecl(comment)
	if err != nil {
		// For backward compatibility, return empty on parse error
		// In the future, we might want to report this error
		return []string{}
	}

	// Evaluate the expression to get the set of effects
	set, err := expr.Eval(nil)
	if err != nil {
		return []string{}
	}

	// Convert to sorted slice
	return set.ToSlice()
}
