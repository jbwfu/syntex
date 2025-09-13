package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/jbwfu/syntex/cmd/syntex/options"
	"github.com/jbwfu/syntex/internal/clipboard"
	"github.com/jbwfu/syntex/internal/filter"
	"github.com/jbwfu/syntex/internal/language"
	"github.com/jbwfu/syntex/internal/packer"
	"github.com/spf13/pflag"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run executes the main logic of the syntex command-line tool.
// It parses flags, initializes dependencies, plans the file processing,
// and executes the plan, writing the output to stdout, a specified file, or the clipboard.
func run(args []string, stdout, stderr io.Writer) error {
	opts, err := options.ParseFlags(args, stderr)
	if err != nil {
		return err
	}

	// Collect all output writers
	var outputWriters []io.Writer

	outputFile := opts.OutputFile
	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file %q: %w", outputFile, err)
		}
		defer f.Close()
		outputWriters = append(outputWriters, f)
	}

	if opts.ToClipboard {
		cw := clipboard.NewWriter()
		defer func() {
			if err := cw.Close(); err != nil {
				fmt.Fprintf(stderr, "warning: failed to write to clipboard: %v\n", err)
			}
		}()
		outputWriters = append(outputWriters, cw)
	}

	if len(outputWriters) == 0 {
		outputWriters = append(outputWriters, stdout)
	}

	// Combine all writers into a single MultiWriter
	finalOutputWriter := io.MultiWriter(outputWriters...)

	filterOpts := filter.Options{
		DisableGitignore: opts.NoGitignore,
		ExcludePatterns:  opts.ExcludePatterns,
		IncludePatterns:  opts.IncludePatterns,
		AllowDotfiles:    opts.IncludeHidden,
	}
	filterManager, err := filter.NewManager(filterOpts)
	if err != nil {
		return fmt.Errorf("failed to initialize filter manager: %w", err)
	}

	formatter, err := packer.NewFormatter(opts.OutputFormat)
	if err != nil {
		return err
	}

	languageDetector := language.NewDetector()
	p := packer.NewPacker(formatter, finalOutputWriter, filterManager, languageDetector)

	plan, err := p.Plan(opts.Targets)
	if err != nil {
		return fmt.Errorf("planning phase failed: %w", err)
	}

	if opts.DryRun {
		return printDryRun(stdout, plan, opts.OutputFormat)
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
