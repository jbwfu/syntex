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
	"github.com/jbwfu/syntex/internal/filetype"
	"github.com/jbwfu/syntex/internal/filter"
	"github.com/jbwfu/syntex/internal/language"
)

// PlannedFile holds pre-calculated information and content for a file to be processed.
type PlannedFile struct {
	// Path is the original path from user input or glob match, preserved for display.
	Path string
	// Language is the detected language identifier for syntax highlighting.
	Language string
	// Content holds the raw byte content of the file.
	Content []byte
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

// Plan discovers, filters, and reads files based on patterns and targets.
// It returns a sorted slice of files (with their content) ready for processing.
func (p *Packer) Plan(targets []string) ([]PlannedFile, error) {
	uniqueFiles := make(map[string]string)

	for _, pattern := range p.filter.GetIncludePatterns() {
		if err := p.processPattern(pattern, uniqueFiles, true); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not process include pattern %q: %v\n", pattern, err)
		}
	}

	for _, pattern := range targets {
		if err := p.processPattern(pattern, uniqueFiles, false); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not process target pattern %q: %v\n", pattern, err)
		}
	}

	result := make([]PlannedFile, 0, len(uniqueFiles))
	for absPath, originalPath := range uniqueFiles {
		content, err := os.ReadFile(absPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping unreadable file %s: %v\n", originalPath, err)
			continue
		}

		result = append(result, PlannedFile{
			Path:     originalPath,
			Language: p.detector.GetLanguage(originalPath, content),
			Content:  content,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Path < result[j].Path
	})

	return result, nil
}

// Execute processes a list of PlannedFile items, formats them, and writes the result.
func (p *Packer) Execute(plan []PlannedFile) error {
	for _, file := range plan {
		// Content is now pre-loaded in the plan.
		formatted, err := p.formatter.Format(file.Path, file.Language, file.Content)
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

	if p.filter.IsDotfileIgnored(path, pattern) {
		return
	}

	isBinary, err := filetype.IsBinary(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not check file type for %s, skipping: %v\n", path, err)
		return
	}
	if isBinary {
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

// preparePattern expands a tilde prefix and converts directory paths into
// recursive glob patterns (e.g., "mydir/" becomes "mydir/**").
func preparePattern(pattern string) (string, error) {
	expanded, err := expandTilde(pattern)
	if err != nil {
		return "", err
	}

	if strings.HasSuffix(expanded, string(os.PathSeparator)) {
		return filepath.Join(expanded, "**"), nil
	}

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
