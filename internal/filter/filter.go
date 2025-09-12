package filter

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

type gitignoreFilter struct {
	matcher gitignore.Matcher
}

func newGitignoreFilter(rootPath string) (*gitignoreFilter, error) {
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

// Manager orchestrates filtering logic.
type Manager struct {
	includePatterns []string
	excludePatterns []string
	gitignoreFilter *gitignoreFilter
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

// GetIncludePatterns returns the configured include patterns, expanding directory
// paths to match all contents.
func (m *Manager) GetIncludePatterns() []string {
	patterns := make([]string, 0, len(m.includePatterns))
	for _, p := range m.includePatterns {
		if strings.HasSuffix(p, string(os.PathSeparator)) {
			patterns = append(patterns, filepath.Join(p, "**"))
		} else {
			patterns = append(patterns, p)
		}
	}
	return patterns
}

// IsExcluded checks if a path matches any exclusion pattern (.gitignore or --exclude).
func (m *Manager) IsExcluded(path string, isDir bool) bool {
	osPath := filepath.FromSlash(path)
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
