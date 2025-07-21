package analyzer

// NewStringSetFromSlice creates a StringSet from a slice of strings
func NewStringSetFromSlice(items []string) StringSet {
	return NewStringSet(items...)
}
