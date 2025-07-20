package analyzer

import (
	"go/ast"
	"go/token"
	"sort"
)

// FunctionInfo represents information about a function
type FunctionInfo struct {
	Name            string
	Package         string
	DeclaredEffects StringSet // Effects declared via //dirty: comment
	ComputedEffects StringSet // Actual effects including those from called functions
	HasDeclaration  bool      // Whether function has //dirty: comment
	Decl            *ast.FuncDecl
	CallSites       []CallSite // Functions called by this function
}

// CallSite represents a function call location
type CallSite struct {
	Callee   string
	Position token.Pos
}

// CallGraph represents the function call relationships
type CallGraph struct {
	// Calls[A] = [B, C] means function A calls functions B and C
	Calls map[string][]CallSite
	// CalledBy[B] = [A] means function B is called by function A
	CalledBy map[string][]string
}

// NewCallGraph creates a new CallGraph
func NewCallGraph() *CallGraph {
	return &CallGraph{
		Calls:    make(map[string][]CallSite),
		CalledBy: make(map[string][]string),
	}
}

// AddCall records that caller calls callee at position
func (g *CallGraph) AddCall(caller, callee string, pos token.Pos) {
	g.Calls[caller] = append(g.Calls[caller], CallSite{
		Callee:   callee,
		Position: pos,
	})

	// Update reverse mapping
	found := false
	for _, c := range g.CalledBy[callee] {
		if c == caller {
			found = true
			break
		}
	}
	if !found {
		g.CalledBy[callee] = append(g.CalledBy[callee], caller)
	}
}

// StringSet represents a set of strings
type StringSet map[string]struct{}

// NewStringSet creates a new StringSet
func NewStringSet(items ...string) StringSet {
	s := make(StringSet)
	for _, item := range items {
		s[item] = struct{}{}
	}
	return s
}

// Add adds an item to the set
func (s StringSet) Add(item string) {
	s[item] = struct{}{}
}

// AddAll adds all items from another set
func (s StringSet) AddAll(other StringSet) {
	for item := range other {
		s[item] = struct{}{}
	}
}

// Contains checks if an item is in the set
func (s StringSet) Contains(item string) bool {
	_, ok := s[item]
	return ok
}

// Clone creates a copy of the set
func (s StringSet) Clone() StringSet {
	clone := make(StringSet)
	for item := range s {
		clone[item] = struct{}{}
	}
	return clone
}

// Equals checks if two sets are equal
func (s StringSet) Equals(other StringSet) bool {
	if len(s) != len(other) {
		return false
	}
	for item := range s {
		if !other.Contains(item) {
			return false
		}
	}
	return true
}

// ToSlice converts the set to a sorted slice
func (s StringSet) ToSlice() []string {
	result := make([]string, 0, len(s))
	for item := range s {
		result = append(result, item)
	}
	sort.Strings(result)
	return result
}

// IsSubsetOf checks if this set is a subset of another
func (s StringSet) IsSubsetOf(other StringSet) bool {
	for item := range s {
		if !other.Contains(item) {
			return false
		}
	}
	return true
}

// Difference returns items in s that are not in other
func (s StringSet) Difference(other StringSet) StringSet {
	diff := make(StringSet)
	for item := range s {
		if !other.Contains(item) {
			diff[item] = struct{}{}
		}
	}
	return diff
}
