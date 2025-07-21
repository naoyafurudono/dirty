package analyzer

import (
	"go/ast"
	"path/filepath"
	"strings"
)

// EnhanceWithCrossPackageSupport adds cross-package analysis capabilities
func EnhanceWithCrossPackageSupport(ea *EffectAnalysis) {
	// Analyze imports for all files
	imports := make(map[string]string)
	for _, file := range ea.Pass.Files {
		for _, imp := range file.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			alias := ""
			if imp.Name != nil {
				alias = imp.Name.Name
			} else {
				// Use the last component of the path as the alias
				alias = filepath.Base(path)
			}
			imports[alias] = path
		}
	}

	// Get current package path (unused for now but will be needed later)
	// packagePath := ""
	// if ea.Pass.Pkg != nil {
	// 	packagePath = ea.Pass.Pkg.Path()
	// }

	// Re-analyze function bodies to find cross-package calls
	for funcName, info := range ea.Functions {
		if info.Decl == nil {
			continue
		}
		ast.Inspect(info.Decl, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			// Resolve cross-package calls
			var resolvedName string
			switch fun := call.Fun.(type) {
			case *ast.SelectorExpr:
				if ident, ok := fun.X.(*ast.Ident); ok {
					// Check if it's an imported package
					if pkgPath, isImport := imports[ident.Name]; isImport {
						// It's an imported package function
						resolvedName = pkgPath + "." + fun.Sel.Name

						// Add to call graph
						ea.CallGraph.AddCall(funcName, resolvedName, call.Pos())

						// Look up effects from JSON
						var effects StringSet = NewStringSet()
						if ea.JSONEffects != nil {
							// Try with full path
							if effectExpr, ok := ea.JSONEffects[resolvedName]; ok {
								if evalEffects, err := effectExpr.Eval(nil); err == nil {
									effects = evalEffects
								}
							} else if effectExpr, ok := ea.JSONEffects[fun.Sel.Name]; ok {
								// Try with just function name for backward compatibility
								if evalEffects, err := effectExpr.Eval(nil); err == nil {
									effects = evalEffects
								}
							}
						}

						// Create a synthetic function info for the imported function
						if _, exists := ea.Functions[resolvedName]; !exists {
							ea.Functions[resolvedName] = &FunctionInfo{
								Name:            resolvedName,
								Package:         pkgPath,
								DeclaredEffects: effects,
								ComputedEffects: effects,
								HasDeclaration:  true, // Treat as declared since it's from JSON
								CallSites:       []CallSite{},
							}
						}
					}
				}
			}

			return true
		})
	}
}
