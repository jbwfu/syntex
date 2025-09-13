package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jbwfu/syntex/internal/filter"
	"github.com/jbwfu/syntex/internal/language"
	"github.com/jbwfu/syntex/internal/packer"
	"github.com/spf13/pflag"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		// pflag.ErrHelp is a special case where the program should exit cleanly.
		if errors.Is(err, pflag.ErrHelp) {
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run executes the main logic of the syntex command-line tool.
// It parses flags, initializes dependencies, plans the file processing,
// and executes the plan, writing the output to stdout.
func run(args []string, stdout, stderr io.Writer) error {
	fs := pflag.NewFlagSet("syntex", pflag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		noGitignore     bool
		excludePatterns []string
		includePatterns []string
		outputFormat    string
		dryRun          bool
		includeHidden   bool
	)

	fs.BoolVar(&noGitignore, "no-gitignore", false, "Disable the use of .gitignore files for filtering.")
	fs.StringSliceVar(&excludePatterns, "exclude", nil, "Patterns to exclude files or directories.")
	fs.StringSliceVar(&includePatterns, "include", nil, "Patterns to force include files or to specify input paths.")
	fs.StringVarP(&outputFormat, "format", "f", "markdown", "Output format (markdown, md, org).")
	fs.BoolVar(&dryRun, "dry-run", false, "Print the list of files to be processed without generating output.")
	fs.BoolVar(&includeHidden, "include-hidden", false, "Include dotfiles and files in dot-directories in the output.")

	fs.Usage = func() {
		progName := filepath.Base(os.Args[0])
		fmt.Fprintf(stderr, "Usage: %s [flags] [path_or_glob...]\n", progName)
		fmt.Fprintf(stderr, "\nOptions:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	targets := fs.Args()
	if len(targets) == 0 && len(includePatterns) == 0 {
		fs.Usage()
		return fmt.Errorf("no target paths or globs provided")
	}

	filterOpts := filter.Options{
		DisableGitignore: noGitignore,
		ExcludePatterns:  excludePatterns,
		IncludePatterns:  includePatterns,
		AllowDotfiles:    includeHidden,
	}
	filterManager, err := filter.NewManager(filterOpts)
	if err != nil {
		return fmt.Errorf("failed to initialize filter manager: %w", err)
	}

	formatter, err := packer.NewFormatter(outputFormat)
	if err != nil {
		return err
	}

	languageDetector := language.NewDetector()
	p := packer.NewPacker(formatter, stdout, filterManager, languageDetector)

	plan, err := p.Plan(targets)
	if err != nil {
		return fmt.Errorf("planning phase failed: %w", err)
	}

	if dryRun {
		return printDryRun(stdout, plan, outputFormat)
	}

	if err := p.Execute(plan); err != nil {
		return fmt.Errorf("execution phase failed: %w", err)
	}

	return nil
}

// printDryRun displays the planned files to be processed without actually processing them.
func printDryRun(w io.Writer, plan []packer.PlannedFile, format string) error {
	if len(plan) == 0 {
		fmt.Fprintln(w, "[Dry Run] No files to be processed.")
		return nil
	}

	fmt.Fprintf(w, "[Dry Run] Planning to process files using the '%s' format:\n", format)

	maxLangLen := 0
	for _, file := range plan {
		if len(file.Language) > maxLangLen {
			maxLangLen = len(file.Language)
		}
	}

	for _, file := range plan {
		fmt.Fprintf(w, "%-*s  %s\n", maxLangLen, file.Language, file.Path)
	}

	fmt.Fprintf(w, "\n[Dry Run] Total: %d\n", len(plan))
	return nil
}
