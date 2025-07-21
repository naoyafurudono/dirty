// Command dirty-cross demonstrates cross-package analysis
package main

import (
	"github.com/naoyafurudono/dirty/analyzer"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(analyzer.Analyzer)
}
