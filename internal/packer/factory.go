package packer

import "fmt"

// NewFormatter acts as a factory for creating Formatter instances based on a name.
func NewFormatter(formatName string) (Formatter, error) {
	switch formatName {
	case "markdown", "md":
		return NewMarkdownFormatter(), nil
	case "org":
		return NewOrgFormatter(), nil
	default:
		return nil, fmt.Errorf("unknown format: %q. Supported formats: markdown, md, org", formatName)
	}
}
