package analyzer

import (
	"encoding/json"
	"fmt"
	"os"
)

// EffectDeclarations represents JSON-based effect declarations
type EffectDeclarations struct {
	Version string            `json:"version"`
	Effects map[string]string `json:"effects"`
}

// LoadEffectDeclarations loads effect declarations from JSON
func LoadEffectDeclarations(path string) (*EffectDeclarations, error) {
	// #nosec G304 - path is controlled by DIRTY_EFFECTS_JSON env var or package directory
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var decls EffectDeclarations
	if err := json.Unmarshal(data, &decls); err != nil {
		return nil, err
	}

	// Validate version
	if decls.Version != "1.0" {
		return nil, fmt.Errorf("unsupported version: %s", decls.Version)
	}

	return &decls, nil
}

// ParsedEffects holds parsed effect expressions
type ParsedEffects map[string]EffectExpr

// ParseAll parses all effect declarations
func (d *EffectDeclarations) ParseAll() (ParsedEffects, error) {
	result := make(ParsedEffects)
	for funcName, effectStr := range d.Effects {
		expr, err := ParseEffectDecl("// dirty: " + effectStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing effects for %s: %w", funcName, err)
		}
		result[funcName] = expr
	}
	return result, nil
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
