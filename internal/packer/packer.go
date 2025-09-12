package packer

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

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
	rootPath  string
}

// NewPacker creates a new Packer instance with all its dependencies injected.
func NewPacker(
	f Formatter,
	out io.Writer,
	filter *filter.Manager,
	detector *language.Detector,
	rootPath string,
) *Packer {
	return &Packer{
		formatter: f,
		output:    out,
		filter:    filter,
		detector:  detector,
		rootPath:  rootPath,
	}
}

// Plan discovers and filters files based on include patterns and target paths.
func (p *Packer) Plan(targets []string) ([]PlannedFile, error) {
	planFiles := make(map[string]PlannedFile)

	p.processIncludes(planFiles)
	p.processTargets(targets, planFiles)

	result := make([]PlannedFile, 0, len(planFiles))
	for _, pf := range planFiles {
		result = append(result, pf)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Path < result[j].Path
	})

	return result, nil
}

func (p *Packer) processIncludes(plan map[string]PlannedFile) {
	for _, pattern := range p.filter.GetIncludePatterns() {
		matches, _ := doublestar.Glob(os.DirFS("."), pattern)
		for _, match := range matches {
			if info, err := os.Stat(match); err == nil && !info.IsDir() {
				plan[match] = PlannedFile{
					Path:     match,
					Language: p.detector.GetLanguage(match),
				}
			}
		}
	}
}

func (p *Packer) processTargets(targets []string, plan map[string]PlannedFile) {
	for _, target := range targets {
		info, err := os.Stat(target)
		if err == nil && info.IsDir() {
			p.walkDirectory(target, plan)
		} else {
			matches, _ := doublestar.Glob(os.DirFS("."), target)
			for _, match := range matches {
				p.addFileIfValid(match, plan)
			}
		}
	}
}

func (p *Packer) walkDirectory(dir string, plan map[string]PlannedFile) {
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return nil
			}
			relPath, err := filepath.Rel(p.rootPath, absPath)
			if err != nil {
				return nil
			}
			if p.filter.IsExcluded(relPath, true) {
				return filepath.SkipDir
			}
		} else {
			p.addFileIfValid(path, plan)
		}
		return nil
	})
}

func (p *Packer) addFileIfValid(path string, plan map[string]PlannedFile) {
	if _, exists := plan[path]; exists {
		return
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return
	}
	relPath, err := filepath.Rel(p.rootPath, absPath)
	if err != nil {
		return
	}

	if !p.filter.IsExcluded(relPath, false) {
		plan[path] = PlannedFile{
			Path:     path,
			Language: p.detector.GetLanguage(path),
		}
	}
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
