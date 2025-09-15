package language

import (
	"strings"

	"github.com/go-enry/go-enry/v2"
)

// Detector determines the language identifier for a given filename and its content.
type Detector struct{}

// NewDetector creates and initializes a new Detector.
func NewDetector() *Detector {
	return &Detector{}
}

// GetLanguage determines the language using a cascade of strategies for accuracy.
// It returns a string suitable for use in Markdown or Org-mode code blocks.
func (d *Detector) GetLanguage(filename string, content []byte) string {
	lang, ok := enry.GetLanguageByFilename(filename)
	if ok {
		return normalizeLanguage(lang)
	}

	lang, ok = enry.GetLanguageByExtension(filename)
	if ok {
		return normalizeLanguage(lang)
	}

	lang = enry.GetLanguage(filename, content)
	if lang != "" && lang != "Other" {
		return normalizeLanguage(lang)
	}

	return "text"
}

// normalizeLanguage converts a language name from enry (e.g., "C++", "C#")
// into a format suitable for code block syntax highlighting (e.g., "cpp", "csharp").
func normalizeLanguage(lang string) string {
	group := enry.GetLanguageGroup(lang)
	if group != "" {
		lang = group
	}

	// Handle special cases that are common in code blocks.
	switch lang {
	case "C++":
		return "cpp"
	case "C#":
		return "csharp"
	case "F#":
		return "fsharp"
	case "Protocol Buffer":
		return "protobuf"
	}

	s := strings.ToLower(lang)
	s = strings.ReplaceAll(s, " ", "-")
	return s
}
