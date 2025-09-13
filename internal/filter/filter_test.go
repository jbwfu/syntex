// internal/filter/filter_test.go (or create a new file like filter_dotfile_test.go if preferred)
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
	os.MkdirAll(filepath.Join(root, ".test", "kkk", "oo"), 0755) // Hidden directory
	os.MkdirAll(filepath.Join(root, ".config"), 0755)            // Hidden directory

	os.WriteFile(filepath.Join(root, "app.go"), nil, 0644)
	os.WriteFile(filepath.Join(root, "error.log"), nil, 0644)
	os.WriteFile(filepath.Join(root, "src", "main.go"), nil, 0644)
	os.WriteFile(filepath.Join(root, "node_modules", "some-lib", "index.js"), nil, 0644)
	os.WriteFile(filepath.Join(root, "build", "app.exe"), nil, 0644)
	os.WriteFile(filepath.Join(root, ".test", "kkk", "oo", "ll.go"), nil, 0644) // File in hidden dir
	os.WriteFile(filepath.Join(root, ".test", ".xxx"), nil, 0644)               // Hidden file in hidden dir
	os.WriteFile(filepath.Join(root, ".env"), nil, 0644)                        // Hidden file at root
	os.WriteFile(filepath.Join(root, ".config", "settings.toml"), nil, 0644)    // File in hidden dir

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
	t.Run("caching", func(t *testing.T) {
		absPath := filepath.Join(repoRoot, "error.log")
		if !manager.IsGitIgnored(absPath, false) {
			t.Error("IsGitIgnored() should still return true on second call (cache hit)")
		}
	})
}

func TestFilterManager_IsDotfileIgnored(t *testing.T) {
	testCases := []struct {
		name          string
		allowDotfiles bool
		filePath      string
		globPattern   string
		wantIgnored   bool
	}{
		// allowDotfiles = false (default behavior: ignore hidden unless explicit)
		{
			name:          "default: regular file, generic pattern",
			allowDotfiles: false,
			filePath:      "foo/bar.go",
			globPattern:   "**",
			wantIgnored:   false,
		},
		{
			name:          "default: hidden file at root, generic pattern",
			allowDotfiles: false,
			filePath:      ".env",
			globPattern:   "**",
			wantIgnored:   true,
		},
		{
			name:          "default: file in hidden dir, generic pattern",
			allowDotfiles: false,
			filePath:      ".test/foo.go",
			globPattern:   "**",
			wantIgnored:   true,
		},
		{
			name:          "default: hidden file in hidden dir, generic pattern",
			allowDotfiles: false,
			filePath:      ".test/.xxx",
			globPattern:   "**",
			wantIgnored:   true,
		},
		{
			name:          "default: file in hidden dir, explicit dir pattern",
			allowDotfiles: false,
			filePath:      ".test/foo.go",
			globPattern:   ".test/**", // .test is explicit, foo.go is not hidden
			wantIgnored:   false,
		},
		{
			name:          "default: hidden file in hidden dir, explicit dir pattern (inner file not explicit)",
			allowDotfiles: false,
			filePath:      ".test/.xxx",
			globPattern:   ".test/**", // .test is explicit, but ** is not explicit for .xxx
			wantIgnored:   true,
		},
		{
			name:          "default: fully explicit hidden file pattern",
			allowDotfiles: false,
			filePath:      ".test/.xxx",
			globPattern:   ".test/.xxx", // Both components are explicit
			wantIgnored:   false,
		},
		{
			name:          "default: explicit dotfile at root",
			allowDotfiles: false,
			filePath:      ".env",
			globPattern:   ".env",
			wantIgnored:   false,
		},
		{
			name:          "default: file in hidden dir, explicit dir and file pattern",
			allowDotfiles: false,
			filePath:      ".config/settings.toml",
			globPattern:   ".config/*.toml",
			wantIgnored:   false,
		},

		// allowDotfiles = true (override default behavior: never ignore hidden by this rule)
		{
			name:          "include-hidden: hidden file at root, generic pattern",
			allowDotfiles: true,
			filePath:      ".env",
			globPattern:   "**",
			wantIgnored:   false,
		},
		{
			name:          "include-hidden: file in hidden dir, generic pattern",
			allowDotfiles: true,
			filePath:      ".test/foo.go",
			globPattern:   "**",
			wantIgnored:   false,
		},
		{
			name:          "include-hidden: hidden file in hidden dir, generic pattern",
			allowDotfiles: true,
			filePath:      ".test/.xxx",
			globPattern:   "**",
			wantIgnored:   false,
		},
		{
			name:          "include-hidden: hidden file in hidden dir, explicit dir pattern",
			allowDotfiles: true,
			filePath:      ".test/.xxx",
			globPattern:   ".test/**",
			wantIgnored:   false,
		},
		{
			name:          "include-hidden: fully explicit hidden file pattern",
			allowDotfiles: true,
			filePath:      ".test/.xxx",
			globPattern:   ".test/.xxx",
			wantIgnored:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manager := &Manager{allowDotfiles: tc.allowDotfiles}
			gotIgnored := manager.IsDotfileIgnored(tc.filePath, tc.globPattern)
			if gotIgnored != tc.wantIgnored {
				t.Errorf("IsDotfileIgnored(%q, %q, allowDotfiles=%t) = %v, want %v",
					tc.filePath, tc.globPattern, tc.allowDotfiles, gotIgnored, tc.wantIgnored)
			}
		})
	}
}
