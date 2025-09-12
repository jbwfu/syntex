package project

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

// FindRoot uses go-git to find the root of the repository.
// It searches upwards from the given path for a .git directory.
// If the path is not part of a git repository, it returns the directory of the
// path as a fallback, allowing the tool to function in non-git environments.
func FindRoot(path string) (string, error) {
	repo, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		if errors.Is(err, git.ErrRepositoryNotExists) {
			return fallbackRoot(path)
		}
		return "", err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return "", err
	}

	return wt.Filesystem.Root(), nil
}

// fallbackRoot determines a sensible root directory when not inside a git repo.
func fallbackRoot(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", err
	}

	if info.IsDir() {
		return absPath, nil
	}

	return filepath.Dir(absPath), nil
}
