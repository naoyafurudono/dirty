// Package main implements the dirty effect analyzer CLI
package main

import (
	"fmt"
	"github.com/naoyafurudono/dirty/analyzer"
	"golang.org/x/tools/go/analysis/singlechecker"
	"os"
)

func main() {
	singlechecker.Main(analyzer.Analyzer)
}
