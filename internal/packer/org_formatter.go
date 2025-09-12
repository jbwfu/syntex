package packer

import (
	"bufio"
	"bytes"
	"fmt"
)

// OrgFormatter implements the Formatter interface for Org Mode.
type OrgFormatter struct{}

// NewOrgFormatter creates a new OrgFormatter.
func NewOrgFormatter() *OrgFormatter {
	return &OrgFormatter{}
}

// escapeOrgContent prepends a comma to lines within an Org source block
// that could be misinterpreted by the Org parser, such as headings or directives.
// This is the standard way to escape content within a source block.
func escapeOrgContent(content []byte) []byte {
	var out bytes.Buffer
	scanner := bufio.NewScanner(bytes.NewReader(content))

	for i := 0; scanner.Scan(); i++ {
		if i > 0 {
			out.WriteByte('\n')
		}
		line := scanner.Bytes()

		if len(line) > 0 {
			firstChar := line[0]
			isDirective := len(line) > 1 && firstChar == '#' && line[1] == '+'

			if firstChar == '*' || firstChar == ',' || isDirective {
				out.WriteByte(',')
			}
		}
		out.Write(line)
	}
	return out.Bytes()
}

// Format takes file details and content, and returns it formatted as an Org Mode source block.
// It includes special handling for .org files to prevent parsing issues.
func (f *OrgFormatter) Format(filename, language string, content []byte) ([]byte, error) {
	var out bytes.Buffer

	if _, err := fmt.Fprintf(&out, "- %s\n", filename); err != nil {
		return nil, err
	}

	if _, err := fmt.Fprintf(&out, "#+BEGIN_SRC %s\n", language); err != nil {
		return nil, err
	}

	processedContent := content
	if language == "org" {
		processedContent = escapeOrgContent(content)
	}

	if _, err := out.Write(processedContent); err != nil {
		return nil, err
	}

	if len(processedContent) > 0 && processedContent[len(processedContent)-1] != '\n' {
		if _, err := out.WriteRune('\n'); err != nil {
			return nil, err
		}
	}

	if _, err := out.WriteString("#+END_SRC\n\n"); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}
