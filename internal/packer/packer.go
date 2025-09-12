package packer

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/jbwfu/syntex/internal/filter"
)

// Packer handles the logic of processing files and directories.
type Packer struct {
	formatter Formatter
	output    io.Writer
	filter    *filter.Manager
	rootPath  string
}

// NewPacker creates a new Packer instance.
func NewPacker(f Formatter, out io.Writer, filter *filter.Manager, rootPath string) *Packer {
	return &Packer{
		formatter: f,
		output:    out,
		filter:    filter,
		rootPath:  rootPath,
	}
}

// ProcessPath evaluates a path and processes it as a file or directory.
func (p *Packer) ProcessPath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot access path %s: %w", path, err)
	}

	if info.IsDir() {
		return p.processDirectory(path)
	}
	return p.processFile(path)
}

// processDirectory recursively walks a directory and processes each file within.
func (p *Packer) processDirectory(rootPath string) error {
	return filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(p.rootPath, path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not compute relative path for %s: %v\n", path, err)
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

		if d.IsDir() {
			return nil
		}

		if err := p.processFile(path); err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping file %s: %v\n", path, err)
		}

		return nil
	})
}

// processFile reads a single file and writes its formatted content to the output.
func (p *Packer) processFile(filePath string) error {
	if p.filter != nil {
		relPath, err := filepath.Rel(p.rootPath, filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not compute relative path for %s: %v\n", filePath, err)
		} else {
			info, statErr := os.Stat(filePath)
			if statErr != nil {
				return statErr
			}
			if p.filter.IsIgnored(relPath, info.IsDir()) {
				return nil
			}
		}
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	lang := GetLanguage(filePath)
	formatted, err := p.formatter.Format(filePath, lang, content)
	if err != nil {
		return err
	}

	_, err = p.output.Write(formatted)
	return err
}
