package analyzer

import (
	"fmt"
	"strings"
)

// EffectExpr is the base interface for effect expressions
type EffectExpr interface {
	// Eval evaluates the expression and returns a set of effect labels
	Eval(resolver EffectResolver) (StringSet, error)
	// String returns a string representation for debugging
	String() string
}

// EffectLabel represents a single effect label (leaf node)
// e.g., select[users], insert[logs]
type EffectLabel struct {
	Operation string // "select", "insert", "update", "delete"
	Target    string // "users", "logs", etc.
}

// Eval returns a set containing just this effect label
func (e *EffectLabel) Eval(_ EffectResolver) (StringSet, error) {
	return NewStringSet(e.String()), nil
}

func (e *EffectLabel) String() string {
	if e.Target == "" {
		return e.Operation
	}
	return fmt.Sprintf("%s[%s]", e.Operation, e.Target)
}

// LiteralSet represents a literal set of effects
// e.g., { a | b | c }
type LiteralSet struct {
	Elements []EffectExpr // Slice of elements in the set
}

// Eval evaluates the literal set and returns the union of all elements
func (s *LiteralSet) Eval(resolver EffectResolver) (StringSet, error) {
	result := NewStringSet()
	for _, elem := range s.Elements {
		set, err := elem.Eval(resolver)
		if err != nil {
			return nil, err
		}
		result = result.Union(set)
	}
	return result, nil
}

func (s *LiteralSet) String() string {
	if len(s.Elements) == 0 {
		return "{ }"
	}
	parts := make([]string, len(s.Elements))
	for i, elem := range s.Elements {
		parts[i] = elem.String()
	}
	return fmt.Sprintf("{ %s }", strings.Join(parts, " | "))
}

// EffectRef represents a reference to a named effect (Phase 2)
// e.g., userOps
type EffectRef struct {
	Name string
}

// Eval evaluates the effect reference using the resolver
func (r *EffectRef) Eval(resolver EffectResolver) (StringSet, error) {
	if resolver == nil {
		return nil, fmt.Errorf("cannot resolve effect reference '%s': no resolver provided", r.Name)
	}
	return resolver.Resolve(r.Name)
}

func (r *EffectRef) String() string {
	return r.Name
}

// EffectResolver is the interface for resolving named effects
type EffectResolver interface {
	Resolve(name string) (StringSet, error)
}

// NilResolver is a resolver that always returns an error
type NilResolver struct{}

// Resolve always returns an error for any name
func (NilResolver) Resolve(name string) (StringSet, error) {
	return nil, fmt.Errorf("effect reference '%s' not supported in Phase 1", name)
}
