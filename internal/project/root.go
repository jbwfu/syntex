package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

// FindRoot searches upwards from a given path to find the root of a Git repository.
// It returns the absolute path to the repository root and a boolean `isRepo` which is true
// if a repository was found. If no repository is found, it returns a sensible fallback
// directory (e.g., the directory of the given path) and `isRepo` will be false.
func FindRoot(path string) (root string, isRepo bool, err error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", false, err
	}

	repo, err := git.PlainOpenWithOptions(absPath, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		if errors.Is(err, git.ErrRepositoryNotExists) {
			fallback, err := fallbackRoot(absPath)
			return fallback, false, err
		}
		return "", false, err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return "", false, fmt.Errorf("failed to get worktree: %w", err)
	}

	return wt.Filesystem.Root(), true, nil
}

// fallbackRoot determines a sensible root directory when not inside a git repo.
func fallbackRoot(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		// If the path doesn't exist (e.g., it's part of a glob pattern),
		// its parent directory is the best guess for a root.
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
