package analyzer

import (
	"encoding/gob"
	"fmt"
)

// PackageEffectsFact holds effect information for all functions in a package.
// This fact is exported by each package and imported by dependent packages
// to enable cross-package effect analysis.
type PackageEffectsFact struct {
	// Map from function name to its effects
	// Key format: "FunctionName" for functions, "(*Type).Method" for methods
	FunctionEffects map[string][]string
}

// AFact marks PackageEffectsFact as a fact type for the analysis framework
func (*PackageEffectsFact) AFact() {}

// String returns a human-readable representation of the package effects
func (f *PackageEffectsFact) String() string {
	return fmt.Sprintf("PackageEffectsFact{%d functions}", len(f.FunctionEffects))
}

// FunctionEffectsFact holds effect information for a specific function.
// This is attached to individual function objects.
type FunctionEffectsFact struct {
	Effects []string
}

// AFact marks FunctionEffectsFact as a fact type
func (*FunctionEffectsFact) AFact() {}

// String returns a human-readable representation of the function effects
func (f *FunctionEffectsFact) String() string {
	return fmt.Sprintf("FunctionEffectsFact%v", f.Effects)
}

// Register fact types for gob encoding/decoding
func init() {
	gob.Register(&PackageEffectsFact{})
	gob.Register(&FunctionEffectsFact{})
}
