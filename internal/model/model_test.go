package model

import (
	"codeberg.org/devcarlosmolero/vibewatch/internal/types"
	"testing"
	"time"
)

func TestFileNavigation(t *testing.T) {
	// Create a test model
	m := Model{
		entries: []types.DiffEntry{
			{
				FilePath:  "file1.go",
				Timestamp: time.Now(),
				Diff:      "diff --git a/file1.go b/file1.go\n+++ b/file1.go\n@@ -1,1 +1,1 @@\n-line1\n+line1 changed\n",
			},
			{
				FilePath:  "file2.go",
				Timestamp: time.Now(),
				Diff:      "diff --git a/file2.go b/file2.go\n+++ b/file2.go\n@@ -1,1 +1,1 @@\n-line1\n+line1 changed\n",
			},
		},
		visibleFiles: make(map[string]bool),
	}

	// First navigation down should select first file (index 0)
	m, _ = m.navigateFiles(1)
	if m.selectedFileIndex != 0 {
		t.Errorf("Expected selected index 0, got %d", m.selectedFileIndex)
	}

	// Test navigation down to second file
	m, _ = m.navigateFiles(1)
	if m.selectedFileIndex != 1 {
		t.Errorf("Expected selected index 1, got %d", m.selectedFileIndex)
	}

	// Test wrapping around - moving down from last should go to first
	m, _ = m.navigateFiles(1)
	if m.selectedFileIndex != 0 {
		t.Errorf("Expected selected index 0 (wrapped), got %d", m.selectedFileIndex)
	}
}

func TestToggleAndUntoggle(t *testing.T) {
	m := Model{
		entries: []types.DiffEntry{
			{
				FilePath:  "file1.go",
				Timestamp: time.Now(),
			},
		},
		visibleFiles: make(map[string]bool),
	}

	// Test toggle
	cmd := m.toggleFileVisibility("file1.go")
	msg := cmd()
	m.Update(msg)

	if m.isFileVisible("file1.go") {
		t.Error("Expected file to be hidden after toggle")
	}

	// Test untoggle
	cmd = m.untoggleFileVisibility("file1.go")
	msg = cmd()
	m.Update(msg)

	if !m.isFileVisible("file1.go") {
		t.Error("Expected file to be visible after untoggle")
	}
}

func TestToggleUntoggleWithNavigation(t *testing.T) {
	m := Model{
		entries: []types.DiffEntry{
			{
				FilePath:  "file1.go",
				Timestamp: time.Now(),
				Diff:      "diff --git a/file1.go b/file1.go\n+++ b/file1.go\n@@ -1,1 +1,1 @@\n-line1\n+line1 changed\n",
			},
			{
				FilePath:  "file2.go",
				Timestamp: time.Now(),
				Diff:      "diff --git a/file2.go b/file2.go\n+++ b/file2.go\n@@ -1,1 +1,1 @@\n-line1\n+line1 changed\n",
			},
		},
		visibleFiles: make(map[string]bool),
	}

	// First navigation selects first file
	m, _ = m.navigateFiles(1)
	// Navigate to second file (index 1)
	m, _ = m.navigateFiles(1)
	if m.selectedFilePath != "file2.go" {
		t.Errorf("Expected selected file to be file2.go, got %s", m.selectedFilePath)
	}

	// Toggle the selected file (file2.go) - should hide diff but keep file entry
	cmd := m.toggleFileVisibility("file2.go")
	msg := cmd()
	m.Update(msg)

	if m.isDiffVisible("file2.go") {
		t.Error("Expected file2.go diff to be hidden after toggle")
	}
	// File entries are always visible now, we only hide diffs

	// Untoggle the selected file (file2.go) - should show diff again
	cmd = m.untoggleFileVisibility("file2.go")
	msg = cmd()
	m.Update(msg)

	if !m.isDiffVisible("file2.go") {
		t.Error("Expected file2.go diff to be visible after untoggle")
	}
}

func TestNavigationWrapping(t *testing.T) {
	m := Model{
		entries: []types.DiffEntry{
			{
				FilePath:  "file1.go",
				Timestamp: time.Now(),
				Diff:      "diff --git a/file1.go b/file1.go\n+++ b/file1.go\n@@ -1,1 +1,1 @@\n-line1\n+line1 changed\n",
			},
			{
				FilePath:  "file2.go",
				Timestamp: time.Now(),
				Diff:      "diff --git a/file2.go b/file2.go\n+++ b/file2.go\n@@ -1,1 +1,1 @@\n-line1\n+line1 changed\n",
			},
		},
		visibleFiles: make(map[string]bool),
	}

	// First navigation down should select first file
	m, _ = m.navigateFiles(1)
	if m.selectedFileIndex != 0 {
		t.Errorf("Expected selected index 0, got %d", m.selectedFileIndex)
	}
	if m.selectedFilePath != "file1.go" {
		t.Errorf("Expected selected file file1.go, got %s", m.selectedFilePath)
	}

	// Move down to second file
	m, _ = m.navigateFiles(1)
	if m.selectedFileIndex != 1 {
		t.Errorf("Expected selected index 1, got %d", m.selectedFileIndex)
	}
	if m.selectedFilePath != "file2.go" {
		t.Errorf("Expected selected file file2.go, got %s", m.selectedFilePath)
	}

	// Move down past end - should wrap to first file
	m, _ = m.navigateFiles(1)
	if m.selectedFileIndex != 0 {
		t.Errorf("Expected selected index 0 (wrapped), got %d", m.selectedFileIndex)
	}
	if m.selectedFilePath != "file1.go" {
		t.Errorf("Expected selected file file1.go (wrapped), got %s", m.selectedFilePath)
	}

	// Move up before beginning - should wrap to last file
	m, _ = m.navigateFiles(-1)
	if m.selectedFileIndex != 1 {
		t.Errorf("Expected selected index 1 (wrapped), got %d", m.selectedFileIndex)
	}
	if m.selectedFilePath != "file2.go" {
		t.Errorf("Expected selected file file2.go (wrapped), got %s", m.selectedFilePath)
	}
}

// Test git operation detection - this tests the core functionality
func TestGitOperationDetection(t *testing.T) {
	m := Model{
		entries: []types.DiffEntry{
			{
				FilePath:  "file1.go",
				Timestamp: time.Now(),
				Diff:      "diff --git a/file1.go b/file1.go\n+++ b/file1.go\n@@ -1,1 +1,1 @@\n-line1\n+line1 changed\n",
			},
			{
				FilePath:  "file2.go",
				Timestamp: time.Now(),
				Diff:      "diff --git a/file2.go b/file2.go\n+++ b/file2.go\n@@ -1,1 +1,1 @@\n-line1\n+line1 changed\n",
			},
		},
		visibleFiles: make(map[string]bool),
	}

	// Simulate receiving a git operation marker
	gitOpEntry := types.DiffEntry{
		FilePath:  "__GIT_OPERATION__",
		Timestamp: time.Now(),
		Diff:      "",
		Error:     "",
		IsNew:     false,
	}

	// This should trigger a refresh and clear entries
	msg := FileChangedMsg(gitOpEntry)
	m.Update(msg)

	// After git operation, entries should be cleared (simulating reload)
	// In real scenario, loadInitialEntries would be called to refresh from git
	if len(m.entries) != 0 {
		t.Errorf("Expected entries to be cleared after git operation, got %d entries", len(m.entries))
	}
}

// Test file removal after commit
func TestCommittedFileRemoval(t *testing.T) {
	m := Model{
		entries: []types.DiffEntry{
			{
				FilePath:  "committed_file.go",
				Timestamp: time.Now(),
				Diff:      "diff --git a/committed_file.go b/committed_file.go\n+++ b/committed_file.go\n@@ -1,1 +1,1 @@\n-line1\n+line1 changed\n",
			},
		},
		visibleFiles: make(map[string]bool),
	}

	// Simulate file being committed (empty diff, no error, not new)
	committedEntry := types.DiffEntry{
		FilePath:  "committed_file.go",
		Timestamp: time.Now(),
		Diff:      "",
		Error:     "",
		IsNew:     false,
	}

	msg := FileChangedMsg(committedEntry)
	m.Update(msg)

	// File should be removed from entries
	if len(m.entries) != 0 {
		t.Errorf("Expected committed file to be removed, still have %d entries", len(m.entries))
	}
}

// Test that regular file changes still work
func TestRegularFileChange(t *testing.T) {
	m := Model{
		entries:      []types.DiffEntry{},
		visibleFiles: make(map[string]bool),
		maxEntries:   10,
	}

	// Add a file change
	fileEntry := types.DiffEntry{
		FilePath:  "changed_file.go",
		Timestamp: time.Now(),
		Diff:      "diff --git a/changed_file.go b/changed_file.go\n+++ b/changed_file.go\n@@ -1,1 +1,1 @@\n-line1\n+line1 changed\n",
		Error:     "",
		IsNew:     false,
	}

	msg := FileChangedMsg(fileEntry)
	m.Update(msg)

	// Should have one entry now
	if len(m.entries) != 1 {
		t.Errorf("Expected 1 entry after file change, got %d", len(m.entries))
	}

	// Should be the file we added
	if m.entries[0].FilePath != "changed_file.go" {
		t.Errorf("Expected changed_file.go, got %s", m.entries[0].FilePath)
	}
}

// Test error handling for file changes
func TestFileChangeWithError(t *testing.T) {
	m := Model{
		entries:      []types.DiffEntry{},
		visibleFiles: make(map[string]bool),
		maxEntries:   10,
	}

	// Add a file change with error
	errorEntry := types.DiffEntry{
		FilePath:  "error_file.go",
		Timestamp: time.Now(),
		Diff:      "",
		Error:     "git diff failed: file not found",
		IsNew:     false,
	}

	msg := FileChangedMsg(errorEntry)
	m.Update(msg)

	// Should still have the entry with error
	if len(m.entries) != 1 {
		t.Errorf("Expected 1 entry with error, got %d", len(m.entries))
	}

	// Should preserve the error
	if m.entries[0].Error != "git diff failed: file not found" {
		t.Errorf("Expected error to be preserved, got: %s", m.entries[0].Error)
	}
}

// Test entry limit enforcement
func TestMaxEntriesEnforcement(t *testing.T) {
	m := Model{
		entries:      []types.DiffEntry{},
		visibleFiles: make(map[string]bool),
		maxEntries:   2, // Set low limit for testing
	}

	// Add multiple files
	for i := 0; i < 5; i++ {
		fileName := fmt.Sprintf("file%d.go", i)
		fileEntry := types.DiffEntry{
			FilePath:  fileName,
			Timestamp: time.Now(),
			Diff:      fmt.Sprintf("diff --git a/%s b/%s\n+++ b/%s\n@@ -1,1 +1,1 @@\n-line1\n+line1 changed\n", fileName, fileName, fileName),
			Error:     "",
			IsNew:     false,
		}
		msg := FileChangedMsg(fileEntry)
		m.Update(msg)
	}

	// Should not exceed maxEntries
	if len(m.entries) > m.maxEntries {
		t.Errorf("Expected at most %d entries, got %d", m.maxEntries, len(m.entries))
	}
}
