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
		os.Setenv("DIRTY_VERBOSE", "1")
	}
	
	// Print usage if verbose
	if *verbose && len(flag.Args()) == 0 {
		fmt.Fprintf(os.Stderr, "dirty-verbose: verbose effect analyzer\n")
		fmt.Fprintf(os.Stderr, "Usage: dirty-verbose [-v] package...\n")
		os.Exit(1)
	}
	
	singlechecker.Main(analyzer.Analyzer)
}