package watcher

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"codeberg.org/devcarlosmolero/vibewatch/internal/differ"
)

var builtinIgnores = []string{
	".git",
	"node_modules",
	".next",
	"dist",
	"build",
	"__pycache__",
	".venv",
	"venv",
	".tox",
	".mypy_cache",
	".pytest_cache",
	".DS_Store",
	".swp",
	".swo",
	"~",
	// "vibewatch", // Removed - was causing issues with directory filtering
}

var ignoredExtensions = []string{
	".log",
	".tmp",
	".bak",
	".pid",
	".lock",
	".exe",   // Binary executables
	".dll",   // DLL files
	".so",    // Shared libraries
	".dylib", // macOS libraries
	".a",     // Static libraries
}

// Filter decides which paths should be ignored by the watcher.
type Filter struct {
	root      string
	repoRoots []string // git repo roots for git check-ignore
}

// NewFilter creates a filter that respects each repo's .gitignore and built-in exclusions.
func NewFilter(root string, repoRoots []string) *Filter {
	return &Filter{root: root, repoRoots: repoRoots}
}

// ShouldIgnore returns true if the path should be excluded from watching.
func (f *Filter) ShouldIgnore(path string) bool {
	base := filepath.Base(path)

	// Special case: allow .git/HEAD to detect git operations like commits
	if base == "HEAD" && strings.Contains(path, ".git") {
		return false
	}

	// Check if this is a Go binary or build artifact
	if strings.HasSuffix(path, ".exe") || strings.HasSuffix(path, ".so") ||
		strings.HasSuffix(path, ".dylib") || strings.HasSuffix(path, ".a") ||
		strings.HasSuffix(path, ".o") || strings.HasSuffix(path, ".out") {
		return true
	}

	// Check for files with long numbers (like watcher.go123456789)
	// Pattern: filename + numbers (8+ digits)
	if match, _ := regexp.MatchString(`\.[0-9]{8,}$`, base); match {
		return true
	}

	// Check for executable files without extensions (common for Unix binaries)
	if !strings.Contains(base, ".") {
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() && info.Mode()&0111 != 0 {
			// This is an executable file without extension, likely a binary
			return true
		}
	}

	for _, pattern := range builtinIgnores {
		if base == pattern || strings.HasSuffix(base, pattern) {
			return true
		}
	}

	lowerBase := strings.ToLower(base)
	for _, ie := range ignoredExtensions {
		if strings.HasSuffix(lowerBase, ie) {
			return true
		}
	}

	// Check any path component against directory ignores
	// EXCEPT for the root directory itself (e.g., "vibewatch" repo)
	rel, err := filepath.Rel(f.root, path)
	if err == nil && rel != "." {
		parts := strings.Split(rel, string(os.PathSeparator))
		for _, part := range parts {
			for _, pattern := range builtinIgnores {
				// Skip if this is the first component (the repo name itself)
				if part == parts[0] && part == "vibewatch" {
					continue
				}
				if part == pattern {
					return true
				}
			}
		}
	}

	// Use git check-ignore for the matching repo
	repoRoot := differ.FindRepoRoot(path, f.repoRoots)
	if repoRoot != "" {
		if differ.IsGitIgnored(repoRoot, path) {
			return true
		}
	}

	return false
}
