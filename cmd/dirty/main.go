// Package main implements the dirty effect analyzer CLI
package main

import (
	"os"

	"github.com/naoyafurudono/dirty/analyzer"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	// Check if running in multi-package mode
	if len(os.Args) > 1 && (containsPattern(os.Args[1:], "...") || len(os.Args) > 2) {
		// Use multichecker for multiple packages
		multichecker.Main(analyzer.Analyzer)
	} else {
		// Use singlechecker for single package
		singlechecker.Main(analyzer.Analyzer)
	}
}

func containsPattern(args []string, pattern string) bool {
	for _, arg := range args {
		if arg == pattern || (len(arg) > 3 && arg[len(arg)-3:] == pattern) {
			return true
		}
	}
	return false
}
