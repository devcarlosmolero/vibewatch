package model

import (
	"codeberg.org/devcarlosmolero/vibewatch/internal/types"
	"fmt"
	"testing"
	"time"
)

func TestFileNavigation(t *testing.T) {
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

	result, _ := m.navigateFiles(1)
	if result != nil {
		m = *result
	}
	if m.selectedFileIndex != 0 {
		t.Errorf("Expected selected index 0, got %d", m.selectedFileIndex)
	}

	result, _ = m.navigateFiles(1)
	if result != nil {
		m = *result
	}
	if m.selectedFileIndex != 1 {
		t.Errorf("Expected selected index 1, got %d", m.selectedFileIndex)
	}

	result, _ = m.navigateFiles(1)
	if result != nil {
		m = *result
	}
	if m.selectedFileIndex != 0 {
		t.Errorf("Expected selected index 0 (wrapped), got %d", m.selectedFileIndex)
	}

	result, _ = m.navigateFiles(1)
	if result != nil {
		m = *result
	}
	if m.selectedFileIndex != 1 {
		t.Errorf("Expected selected index 1, got %d", m.selectedFileIndex)
	}

	result, _ = m.navigateFiles(1)
	if result != nil {
		m = *result
	}
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

	cmd := m.toggleFileVisibility("file1.go")
	msg := cmd()
	m.Update(msg)

	if m.isFileVisible("file1.go") {
		t.Error("Expected file to be hidden after toggle")
	}

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

	result, _ := m.navigateFiles(1)
	if result != nil {
		m = *result
	}
	result, _ = m.navigateFiles(1)
	if result != nil {
		m = *result
	}
	if m.selectedFilePath != "file2.go" {
		t.Errorf("Expected selected file to be file2.go, got %s", m.selectedFilePath)
	}

	cmd := m.toggleFileVisibility("file2.go")
	msg := cmd()
	m.Update(msg)

	if m.isDiffVisible("file2.go") {
		t.Error("Expected file2.go diff to be hidden after toggle")
	}

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

	result, _ := m.navigateFiles(1)
	if result != nil {
		m = *result
	}
	if m.selectedFileIndex != 0 {
		t.Errorf("Expected selected index 0, got %d", m.selectedFileIndex)
	}
	if m.selectedFilePath != "file1.go" {
		t.Errorf("Expected selected file file1.go, got %s", m.selectedFilePath)
	}

	result, _ = m.navigateFiles(1)
	if result != nil {
		m = *result
	}
	if m.selectedFileIndex != 1 {
		t.Errorf("Expected selected index 1, got %d", m.selectedFileIndex)
	}
	if m.selectedFilePath != "file2.go" {
		t.Errorf("Expected selected file file2.go, got %s", m.selectedFilePath)
	}

	result, _ = m.navigateFiles(1)
	if result != nil {
		m = *result
	}
	if m.selectedFileIndex != 0 {
		t.Errorf("Expected selected index 0 (wrapped), got %d", m.selectedFileIndex)
	}
	if m.selectedFilePath != "file1.go" {
		t.Errorf("Expected selected file file1.go (wrapped), got %s", m.selectedFilePath)
	}

	result, _ = m.navigateFiles(-1)
	if result != nil {
		m = *result
	}
	if m.selectedFileIndex != 1 {
		t.Errorf("Expected selected index 1 (wrapped), got %d", m.selectedFileIndex)
	}
	if m.selectedFilePath != "file2.go" {
		t.Errorf("Expected selected file file2.go (wrapped), got %s", m.selectedFilePath)
	}
}

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

	gitOpEntry := types.DiffEntry{
		FilePath:  "__GIT_OPERATION__",
		Timestamp: time.Now(),
		Diff:      "",
		Error:     "",
		IsNew:     false,
	}

	msg := FileChangedMsg(gitOpEntry)
	_, result := m.Update(msg)

	if result == nil {
		t.Errorf("Expected non-nil command after git operation")
	}

	if len(m.entries) != 2 {
		t.Errorf("Expected entries to remain until refresh completes, got %d entries", len(m.entries))
	}
}

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

	committedEntry := types.DiffEntry{
		FilePath:  "committed_file.go",
		Timestamp: time.Now(),
		Diff:      "",
		Error:     "",
		IsNew:     false,
	}

	msg := FileChangedMsg(committedEntry)
	m.Update(msg)

	if len(m.entries) != 0 {
		t.Errorf("Expected committed file to be removed, still have %d entries", len(m.entries))
	}
}

func TestRegularFileChange(t *testing.T) {
	m := Model{
		entries:      []types.DiffEntry{},
		visibleFiles: make(map[string]bool),
		maxEntries:   10,
	}

	fileEntry := types.DiffEntry{
		FilePath:  "changed_file.go",
		Timestamp: time.Now(),
		Diff:      "diff --git a/changed_file.go b/changed_file.go\n+++ b/changed_file.go\n@@ -1,1 +1,1 @@\n-line1\n+line1 changed\n",
		Error:     "",
		IsNew:     false,
	}

	msg := FileChangedMsg(fileEntry)
	m.Update(msg)

	if len(m.entries) != 1 {
		t.Errorf("Expected 1 entry after file change, got %d", len(m.entries))
	}

	if m.entries[0].FilePath != "changed_file.go" {
		t.Errorf("Expected changed_file.go, got %s", m.entries[0].FilePath)
	}
}

func TestFileChangeWithError(t *testing.T) {
	m := Model{
		entries:      []types.DiffEntry{},
		visibleFiles: make(map[string]bool),
		maxEntries:   10,
	}

	errorEntry := types.DiffEntry{
		FilePath:  "error_file.go",
		Timestamp: time.Now(),
		Diff:      "",
		Error:     "git diff failed: file not found",
		IsNew:     false,
	}

	msg := FileChangedMsg(errorEntry)
	m.Update(msg)

	if len(m.entries) != 1 {
		t.Errorf("Expected 1 entry with error, got %d", len(m.entries))
	}

	if m.entries[0].Error != "git diff failed: file not found" {
		t.Errorf("Expected error to be preserved, got: %s", m.entries[0].Error)
	}
}

func TestMaxEntriesEnforcement(t *testing.T) {
	m := Model{
		entries:      []types.DiffEntry{},
		visibleFiles: make(map[string]bool),
		maxEntries:   2,
	}

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

	if len(m.entries) > m.maxEntries {
		t.Errorf("Expected at most %d entries, got %d", m.maxEntries, len(m.entries))
	}
}
