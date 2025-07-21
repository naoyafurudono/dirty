package main

import (
	"github.com/naoyafurudono/dirty/testdata/src/json_conflict/a"
	"github.com/naoyafurudono/dirty/testdata/src/json_conflict/b"
)

// dirty: { print }
func main() {
	// Both packages have Process function
	a.Process()
	b.Process()
}
