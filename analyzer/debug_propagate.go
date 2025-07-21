package analyzer

import "os"

// debugPropagateEffects prints detailed information during effect propagation
func (ea *EffectAnalysis) debugPropagateEffects() {
	if os.Getenv("DIRTY_VERBOSE") != "1" {
		return
	}

	debugLog("=== Propagating Effects ===")

	// Show initial state
	debugLog("Initial function effects:")
	for fname, fn := range ea.Functions {
		debugLog("  %s: declared=%v, computed=%v",
			fname,
			fn.DeclaredEffects.ToSlice(),
			fn.ComputedEffects.ToSlice())
	}

	// Show call graph
	debugLog("Call graph:")
	for caller, fn := range ea.Functions {
		if len(fn.CallSites) > 0 {
			debugLog("  %s calls:", caller)
			for _, call := range fn.CallSites {
				debugLog("    -> %s", call.Callee)
			}
		}
	}
}
