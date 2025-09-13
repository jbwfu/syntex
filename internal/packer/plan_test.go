package packer

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/jbwfu/syntex/internal/filter"
	"github.com/jbwfu/syntex/internal/language"
)

// setupTestEnvironment creates a temporary directory structure with a git repo
// and a .gitignore file, similar to the filter test, but tailored for planning.
func setupTestEnvironment(t *testing.T) (string, func()) {
	t.Helper()

	root, err := os.MkdirTemp("", "syntex-plan-test-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Change working directory to the temp dir to simulate real CLI usage
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	os.Chdir(root)

	_, err = git.PlainInit(".", false)
	if err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	gitignoreContent := `
vendor/
*.log
`
	err = os.WriteFile(".gitignore", []byte(gitignoreContent), 0644)
	if err != nil {
		t.Fatalf("failed to write .gitignore: %v", err)
	}

	// Create test file structure
	os.MkdirAll("src", 0755)
	os.MkdirAll("vendor/lib", 0755)
	os.MkdirAll(".test/kkk/oo", 0755) // Hidden directory
	os.MkdirAll(".config", 0755)      // Hidden directory

	os.WriteFile("main.go", nil, 0644)
	os.WriteFile("main.log", nil, 0644)
	os.WriteFile("src/app.go", nil, 0644)
	os.WriteFile("vendor/lib/lib.go", nil, 0644)
	os.WriteFile("README.md", nil, 0644)
	os.WriteFile(".test/kkk/oo/ll.go", nil, 0644)    // File in hidden dir
	os.WriteFile(".test/.xxx", nil, 0644)            // Hidden file in hidden dir
	os.WriteFile(".env", nil, 0644)                  // Hidden file at root
	os.WriteFile(".config/settings.toml", nil, 0644) // File in hidden dir

	cleanup := func() {
		os.Chdir(originalWD)
		os.RemoveAll(root)
	}

	return root, cleanup
}

func TestPacker_Plan(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create a non-repo file outside the project to test absolute paths
	outsideFile, err := os.CreateTemp("", "outside-*.js")
	if err != nil {
		t.Fatalf("Failed to create outside file: %v", err)
	}
	defer os.Remove(outsideFile.Name())

	testCases := []struct {
		name            string
		targets         []string
		includePatterns []string
		excludePatterns []string
		allowDotfiles   bool     // Test with allowDotfiles option
		expectedPlan    []string // Just the paths for easier comparison
	}{
		{
			name:         "should include all non-ignored files with globstar",
			targets:      []string{"**/*.go"},
			expectedPlan: []string{"main.go", "src/app.go"},
		},
		{
			name:         "should ignore files from .gitignore and also dotfiles by default",
			targets:      []string{"**"}, // Matches everything, but dotfiles are ignored by default
			expectedPlan: []string{"main.go", "README.md", "src/app.go"},
		},
		{
			name:            "should force include gitignored file with --include",
			targets:         []string{"**/*.go"},
			includePatterns: []string{"vendor/lib/lib.go"},
			expectedPlan:    []string{"main.go", "src/app.go", "vendor/lib/lib.go"},
		},
		{
			name:            "should exclude files with --exclude",
			targets:         []string{"**/*.go"},
			excludePatterns: []string{"**/app.go"},
			expectedPlan:    []string{"main.go"},
		},
		{
			name:         "should handle directory target",
			targets:      []string{"src/"},
			expectedPlan: []string{filepath.Join("src", "app.go")},
		},
		{
			name:            "should handle absolute path from --include",
			targets:         []string{"main.go"},
			includePatterns: []string{outsideFile.Name()},
			expectedPlan:    []string{"main.go", outsideFile.Name()},
		},
		{
			name:         "should include dotfiles when pattern explicitly starts with a dot",
			targets:      []string{".*"}, // This pattern explicitly targets dotfiles
			expectedPlan: []string{".env", ".gitignore"},
		},
		{
			name:         "should include files in explicit dot-directory but not its hidden files",
			targets:      []string{".test/**"},
			expectedPlan: []string{filepath.Join(".test", "kkk", "oo", "ll.go")}, // .test/.xxx should be ignored
		},
		{
			name:         "should include explicit hidden file",
			targets:      []string{".test/.xxx"},
			expectedPlan: []string{filepath.Join(".test", ".xxx")},
		},
		{
			name:          "include-hidden: generic globstar should include all hidden files (excluding gitignored)",
			targets:       []string{"**"},
			allowDotfiles: true,
			// Corrected: vendor/lib/lib.go is gitignored and --no-gitignore is not set.
			expectedPlan: []string{".config/settings.toml", ".env", ".gitignore", ".test/.xxx", ".test/kkk/oo/ll.go", "main.go", "README.md", "src/app.go"},
		},
		{
			name:          "include-hidden: explicit dot-directory pattern should include its hidden files",
			targets:       []string{".test/**"},
			allowDotfiles: true,
			expectedPlan:  []string{filepath.Join(".test", ".xxx"), filepath.Join(".test", "kkk", "oo", "ll.go")},
		},
		{
			name:            "include-hidden: combined with exclude pattern (excluding gitignored)",
			targets:         []string{"**"},
			excludePatterns: []string{"**/*.go"},
			allowDotfiles:   true,
			// Corrected: vendor/lib/lib.go is gitignored and --no-gitignore is not set.
			expectedPlan: []string{".config/settings.toml", ".env", ".gitignore", ".test/.xxx", "README.md"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := filter.Options{
				IncludePatterns: tc.includePatterns,
				ExcludePatterns: tc.excludePatterns,
				AllowDotfiles:   tc.allowDotfiles,
			}
			filterManager, _ := filter.NewManager(opts)
			detector := language.NewDetector()
			packer := NewPacker(nil, nil, filterManager, detector)

			plan, err := packer.Plan(tc.targets)
			if err != nil {
				t.Fatalf("Plan() returned an unexpected error: %v", err)
			}

			actualPaths := make([]string, len(plan))
			for i, p := range plan {
				actualPaths[i] = p.Path
			}

			sort.Strings(actualPaths)
			sort.Strings(tc.expectedPlan)

			if !reflect.DeepEqual(actualPaths, tc.expectedPlan) {
				t.Errorf("Plan() mismatch:\ngot:  %v\nwant: %v", actualPaths, tc.expectedPlan)
			}
		})
	}
}
