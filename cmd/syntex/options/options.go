package options

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
)

// Options holds all parsed command-line flags for the syntex tool.
type Options struct {
	// Filtering options
	NoIgnore        bool
	Hidden          bool
	Unrestricted    bool
	ExcludePatterns []string
	IncludePatterns []string

	// Input/Output options
	OutputFormat  string
	OutputFile    string
	ToClipboard   bool
	FromStdin0    bool
	FromStdinLine bool

	// Behavior options
	DryRun      bool
	ShowVersion bool

	// Positional arguments
	Targets []string
}

// ParseFlags parses the command-line arguments and populates the Options struct.
// It returns a pointer to Options and an error if parsing fails or required arguments are missing.
func ParseFlags(args []string, stderr io.Writer) (*Options, error) {
	opts := &Options{}
	fs := pflag.NewFlagSet("syntex", pflag.ContinueOnError)
	fs.SetOutput(stderr)

	// Filtering Flags
	fs.BoolVarP(&opts.Hidden, "hidden", "H", false, "Include hidden files and directories.")
	fs.BoolVarP(&opts.NoIgnore, "no-ignore", "I", false, "Do not respect .gitignore files.")
	fs.BoolVarP(&opts.Unrestricted, "unrestricted", "u", false, "Perform an unrestricted search, alias for --hidden --no-ignore.")
	fs.StringSliceVarP(&opts.ExcludePatterns, "exclude", "E", nil, "Exclude files/directories matching the given glob pattern.")
	fs.StringSliceVar(&opts.IncludePatterns, "include", nil, "Force-include files matching the given glob, bypassing ignore rules.")

	// Input/Output Flags
	fs.StringVarP(&opts.OutputFormat, "format", "f", "markdown", "Output format (markdown, md, org).")
	fs.StringVarP(&opts.OutputFile, "output", "o", "", "Write output to a file instead of stdout.")
	fs.BoolVarP(&opts.ToClipboard, "clipboard", "c", false, "Copy output to the system clipboard.")
	fs.BoolVarP(&opts.FromStdin0, "from-stdin-0", "0", false, "Read NUL-separated paths from stdin (e.g., 'find . -print0').")
	fs.BoolVarP(&opts.FromStdinLine, "from-stdin-line", "l", false, "Read newline-separated paths from stdin (e.g., 'ls -1').")

	// Behavior Flags
	fs.BoolVar(&opts.DryRun, "dry-run", false, "Print the list of files to be processed without generating output.")
	fs.BoolVarP(&opts.ShowVersion, "version", "V", false, "Print version information and exit.")

	// Custom usage template
	fs.Usage = func() {
		output := fs.Output()
		progName := filepath.Base(os.Args[0])
		var b strings.Builder

		fmt.Fprintf(&b, "A tool to pack multiple source files into a single context file.\n\n")
		fmt.Fprintf(&b, "Usage:\n  %s [OPTIONS] [path_or_glob...]\n\n", progName)
		fmt.Fprintf(&b, "Arguments:\n")
		fmt.Fprintf(&b, "  [path_or_glob...]   Paths or glob patterns to search for files (optional).\n")
		fmt.Fprintf(&b, "                        If omitted, input must be provided via stdin flags.\n\n")
		fmt.Fprintf(&b, "Options:\n")

		fmt.Fprint(output, b.String())
		fmt.Fprint(output, fs.FlagUsages())
	}

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	// Post-processing for combined flags
	if opts.Unrestricted {
		opts.Hidden = true
		opts.NoIgnore = true
	}

	if opts.FromStdin0 && opts.FromStdinLine {
		return nil, fmt.Errorf("cannot use both -0/--from-stdin-0 and -l/--from-stdin-line flags simultaneously")
	}

	if !opts.ShowVersion {
		opts.Targets = fs.Args()
		if len(opts.Targets) == 0 && len(opts.IncludePatterns) == 0 && !opts.FromStdin0 && !opts.FromStdinLine {
			fs.Usage()
			return nil, fmt.Errorf("no target paths provided, and no input from stdin specified")
		}
	}

	return opts, nil
}
