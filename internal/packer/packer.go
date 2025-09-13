package packer

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/jbwfu/syntex/internal/filter"
	"github.com/jbwfu/syntex/internal/language"
)

// PlannedFile holds pre-calculated information for a file to be processed.
type PlannedFile struct {
	Path     string
	Language string
}

// Packer handles the logic of processing files and directories.
type Packer struct {
	formatter Formatter
	output    io.Writer
	filter    *filter.Manager
	detector  *language.Detector
}

// NewPacker creates a new Packer instance with all its dependencies injected.
func NewPacker(
	f Formatter,
	out io.Writer,
	filter *filter.Manager,
	detector *language.Detector,
) *Packer {
	return &Packer{
		formatter: f,
		output:    out,
		filter:    filter,
		detector:  detector,
	}
}

// expandTilde expands the '~' prefix in a path to the user's home directory.
func expandTilde(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, path[1:]), nil
}

// Plan discovers and filters files based on include patterns and target paths.
func (p *Packer) Plan(targets []string) ([]PlannedFile, error) {
	uniqueFiles := make(map[string]string)

	// Process --include patterns first. These bypass .gitignore.
	for _, pattern := range p.filter.GetIncludePatterns() {
		processedPattern, err := p.preparePattern(pattern)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not process pattern %q: %v\n", pattern, err)
			continue
		}
		matches, _ := doublestar.FilepathGlob(processedPattern)
		for _, match := range matches {
			p.addFile(match, uniqueFiles, true) // isFromInclude = true
		}
	}

	// Process positional target patterns. These respect .gitignore.
	for _, pattern := range targets {
		processedPattern, err := p.preparePattern(pattern)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not process pattern %q: %v\n", pattern, err)
			continue
		}
		matches, _ := doublestar.FilepathGlob(processedPattern)
		for _, match := range matches {
			p.addFile(match, uniqueFiles, false) // isFromInclude = false
		}
	}

	result := make([]PlannedFile, 0, len(uniqueFiles))
	for _, originalPath := range uniqueFiles {
		result = append(result, PlannedFile{
			Path:     originalPath,
			Language: p.detector.GetLanguage(originalPath),
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Path < result[j].Path
	})

	return result, nil
}

// preparePattern expands tilde and converts directory paths to recursive globs.
func (p *Packer) preparePattern(pattern string) (string, error) {
	expanded, err := expandTilde(pattern)
	if err != nil {
		return "", err
	}

	if strings.HasSuffix(expanded, string(os.PathSeparator)) {
		return filepath.Join(expanded, "**"), nil
	}

	// Also check if the path exists and is a directory, even without a trailing slash.
	info, err := os.Stat(expanded)
	if err == nil && info.IsDir() {
		return filepath.Join(expanded, "**"), nil
	}

	return expanded, nil
}

// addFile checks and adds a file to the plan, respecting uniqueness and filtering rules.
func (p *Packer) addFile(path string, uniqueFiles map[string]string, isFromInclude bool) {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return
	}

	if _, exists := uniqueFiles[absPath]; exists {
		return
	}

	if p.filter.IsGloballyExcluded(absPath) {
		return
	}

	if !isFromInclude && p.filter.IsGitIgnored(absPath, false) {
		return
	}

	uniqueFiles[absPath] = path
}

// Execute processes a list of PlannedFile items, formats them, and writes to output.
func (p *Packer) Execute(plan []PlannedFile) error {
	for _, file := range plan {
		content, err := os.ReadFile(file.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping file %s: %v\n", file.Path, err)
			continue
		}

		formatted, err := p.formatter.Format(file.Path, file.Language, content)
		if err != nil {
			return fmt.Errorf("formatting file %s: %w", file.Path, err)
		}

		if _, err := p.output.Write(formatted); err != nil {
			return fmt.Errorf("writing to output for file %s: %w", file.Path, err)
		}
	}
	return nil
}
