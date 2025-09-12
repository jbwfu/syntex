package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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
	var dryRun bool

	fs.BoolVar(&noGitignore, "no-gitignore", false, "Disable the use of .gitignore files for filtering.")
	fs.StringSliceVar(&excludePatterns, "exclude", nil, "Patterns to exclude files or directories. Can be used multiple times.")
	fs.StringSliceVar(&includePatterns, "include", nil, "Patterns to force include files or to specify input paths. Can be used multiple times.")
	fs.StringVarP(&outputFormat, "format", "f", "markdown", "Output format (markdown, md, org).")
	fs.BoolVar(&dryRun, "dry-run", false, "Print the list of files to be processed without generating output.")

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

	plan, err := p.Plan(targets)
	if err != nil {
		return fmt.Errorf("planning phase failed: %w", err)
	}

	if dryRun {
		if len(plan) == 0 {
			fmt.Fprintln(stdout, "[Dry Run] No files to be processed.")
			return nil
		}

		fmt.Fprintf(stdout, "[Dry Run] Planning to process files using the '%s' format:\n", outputFormat)

		// For aligned output, find the longest path
		maxLen := 0
		for _, file := range plan {
			if len(file.Path) > maxLen {
				maxLen = len(file.Path)
			}
		}

		for _, file := range plan {
			padding := strings.Repeat(" ", maxLen-len(file.Path)+2)
			fmt.Fprintf(stdout, "%s%s(%s)\n", file.Path, padding, file.Language)
		}
		fmt.Fprintf(stdout, "\n[Dry Run] Total files to be processed: %d\n", len(plan))
		return nil
	}

	if err := p.Execute(plan); err != nil {
		return fmt.Errorf("execution phase failed: %w", err)
	}

	return nil
}
