package main

import (
	"github.com/naoyafurudono/dirty/analyzer"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() {
	// unitcheckerはgo vetと同じインターフェースを提供
	unitchecker.Main(analyzer.Analyzer)
}