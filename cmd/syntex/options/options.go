package options

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
)

// Options holds all parsed command-line flags for the syntex tool.
type Options struct {
	NoGitignore     bool
	ExcludePatterns []string
	IncludePatterns []string
	OutputFormat    string
	DryRun          bool
	IncludeHidden   bool
	OutputFile      string
	ToClipboard     bool
	Targets         []string
	FromStdin0      bool
	FromStdinLine   bool
	ShowVersion     bool
}

// ParseFlags parses the command-line arguments and populates the Options struct.
// It returns a pointer to Options and an error if parsing fails or required arguments are missing.
func ParseFlags(args []string, stderr io.Writer) (*Options, error) {
	opts := &Options{}
	fs := pflag.NewFlagSet("syntex", pflag.ContinueOnError)
	fs.SetOutput(stderr)

	fs.BoolVar(&opts.NoGitignore, "no-gitignore", false, "Disable the use of .gitignore files for filtering.")
	fs.StringSliceVar(&opts.ExcludePatterns, "exclude", nil, "Patterns to exclude files or directories.")
	fs.StringSliceVar(&opts.IncludePatterns, "include", nil, "Patterns to force include files or to specify input paths.")
	fs.StringVarP(&opts.OutputFormat, "format", "f", "markdown", "Output format (markdown, md, org).")
	fs.BoolVar(&opts.DryRun, "dry-run", false, "Print the list of files to be processed without generating output.")
	fs.BoolVar(&opts.IncludeHidden, "include-hidden", false, "Include dotfiles and files in dot-directories in the output.")
	fs.StringVarP(&opts.OutputFile, "output", "o", "", "Write output to a file instead of stdout.")
	fs.BoolVarP(&opts.ToClipboard, "clipboard", "c", false, "Write output to the system clipboard.")
	fs.BoolVarP(&opts.FromStdin0, "from-stdin-0", "0", false, "Read NUL-separated file paths from stdin (e.g., from find . -print0).")
	fs.BoolVarP(&opts.FromStdinLine, "from-stdin-line", "l", false, "Read newline-separated file paths from stdin (e.g., from ls -1). NOTE: This mode is NOT safe for filenames containing newlines.")
	fs.BoolVarP(&opts.ShowVersion, "version", "V", false, "Print version information and exit.")

	fs.Usage = func() {
		progName := filepath.Base(os.Args[0])
		fmt.Fprintf(stderr, "Usage: %s [flags] [path_or_glob...]\n", progName)
		fmt.Fprintf(stderr, "\nOptions:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if opts.FromStdin0 && opts.FromStdinLine {
		return nil, fmt.Errorf("cannot use both -0/--from-stdin-0 and -l/--from-stdin-line flags simultaneously")
	}

	if !opts.ShowVersion {
		opts.Targets = fs.Args()
		if len(opts.Targets) == 0 && len(opts.IncludePatterns) == 0 && !opts.FromStdin0 && !opts.FromStdinLine {
			fs.Usage()
			return nil, fmt.Errorf("no target paths or globs provided, and no stdin input flag (-0/-l) was specified")
		}
	}

	return opts, nil
}
