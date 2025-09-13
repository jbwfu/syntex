// internal/packer/packer_test.go
package packer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPreparePattern(t *testing.T) {
	// Create a temporary directory to test the case where a path is a directory
	// but doesn't have a trailing slash.
	tempDir, err := os.MkdirTemp("", "syntex-test-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Define test cases in a table, which is a Go community best practice.
	testCases := []struct {
		name     string // A descriptive name for the test case.
		input    string // The input to the function being tested.
		expected string // The expected output.
		wantErr  bool   // Whether we expect an error.
	}{
		{
			name:     "should return simple file path unchanged",
			input:    "path/to/file.go",
			expected: "path/to/file.go",
		},
		{
			name:     "should append globstar to directory with trailing slash",
			input:    "path/to/dir/",
			expected: filepath.Join("path/to/dir/", "**"),
		},
		{
			name:     "should append globstar to existing directory without trailing slash",
			input:    tempDir,
			expected: filepath.Join(tempDir, "**"),
		},
		{
			name:     "should not modify non-existent path without trailing slash",
			input:    "non/existent/path",
			expected: "non/existent/path",
		},
		{
			name:     "should not modify glob pattern",
			input:    "*.go",
			expected: "*.go",
		},
		{
			name:  "should expand tilde",
			input: "~/.config",
			// The exact expansion depends on the user, so we'll do a special check.
			expected: "special_check_tilde",
		},
	}

	for _, tc := range testCases {
		// t.Run creates a sub-test, which gives clearer output on failure.
		t.Run(tc.name, func(t *testing.T) {
			actual, err := preparePattern(tc.input)

			if (err != nil) != tc.wantErr {
				t.Errorf("preparePattern() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			// Special handling for the tilde expansion test case.
			if tc.expected == "special_check_tilde" {
				if strings.HasPrefix(actual, "~") {
					t.Errorf("tilde was not expanded, got: %s", actual)
				}
				// We don't check the full path as it varies by user/environment.
				return
			}

			if actual != tc.expected {
				t.Errorf("preparePattern() got = %v, want %v", actual, tc.expected)
			}
		})
	}
}
