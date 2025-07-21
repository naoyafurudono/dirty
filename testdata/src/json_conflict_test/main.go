package main

import (
	"json_conflict/a"
	"json_conflict/b"
)

// dirty: { print }
func TestConflict() {
	// Test with DIRTY_EFFECTS_JSON pointing to our JSON file
	a.Process()
	b.Process()
}
