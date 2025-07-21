package analyzer

import (
	"fmt"
	"os"
)

// debugLog prints debug information if DIRTY_VERBOSE is set
func debugLog(format string, args ...interface{}) {
	if os.Getenv("DIRTY_VERBOSE") == "1" {
		fmt.Fprintf(os.Stderr, "[DIRTY DEBUG] "+format+"\n", args...)
	}
}

// debugPackageFacts prints all facts for a package
func (ea *EffectAnalysis) debugPackageFacts() {
	if os.Getenv("DIRTY_VERBOSE") != "1" {
		return
	}

	debugLog("=== Package: %s ===", ea.Pass.Pkg.Path())

	// Debug imports
	debugLog("Imports:")
	for _, imp := range ea.Pass.Pkg.Imports() {
		debugLog("  - %s", imp.Path())

		// Try to import facts from this package
		var fact PackageEffectsFact
		if ea.Pass.ImportPackageFact(imp, &fact) {
			debugLog("    Found PackageEffectsFact with %d functions", len(fact.FunctionEffects))
			for fname, effects := range fact.FunctionEffects {
				debugLog("      %s: %v", fname, effects)
			}
		} else {
			debugLog("    No PackageEffectsFact found")
		}
	}

	// Debug functions in current package
	debugLog("Functions in current package:")
	for fname, info := range ea.Functions {
		debugLog("  - %s: computed=%v, declared=%v",
			fname,
			info.ComputedEffects.ToSlice(),
			info.DeclaredEffects.ToSlice())
	}
}
