package filetype

import (
	"os"
	"path/filepath"
	"testing"
)

// createTestFile creates a temporary file with the given content for testing.
func createTestFile(t *testing.T, name, content string, isBinary bool) string {
	t.Helper()

	file, err := os.CreateTemp("", name)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	var data []byte
	if isBinary {
		data = []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01} // JPEG magic bytes
		data = append(data, content...)
	} else {
		data = []byte(content)
	}

	if _, err := file.Write(data); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	t.Cleanup(func() { os.Remove(file.Name()) })

	return file.Name()
}

func TestIsBinary(t *testing.T) {
	testCases := []struct {
		name       string
		filename   string
		content    string
		isBinary   bool
		wantBinary bool
		wantErr    bool
	}{
		{
			name:       "plain text file",
			filename:   "test.txt",
			content:    "this is a plain text file.",
			isBinary:   false,
			wantBinary: false,
		},
		{
			name:       "shell script file",
			filename:   "test.sh",
			content:    "#!/bin/bash\necho 'hello'",
			isBinary:   false,
			wantBinary: false,
		},
		{
			name:       "json file",
			filename:   "test.json",
			content:    `{"key": "value"}`,
			isBinary:   false,
			wantBinary: false,
		},
		{
			name:       "binary file (jpeg magic bytes)",
			filename:   "test.jpg",
			content:    "some image data",
			isBinary:   true,
			wantBinary: true,
		},
		{
			name:       "empty file",
			filename:   "empty.txt",
			content:    "",
			isBinary:   false,
			wantBinary: false,
		},
		{
			name:       "non-existent file",
			filename:   filepath.Join("non", "existent", "path.txt"),
			content:    "",
			isBinary:   false,
			wantBinary: false,
			wantErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var path string
			if !tc.wantErr {
				path = createTestFile(t, tc.filename, tc.content, tc.isBinary)
			} else {
				path = tc.filename
			}

			gotBinary, err := IsBinary(path)

			if (err != nil) != tc.wantErr {
				t.Fatalf("IsBinary() error = %v, wantErr %v", err, tc.wantErr)
			}

			if gotBinary != tc.wantBinary {
				t.Errorf("IsBinary() got = %v, want %v", gotBinary, tc.wantBinary)
			}
		})
	}
}
