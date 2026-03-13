package differ

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"codeberg.org/devcarlosmolero/vibewatch/internal/types"
)

func TestNewGit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-git-repo")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	gitDiffer, err := NewGit(tmpDir)
	if err != nil {
		t.Errorf("Unexpected error creating GitDiffer: %v", err)
	}
	if gitDiffer == nil {
		t.Error("Expected non-nil GitDiffer")
	}

	_, err = NewGit("/tmp/non-existent")
	if err == nil {
		t.Error("Expected error for non-git directory")
	}
}

func TestGitDifferDiff(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-git-diff")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	gitDiffer, err := NewGit(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create GitDiffer: %v", err)
	}

	var entry types.DiffEntry
	entry, err = gitDiffer.Diff("/tmp/outside.txt")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	entry, err = gitDiffer.Diff("__GIT_OPERATION__")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if entry.FilePath != "__GIT_OPERATION__" {
		t.Errorf("Expected git operation marker, got %s", entry.FilePath)
	}
	if entry.Diff != "" {
		t.Errorf("Expected empty diff for git operation, got %s", entry.Diff)
	}

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("initial content\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set git email: %v", err)
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set git name: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	entry, err = gitDiffer.Diff(testFile)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if entry.Diff != "" {
		t.Errorf("Expected empty diff for committed file, got %s", entry.Diff)
	}

	time.Sleep(1100 * time.Millisecond)

	if err := os.WriteFile(testFile, []byte("modified content\n"), 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	entry, err = gitDiffer.Diff(testFile)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if entry.Diff == "" {
		t.Error("Expected non-empty diff for modified file")
	}

	untrackedFile := filepath.Join(tmpDir, "untracked.txt")
	if err := os.WriteFile(untrackedFile, []byte("untracked content\n"), 0644); err != nil {
		t.Fatalf("Failed to create untracked file: %v", err)
	}

	time.Sleep(1100 * time.Millisecond)

	entry, err = gitDiffer.Diff(untrackedFile)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if entry.Diff == "" {
		t.Error("Expected non-empty diff for untracked file")
	}
	if !entry.IsNew {
		t.Error("Expected IsNew to be true for untracked file")
	}

	nonExistentFile := filepath.Join(tmpDir, "non-existent.txt")
	time.Sleep(1100 * time.Millisecond)
	entry, err = gitDiffer.Diff(nonExistentFile)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestGitDifferDirtyFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-dirty-files")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	gitDiffer, err := NewGit(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create GitDiffer: %v", err)
	}

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("initial content\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set git email: %v", err)
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set git name: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	entries, err := gitDiffer.DirtyFiles()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected 0 dirty files, got %d", len(entries))
	}

	if err := os.WriteFile(testFile, []byte("modified content\n"), 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	entries, err = gitDiffer.DirtyFiles()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 dirty file, got %d", len(entries))
	}
}

func TestGitDifferCache(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-cache")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	gitDiffer, err := NewGit(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create GitDiffer: %v", err)
	}

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("initial content\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set git email: %v", err)
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set git name: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	_, err = gitDiffer.Diff(testFile)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(gitDiffer.diffCache) != 1 {
		t.Errorf("Expected cache to have 1 entry, got %d", len(gitDiffer.diffCache))
	}

	entry, err := gitDiffer.Diff(testFile)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if entry.Diff != "" {
		t.Errorf("Expected empty diff from cache, got %s", entry.Diff)
	}
}

func TestGitDifferClearCache(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-clear-cache")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	gitDiffer, err := NewGit(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create GitDiffer: %v", err)
	}

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("initial content\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set git email: %v", err)
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set git name: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	_, err = gitDiffer.Diff(testFile)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(gitDiffer.diffCache) != 1 {
		t.Errorf("Expected cache to have 1 entry before clear, got %d", len(gitDiffer.diffCache))
	}

	gitDiffer.ClearCache()

	if len(gitDiffer.diffCache) != 0 {
		t.Errorf("Expected cache to be empty after clear, got %d entries", len(gitDiffer.diffCache))
	}
}

func TestGitDifferRoot(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-root")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	gitDiffer, err := NewGit(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create GitDiffer: %v", err)
	}

	root := gitDiffer.Root()
	if root != tmpDir {
		t.Errorf("Expected root %s, got %s", tmpDir, root)
	}
}

func TestGitDifferRepoRoots(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-repo-roots")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	gitDiffer, err := NewGit(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create GitDiffer: %v", err)
	}

	roots := gitDiffer.RepoRoots()
	if len(roots) != 1 {
		t.Errorf("Expected 1 repo root, got %d", len(roots))
	}
	if roots[0] != tmpDir {
		t.Errorf("Expected root %s, got %s", tmpDir, roots[0])
	}
}

func TestGitDifferRepoRootsWithNames(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-repo-roots-names")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	gitDiffer, err := NewGit(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create GitDiffer: %v", err)
	}

	roots := gitDiffer.RepoRootsWithNames()
	if len(roots) != 1 {
		t.Errorf("Expected 1 repo root, got %d", len(roots))
	}
	expectedName := filepath.Base(tmpDir)
	if roots[tmpDir] != expectedName {
		t.Errorf("Expected repo name %s, got %s", expectedName, roots[tmpDir])
	}
}

func TestNewMulti(t *testing.T) {
	repo1, err := os.MkdirTemp("", "test-multi-repo1")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(repo1)

	repo2, err := os.MkdirTemp("", "test-multi-repo2")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(repo2)

	for _, repo := range []string{repo1, repo2} {
		cmd := exec.Command("git", "init")
		cmd.Dir = repo
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to initialize git repo: %v", err)
		}
	}

	repos := map[string]string{
		repo1: "repo1",
		repo2: "repo2",
	}

	multiDiffer, err := NewMulti(repos)
	if err != nil {
		t.Errorf("Unexpected error creating MultiDiffer: %v", err)
	}
	if multiDiffer == nil {
		t.Error("Expected non-nil MultiDiffer")
	}

	_, err = NewMulti(map[string]string{})
	if err == nil {
		t.Error("Expected error for empty repos map")
	}
}

func TestMultiDifferDiff(t *testing.T) {
	repo1, err := os.MkdirTemp("", "test-multi-diff1")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(repo1)

	cmd := exec.Command("git", "init")
	cmd.Dir = repo1
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	repos := map[string]string{
		repo1: "repo1",
	}

	multiDiffer, err := NewMulti(repos)
	if err != nil {
		t.Fatalf("Failed to create MultiDiffer: %v", err)
	}

	testFile := filepath.Join(repo1, "test.txt")
	if err := os.WriteFile(testFile, []byte("initial content\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = repo1
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = repo1
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set git email: %v", err)
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = repo1
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set git name: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = repo1
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	entry, err := multiDiffer.Diff(testFile)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if entry.Repo != "repo1" {
		t.Errorf("Expected repo name 'repo1', got %s", entry.Repo)
	}

	entry, err = multiDiffer.Diff("/tmp/non-existent.txt")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if entry.Error != "file not inside any known git repository" {
		t.Errorf("Expected error message, got %s", entry.Error)
	}
}

func TestMultiDifferDirtyFiles(t *testing.T) {
	repo1, err := os.MkdirTemp("", "test-multi-dirty1")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(repo1)

	cmd := exec.Command("git", "init")
	cmd.Dir = repo1
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	repos := map[string]string{
		repo1: "repo1",
	}

	multiDiffer, err := NewMulti(repos)
	if err != nil {
		t.Fatalf("Failed to create MultiDiffer: %v", err)
	}

	testFile := filepath.Join(repo1, "test.txt")
	if err := os.WriteFile(testFile, []byte("initial content\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = repo1
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = repo1
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set git email: %v", err)
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = repo1
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set git name: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = repo1
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	entries, err := multiDiffer.DirtyFiles()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected 0 dirty files, got %d", len(entries))
	}

	if err := os.WriteFile(testFile, []byte("modified content\n"), 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	entries, err = multiDiffer.DirtyFiles()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 dirty file, got %d", len(entries))
	}
	if entries[0].Repo != "repo1" {
		t.Errorf("Expected repo name 'repo1', got %s", entries[0].Repo)
	}
}

func TestMultiDifferRepoRoots(t *testing.T) {
	repo1, err := os.MkdirTemp("", "test-multi-roots1")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(repo1)

	repo2, err := os.MkdirTemp("", "test-multi-roots2")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(repo2)

	for _, repo := range []string{repo1, repo2} {
		cmd := exec.Command("git", "init")
		cmd.Dir = repo
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to initialize git repo: %v", err)
		}
	}

	repos := map[string]string{
		repo1: "repo1",
		repo2: "repo2",
	}

	multiDiffer, err := NewMulti(repos)
	if err != nil {
		t.Fatalf("Failed to create MultiDiffer: %v", err)
	}

	roots := multiDiffer.RepoRoots()
	if len(roots) != 2 {
		t.Errorf("Expected 2 repo roots, got %d", len(roots))
	}
}

func TestMultiDifferRepoRootsWithNames(t *testing.T) {
	repo1, err := os.MkdirTemp("", "test-multi-roots-names1")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(repo1)

	repo2, err := os.MkdirTemp("", "test-multi-roots-names2")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(repo2)

	for _, repo := range []string{repo1, repo2} {
		cmd := exec.Command("git", "init")
		cmd.Dir = repo
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to initialize git repo: %v", err)
		}
	}

	repos := map[string]string{
		repo1: "repo1",
		repo2: "repo2",
	}

	multiDiffer, err := NewMulti(repos)
	if err != nil {
		t.Fatalf("Failed to create MultiDiffer: %v", err)
	}

	roots := multiDiffer.RepoRootsWithNames()
	if len(roots) != 2 {
		t.Errorf("Expected 2 repo roots, got %d", len(roots))
	}
	if roots[repo1] != "repo1" || roots[repo2] != "repo2" {
		t.Errorf("Expected repo names to match input map")
	}
}

func TestDiscoverRepos(t *testing.T) {
	parentDir, err := os.MkdirTemp("", "test-discover")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(parentDir)

	nested1 := filepath.Join(parentDir, "nested1")
	if err := os.Mkdir(nested1, 0755); err != nil {
		t.Fatalf("Failed to create nested dir: %v", err)
	}

	nested2 := filepath.Join(parentDir, "nested2")
	if err := os.Mkdir(nested2, 0755); err != nil {
		t.Fatalf("Failed to create nested dir: %v", err)
	}

	for _, dir := range []string{nested1, nested2} {
		cmd := exec.Command("git", "init")
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to initialize git repo: %v", err)
		}
	}

	repos, err := DiscoverRepos(parentDir)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(repos) != 2 {
		t.Errorf("Expected 2 repos, got %d", len(repos))
	}
}

func TestGetBranch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-branch")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("initial content\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set git email: %v", err)
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set git name: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	branch := GetBranch(tmpDir)
	if branch != "master" && branch != "main" {
		t.Errorf("Expected 'master' or 'main' branch, got %s", branch)
	}

	branch = GetBranch("/tmp/non-existent")
	if branch != "" {
		t.Errorf("Expected empty branch for non-git directory, got %s", branch)
	}
}

func TestIsGitRepo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-is-git")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if IsGitRepo(tmpDir) {
		t.Error("Expected false for non-git directory")
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	if !IsGitRepo(tmpDir) {
		t.Error("Expected true for git directory")
	}
}

func TestIsGitIgnored(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-git-ignore")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("content\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if IsGitIgnored(tmpDir, testFile) {
		t.Error("Expected false for non-ignored file")
	}

	gitignoreFile := filepath.Join(tmpDir, ".gitignore")
	if err := os.WriteFile(gitignoreFile, []byte("test.txt\n"), 0644); err != nil {
		t.Fatalf("Failed to create .gitignore file: %v", err)
	}

	if !IsGitIgnored(tmpDir, testFile) {
		t.Error("Expected true for ignored file")
	}
}

func TestFindRepoRoot(t *testing.T) {
	repo1, err := os.MkdirTemp("", "test-find-root1")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(repo1)

	repo2, err := os.MkdirTemp("", "test-find-root2")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(repo2)

	for _, repo := range []string{repo1, repo2} {
		cmd := exec.Command("git", "init")
		cmd.Dir = repo
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to initialize git repo: %v", err)
		}
	}

	repoRoots := []string{repo1, repo2}

	testFile1 := filepath.Join(repo1, "test.txt")
	foundRoot := FindRepoRoot(testFile1, repoRoots)
	if foundRoot != repo1 {
		t.Errorf("Expected repo root %s, got %s", repo1, foundRoot)
	}

	testFile2 := filepath.Join(repo2, "test.txt")
	foundRoot = FindRepoRoot(testFile2, repoRoots)
	if foundRoot != repo2 {
		t.Errorf("Expected repo root %s, got %s", repo2, foundRoot)
	}

	foundRoot = FindRepoRoot("/tmp/non-existent.txt", repoRoots)
	if foundRoot != "" {
		t.Errorf("Expected empty string for file not in any repo, got %s", foundRoot)
	}
}
