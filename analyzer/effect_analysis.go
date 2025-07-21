package analyzer

import (
	"go/ast"
	"os"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ast/inspector"
)

// EffectAnalysis holds the state for effect analysis
type EffectAnalysis struct {
	Pass      *analysis.Pass
	Inspector *inspector.Inspector
	Functions map[string]*FunctionInfo
	CallGraph *CallGraph

	// JSON effect declarations
	JSONEffects ParsedEffects

	// DisableFacts disables fact export (for testing)
	DisableFacts bool

	// UnifiedEffectResolver provides unified effect resolution
	Resolver *UnifiedEffectResolver
}

// NewEffectAnalysis creates a new EffectAnalysis
func NewEffectAnalysis(pass *analysis.Pass, inspect *inspector.Inspector) *EffectAnalysis {
	return &EffectAnalysis{
		Pass:      pass,
		Inspector: inspect,
		Functions: make(map[string]*FunctionInfo),
		CallGraph: NewCallGraph(),
		Resolver:  NewUnifiedEffectResolver(),
	}
}

// CollectFunctions collects all function declarations and their effects
func (ea *EffectAnalysis) CollectFunctions() {
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	ea.Inspector.Preorder(nodeFilter, func(n ast.Node) {
		fn := n.(*ast.FuncDecl)
		if fn.Name == nil {
			return
		}

		funcName := fn.Name.Name
		info := &FunctionInfo{
			Name:            funcName,
			Package:         ea.Pass.Pkg.Path(),
			DeclaredEffects: NewStringSet(),
			ComputedEffects: NewStringSet(),
			HasDeclaration:  false,
			Decl:            fn,
			CallSites:       []CallSite{},
		}

		// Extract effects from // dirty: comment
		if fn.Doc != nil {
			for _, comment := range fn.Doc.List {
				if effects := ParseEffects(comment.Text); effects != nil {
					info.HasDeclaration = true
					info.DeclaredEffects = NewStringSet(effects...)
					info.ComputedEffects = NewStringSet(effects...)
					break
				}
			}
		}

		ea.Functions[funcName] = info
		ea.Resolver.AddLocalFunction(funcName, info)
	})
}

// BuildCallGraph analyzes function bodies to build the call graph
func (ea *EffectAnalysis) BuildCallGraph() {
	for funcName, info := range ea.Functions {
		// Analyze function body for calls
		ast.Inspect(info.Decl, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			// Extract called function name
			var calleeName string
			switch fun := call.Fun.(type) {
			case *ast.Ident:
				calleeName = fun.Name
			case *ast.SelectorExpr:
				// Handle method calls
				calleeName = fun.Sel.Name
			}

			if calleeName != "" {
				// Check if the called function is in our analysis
				if _, exists := ea.Functions[calleeName]; exists {
					info.CallSites = append(info.CallSites, CallSite{
						Callee:   calleeName,
						Position: call.Pos(),
					})
					ea.CallGraph.AddCall(funcName, calleeName, call.Pos())
				}

				// Always check JSON effects, even if function exists
				if ea.JSONEffects != nil {
					if effectExpr, ok := ea.JSONEffects[calleeName]; ok {
						// Evaluate the effect expression
						effectSet, err := effectExpr.Eval(nil)
						if err == nil {
							// If function already exists, update its effects
							if existingFunc, exists := ea.Functions[calleeName]; exists {
								// Only update if it doesn't have a declaration
								if !existingFunc.HasDeclaration {
									existingFunc.DeclaredEffects = effectSet
									existingFunc.ComputedEffects = effectSet
									existingFunc.HasDeclaration = true // Treat JSON as declaration
								}
							} else {
								// Create a synthetic function info for JSON function
								jsonFunc := &FunctionInfo{
									Name:            calleeName,
									Package:         ea.Pass.Pkg.Path(),
									DeclaredEffects: effectSet,
									ComputedEffects: effectSet,
									HasDeclaration:  true, // Treat as if it has declaration
								}
								ea.Functions[calleeName] = jsonFunc

								// Add to call graph
								info.CallSites = append(info.CallSites, CallSite{
									Callee:   calleeName,
									Position: call.Pos(),
								})
								ea.CallGraph.AddCall(funcName, calleeName, call.Pos())
							}
						}
					}
				}
			}

			return true
		})
	}
}

// PropagateEffects computes implicit effects using a worklist algorithm
func (ea *EffectAnalysis) PropagateEffects() {
	// Initialize worklist with all functions
	worklist := make([]string, 0, len(ea.Functions))
	inWorklist := make(map[string]bool)

	for name := range ea.Functions {
		worklist = append(worklist, name)
		inWorklist[name] = true
	}

	// Process until worklist is empty
	for len(worklist) > 0 {
		// Pop from worklist
		funcName := worklist[len(worklist)-1]
		worklist = worklist[:len(worklist)-1]
		inWorklist[funcName] = false

		fn := ea.Functions[funcName]
		oldEffects := fn.ComputedEffects.Clone()

		// Collect effects from all called functions
		for _, call := range fn.CallSites {
			if callee, ok := ea.Functions[call.Callee]; ok {
				fn.ComputedEffects.AddAll(callee.ComputedEffects)
			}
		}

		// If effects changed, add callers to worklist
		if !oldEffects.Equals(fn.ComputedEffects) {
			for _, caller := range ea.CallGraph.CalledBy[funcName] {
				if !inWorklist[caller] {
					worklist = append(worklist, caller)
					inWorklist[caller] = true
				}
			}
		}
	}
}

// CheckEffects verifies that declared effects match computed effects
func (ea *EffectAnalysis) CheckEffects() {
	for _, fn := range ea.Functions {
		// Only check functions with explicit declarations
		if !fn.HasDeclaration {
			continue
		}

		// Check each call site
		for _, call := range fn.CallSites {
			if callee, ok := ea.Functions[call.Callee]; ok {
				// Check if called function's effects are declared
				if !callee.ComputedEffects.IsSubsetOf(fn.DeclaredEffects) {
					missingEffects := callee.ComputedEffects.Difference(fn.DeclaredEffects)

					// Build detailed error
					err := &EffectError{
						CallSite:       call.Position,
						Caller:         fn.Name,
						Callee:         call.Callee,
						CallerEffects:  fn.DeclaredEffects.ToSlice(),
						CalleeEffects:  callee.ComputedEffects.ToSlice(),
						MissingEffects: missingEffects.ToSlice(),
					}

					// Add propagation path if callee has no declaration
					if !callee.HasDeclaration {
						visited := make(map[string]bool)
						err.PropagationPath = BuildPropagationPath(call.Callee, ea.Functions, visited)
					}

					// Check if verbose mode is enabled
					if os.Getenv("DIRTY_VERBOSE") == "1" {
						// Use detailed error format
						ea.Pass.Report(analysis.Diagnostic{
							Pos:     call.Position,
							Message: err.Format(),
						})
					} else {
						// Use simple format
						ea.Pass.Reportf(call.Position,
							"function calls %s which has effects [%s] not declared in this function",
							call.Callee, joinEffects(callee.ComputedEffects.ToSlice()))
					}
				}
			}
		}
	}
}

// joinEffects joins effect strings for error messages
func joinEffects(effects []string) string {
	result := ""
	for i, effect := range effects {
		if i > 0 {
			result += ", "
		}
		result += effect
	}
	return result
}
