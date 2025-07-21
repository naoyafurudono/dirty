package analyzer

// EffectSource indicates where effect information came from
type EffectSource int

const (
	// SourceUnknown indicates the function is not found
	SourceUnknown EffectSource = iota
	// SourceLocal indicates effects from current package analysis
	SourceLocal
	// SourceFacts indicates effects from imported Facts
	SourceFacts
	// SourceJSON indicates effects from JSON declarations
	SourceJSON
)

// UnifiedEffectResolver provides a unified interface for resolving function effects
// from various sources (local analysis, Facts, JSON declarations)
type UnifiedEffectResolver struct {
	// localEffects contains effects for functions in the current package
	localEffects map[string]*FunctionInfo

	// importedFacts contains effects imported from other packages via Facts
	importedFacts map[string]StringSet

	// jsonEffects contains effects declared in JSON files
	jsonEffects ParsedEffects
}

// NewUnifiedEffectResolver creates a new unified effect resolver
func NewUnifiedEffectResolver() *UnifiedEffectResolver {
	return &UnifiedEffectResolver{
		localEffects:  make(map[string]*FunctionInfo),
		importedFacts: make(map[string]StringSet),
		jsonEffects:   make(ParsedEffects),
	}
}

// ResolveEffects returns the effects for a function from any available source
// Priority: local > facts > json
func (r *UnifiedEffectResolver) ResolveEffects(funcName string) (StringSet, EffectSource) {
	// Check local effects first (current package)
	if info, ok := r.localEffects[funcName]; ok {
		return info.ComputedEffects, SourceLocal
	}

	// Check imported Facts
	if effects, ok := r.importedFacts[funcName]; ok {
		return effects, SourceFacts
	}

	// Check JSON declarations
	if effectExpr, ok := r.jsonEffects[funcName]; ok {
		effects, err := effectExpr.Eval(nil)
		if err == nil {
			return effects, SourceJSON
		}
	}

	// Function not found
	return nil, SourceUnknown
}

// AddLocalFunction adds a function from the current package
func (r *UnifiedEffectResolver) AddLocalFunction(funcName string, info *FunctionInfo) {
	r.localEffects[funcName] = info
}

// AddImportedEffects adds effects imported from Facts
func (r *UnifiedEffectResolver) AddImportedEffects(funcName string, effects StringSet) {
	r.importedFacts[funcName] = effects
}

// SetJSONEffects sets the JSON effect declarations
func (r *UnifiedEffectResolver) SetJSONEffects(effects ParsedEffects) {
	r.jsonEffects = effects
}

// GetLocalFunctions returns all functions in the current package
func (r *UnifiedEffectResolver) GetLocalFunctions() map[string]*FunctionInfo {
	return r.localEffects
}

// HasEffects checks if effects are available for a function from any source
func (r *UnifiedEffectResolver) HasEffects(funcName string) bool {
	_, source := r.ResolveEffects(funcName)
	return source != SourceUnknown
}
