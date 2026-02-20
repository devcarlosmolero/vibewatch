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
