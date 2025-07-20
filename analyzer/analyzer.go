package analyzer

import (
	"go/ast"
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

	// Map from function name to its effects
	functionEffects := make(map[string][]string)

	// First pass: collect all function effects
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		fn := n.(*ast.FuncDecl)
		if fn.Name == nil {
			return
		}
		
		if fn.Doc != nil {
			for _, comment := range fn.Doc.List {
				if effects := ParseEffects(comment.Text); effects != nil {
					functionEffects[fn.Name.Name] = effects
					break
				}
			}
		}
	})

	// Second pass: check calls within functions
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		fn := n.(*ast.FuncDecl)
		if fn.Name == nil {
			return
		}
		
		// Get declared effects for this function
		declaredEffects := functionEffects[fn.Name.Name]
		
		// Check all function calls within this function
		ast.Inspect(fn, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			
			// Get called function name
			var calledName string
			switch fun := call.Fun.(type) {
			case *ast.Ident:
				calledName = fun.Name
			case *ast.SelectorExpr:
				calledName = fun.Sel.Name
			}
			
			if calledName == "" {
				return true
			}
			
			// Get effects of called function
			calledEffects, ok := functionEffects[calledName]
			if !ok || len(calledEffects) == 0 {
				// No effects declared for called function
				return true
			}
			
			// Check if declared effects is nil (function has no //dirty: comment)
			if declaredEffects == nil {
				// Function doesn't declare effects, so we don't check it
				return true
			}
			
			// Find missing effects
			missingEffects := findMissingEffects(calledEffects, declaredEffects)
			if len(missingEffects) > 0 {
				pass.Reportf(call.Pos(), "function calls %s which has effects [%s] not declared in this function",
					calledName, strings.Join(calledEffects, ", "))
			}
			
			return true
		})
	})

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