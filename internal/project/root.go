package project

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

// FindRoot uses go-git to find the root of the repository.
// It searches upwards from the given path for a .git directory.
// If a repository is found, it returns the repository root and isRepo=true.
// If not, it returns a fallback directory and isRepo=false.
func FindRoot(path string) (string, bool, error) {
	// Ensure we start with an absolute path to make discovery consistent.
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", false, err
	}

	repo, err := git.PlainOpenWithOptions(absPath, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		if errors.Is(err, git.ErrRepositoryNotExists) {
			// Not a git repo, return a sensible fallback root.
			fallback, err := fallbackRoot(absPath)
			return fallback, false, err
		}
		return "", false, err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return "", false, err
	}

	return wt.Filesystem.Root(), true, nil
}

// fallbackRoot determines a sensible root directory when not inside a git repo.
func fallbackRoot(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		// If path doesn't exist, its directory might, for glob patterns.
		if os.IsNotExist(err) {
			return filepath.Dir(path), nil
		}
		return "", err
	}

	if info.IsDir() {
		return path, nil
	}

	return filepath.Dir(path), nil
}
