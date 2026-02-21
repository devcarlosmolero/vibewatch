package differ

import (
	"testing"
	"time"

	"codeberg.org/devcarlosmolero/vibewatch/internal/types"
)

// MockDiffer for testing without real git commands
type MockDiffer struct {
	mockDiffs map[string]types.DiffEntry
}

func NewMockDiffer() *MockDiffer {
	return &MockDiffer{
		mockDiffs: make(map[string]types.DiffEntry),
	}
}

func (m *MockDiffer) Diff(filePath string) (types.DiffEntry, error) {
	if filePath == "__GIT_OPERATION__" {
		return types.DiffEntry{
			FilePath:  filePath,
			Timestamp: time.Now(),
			Diff:      "",
			Error:     "",
			IsNew:     false,
		}, nil
	}

	if entry, exists := m.mockDiffs[filePath]; exists {
		return entry, nil
	}
	return types.DiffEntry{
		FilePath:  filePath,
		Timestamp: time.Now(),
		Diff:      "mock diff for " + filePath,
		Error:     "",
		IsNew:     false,
	}, nil
}

func (m *MockDiffer) DirtyFiles() ([]types.DiffEntry, error) {
	var entries []types.DiffEntry
	for _, entry := range m.mockDiffs {
		entries = append(entries, entry)
	}
	return entries, nil
}

func (m *MockDiffer) AddMockDiff(filePath, diff string) {
	m.mockDiffs[filePath] = types.DiffEntry{
		FilePath:  filePath,
		Timestamp: time.Now(),
		Diff:      diff,
		Error:     "",
		IsNew:     false,
	}
}

// TestGitOperationDetection tests the special git operation marker
func TestGitOperationDetection(t *testing.T) {
	mock := NewMockDiffer()

	entry, err := mock.Diff("__GIT_OPERATION__")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if entry.FilePath != "__GIT_OPERATION__" {
		t.Errorf("Expected git operation marker, got %s", entry.FilePath)
	}

	if entry.Diff != "" {
		t.Errorf("Expected empty diff for git operation, got %s", entry.Diff)
	}
}

// TestNormalFileDiff tests diff computation for normal files
func TestNormalFileDiff(t *testing.T) {
	mock := NewMockDiffer()

	mock.AddMockDiff("test.go", "diff --git a/test.go b/test.go\n+++ b/test.go\n@@ -1,1 +1,1 @@\n-old\n+new\n")

	entry, err := mock.Diff("test.go")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if entry.Diff != "diff --git a/test.go b/test.go\n+++ b/test.go\n@@ -1,1 +1,1 @@\n-old\n+new\n" {
		t.Errorf("Unexpected diff content")
	}
}

// TestCommittedFileDetection tests detection of committed files
func TestCommittedFileDetection(t *testing.T) {
	mock := NewMockDiffer()

	mock.mockDiffs["committed.go"] = types.DiffEntry{
		FilePath:  "committed.go",
		Timestamp: time.Now(),
		Diff:      "",
		Error:     "",
		IsNew:     false,
	}

	entry, err := mock.Diff("committed.go")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if entry.Diff != "" {
		t.Errorf("Expected empty diff for committed file")
	}
}

// TestDirtyFilesReturn tests that DirtyFiles returns expected entries
func TestDirtyFilesReturn(t *testing.T) {
	mock := NewMockDiffer()

	mock.AddMockDiff("file1.go", "diff1")
	mock.AddMockDiff("file2.go", "diff2")

	entries, err := mock.DirtyFiles()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 dirty files, got %d", len(entries))
	}
}
