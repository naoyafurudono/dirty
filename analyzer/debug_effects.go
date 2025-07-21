package analyzer

import "os"

// debugCheckEffects prints detailed information during effect checking
func (ea *EffectAnalysis) debugCheckEffects() {
	if os.Getenv("DIRTY_VERBOSE") != "1" {
		return
	}

	debugLog("=== Checking Effects ===")
	for fname, fn := range ea.Functions {
		if !fn.HasDeclaration {
			debugLog("Skipping %s (no declaration)", fname)
			continue
		}

		debugLog("Checking function: %s", fname)
		debugLog("  Declared effects: %v", fn.DeclaredEffects.ToSlice())
		debugLog("  Computed effects: %v", fn.ComputedEffects.ToSlice())
		debugLog("  Call sites: %d", len(fn.CallSites))

		for _, call := range fn.CallSites {
			debugLog("    Call to: %s at %v", call.Callee, call.Position)
			if callee, ok := ea.Functions[call.Callee]; ok {
				debugLog("      Callee effects: %v", callee.ComputedEffects.ToSlice())

				// Check if effects are missing
				missingEffects := NewStringSet()
				for effect := range callee.ComputedEffects {
					if !fn.DeclaredEffects.Contains(effect) {
						missingEffects.Add(effect)
					}
				}

				if len(missingEffects) > 0 {
					debugLog("      MISSING EFFECTS: %v", missingEffects.ToSlice())
				} else {
					debugLog("      All effects declared")
				}
			} else {
				debugLog("      Callee not found in function map")
			}
		}
	}
}
