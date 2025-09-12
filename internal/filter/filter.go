package filter

import (
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

// Filter defines the interface for any path filtering logic.
type Filter interface {
	IsIgnored(path string, isDir bool) bool
}

type gitignoreFilter struct {
	matcher gitignore.Matcher
}

func newGitignoreFilter(rootPath string) (Filter, error) {
	fs := osfs.New(rootPath)
	patterns, err := gitignore.ReadPatterns(fs, nil)
	if err != nil {
		return nil, err
	}
	patterns = append(patterns, gitignore.ParsePattern(".git", nil))
	return &gitignoreFilter{matcher: gitignore.NewMatcher(patterns)}, nil
}

func (f *gitignoreFilter) IsIgnored(path string, isDir bool) bool {
	components := strings.Split(filepath.ToSlash(path), "/")
	return f.matcher.Match(components, isDir)
}

// Manager orchestrates multiple filters according to predefined precedence.
type Manager struct {
	includePatterns []string
	excludePatterns []string
	gitignoreFilter Filter
}

// Options configures the behavior of the filter Manager.
type Options struct {
	DisableGitignore bool
	ExcludePatterns  []string
	IncludePatterns  []string
}

// NewManager creates a new filter Manager based on the provided options.
func NewManager(rootPath string, opts Options) (*Manager, error) {
	m := &Manager{
		includePatterns: opts.IncludePatterns,
		excludePatterns: opts.ExcludePatterns,
	}

	if !opts.DisableGitignore {
		var err error
		m.gitignoreFilter, err = newGitignoreFilter(rootPath)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}

// IsIgnored checks the path against all configured filters with correct precedence.
// Precedence: Inclusion > Exclusion > Gitignore
func (m *Manager) IsIgnored(path string, isDir bool) bool {
	// doublestar expects OS-specific separators for matching.
	osPath := filepath.FromSlash(path)

	for _, pattern := range m.includePatterns {
		if match, _ := doublestar.Match(pattern, osPath); match {
			return false
		}
	}

	for _, pattern := range m.excludePatterns {
		if match, _ := doublestar.Match(pattern, osPath); match {
			return true
		}
	}

	if m.gitignoreFilter != nil && m.gitignoreFilter.IsIgnored(path, isDir) {
		return true
	}

	return false
}
