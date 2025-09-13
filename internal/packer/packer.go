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
	// Path is the original path from user input or glob match, preserved for display.
	Path string
	// Language is the detected language identifier for syntax highlighting.
	Language string
}

// Packer handles the logic of discovering, filtering, and planning which files
// to include in the final output.
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

// Plan discovers and filters files based on include patterns and target paths.
// It returns a sorted slice of files that are ready to be processed.
// Files from `--include` patterns bypass .gitignore rules.
func (p *Packer) Plan(targets []string) ([]PlannedFile, error) {
	// Use a map with absolute paths as keys to ensure file uniqueness.
	// The value is the original matched path, which we preserve for output.
	uniqueFiles := make(map[string]string)

	// Process --include patterns first, as they have override priority.
	for _, pattern := range p.filter.GetIncludePatterns() {
		if err := p.processPattern(pattern, uniqueFiles, true); err != nil {
			// Log warnings for non-critical errors and continue.
			fmt.Fprintf(os.Stderr, "warning: could not process include pattern %q: %v\n", pattern, err)
		}
	}

	// Process positional target patterns, which respect .gitignore rules.
	for _, pattern := range targets {
		if err := p.processPattern(pattern, uniqueFiles, false); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not process target pattern %q: %v\n", pattern, err)
		}
	}

	// Convert the map of unique files into a sorted slice of PlannedFile.
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

// Execute processes a list of PlannedFile items, formats them using the
// configured formatter, and writes the result to the output writer.
func (p *Packer) Execute(plan []PlannedFile) error {
	for _, file := range plan {
		content, err := os.ReadFile(file.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping unreadable file %s: %v\n", file.Path, err)
			continue
		}

		formatted, err := p.formatter.Format(file.Path, file.Language, content)
		if err != nil {
			return fmt.Errorf("formatting file %q: %w", file.Path, err)
		}

		if _, err := p.output.Write(formatted); err != nil {
			return fmt.Errorf("writing output for file %q: %w", file.Path, err)
		}
	}
	return nil
}

// processPattern finds all files matching a pattern and adds them to the plan.
func (p *Packer) processPattern(pattern string, uniqueFiles map[string]string, isFromInclude bool) error {
	processedPattern, err := preparePattern(pattern)
	if err != nil {
		return err
	}

	matches, err := doublestar.FilepathGlob(processedPattern)
	if err != nil {
		return fmt.Errorf("glob pattern %q failed: %w", processedPattern, err)
	}

	for _, match := range matches {
		p.addFileToPlan(match, processedPattern, uniqueFiles, isFromInclude)
	}
	return nil
}

// addFileToPlan validates a single file path and, if it passes all checks,
// adds it to the map of unique files for processing.
func (p *Packer) addFileToPlan(path, pattern string, uniqueFiles map[string]string, isFromInclude bool) {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return
	}

	if shouldIgnoreDotfile(path, pattern) {
		return
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return // Cannot get an absolute path, so we cannot safely process it.
	}

	// Ensure each file is processed only once.
	if _, exists := uniqueFiles[absPath]; exists {
		return
	}

	// Always apply global --exclude patterns.
	if p.filter.IsGloballyExcluded(absPath) {
		return
	}

	// Apply .gitignore filter only if the file is not from a direct --include pattern.
	if !isFromInclude && p.filter.IsGitIgnored(absPath, false) {
		return
	}

	uniqueFiles[absPath] = path
}

// shouldIgnoreDotfile contains the business logic for deciding if a dotfile
// should be ignored based on the context of how it was found.
func shouldIgnoreDotfile(path, pattern string) bool {
	baseName := filepath.Base(path)
	isDotfile := strings.HasPrefix(baseName, ".") && baseName != "." && baseName != ".."
	if !isDotfile {
		return false
	}

	patternBase := filepath.Base(pattern)
	if strings.HasPrefix(patternBase, ".") {
		return false
	}

	return true
}

// preparePattern expands a tilde prefix and converts directory paths into
// recursive glob patterns (e.g., "mydir/" becomes "mydir/**").
func preparePattern(pattern string) (string, error) {
	expanded, err := expandTilde(pattern)
	if err != nil {
		return "", err
	}

	// If the pattern ends with a path separator, it's explicitly a directory.
	if strings.HasSuffix(expanded, string(os.PathSeparator)) {
		return filepath.Join(expanded, "**"), nil
	}

	// Also, check if the path points to an existing directory.
	info, err := os.Stat(expanded)
	if err == nil && info.IsDir() {
		return filepath.Join(expanded, "**"), nil
	}

	return expanded, nil
}

// expandTilde expands a leading tilde (~) in a path to the user's home directory.
func expandTilde(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("could not get current user for tilde expansion: %w", err)
	}
	return filepath.Join(usr.HomeDir, path[1:]), nil
}
