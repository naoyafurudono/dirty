package analyzer

import (
	"go/ast"
	
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ast/inspector"
)

// EffectAnalysis holds the state for effect analysis
type EffectAnalysis struct {
	Pass      *analysis.Pass
	Inspector *inspector.Inspector
	Functions map[string]*FunctionInfo
	CallGraph *CallGraph
}

// NewEffectAnalysis creates a new EffectAnalysis
func NewEffectAnalysis(pass *analysis.Pass, inspect *inspector.Inspector) *EffectAnalysis {
	return &EffectAnalysis{
		Pass:      pass,
		Inspector: inspect,
		Functions: make(map[string]*FunctionInfo),
		CallGraph: NewCallGraph(),
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
		
		// Extract effects from //dirty: comment
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
					ea.Pass.Reportf(call.Position,
						"function calls %s which has effects [%s] not declared in this function",
						call.Callee, joinEffects(callee.ComputedEffects.ToSlice()))
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