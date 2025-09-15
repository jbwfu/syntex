package language

import (
	"io"
	"os"
	"strings"

	"github.com/go-enry/go-enry/v2"
)

const readBufferSize = 8192

// DetectionResult holds the outcome of a file analysis.
type DetectionResult struct {
	Language string
	IsBinary bool
}

// Detector determines the language identifier for a given filename.
type Detector struct{}

// NewDetector creates and initializes a new Detector.
func NewDetector() *Detector {
	return &Detector{}
}

// AnalyzeFile determines the language and binary status of a file by its path.
func (d *Detector) AnalyzeFile(path string) (*DetectionResult, error) {
	var lang string
	var ok bool

	lang, ok = enry.GetLanguageByExtension(path)
	if !ok {
		lang, ok = enry.GetLanguageByFilename(path)
	}

	if !ok {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		buffer := make([]byte, readBufferSize)
		n, err := io.ReadFull(file, buffer)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return nil, err
		}
		contentPrefix := buffer[:n]

		if enry.IsBinary(contentPrefix) {
			return &DetectionResult{IsBinary: true}, nil
		}

		lang = enry.GetLanguage(path, contentPrefix)
	}

	result := &DetectionResult{
		Language: d.normalizeLanguage(path, lang),
		IsBinary: false,
	}

	return result, nil
}

// normalizeLanguage converts a language name from enry into a code block-friendly format.
func (d *Detector) normalizeLanguage(filename, lang string) string {
	if lang == "" || lang == "Other" {
		return "unknown"

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
