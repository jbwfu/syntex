package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jbwfu/syntex/internal/packer"
)

func main() {
	flag.Usage = func() {
		progName := filepath.Base(os.Args[0])
		fmt.Fprintf(os.Stderr, "Usage: %s <directory_or_file>\n", progName)
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	targetPath := flag.Arg(0)

	// Setup dependencies
	formatter := packer.NewMarkdownFormatter()
	// Packer will write to standard output.
	p := packer.NewPacker(formatter, os.Stdout)

	// Execute core logic
	if err := p.ProcessPath(targetPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
