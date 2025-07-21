// Package main provides the verbose dirty effect analyzer CLI tool
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/naoyafurudono/dirty/analyzer"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	// Add verbose flag
	verbose := flag.Bool("v", false, "verbose error messages")
	flag.Parse()

	if *verbose {
		// Set environment variable for analyzer to detect
		if err := os.Setenv("DIRTY_VERBOSE", "1"); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to set DIRTY_VERBOSE: %v\n", err)
		}
	}

	// Print usage if verbose
	if *verbose && len(flag.Args()) == 0 {
		fmt.Fprintf(os.Stderr, "dirty-verbose: verbose effect analyzer\n")
		fmt.Fprintf(os.Stderr, "Usage: dirty-verbose [-v] package...\n")
		os.Exit(1)
	}

	singlechecker.Main(analyzer.Analyzer)
}
