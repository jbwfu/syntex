// internal/filter/filter_test.go
package filter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
)

// setupTestRepo creates a temporary directory with a git repository,
// a .gitignore file, and some files/directories to test against.
func setupTestRepo(t *testing.T) (string, func()) {
	t.Helper()

	root, err := os.MkdirTemp("", "syntex-filter-test-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Initialize a git repository.
	_, err = git.PlainInit(root, false)
	if err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Create .gitignore file.
	gitignoreContent := `
# Comments should be ignored
*.log
/node_modules/
build/
`
	err = os.WriteFile(filepath.Join(root, ".gitignore"), []byte(gitignoreContent), 0644)
	if err != nil {
		t.Fatalf("failed to write .gitignore: %v", err)
	}

	// Create test file structure.
	os.MkdirAll(filepath.Join(root, "src"), 0755)
	os.MkdirAll(filepath.Join(root, "node_modules", "some-lib"), 0755)
	os.MkdirAll(filepath.Join(root, "build", "output"), 0755)

	os.WriteFile(filepath.Join(root, "app.go"), nil, 0644)
	os.WriteFile(filepath.Join(root, "error.log"), nil, 0644)
	os.WriteFile(filepath.Join(root, "src", "main.go"), nil, 0644)
	os.WriteFile(filepath.Join(root, "node_modules", "some-lib", "index.js"), nil, 0644)
	os.WriteFile(filepath.Join(root, "build", "app.exe"), nil, 0644)

	cleanup := func() {
		os.RemoveAll(root)
	}

	return root, cleanup
}

func TestFilterManager(t *testing.T) {
	repoRoot, cleanup := setupTestRepo(t)
	defer cleanup()

	opts := Options{
		ExcludePatterns: []string{"**/*.exe"},
	}
	manager, err := NewManager(opts)
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	testCases := []struct {
		name             string
		path             string // Path relative to repoRoot
		isDir            bool
		wantGitIgnored   bool
		wantGloballyExcl bool
	}{
		{"go file", "app.go", false, false, false},
		{"log file", "error.log", false, true, false},
		{"nested go file", "src/main.go", false, false, false},
		{"ignored dir", "node_modules", true, true, false},
		{"file in ignored dir", "node_modules/some-lib/index.js", false, true, false},
		{"build output dir", "build", true, true, false},
		{"globally excluded file", "build/app.exe", false, true, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			absPath := filepath.Join(repoRoot, tc.path)

			// Test IsGitIgnored
			gotGitIgnored := manager.IsGitIgnored(absPath, tc.isDir)
			if gotGitIgnored != tc.wantGitIgnored {
				t.Errorf("IsGitIgnored(%q) = %v, want %v", tc.path, gotGitIgnored, tc.wantGitIgnored)
			}

			// Test IsGloballyExcluded
			gotGloballyExcl := manager.IsGloballyExcluded(absPath)
			if gotGloballyExcl != tc.wantGloballyExcl {
				t.Errorf("IsGloballyExcluded(%q) = %v, want %v", tc.path, gotGloballyExcl, tc.wantGloballyExcl)
			}
		})
	}

	// Test caching: the second call should hit the cache.
	// A simple way to test is just to call it again and ensure it still works.
	// More complex tests could use mocks to verify cache hits.
	t.Run("caching", func(t *testing.T) {
		absPath := filepath.Join(repoRoot, "error.log")
		if !manager.IsGitIgnored(absPath, false) {
			t.Error("IsGitIgnored() should still return true on second call (cache hit)")
		}
	})
}
