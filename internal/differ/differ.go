package differ

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"codeberg.org/devcarlosmolero/vibewatch/internal/types"
)

var (
	debugFile  *os.File
	debugMutex sync.Mutex
)

// Differ computes diffs for changed files.
type Differ interface {
	Diff(filePath string) (types.DiffEntry, error)
	// DirtyFiles returns DiffEntries for all files with uncommitted changes.
	DirtyFiles() ([]types.DiffEntry, error)
}

// cacheEntry represents a cached diff result
type cacheEntry struct {
	diff      string
	timestamp time.Time
	error     string
	isNew     bool
}

// GitDiffer uses git to compute diffs.
type GitDiffer struct {
	root       string
	diffCache  map[string]cacheEntry
	cacheMutex sync.Mutex
}

// logMessage writes a debug message to the debug file
func logMessage(message string) {
	debugMutex.Lock()
	defer debugMutex.Unlock()

	if debugFile != nil {
		timestamp := time.Now().Format("15:04:05")
		debugFile.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, message))
		debugFile.Sync()
	}
}

// NewGit creates a GitDiffer after verifying the directory is a git repo.
func NewGit(root string) (*GitDiffer, error) {
	cmd := exec.Command("git", "-C", root, "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("not a git repository: %s", root)
	}
	return &GitDiffer{
		root:      root,
		diffCache: make(map[string]cacheEntry),
	}, nil
}

// Diff computes the diff for a single file.
func (g *GitDiffer) Diff(filePath string) (types.DiffEntry, error) {
	if filePath == "__GIT_OPERATION__" {
		return types.DiffEntry{
			FilePath:  filePath,
			Timestamp: time.Now(),
			Repo:      filepath.Base(g.root),
			Diff:      "",
			Error:     "",
			IsNew:     false,
		}, nil
	}

	entry := types.DiffEntry{
		FilePath:  filePath,
		Timestamp: time.Now(),
		Repo:      filepath.Base(g.root),
	}

	g.cacheMutex.Lock()
	if cached, exists := g.diffCache[filePath]; exists {
		if time.Since(cached.timestamp) < 1*time.Second {
			entry.Diff = cached.diff
			entry.Error = cached.error
			entry.IsNew = cached.isNew
			g.cacheMutex.Unlock()
			return entry, nil
		}
	}
	g.cacheMutex.Unlock()

	rel, err := filepath.Rel(g.root, filePath)
	if err != nil {
		rel = filePath
	}

	diff, err := g.gitDiff(rel)
	if err != nil {
		entry.Error = err.Error()
		g.cacheResult(filePath, entry)
		return entry, nil
	}

	if diff == "" {
		diff, err = g.gitDiffStaged(rel)
		if err != nil {
			entry.Error = err.Error()
			g.cacheResult(filePath, entry)
			return entry, nil
		}
	}

	if diff == "" {
		tracked, _ := g.isTracked(rel)
		if !tracked {
			diff, err = g.gitDiffUntracked(filePath)
			if err != nil {
				entry.Error = err.Error()
				g.cacheResult(filePath, entry)
				return entry, nil
			}
			entry.IsNew = true
		} else {
			logMessage(fmt.Sprintf("Differ: File committed/clean, clearing diff: %s", filePath))
			entry.Diff = ""
			entry.Error = ""
			entry.IsNew = false
			g.cacheResult(filePath, entry)
			return entry, nil
		}
	}

	entry.Diff = diff
	g.cacheResult(filePath, entry)
	return entry, nil
}

// DirtyFiles returns DiffEntries for all files with uncommitted changes in this repo.
func (g *GitDiffer) DirtyFiles() ([]types.DiffEntry, error) {
	cmd := exec.Command("git", "-C", g.root, "diff", "--name-only", "HEAD")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &bytes.Buffer{}
	if err := cmd.Run(); err != nil {
		cmd = exec.Command("git", "-C", g.root, "diff", "--name-only")
		out.Reset()
		cmd.Stdout = &out
		cmd.Stderr = &bytes.Buffer{}
		cmd.Run()
	}

	cmd2 := exec.Command("git", "-C", g.root, "diff", "--name-only")
	var out2 bytes.Buffer
	cmd2.Stdout = &out2
	cmd2.Stderr = &bytes.Buffer{}
	cmd2.Run()

	cmd3 := exec.Command("git", "-C", g.root, "ls-files", "--others", "--exclude-standard")
	var out3 bytes.Buffer
	cmd3.Stdout = &out3
	cmd3.Stderr = &bytes.Buffer{}
	cmd3.Run()

	seen := make(map[string]bool)
	var relPaths []string
	for _, o := range []string{out.String(), out2.String(), out3.String()} {
		for _, line := range strings.Split(strings.TrimSpace(o), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || seen[line] {
				continue
			}
			seen[line] = true
			relPaths = append(relPaths, line)
		}
	}

	var entries []types.DiffEntry
	for _, relPath := range relPaths {
		lower := strings.ToLower(relPath)
		if strings.HasSuffix(lower, ".tmp") || strings.HasSuffix(lower, ".log") ||
			strings.HasSuffix(lower, ".bak") || strings.HasSuffix(lower, ".swp") {
			continue
		}
		absPath := filepath.Join(g.root, relPath)
		entry, _ := g.Diff(absPath)
		if entry.Diff != "" || entry.IsNew {
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

func (g *GitDiffer) gitDiff(relPath string) (string, error) {
	cmd := exec.Command("git", "-C", g.root, "diff", "--no-color", "--unified=3", "--", relPath)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &bytes.Buffer{}
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

func (g *GitDiffer) gitDiffStaged(relPath string) (string, error) {
	cmd := exec.Command("git", "-C", g.root, "diff", "--no-color", "--unified=3", "--cached", "--", relPath)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &bytes.Buffer{}
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

func (g *GitDiffer) isTracked(relPath string) (bool, error) {
	cmd := exec.Command("git", "-C", g.root, "ls-files", "--error-unmatch", relPath)
	cmd.Stdout = &bytes.Buffer{}
	cmd.Stderr = &bytes.Buffer{}
	err := cmd.Run()
	return err == nil, nil
}

func (g *GitDiffer) cacheResult(filePath string, entry types.DiffEntry) {
	g.cacheMutex.Lock()
	defer g.cacheMutex.Unlock()
	g.diffCache[filePath] = cacheEntry{
		diff:      entry.Diff,
		timestamp: time.Now(),
		error:     entry.Error,
		isNew:     entry.IsNew,
	}
}

func (g *GitDiffer) gitDiffUntracked(absPath string) (string, error) {
	cmd := exec.Command("git", "diff", "--no-color", "--no-index", "--", "/dev/null", absPath)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &bytes.Buffer{}
	cmd.Run()
	return strings.TrimSpace(out.String()), nil
}

// ClearCache clears the diff cache, useful when git state changes (commit, reset, etc.)
func (g *GitDiffer) ClearCache() {
	g.cacheMutex.Lock()
	g.diffCache = make(map[string]cacheEntry)
	g.cacheMutex.Unlock()
}

// MultiDiffer routes file paths to the correct GitDiffer based on which repo
// the file belongs to. Used when watching a parent directory containing multiple repos.
type MultiDiffer struct {
	repos []repoEntry // sorted longest-path-first for correct matching
}

type repoEntry struct {
	root   string
	name   string
	differ *GitDiffer
}

// NewMulti creates a MultiDiffer from a list of discovered repo roots.
func NewMulti(repos map[string]string) (*MultiDiffer, error) {
	if len(repos) == 0 {
		return nil, fmt.Errorf("no git repositories found")
	}

	var entries []repoEntry
	for root, name := range repos {
		d, err := NewGit(root)
		if err != nil {
			continue // skip repos that fail init
		}
		entries = append(entries, repoEntry{root: root, name: name, differ: d})
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("no valid git repositories found")
	}

	// Sort longest path first so nested repos match before parents
	sort.Slice(entries, func(i, j int) bool {
		return len(entries[i].root) > len(entries[j].root)
	})

	return &MultiDiffer{repos: entries}, nil
}

// Diff finds the matching repo for the file path and computes the diff.
func (m *MultiDiffer) Diff(filePath string) (types.DiffEntry, error) {
	for _, repo := range m.repos {
		if strings.HasPrefix(filePath, repo.root+string(filepath.Separator)) || filePath == repo.root {
			entry, err := repo.differ.Diff(filePath)
			entry.Repo = repo.name
			return entry, err
		}
	}
	return types.DiffEntry{
		FilePath:  filePath,
		Timestamp: time.Now(),
		Error:     "file not inside any known git repository",
	}, nil
}

// DirtyFiles returns DiffEntries for all dirty files across all repos.
func (m *MultiDiffer) DirtyFiles() ([]types.DiffEntry, error) {
	var all []types.DiffEntry
	for _, repo := range m.repos {
		entries, err := repo.differ.DirtyFiles()
		if err != nil {
			continue
		}
		for i := range entries {
			entries[i].Repo = repo.name
		}
		all = append(all, entries...)
	}
	return all, nil
}

// RepoRoots returns the root paths of all discovered repos.
func (m *MultiDiffer) RepoRoots() []string {
	roots := make([]string, len(m.repos))
	for i, r := range m.repos {
		roots[i] = r.root
	}
	return roots
}

// DiscoverRepos walks a directory and returns a map of
// repo root path -> repo name for all directories containing a .git folder.
// Once a repo is found, it does not descend further into it.
func DiscoverRepos(parentDir string) (map[string]string, error) {
	repos := make(map[string]string)

	err := filepath.WalkDir(parentDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		// Don't descend into .git directories themselves
		if d.Name() == ".git" {
			return filepath.SkipDir
		}
		// Check if this directory contains a .git folder (i.e. is a repo root)
		gitDir := filepath.Join(path, ".git")
		if _, statErr := os.Stat(gitDir); statErr == nil {
			repos[path] = filepath.Base(path)
			return filepath.SkipDir // don't descend into repos
		}
		return nil
	})

	return repos, err
}

// GetBranch returns the current branch name for a git repo.
func GetBranch(repoRoot string) string {
	cmd := exec.Command("git", "-C", repoRoot, "rev-parse", "--abbrev-ref", "HEAD")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &bytes.Buffer{}
	if err := cmd.Run(); err != nil {
		return ""
	}
	return strings.TrimSpace(out.String())
}

// IsGitRepo checks whether a directory is itself a git repository.
func IsGitRepo(dir string) bool {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--git-dir")
	cmd.Stdout = &bytes.Buffer{}
	cmd.Stderr = &bytes.Buffer{}
	return cmd.Run() == nil
}

// IsGitIgnored uses git check-ignore to determine if a path is ignored
// by any .gitignore in the repo hierarchy. Returns true if ignored.
func IsGitIgnored(repoRoot, absPath string) bool {
	cmd := exec.Command("git", "-C", repoRoot, "check-ignore", "-q", absPath)
	cmd.Stdout = &bytes.Buffer{}
	cmd.Stderr = &bytes.Buffer{}
	return cmd.Run() == nil // exit 0 = ignored, exit 1 = not ignored
}

// FindRepoRoot returns the git repo root for a given file path,
// or empty string if the file is not inside any of the provided repo roots.
func FindRepoRoot(filePath string, repoRoots []string) string {
	// Check longest paths first (already sorted in MultiDiffer)
	for _, root := range repoRoots {
		if strings.HasPrefix(filePath, root+string(filepath.Separator)) {
			return root
		}
	}
	return ""
}
