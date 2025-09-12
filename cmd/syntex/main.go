package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jbwfu/syntex/internal/filter"
	"github.com/jbwfu/syntex/internal/language"
	"github.com/jbwfu/syntex/internal/packer"
	"github.com/jbwfu/syntex/internal/project"
	"github.com/spf13/pflag"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	fs := pflag.NewFlagSet("syntex", pflag.ContinueOnError)
	fs.SetOutput(stderr)

	var noGitignore bool
	var excludePatterns []string
	var includePatterns []string
	var outputFormat string

	fs.BoolVar(&noGitignore, "no-gitignore", false, "Disable the use of .gitignore files for filtering.")
	fs.StringSliceVar(&excludePatterns, "exclude", nil, "Patterns to exclude files or directories. Can be used multiple times.")
	fs.StringSliceVar(&includePatterns, "include", nil, "Patterns to force include files or to specify input paths. Can be used multiple times.")
	fs.StringVarP(&outputFormat, "format", "f", "markdown", "Output format (markdown, md, org).")

	fs.Usage = func() {
		progName := filepath.Base(os.Args[0])
		fmt.Fprintf(stderr, "Usage: %s [flags] [directory_or_file...]\n", progName)
		fmt.Fprintf(stderr, "\nFlags:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	targets := fs.Args()
	if len(targets) == 0 && len(includePatterns) > 0 {
		targets = includePatterns
	}

	if len(targets) == 0 {
		fs.Usage()
		return fmt.Errorf("no target paths provided")
	}

	rootDiscoveryPath := "."
	if fs.NArg() > 0 {
		rootDiscoveryPath = targets[0]
	}
	projectRoot, err := project.FindRoot(rootDiscoveryPath)
	if err != nil {
		return fmt.Errorf("failed to determine project root: %w", err)
	}

	filterOpts := filter.Options{
		DisableGitignore: noGitignore,
		ExcludePatterns:  excludePatterns,
		IncludePatterns:  includePatterns,
	}
	filterManager, err := filter.NewManager(projectRoot, filterOpts)
	if err != nil {
		return fmt.Errorf("failed to initialize filter manager: %w", err)
	}

	formatter, err := packer.NewFormatter(outputFormat)
	if err != nil {
		return err
	}

	languageDetector := language.NewDetector()

	p := packer.NewPacker(formatter, stdout, filterManager, languageDetector, projectRoot)

	var absTargets []string
	for _, target := range targets {
		abs, err := filepath.Abs(target)
		if err != nil {
			return fmt.Errorf("invalid path %q: %w", target, err)
		}
		absTargets = append(absTargets, abs)
	}

	plan, err := p.Plan(absTargets)
	if err != nil {
		return fmt.Errorf("planning phase failed: %w", err)
	}

	if err := p.Execute(plan); err != nil {
		return fmt.Errorf("execution phase failed: %w", err)
	}

	return nil
}
