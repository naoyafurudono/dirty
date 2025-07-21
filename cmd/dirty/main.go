// Package main provides the main dirty effect analyzer CLI tool
package main

import (
	"github.com/naoyafurudono/dirty/analyzer"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(analyzer.Analyzer)
}
