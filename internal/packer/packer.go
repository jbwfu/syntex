package packer

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/jbwfu/syntex/internal/filter"
	"github.com/jbwfu/syntex/internal/language"
)

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

// Plan discovers and filters all files from the given target paths,
// returning a list of absolute file paths to be processed.
func (p *Packer) Plan(targets []string) ([]string, error) {
	var plan []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, target := range targets {
		wg.Add(1)
		go func(t string) {
			defer wg.Done()
			var files []string
			info, err := os.Stat(t)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: cannot access path %s: %v\n", t, err)
				return
			}

			if info.IsDir() {
				files, err = p.planDirectory(t)
			} else {
				files, err = p.planFile(t)
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: error planning path %s: %v\n", t, err)
				return
			}
			mu.Lock()
			plan = append(plan, files...)
			mu.Unlock()
		}(target)
	}

	wg.Wait()
	return plan, nil
}

func (p *Packer) planDirectory(rootTarget string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(rootTarget, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(p.rootPath, path)
		if err != nil {
			return nil
		}

		if p.filter != nil && relPath != "." {
			if p.filter.IsIgnored(relPath, d.IsDir()) {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func (p *Packer) planFile(filePath string) ([]string, error) {
	relPath, err := filepath.Rel(p.rootPath, filePath)
	if err != nil {
		return nil, nil
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	if p.filter != nil && p.filter.IsIgnored(relPath, info.IsDir()) {
		return nil, nil
	}
	return []string{filePath}, nil
}

// Execute processes a list of file paths (the plan), formats them, and writes to output.
func (p *Packer) Execute(plan []string) error {
	for _, filePath := range plan {
		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping file %s: %v\n", filePath, err)
			continue
		}

		lang := p.detector.GetLanguage(filePath)
		formatted, err := p.formatter.Format(filePath, lang, content)
		if err != nil {
			return fmt.Errorf("formatting file %s: %w", filePath, err)
		}

		if _, err := p.output.Write(formatted); err != nil {
			return fmt.Errorf("writing to output for file %s: %w", filePath, err)
		}
	}
	return nil
}
