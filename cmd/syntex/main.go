package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jbwfu/syntex/internal/filter"
	"github.com/jbwfu/syntex/internal/language"
	"github.com/jbwfu/syntex/internal/packer"
	"github.com/jbwfu/syntex/internal/project"
	"github.com/spf13/pflag"
)

var (
	noGitignore     bool
	excludePatterns []string
	includePatterns []string
	outputFormat    string
)

func main() {
	pflag.BoolVar(&noGitignore, "no-gitignore", false, "Disable the use of .gitignore files for filtering.")
	pflag.StringSliceVar(&excludePatterns, "exclude", nil, "Patterns to exclude files or directories. Can be used multiple times.")
	pflag.StringSliceVar(&includePatterns, "include", nil, "Patterns to force include files or to specify input paths. Can be used multiple times.")
	pflag.StringVarP(&outputFormat, "format", "f", "markdown", "Output format (markdown, md, org).")

	pflag.Usage = func() {
		progName := filepath.Base(os.Args[0])
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] [directory_or_file...]\n", progName)
		fmt.Fprintf(os.Stderr, "\nFlags:\n")
		pflag.PrintDefaults()
	}
	pflag.Parse()

	targets := pflag.Args()
	if len(targets) == 0 && len(includePatterns) > 0 {
		targets = includePatterns
	}

	if len(targets) == 0 {
		pflag.Usage()
		os.Exit(1)
	}

	rootDiscoveryPath := "."
	if pflag.NArg() > 0 {
		rootDiscoveryPath = targets[0]
	}
	projectRoot, err := project.FindRoot(rootDiscoveryPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to determine project root: %v\n", err)
		os.Exit(1)
	}

	filterOpts := filter.Options{
		DisableGitignore: noGitignore,
		ExcludePatterns:  excludePatterns,
		IncludePatterns:  includePatterns,
	}
	filterManager, err := filter.NewManager(projectRoot, filterOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing filter manager: %v\n", err)
		os.Exit(1)
	}

	formatter, err := packer.NewFormatter(outputFormat)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	languageDetector := language.NewDetector()

	p := packer.NewPacker(formatter, os.Stdout, filterManager, languageDetector, projectRoot)

	var absTargets []string
	for _, target := range targets {
		abs, err := filepath.Abs(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid path %s: %v\n", target, err)
			continue
		}
		absTargets = append(absTargets, abs)
	}

	plan, err := p.Plan(absTargets)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during planning phase: %v\n", err)
		os.Exit(1)
	}

	if err := p.Execute(plan); err != nil {
		fmt.Fprintf(os.Stderr, "Error during execution phase: %v\n", err)
		os.Exit(1)
	}
}
