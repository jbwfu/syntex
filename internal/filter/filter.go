package filter

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/jbwfu/syntex/internal/project"
)

// gitignoreFilter internal struct to hold a matcher for a specific root.
type gitignoreFilter struct {
	matcher gitignore.Matcher
}

func newGitignoreFilter(rootPath string) (*gitignoreFilter, error) {
	fs := osfs.New(rootPath)
	patterns, err := gitignore.ReadPatterns(fs, nil)
	if err != nil {
		// .gitignore not existing is not an error, just means no patterns.
		if os.IsNotExist(err) {
			return &gitignoreFilter{matcher: gitignore.NewMatcher(nil)}, nil
		}
		return nil, err
	}
	// Always ignore the .git directory itself.
	patterns = append(patterns, gitignore.ParsePattern(".git", nil))
	return &gitignoreFilter{matcher: gitignore.NewMatcher(patterns)}, nil
}

func (f *gitignoreFilter) IsIgnored(pathRelativeToRoot string, isDir bool) bool {
	components := strings.Split(filepath.ToSlash(pathRelativeToRoot), "/")
	return f.matcher.Match(components, isDir)
}

// Manager orchestrates filtering logic with caching for multiple git repositories.
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
// It expects an absolute path for consistent matching.
func (m *Manager) IsGloballyExcluded(absPath string) bool {
	for _, pattern := range m.excludePatterns {
		// Match against the absolute path to handle patterns correctly.
		if match, _ := doublestar.Match(pattern, absPath); match {
			return true
		}
	}
	return false
}

// IsGitIgnored checks if a path is ignored by a .gitignore file from its
// containing repository. It uses a cache to avoid re-parsing .gitignore files.
// It expects an absolute path.
func (m *Manager) IsGitIgnored(absPath string, isDir bool) bool {
	if m.disableGitignore {
		return false
	}

	root, isRepo, err := project.FindRoot(filepath.Dir(absPath))
	if err != nil || !isRepo {
		// Not in a git repo or error finding root, so not ignored by git.
		return false
	}

	m.mu.Lock()
	filter, found := m.rootFilters[root]
	if !found {
		// Cache miss, create and store a new filter.
		var creationErr error
		filter, creationErr = newGitignoreFilter(root)
		if creationErr != nil {
			m.mu.Unlock()
			return false
		}
		m.rootFilters[root] = filter
	}
	m.mu.Unlock()

	relPath, err := filepath.Rel(root, absPath)
	if err != nil {
		// Cannot determine relative path, assume not ignored.
		return false
	}

	return filter.IsIgnored(relPath, isDir)
}
