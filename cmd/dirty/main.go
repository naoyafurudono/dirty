// Package main implements the dirty effect analyzer CLI
package main

import (
	"github.com/naoyafurudono/dirty/analyzer"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(analyzer.Analyzer)
}
