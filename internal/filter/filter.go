package filter

import (
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

// Filter defines the interface for any path filtering logic.
type Filter interface {
	IsIgnored(path string, isDir bool) bool
}

// gitignoreFilter implements the Filter interface using .gitignore patterns.
type gitignoreFilter struct {
	matcher gitignore.Matcher
}

// newGitignoreFilter creates a filter by reading .gitignore patterns
// from the specified root path and its subdirectories.
func newGitignoreFilter(rootPath string) (Filter, error) {
	fs := osfs.New(rootPath)

	patterns, err := gitignore.ReadPatterns(fs, nil)
	if err != nil {
		return nil, err
	}
	patterns = append(patterns, gitignore.ParsePattern(".git", nil))

	return &gitignoreFilter{
		matcher: gitignore.NewMatcher(patterns),
	}, nil
}

func (f *gitignoreFilter) IsIgnored(path string, isDir bool) bool {
	components := strings.Split(filepath.ToSlash(path), "/")
	return f.matcher.Match(components, isDir)
}

// Manager orchestrates multiple filters according to predefined precedence.
type Manager struct {
	filters []Filter
}

// Options configures the behavior of the filter Manager.
type Options struct {
	DisableGitignore bool
}

// NewManager creates a new filter Manager based on the provided options.
func NewManager(rootPath string, opts Options) (*Manager, error) {
	var filters []Filter

	if !opts.DisableGitignore {
		gf, err := newGitignoreFilter(rootPath)
		if err != nil {
			// The error is returned to let the caller decide if it's fatal.
			return nil, err
		}
		filters = append(filters, gf)
	}

	return &Manager{
		filters: filters,
	}, nil
}

// IsIgnored checks the path against all configured filters.
// The first filter that ignores the path will cause it to be ignored.
func (m *Manager) IsIgnored(path string, isDir bool) bool {
	for _, f := range m.filters {
		if f.IsIgnored(path, isDir) {
			return true
		}
	}
	return false
}
