package packer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Packer handles the logic of processing files and directories.
type Packer struct {
	formatter Formatter
	output    io.Writer
}

// NewPacker creates a new Packer instance.
// It takes a Formatter for content processing and an io.Writer for output.
func NewPacker(f Formatter, out io.Writer) *Packer {
	return &Packer{
		formatter: f,
		output:    out,
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

// processDirectory iterates over a directory and processes each file within.
// It does not recurse into subdirectories.
func (p *Packer) processDirectory(dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("cannot read directory %s: %w", dirPath, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			filePath := filepath.Join(dirPath, entry.Name())
			if err := p.processFile(filePath); err != nil {
				// Log non-fatal errors to stderr and continue
				fmt.Fprintf(os.Stderr, "warning: skipping file %s: %v\n", filePath, err)
			}
		}
	}
	return nil
}

// processFile reads a single file and writes its formatted content to the output.
func (p *Packer) processFile(filePath string) error {
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
