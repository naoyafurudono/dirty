package analyzer

import (
	"go/types"
)

// ImportAllPackageEffects imports effects from all imported packages into the resolver
func (ea *EffectAnalysis) ImportAllPackageEffects() {
	// Skip if Facts are disabled
	if ea.DisableFacts {
		return
	}

	// Import effects from all imported packages
	for _, pkg := range ea.Pass.Pkg.Imports() {
		ea.ImportPackageEffectsIntoResolver(pkg)
	}
}

// ImportPackageEffectsIntoResolver imports effects from a package into the resolver
func (ea *EffectAnalysis) ImportPackageEffectsIntoResolver(pkg *types.Package) {
	var packageFact PackageEffectsFact
	if !ea.Pass.ImportPackageFact(pkg, &packageFact) {
		return
	}

	// Add all function effects to the resolver
	for funcName, effects := range packageFact.FunctionEffects {
		// Construct the full qualified name
		qualifiedName := pkg.Path() + "." + funcName
		ea.Resolver.AddImportedEffects(qualifiedName, NewStringSetFromSlice(effects))
	}
}
