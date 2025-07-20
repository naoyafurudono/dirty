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
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Create effect analysis
	analysis := NewEffectAnalysis(pass, inspect)

	// Load SQLC effects if available
	sqlcJSON := os.Getenv("DIRTY_SQLC_JSON")
	if sqlcJSON == "" {
		// Try to find query-table-operations.json in current directory
		if _, err := os.Stat("query-table-operations.json"); err == nil {
			sqlcJSON = "query-table-operations.json"
		}
		// Also check package directory
		if sqlcJSON == "" && len(pass.Files) > 0 {
			pkgDir := filepath.Dir(pass.Fset.Position(pass.Files[0].Pos()).Filename)
			jsonPath := filepath.Join(pkgDir, "query-table-operations.json")
			if _, err := os.Stat(jsonPath); err == nil {
				sqlcJSON = jsonPath
			}
		}
	}

	if sqlcJSON != "" {
		if sqlcEffects, err := LoadSQLCEffects(sqlcJSON); err == nil {
			analysis.SQLCEffects = sqlcEffects
		}
		// Silently ignore errors loading SQLC JSON
	}

	// Phase 1: Collect all functions and their declared effects
	analysis.CollectFunctions()

	// Phase 2: Build call graph
	analysis.BuildCallGraph()

	// Phase 3: Propagate effects
	analysis.PropagateEffects()

	// Phase 4: Check effect consistency
	analysis.CheckEffects()

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

// findMissingEffects returns effects that are in called but not in declared
func findMissingEffects(called, declared []string) []string {
	declaredSet := make(map[string]bool)
	for _, effect := range declared {
		declaredSet[effect] = true
	}

	var missing []string
	for _, effect := range called {
		if !declaredSet[effect] {
			missing = append(missing, effect)
		}
	}

	return missing
}
