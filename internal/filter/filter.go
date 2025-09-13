package filter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/jbwfu/syntex/internal/project"
)

// gitignoreFilter holds a gitignore matcher for a specific repository root.
type gitignoreFilter struct {
	matcher gitignore.Matcher
}

// newGitignoreFilter creates a filter by reading .gitignore patterns from a given root path.
func newGitignoreFilter(rootPath string) (*gitignoreFilter, error) {
	fs := osfs.New(rootPath)
	patterns, err := gitignore.ReadPatterns(fs, nil)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("could not read gitignore patterns: %w", err)
	}

	// Always ignore the .git directory itself.
	patterns = append(patterns, gitignore.ParsePattern(".git", nil))
	return &gitignoreFilter{matcher: gitignore.NewMatcher(patterns)}, nil
}

// isIgnored checks if a path relative to the filter's root is ignored.
func (f *gitignoreFilter) isIgnored(pathRelativeToRoot string, isDir bool) bool {
	components := strings.Split(filepath.ToSlash(pathRelativeToRoot), "/")
	return f.matcher.Match(components, isDir)
}

// Manager orchestrates file filtering logic. It handles user-defined exclude
// patterns and dynamically applies .gitignore rules from multiple repositories
// with an internal cache to optimize performance.
type Manager struct {
	includePatterns  []string
	excludePatterns  []string
	disableGitignore bool

	mu          sync.Mutex
	rootFilters map[string]*gitignoreFilter
}

// Options configures the behavior of the filter Manager.
type Options struct {
	DisableGitignore bool
	ExcludePatterns  []string
	IncludePatterns  []string
}

// NewManager creates a new filter Manager.
func NewManager(opts Options) (*Manager, error) {
	return &Manager{
		includePatterns:  opts.IncludePatterns,
		excludePatterns:  opts.ExcludePatterns,
		disableGitignore: opts.DisableGitignore,
		rootFilters:      make(map[string]*gitignoreFilter),
	}, nil
}

// GetIncludePatterns returns the configured include patterns.
func (m *Manager) GetIncludePatterns() []string {
	return m.includePatterns
}

// IsGloballyExcluded checks if a path matches any user-defined --exclude patterns.
// It provides robust filtering by attempting to match patterns against both
// the absolute path and the path relative to the current working directory (CWD).
func (m *Manager) IsGloballyExcluded(absPath string) bool {
	cwd, getCwdErr := os.Getwd()
	if getCwdErr != nil {
		fmt.Fprintf(os.Stderr, "warning: could not get current working directory for global exclusion: %v\n", getCwdErr)
	}

	for _, pattern := range m.excludePatterns {
		// Try matching against the absolute path first. This handles absolute exclude patterns.
		if match, _ := doublestar.Match(pattern, absPath); match {
			return true
		}

		if getCwdErr == nil {
			relPath, err := filepath.Rel(cwd, absPath)
			if err == nil {
				if match, _ := doublestar.Match(pattern, relPath); match {
					return true
				}
			}
		}
	}
	return false
}

// IsGitIgnored checks if a path is ignored by a .gitignore file from its
// containing repository. It expects an absolute path to correctly determine
// the repository context. It uses a cache to avoid re-parsing .gitignore files.
func (m *Manager) IsGitIgnored(absPath string, isDir bool) bool {
	if m.disableGitignore {
		return false
	}

	// Find the git repository root for the given path.
	root, isRepo, err := project.FindRoot(filepath.Dir(absPath))
	if err != nil || !isRepo {
		// Not in a git repo or an error occurred, so it cannot be git-ignored.
		return false
	}

	filter, err := m.getOrCreateFilter(root)
	if err != nil {
		// If filter creation fails, conservatively assume the file is not ignored.
		// Consider logging this error in a real application.
		return false
	}

	relPath, err := filepath.Rel(root, absPath)
	if err != nil {
		// Cannot determine relative path, so cannot apply ignore rules.
		return false
	}

	return filter.isIgnored(relPath, isDir)
}

// getOrCreateFilter retrieves a gitignoreFilter from the cache or creates a new one.
// This method is thread-safe.
func (m *Manager) getOrCreateFilter(root string) (*gitignoreFilter, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if filter, found := m.rootFilters[root]; found {
		return filter, nil
	}

	filter, err := newGitignoreFilter(root)
	if err != nil {
		return nil, err
	}

	m.rootFilters[root] = filter
	return filter, nil
}
