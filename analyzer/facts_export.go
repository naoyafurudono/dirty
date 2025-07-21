package analyzer

import (
	"go/ast"
	"go/types"
)

// ExportPackageEffects exports the effect information for the current package
// as a fact that can be imported by dependent packages.
func (ea *EffectAnalysis) ExportPackageEffects() {
	// Create package fact with all function effects
	packageFact := &PackageEffectsFact{
		FunctionEffects: make(map[string][]string),
	}

	// Collect effects for all functions in the package
	for funcName, info := range ea.Functions {
		if info.ComputedEffects == nil {
			continue
		}

		// Store effects as sorted slice
		effects := info.ComputedEffects.ToSlice()
		if len(effects) > 0 {
			packageFact.FunctionEffects[funcName] = effects
		}

		// Also export individual function facts for direct object queries
		if info.Decl != nil && info.Decl.Name != nil {
			funcObj := ea.Pass.TypesInfo.Defs[info.Decl.Name]
			if funcObj != nil {
				funcFact := &FunctionEffectsFact{
					Effects: effects,
				}
				ea.Pass.ExportObjectFact(funcObj, funcFact)
			}
		}
	}

	// Export the package fact
	if len(packageFact.FunctionEffects) > 0 {
		ea.Pass.ExportPackageFact(packageFact)
	}
}

// ImportPackageEffects imports effect information from a specific package
func (ea *EffectAnalysis) ImportPackageEffects(pkg *types.Package) map[string][]string {
	var packageFact PackageEffectsFact
	if ea.Pass.ImportPackageFact(pkg, &packageFact) {
		return packageFact.FunctionEffects
	}
	return nil
}

// ImportFunctionEffects imports effect information for a specific function object
func (ea *EffectAnalysis) ImportFunctionEffects(obj types.Object) []string {
	var funcFact FunctionEffectsFact
	if ea.Pass.ImportObjectFact(obj, &funcFact) {
		return funcFact.Effects
	}
	return nil
}

// ResolveCrossPackageCall resolves effects for a cross-package function call
func (ea *EffectAnalysis) ResolveCrossPackageCall(call *ast.CallExpr, pkgPath string, funcName string) (StringSet, bool) {
	// First, try to find the package in imports
	var targetPkg *types.Package
	for _, imp := range ea.Pass.Pkg.Imports() {
		if imp.Path() == pkgPath {
			targetPkg = imp
			break
		}
	}

	if targetPkg == nil {
		return nil, false
	}

	// Try to get effects from package fact
	if effects := ea.ImportPackageEffects(targetPkg); effects != nil {
		if funcEffects, ok := effects[funcName]; ok {
			return NewStringSetFromSlice(funcEffects), true
		}
	}

	// Try to resolve the function object directly
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		if obj := ea.Pass.TypesInfo.Uses[sel.Sel]; obj != nil {
			if effects := ea.ImportFunctionEffects(obj); effects != nil {
				return NewStringSetFromSlice(effects), true
			}
		}
	}

	return nil, false
}
