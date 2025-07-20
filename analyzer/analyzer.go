package analyzer

import (
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


// ParseEffects extracts effects from a //dirty: comment
func ParseEffects(comment string) []string {
	comment = strings.TrimSpace(comment)
	if !strings.HasPrefix(comment, "//dirty:") {
		return nil
	}
	
	effectStr := strings.TrimPrefix(comment, "//dirty:")
	effectStr = strings.TrimSpace(effectStr)
	
	if effectStr == "" {
		return []string{}
	}
	
	// Split by comma and trim spaces
	parts := strings.Split(effectStr, ",")
	effects := make([]string, 0, len(parts))
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			effects = append(effects, part)
		}
	}
	
	return effects
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