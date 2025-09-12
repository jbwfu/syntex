package language

import "path/filepath"

// Detector determines the language identifier for a given filename.
type Detector struct {
	languageToExts   map[string][]string
	specialFilenames map[string]string
	extToLanguage    map[string]string
}

// NewDetector creates and initializes a new Detector.
func NewDetector() *Detector {
	d := &Detector{
		languageToExts: map[string][]string{
			"go":         {".go"},
			"elisp":      {".el"},
			"org":        {".org"},
			"python":     {".py", ".pyw"},
			"javascript": {".js", ".mjs", ".cjs"},
			"java":       {".java"},
			"c":          {".c", ".h"},
			"cpp":        {".cpp", ".hpp", ".cc", ".hh"},
			"bash":       {".sh"},
			"html":       {".html", ".htm"},
			"css":        {".css"},
			"json":       {".json"},
			"xml":        {".xml"},
			"markdown":   {".md", ".markdown"},
		},
		specialFilenames: map[string]string{
			"Makefile":    "makefile",
			"Dockerfile":  "dockerfile",
			"Jenkinsfile": "groovy",
			"Vagrantfile": "ruby",
			"go.mod":      "go-mod",
			"go.sum":      "go-sum",
		},
		extToLanguage: make(map[string]string),
	}

	for lang, exts := range d.languageToExts {
		for _, ext := range exts {
			d.extToLanguage[ext] = lang
		}
	}
	return d
}

// GetLanguage determines the language by checking for special filenames first,
// then extensions, and finally defaulting to "text".
// TODO: For more robust language detection, consider using github.com/go-enry/go-enry/v2.
func (d *Detector) GetLanguage(filename string) string {
	base := filepath.Base(filename)
	if lang, ok := d.specialFilenames[base]; ok {
		return lang
	}

	ext := filepath.Ext(base)
	if lang, ok := d.extToLanguage[ext]; ok {
		return lang
	}

	return "text"
}
