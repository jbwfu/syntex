package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
)

// setupTestEnvironment creates a temporary directory and initializes a git repository if requested.
// It returns the root path of the temporary environment and a cleanup function.
func setupTestEnvironment(t *testing.T, initGit bool) (string, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "syntex-project-test-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	if initGit {
		_, err = git.PlainInit(tempDir, false)
		if err != nil {
			os.RemoveAll(tempDir) // Clean up on failure
			t.Fatalf("failed to init git repo in %s: %v", tempDir, err)
		}
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}
	return tempDir, cleanup
}

func TestFindRoot(t *testing.T) {
	t.Run("inside git repository", func(t *testing.T) {
		repoRoot, cleanup := setupTestEnvironment(t, true)
		defer cleanup()

		// Create a nested directory to test finding root from a subdirectory
		nestedDir := filepath.Join(repoRoot, "src", "app")
		err := os.MkdirAll(nestedDir, 0755)
		if err != nil {
			t.Fatalf("failed to create nested dir: %v", err)
		}

		root, isRepo, err := FindRoot(nestedDir)
		if err != nil {
			t.Errorf("FindRoot(%q) returned error: %v", nestedDir, err)
		}
		if !isRepo {
			t.Errorf("FindRoot(%q) expected to be a repo, got isRepo=false", nestedDir)
		}
		if root != repoRoot {
			t.Errorf("FindRoot(%q) got root %q, want %q", nestedDir, root, repoRoot)
		}
	})

	t.Run("not inside git repository (path is dir)", func(t *testing.T) {
		nonRepoDir, cleanup := setupTestEnvironment(t, false)
		defer cleanup()

		root, isRepo, err := FindRoot(nonRepoDir)
		if err != nil {
			t.Errorf("FindRoot(%q) returned error: %v", nonRepoDir, err)
		}
		if isRepo {
			t.Errorf("FindRoot(%q) expected not to be a repo, got isRepo=true", nonRepoDir)
		}
		if root != nonRepoDir {
			t.Errorf("FindRoot(%q) got root %q, want %q", nonRepoDir, root, nonRepoDir)
		}
	})

	t.Run("not inside git repository (path is file)", func(t *testing.T) {
		tempDir, cleanup := setupTestEnvironment(t, false)
		defer cleanup()

		testFile := filepath.Join(tempDir, "test.txt")
		err := os.WriteFile(testFile, []byte("content"), 0644)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		root, isRepo, err := FindRoot(testFile)
		if err != nil {
			t.Errorf("FindRoot(%q) returned error: %v", testFile, err)
		}
		if isRepo {
			t.Errorf("FindRoot(%q) expected not to be a repo, got isRepo=true", testFile)
		}
		if root != tempDir { // Fallback should be the directory of the file
			t.Errorf("FindRoot(%q) got root %q, want %q", testFile, root, tempDir)
		}
	})

	t.Run("path does not exist", func(t *testing.T) {
		nonExistentPath := filepath.Join(os.TempDir(), "non-existent-dir", "non-existent-file.txt")
		expectedFallbackRoot := filepath.Dir(nonExistentPath) // Fallback should be the parent dir

		root, isRepo, err := FindRoot(nonExistentPath)
		if err != nil {
			t.Errorf("FindRoot(%q) returned error: %v", nonExistentPath, err)
		}
		if isRepo {
			t.Errorf("FindRoot(%q) expected not to be a repo, got isRepo=true", nonExistentPath)
		}
		if root != expectedFallbackRoot {
			t.Errorf("FindRoot(%q) got root %q, want %q", nonExistentPath, root, expectedFallbackRoot)
		}
	})
}

func TestFallbackRoot(t *testing.T) {
	t.Run("path is directory", func(t *testing.T) {
		tempDir, cleanup := setupTestEnvironment(t, false)
		defer cleanup()

		root, err := fallbackRoot(tempDir)
		if err != nil {
			t.Errorf("fallbackRoot(%q) returned error: %v", tempDir, err)
		}
		if root != tempDir {
			t.Errorf("fallbackRoot(%q) got root %q, want %q", tempDir, root, tempDir)
		}
	})

	t.Run("path is file", func(t *testing.T) {
		tempDir, cleanup := setupTestEnvironment(t, false)
		defer cleanup()

		testFile := filepath.Join(tempDir, "test.txt")
		err := os.WriteFile(testFile, []byte("content"), 0644)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		root, err := fallbackRoot(testFile)
		if err != nil {
			t.Errorf("fallbackRoot(%q) returned error: %v", testFile, err)
		}
		if root != tempDir { // Fallback should be the directory of the file
			t.Errorf("fallbackRoot(%q) got root %q, want %q", testFile, root, tempDir)
		}
	})

	t.Run("path does not exist", func(t *testing.T) {
		nonExistentPath := filepath.Join(os.TempDir(), "non-existent-dir", "non-existent-file.txt")
		expectedFallbackRoot := filepath.Dir(nonExistentPath) // Fallback should be the parent dir

		root, err := fallbackRoot(nonExistentPath)
		if err != nil {
			t.Errorf("fallbackRoot(%q) returned error: %v", nonExistentPath, err)
		}
		if root != expectedFallbackRoot {
			t.Errorf("fallbackRoot(%q) got root %q, want %q", nonExistentPath, root, expectedFallbackRoot)
		}
	})
}
