package model

import (
	"codeberg.org/devcarlosmolero/vibewatch/internal/types"
	"fmt"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
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

func TestWindowSizeUpdate(t *testing.T) {
	m := Model{
		entries:      []types.DiffEntry{},
		visibleFiles: make(map[string]bool),
		width:        0,
		height:       0,
		ready:        false,
	}

	msg := tea.WindowSizeMsg{
		Width:  80,
		Height: 24,
	}
	m.Update(msg)

	if m.width != 80 {
		t.Errorf("Expected width 80, got %d", m.width)
	}
	if m.height != 24 {
		t.Errorf("Expected height 24, got %d", m.height)
	}
	if !m.ready {
		t.Error("Expected ready to be true after window size update")
	}
}

func TestToggleVisibilityCommands(t *testing.T) {
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
	if cmd == nil {
		t.Error("Expected non-nil command from toggleFileVisibility")
	}

	cmd = m.untoggleFileVisibility("file1.go")
	if cmd == nil {
		t.Error("Expected non-nil command from untoggleFileVisibility")
	}
}

func TestNavigationWithEmptyEntries(t *testing.T) {
	m := Model{
		entries:           []types.DiffEntry{},
		visibleFiles:      make(map[string]bool),
		selectedFileIndex: 0,
		selectedFilePath:  "",
	}

	result, _ := m.navigateFiles(1)
	if result != nil {
		m = *result
	}
	if m.selectedFileIndex != 0 {
		t.Errorf("Expected selected index 0 for empty entries, got %d", m.selectedFileIndex)
	}
	if m.selectedFilePath != "" {
		t.Errorf("Expected empty selected file path for empty entries, got %s", m.selectedFilePath)
	}
}

func TestNavigationWrappingWithSingleEntry(t *testing.T) {
	m := Model{
		entries: []types.DiffEntry{
			{
				FilePath:  "file1.go",
				Timestamp: time.Now(),
				Diff:      "diff --git a/file1.go b/file1.go\n+++ b/file1.go\n@@ -1,1 +1,1 @@\n-line1\n+line1 changed\n",
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

	result, _ = m.navigateFiles(-1)
	if result != nil {
		m = *result
	}
	if m.selectedFileIndex != 0 {
		t.Errorf("Expected selected index 0 (wrapped), got %d", m.selectedFileIndex)
	}
	if m.selectedFilePath != "file1.go" {
		t.Errorf("Expected selected file file1.go (wrapped), got %s", m.selectedFilePath)
	}
}

func TestInitialEntriesUpdate(t *testing.T) {
	m := Model{
		entries:      []types.DiffEntry{},
		visibleFiles: make(map[string]bool),
		ready:        true,
		maxEntries:   10,
	}

	entries := []types.DiffEntry{
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
	}

	msg := InitialEntriesMsg(entries)
	m.Update(msg)

	if len(m.entries) != 2 {
		t.Errorf("Expected 2 entries after initial update, got %d", len(m.entries))
	}
}

func TestToggleFileMsgUpdate(t *testing.T) {
	m := Model{
		entries: []types.DiffEntry{
			{
				FilePath:  "file1.go",
				Timestamp: time.Now(),
			},
		},
		visibleFiles:    make(map[string]bool),
		showHiddenCount: 0,
	}

	msg := ToggleFileMsg("file1.go")
	m.Update(msg)

	if m.showHiddenCount != 1 {
		t.Errorf("Expected showHiddenCount 1, got %d", m.showHiddenCount)
	}

	msg = ToggleFileMsg("file1.go")
	m.Update(msg)

	if m.showHiddenCount != 0 {
		t.Errorf("Expected showHiddenCount 0 after second toggle, got %d", m.showHiddenCount)
	}
}

func TestUntoggleFileMsgUpdate(t *testing.T) {
	m := Model{
		entries: []types.DiffEntry{
			{
				FilePath:  "file1.go",
				Timestamp: time.Now(),
			},
		},
		visibleFiles:    make(map[string]bool),
		showHiddenCount: 1,
	}

	m.visibleFiles["file1.go"] = false

	msg := UntoggleFileMsg("file1.go")
	m.Update(msg)

	if m.showHiddenCount != 0 {
		t.Errorf("Expected showHiddenCount 0 after untoggle, got %d", m.showHiddenCount)
	}
}

func TestUpdateBranchesMsgUpdate(t *testing.T) {
	m := Model{
		branches: make(map[string]string),
		branch:   "",
	}

	branches := map[string]string{
		"repo1": "main",
		"repo2": "develop",
	}

	msg := UpdateBranchesMsg(branches)
	m.Update(msg)

	if len(m.branches) != 2 {
		t.Errorf("Expected 2 branches, got %d", len(m.branches))
	}
	if m.branches["repo1"] != "main" {
		t.Errorf("Expected repo1 branch 'main', got %s", m.branches["repo1"])
	}
}

func TestKeyHandling(t *testing.T) {
	m := Model{
		entries:      []types.DiffEntry{},
		visibleFiles: make(map[string]bool),
		paused:       false,
		showHelp:     false,
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.Update(msg)
	if cmd == nil {
		t.Error("Expected quit command for 'q' key")
	}

	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	m.Update(msg)
	if !m.showHelp {
		t.Error("Expected showHelp to be true after '?' key")
	}

	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	m.Update(msg)
	if !m.paused {
		t.Error("Expected paused to be true after 'p' key")
	}
}

func TestTabNavigation(t *testing.T) {
	m := Model{
		tabs:      []string{"All", "repo1", "repo2"},
		activeTab: 0,
		ready:     true,
		viewport:  viewport.New(80, 20),
	}

	// Initialize viewport content first
	m.viewport.SetContent("test content")

	// Test tab key (need to use the string "tab", not the character)
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t', 'a', 'b'}}
	m.Update(keyMsg)
	if m.activeTab != 1 {
		t.Errorf("Expected activeTab 1 after tab key, got %d", m.activeTab)
	}

	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t', 'a', 'b'}}
	m.Update(keyMsg)
	if m.activeTab != 2 {
		t.Errorf("Expected activeTab 2 after second tab key, got %d", m.activeTab)
	}

	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t', 'a', 'b'}}
	m.Update(keyMsg)
	if m.activeTab != 0 {
		t.Errorf("Expected activeTab 0 (wrapped) after third tab key, got %d", m.activeTab)
	}
}

func TestNumberKeyNavigation(t *testing.T) {
	m := Model{
		tabs:      []string{"All", "repo1", "repo2"},
		activeTab: 0,
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}}
	m.Update(msg)
	if m.activeTab != 0 {
		t.Errorf("Expected activeTab 0 after '1' key, got %d", m.activeTab)
	}

	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}}
	m.Update(msg)
	if m.activeTab != 1 {
		t.Errorf("Expected activeTab 1 after '2' key, got %d", m.activeTab)
	}

	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}}
	m.Update(msg)
	if m.activeTab != 2 {
		t.Errorf("Expected activeTab 2 after '3' key, got %d", m.activeTab)
	}
}

func TestViewportNavigation(t *testing.T) {
	m := Model{
		entries:      []types.DiffEntry{},
		visibleFiles: make(map[string]bool),
		ready:        true,
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	m.Update(msg)
	// Can't easily test viewport position without mocking

	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}}
	m.Update(msg)
	// Can't easily test viewport position without mocking
}

func TestClearEntries(t *testing.T) {
	m := Model{
		entries: []types.DiffEntry{
			{
				FilePath:  "file1.go",
				Timestamp: time.Now(),
				Diff:      "diff --git a/file1.go b/file1.go\n+++ b/file1.go\n@@ -1,1 +1,1 @@\n-line1\n+line1 changed\n",
			},
		},
		visibleFiles: make(map[string]bool),
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
	m.Update(msg)

	if len(m.entries) != 0 {
		t.Errorf("Expected entries to be cleared, got %d entries", len(m.entries))
	}
}

func TestIsFileVisible(t *testing.T) {
	m := Model{
		visibleFiles: make(map[string]bool),
	}

	if !m.isFileVisible("file1.go") {
		t.Error("Expected file to be visible by default")
	}

	m.visibleFiles["file1.go"] = false
	if m.isFileVisible("file1.go") {
		t.Error("Expected file to be hidden after setting to false")
	}

	m.visibleFiles["file1.go"] = true
	if !m.isFileVisible("file1.go") {
		t.Error("Expected file to be visible after setting to true")
	}
}

func TestIsDiffVisible(t *testing.T) {
	m := Model{
		visibleFiles: make(map[string]bool),
	}

	if !m.isDiffVisible("file1.go") {
		t.Error("Expected diff to be visible by default")
	}

	m.visibleFiles["file1.go"] = false
	if m.isDiffVisible("file1.go") {
		t.Error("Expected diff to be hidden after setting to false")
	}

	m.visibleFiles["file1.go"] = true
	if !m.isDiffVisible("file1.go") {
		t.Error("Expected diff to be visible after setting to true")
	}
}

func TestFilteredEntries(t *testing.T) {
	m := Model{
		entries: []types.DiffEntry{
			{
				FilePath:  "file1.go",
				Repo:      "repo1",
				Timestamp: time.Now(),
			},
			{
				FilePath:  "file2.go",
				Repo:      "repo2",
				Timestamp: time.Now(),
			},
		},
		tabs:         []string{"All", "repo1", "repo2"},
		activeTab:    0,
		visibleFiles: make(map[string]bool),
	}

	filtered := m.filteredEntries()
	if len(filtered) != 2 {
		t.Errorf("Expected 2 filtered entries for 'All' tab, got %d", len(filtered))
	}

	m.activeTab = 1
	filtered = m.filteredEntries()
	if len(filtered) != 1 {
		t.Errorf("Expected 1 filtered entry for 'repo1' tab, got %d", len(filtered))
	}
	if filtered[0].Repo != "repo1" {
		t.Errorf("Expected repo1 entry, got %s", filtered[0].Repo)
	}

	m.activeTab = 2
	filtered = m.filteredEntries()
	if len(filtered) != 1 {
		t.Errorf("Expected 1 filtered entry for 'repo2' tab, got %d", len(filtered))
	}
	if filtered[0].Repo != "repo2" {
		t.Errorf("Expected repo2 entry, got %s", filtered[0].Repo)
	}
}

func TestRemoveEntriesForFile(t *testing.T) {
	entries := []types.DiffEntry{
		{
			FilePath:  "file1.go",
			Timestamp: time.Now(),
		},
		{
			FilePath:  "file2.go",
			Timestamp: time.Now(),
		},
		{
			FilePath:  "file1.go",
			Timestamp: time.Now().Add(1 * time.Second),
		},
	}

	result := removeEntriesForFile(entries, "file1.go")
	if len(result) != 1 {
		t.Errorf("Expected 1 entry after removing file1.go, got %d", len(result))
	}
	if result[0].FilePath != "file2.go" {
		t.Errorf("Expected file2.go to remain, got %s", result[0].FilePath)
	}
}
