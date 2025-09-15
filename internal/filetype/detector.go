package filetype

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// List of content type prefixes that should be considered as text.
// http.DetectContentType may misclassify some plain text files (like shell scripts,
// JSON, etc.) as "application/octet-stream", so we maintain an allowlist.
var knownTextContentTypes = []string{
	"text/",
	"application/json",
	"application/javascript",
	"application/xml",
	"application/x-sh",
	"application/x-httpd-php",
}

// IsBinary checks if the file at the given path is likely a binary file.
func IsBinary(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("could not open file %s: %w", path, err)
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false, fmt.Errorf("could not read file %s: %w", path, err)
	}
	buffer = buffer[:n]

	if n == 0 {
		return false, nil
	}

	contentType := http.DetectContentType(buffer)

	for _, textType := range knownTextContentTypes {
		if strings.HasPrefix(contentType, textType) {
			return false, nil
		}
	}

	return true, nil
}
