package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jbwfu/syntex/internal/filter"
	"github.com/jbwfu/syntex/internal/packer"
	"github.com/jbwfu/syntex/internal/project"
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

	projectRoot, err := project.FindRoot(targetPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to determine project root for %s: %v\n", targetPath, err)
		os.Exit(1)
	}

	// For now, we create default options. This will be driven by flags later.
	filterOpts := filter.Options{
		DisableGitignore: false,
	}

	filterManager, err := filter.NewManager(projectRoot, filterOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not initialize filter manager: %v\n", err)
	}

	// Setup dependencies
	formatter := packer.NewMarkdownFormatter()
	p := packer.NewPacker(formatter, os.Stdout, filterManager, projectRoot)

	// Execute core logic
	absTargetPath, err := filepath.Abs(targetPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid path %s: %v\n", targetPath, err)
		os.Exit(1)
	}

	if err := p.ProcessPath(absTargetPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
