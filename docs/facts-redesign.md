# Facts Redesign: Separating Information Sharing from Diagnostics

## Executive Summary

This document analyzes the current use of Facts in the dirty analyzer and proposes a redesign where Facts are used purely for internal information sharing between packages, not for generating diagnostics. This aligns with the pattern used by Go's built-in analyzers like printf.

## Current State Analysis

### How Facts are Currently Used

1. **Export Phase** (`ExportPackageEffects` in `facts_export.go`):
   - Exports `PackageEffectsFact` containing all function effects in a package
   - Also exports individual `FunctionEffectsFact` for each function
   - Only exports if effects are non-empty

2. **Import Phase** (`ResolveCrossPackageCall` in `facts_export.go`):
   - When analyzing a cross-package call, attempts to import Facts
   - First tries package-level Facts, then object-level Facts
   - Returns effects if found, or falls back to JSON declarations

3. **Diagnostic Generation** (`CheckEffects` in `effect_analysis.go`):
   - When a function calls another with undeclared effects
   - Generates diagnostic messages directly based on computed effects
   - **Key issue**: The diagnostic depends on whether effects were found via Facts or JSON

### Problems with Current Approach

1. **Facts Availability Affects Diagnostics**:
   ```go
   // In crosspackage_v3.go
   if foundInFacts {
       // Create synthetic function info
       // This affects whether diagnostics are generated
   }
   ```

2. **Test Fragility**:
   - Tests must explicitly enable/disable Facts
   - Cross-package tests fail without Facts
   - Behavior differs between Facts and JSON fallback

3. **Coupling**:
   - Diagnostic generation is tightly coupled to Facts availability
   - The `foundInFacts` flag determines program behavior

## Printf Analyzer Pattern

The printf analyzer demonstrates the correct pattern:

1. **Facts for Information Only**:
   ```go
   type isWrapper struct{ Kind Kind }
   func (*isWrapper) AFact() {}
   ```
   - Facts identify wrapper functions
   - Used to propagate wrapper information across packages

2. **Diagnostics Independent of Facts**:
   ```go
   func checkPrintf(pass *analysis.Pass, ...) {
       // Diagnostics based on local analysis
       // Facts only help identify what to check
   }
   ```
   - Diagnostics are generated based on the actual code being analyzed
   - Facts only influence *what* to check, not *whether* to report

3. **Graceful Degradation**:
   - If Facts aren't available, analysis still works
   - May miss some wrapper functions, but core functionality remains

## Proposed New Architecture

### Design Principles

1. **Facts are Optional Optimization**:
   - Analysis must work correctly without Facts
   - Facts improve accuracy but don't change core behavior

2. **Diagnostics from Local Analysis**:
   - All diagnostics based on information available in current package
   - Cross-package information enhances but doesn't determine diagnostics

3. **Clear Separation of Concerns**:
   - Facts: Share effect information between packages
   - JSON: Declare effects for external libraries
   - Diagnostics: Based on effect mismatches in current code

### Implementation Design

#### 1. Effect Resolution Layer

```go
// EffectResolver handles all effect lookups
type EffectResolver struct {
    localEffects  map[string]*FunctionInfo  // Current package
    importedFacts map[string]StringSet       // From Facts
    jsonEffects   ParsedEffects             // From JSON
}

// ResolveEffects returns effects for any function
func (r *EffectResolver) ResolveEffects(funcName string) (StringSet, EffectSource) {
    // Priority: local > facts > json
    if local, ok := r.localEffects[funcName]; ok {
        return local.ComputedEffects, SourceLocal
    }
    if facts, ok := r.importedFacts[funcName]; ok {
        return facts, SourceFacts
    }
    if json, ok := r.jsonEffects[funcName]; ok {
        effects, _ := json.Eval(nil)
        return effects, SourceJSON
    }
    return nil, SourceUnknown
}
```

#### 2. Analysis Flow

```go
func AnalyzePackage(pass *analysis.Pass) {
    resolver := NewEffectResolver()
    
    // Phase 1: Import available Facts (best effort)
    resolver.ImportAvailableFacts(pass)
    
    // Phase 2: Load JSON declarations
    resolver.LoadJSONEffects()
    
    // Phase 3: Analyze current package
    for _, fn := range functions {
        // Compute effects using resolver
        effects := computeEffects(fn, resolver)
        resolver.AddLocalFunction(fn.Name, effects)
    }
    
    // Phase 4: Check violations (always runs)
    for _, fn := range functions {
        checkEffectViolations(fn, resolver)
    }
    
    // Phase 5: Export Facts (best effort)
    exportFacts(pass, resolver.localEffects)
}
```

#### 3. Diagnostic Generation

```go
func checkEffectViolations(fn *FunctionInfo, resolver *EffectResolver) {
    for _, call := range fn.Calls {
        calleeEffects, source := resolver.ResolveEffects(call.Callee)
        
        if calleeEffects == nil {
            // Unknown function - skip or warn
            continue
        }
        
        missingEffects := calleeEffects.Difference(fn.DeclaredEffects)
        if !missingEffects.IsEmpty() {
            // Generate diagnostic
            // Note: diagnostic is same regardless of source
            reportMissingEffects(fn, call, missingEffects)
        }
    }
}
```

### Key Changes from Current Implementation

1. **Remove `foundInFacts` Flag**:
   - Don't track where effects came from for diagnostics
   - Use unified effect resolution

2. **Always Create Function Info**:
   - Create synthetic function info for all resolved functions
   - Don't condition on Facts availability

3. **Decouple Cross-Package Analysis**:
   - Move from `EnhanceWithCrossPackageSupport` to integrated resolver
   - Cross-package is just another source of effects

4. **Simplify Test Infrastructure**:
   - Remove `DisableFacts` flag
   - Tests work the same with or without Facts
   - Use JSON for predictable cross-package testing

## Implementation Plan

### Phase 1: Refactor Effect Resolution (Low Risk)
1. Create `EffectResolver` interface
2. Move resolution logic from various places into resolver
3. Keep existing Facts import/export

### Phase 2: Decouple Diagnostics (Medium Risk)
1. Remove `foundInFacts` conditionals
2. Always create function info for resolved effects
3. Update tests to not depend on Facts

### Phase 3: Simplify Architecture (Higher Risk)
1. Remove `DisableFacts` flag
2. Merge cross-package analysis into main flow
3. Clean up test infrastructure

### Phase 4: Optimize (Enhancement)
1. Cache effect resolutions
2. Lazy Facts loading
3. Parallel analysis where possible

## Testing Strategy

1. **Unit Tests**:
   - Test `EffectResolver` in isolation
   - Mock different effect sources

2. **Integration Tests**:
   - Test with Facts available
   - Test without Facts (JSON only)
   - Ensure same diagnostics in both cases

3. **Cross-Package Tests**:
   - Use JSON declarations for predictable testing
   - Optionally verify Facts optimization works

## Migration Guide

For users:
- No changes required
- Existing JSON declarations continue to work
- Facts remain an internal optimization

For developers:
- Update any code that checks `foundInFacts`
- Use `EffectResolver` for all lookups
- Don't assume Facts are available

## Benefits

1. **Robustness**: Analysis works correctly regardless of Facts availability
2. **Simplicity**: Clear separation between information sharing and diagnostics
3. **Testability**: Tests don't need complex Facts setup
4. **Maintainability**: Easier to understand and modify
5. **Performance**: Facts remain as optimization without affecting correctness

## Risks and Mitigations

1. **Risk**: Breaking existing behavior
   - **Mitigation**: Phased implementation with extensive testing

2. **Risk**: Performance regression without Facts
   - **Mitigation**: JSON declarations provide fallback; Facts remain for optimization

3. **Risk**: Complex migration
   - **Mitigation**: Compatibility layer during transition

## Conclusion

By following the printf analyzer pattern and separating Facts from diagnostics, we can create a more robust and maintainable cross-package analysis system. Facts become purely an optimization for sharing information between packages, while diagnostics are always based on the actual effect mismatches detected during analysis.