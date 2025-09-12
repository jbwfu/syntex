package packer

import (
	"bytes"
	"fmt"
)

// Formatter defines the contract for output formatters.
type Formatter interface {
	Format(filename, language string, content []byte) ([]byte, error)
}

// MarkdownFormatter implements the Formatter interface for Markdown.
type MarkdownFormatter struct{}

// NewMarkdownFormatter creates a new MarkdownFormatter.
func NewMarkdownFormatter() *MarkdownFormatter {
	return &MarkdownFormatter{}
}

// Format takes file details and content, and returns it formatted as a Markdown code block.
func (f *MarkdownFormatter) Format(filename, language string, content []byte) ([]byte, error) {
	var out bytes.Buffer

	// Write header: "- filename"
	if _, err := fmt.Fprintf(&out, "- %s\n", filename); err != nil {
		return nil, err
	}

	// Write code fence start: "```language"
	if _, err := fmt.Fprintf(&out, "```%s\n", language); err != nil {
		return nil, err
	}

	// Write content
	if _, err := out.Write(content); err != nil {
		return nil, err
	}

	// Write code fence end: "```"
	if _, err := out.WriteString("\n```\n\n"); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}
