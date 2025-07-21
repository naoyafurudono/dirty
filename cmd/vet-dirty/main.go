package main

import (
	"github.com/naoyafurudono/dirty/analyzer"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() {
	// unitchecker provides the same interface as go vet
	unitchecker.Main(analyzer.Analyzer)
}
