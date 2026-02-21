package model

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"codeberg.org/devcarlosmolero/vibewatch/internal/differ"
	"codeberg.org/devcarlosmolero/vibewatch/internal/types"
)

const maxDiffLines = 100

type Model struct {
	entries           []types.DiffEntry
	viewport          viewport.Model
	width             int
	height            int
	changes           <-chan string
	differ            differ.Differ
	maxEntries        int
	paused            bool
	showHelp          bool
	ready             bool
	dir               string
	tabs              []string
	activeTab         int
	branches          map[string]string
	branch            string
	visibleFiles      map[string]bool
	visibleFilesMu    sync.Mutex
	showHiddenCount   int
	selectedFileIndex int
	selectedFilePath  string
}

func New(changes <-chan string, d differ.Differ, maxEntries int, dir string, repoNames []string, branches map[string]string, branch string) Model {
	var tabs []string
	if len(repoNames) > 1 {
		tabs = []string{"All"}
		tabs = append(tabs, repoNames...)
	}
	return Model{
		changes:         changes,
		differ:          d,
		maxEntries:      maxEntries,
		dir:             dir,
		tabs:            tabs,
		branches:        branches,
		branch:          branch,
		visibleFiles:    make(map[string]bool),
		showHiddenCount: 0,
	}
}

func (m *Model) Init() tea.Cmd {
	initModelDebugLogging("/Users/carlos/Desktop/Git/vibewatch")
	logMessage("Model initialized, waiting for changes...")

	return tea.Batch(
		loadInitialEntries(m.differ),
		waitForChange(m.changes, m.differ),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		case "p":
			m.paused = !m.paused
			return m, nil
		case "c":
			m.entries = nil
			m.viewport.SetContent(m.renderEntries())
			return m, nil
		case "g", "home":
			m.viewport.GotoTop()
			return m, nil
		case "G", "end":
			m.viewport.GotoBottom()
			return m, nil
		case "tab":
			if len(m.tabs) > 0 {
				m.activeTab = (m.activeTab + 1) % len(m.tabs)
				m.viewport.SetContent(m.renderEntries())
				m.viewport.GotoTop()
			}
			return m, nil
		case "shift+tab":
			if len(m.tabs) > 0 {
				m.activeTab = (m.activeTab - 1 + len(m.tabs)) % len(m.tabs)
				m.viewport.SetContent(m.renderEntries())
				m.viewport.GotoTop()
			}
			return m, nil
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			idx := int(msg.String()[0] - '1')
			if len(m.tabs) > 0 && idx < len(m.tabs) {
				m.activeTab = idx
				m.viewport.SetContent(m.renderEntries())
				m.viewport.GotoTop()
			}
			return m, nil
		case "t":
			if len(m.entries) > 0 {
				filePath := m.selectedFilePath
				if filePath == "" {
					filtered := m.filteredEntries()
					if len(filtered) > 0 {
						m.selectedFileIndex = 0
						m.selectedFilePath = filtered[0].FilePath
						filePath = filtered[0].FilePath
					}
				}
				if filePath != "" {
					return m, m.toggleFileVisibility(filePath)
				}
			}
			return m, nil
		case "T":
			if len(m.entries) > 0 {
				filePath := m.selectedFilePath
				if filePath == "" {
					for _, entry := range m.entries {
						if m.isFileVisible(entry.FilePath) {
							filePath = entry.FilePath
							break
						}
					}
				}
				if filePath != "" {
					return m, m.toggleFileVisibility(filePath)
				}
			}
			return m, nil
		case "U":
			if len(m.entries) > 0 {
				filePath := m.selectedFilePath
				if filePath == "" {
					filtered := m.filteredEntries()
					if len(filtered) > 0 {
						m.selectedFileIndex = 0
						m.selectedFilePath = filtered[0].FilePath
						filePath = filtered[0].FilePath
					}
				}
				if filePath != "" && !m.isDiffVisible(filePath) {
					return m, m.untoggleFileVisibility(filePath)
				}
			}
			return m, nil
		case "up", "k":
			return m.navigateFiles(-1)
		case "down", "j":
			return m.navigateFiles(1)

		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		headerHeight := 1
		tabHeight := 0
		if len(m.tabs) > 0 {
			tabHeight = 1
		}
		statusHeight := 1
		vpHeight := m.height - headerHeight - tabHeight - statusHeight

		if !m.ready {
			m.viewport = viewport.New(m.width, vpHeight)
			m.viewport.SetContent(m.renderEntries())
			m.ready = true
		} else {
			m.viewport.Width = m.width
			m.viewport.Height = vpHeight
		}
		return m, nil

	case InitialEntriesMsg:
		entries := []types.DiffEntry(msg)
		m.entries = entries
		if len(m.entries) > m.maxEntries {
			m.entries = m.entries[:m.maxEntries]
		}
		m.viewport.SetContent(m.renderEntries())
		return m, nil

	case FileChangedMsg:
		entry := types.DiffEntry(msg)

		// Special handling for git operations (commit, etc.)
		// When .git/HEAD or .git/index changes, we need to refresh all files
		if entry.FilePath == "__GIT_OPERATION__" {
			logMessage("Model: Git operation detected, refreshing all files")
			cmds = append(cmds, loadInitialEntries(m.differ))
			cmds = append(cmds, waitForChange(m.changes, m.differ))
			return m, tea.Batch(cmds...)
		}

		if entry.Diff == "" && entry.Error == "" && !entry.IsNew {
			logMessage(fmt.Sprintf("Model: Removing committed file: %s", entry.FilePath))
			m.entries = removeEntriesForFile(m.entries, entry.FilePath)
			m.viewport.SetContent(m.renderEntries())
			cmds = append(cmds, waitForChange(m.changes, m.differ))
			return m, tea.Batch(cmds...)
		}

		if entry.Error != "" {
			logMessage(fmt.Sprintf("Model: Error getting diff for %s: %s", entry.FilePath, entry.Error))
		}

		m.entries = removeEntriesForFile(m.entries, entry.FilePath)
		m.entries = append([]types.DiffEntry{entry}, m.entries...)
		if len(m.entries) > m.maxEntries {
			m.entries = m.entries[:m.maxEntries]
		}
		m.viewport.SetContent(m.renderEntries())
		if !m.paused {
			m.viewport.GotoTop()
		}
		cmds = append(cmds, waitForChange(m.changes, m.differ))
		return m, tea.Batch(cmds...)

		if entry.Diff == "" && entry.Error == "" && !entry.IsNew {
			logMessage(fmt.Sprintf("Model: Removing committed file: %s", entry.FilePath))
			m.entries = removeEntriesForFile(m.entries, entry.FilePath)
			m.viewport.SetContent(m.renderEntries())
			cmds = append(cmds, waitForChange(m.changes, m.differ))
			return m, tea.Batch(cmds...)
		}

		if entry.Error != "" {
			logMessage(fmt.Sprintf("Model: Error getting diff for %s: %s", entry.FilePath, entry.Error))
		}
		m.entries = removeEntriesForFile(m.entries, entry.FilePath)
		m.entries = append([]types.DiffEntry{entry}, m.entries...)
		if len(m.entries) > m.maxEntries {
			m.entries = m.entries[:m.maxEntries]
		}
		m.viewport.SetContent(m.renderEntries())
		if !m.paused {
			m.viewport.GotoTop()
		}
		cmds = append(cmds, waitForChange(m.changes, m.differ))
		return m, tea.Batch(cmds...)

	case ToggleFileMsg:
		filePath := string(msg)
		m.visibleFilesMu.Lock()
		if visible, exists := m.visibleFiles[filePath]; exists {
			m.visibleFiles[filePath] = !visible
			if !visible {
				m.showHiddenCount--
			} else {
				m.showHiddenCount++
			}
		} else {
			m.visibleFiles[filePath] = false
			m.showHiddenCount++
		}
		m.visibleFilesMu.Unlock()
		m.viewport.SetContent(m.renderEntries())
		if filePath == m.selectedFilePath {
			m.ensureSelectedFileVisible()
		}
		return m, nil
	case UntoggleFileMsg:
		filePath := string(msg)
		m.visibleFilesMu.Lock()
		if _, exists := m.visibleFiles[filePath]; exists {
			m.visibleFiles[filePath] = true
			m.showHiddenCount--
		}
		m.visibleFilesMu.Unlock()
		m.viewport.SetContent(m.renderEntries())
		if filePath == m.selectedFilePath {
			m.ensureSelectedFileVisible()
		}
		return m, nil
	}

	// Forward remaining keys to viewport for scrolling
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Header
	headerText := fmt.Sprintf(" vibewatch — watching %s", m.dir)
	if m.branch != "" {
		headerText += "  " + BranchStyle.Render(m.branch)
	}
	header := HeaderStyle.Width(m.width).Render(headerText)

	// Status bar
	filtered := m.filteredEntries()
	totalVisible := len(filtered)
	totalFiles := len(m.entries)
	status := fmt.Sprintf(" %d changes", totalVisible)
	if len(m.tabs) > 0 && totalVisible != totalFiles {
		status += fmt.Sprintf(" (of %d total)", totalFiles)
	}
	if m.showHiddenCount > 0 {
		status += fmt.Sprintf("  %d hidden", m.showHiddenCount)
	}
	if m.paused {
		status += "  " + PausedStyle.Render("[PAUSED]")
	}
	if len(m.tabs) > 0 {
		status += "  tab switch"
	}
	status += "  t toggle  ? help  q quit"
	statusBar := StatusBarStyle.Width(m.width).Render(status)

	// Help overlay
	if m.showHelp {
		helpText := renderHelp()
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, helpText)
	}

	mainContent := ""
	if len(m.tabs) > 0 {
		tabBar := m.renderTabs()
		mainContent = header + "\n" + tabBar + "\n" + m.viewport.View()
	} else {
		mainContent = header + "\n" + m.viewport.View()
	}

	return mainContent + "\n" + statusBar
}

func (m *Model) renderTabs() string {
	var tabs []string
	for i, tab := range m.tabs {
		// Count entries for this tab
		count := 0
		if i == 0 {
			count = len(m.entries)
		} else {
			for _, e := range m.entries {
				if e.Repo == tab {
					count++
				}
			}
		}

		label := fmt.Sprintf(" %s (%d)", tab, count)
		if i > 0 {
			if br, ok := m.branches[tab]; ok && br != "" {
				label += " " + BranchLabelStyle.Render(br)
			}
		}
		label += " "
		if i == m.activeTab {
			tabs = append(tabs, ActiveTabStyle.Render(label))
		} else if count > 0 {
			tabs = append(tabs, TabWithChangesStyle.Render(label))
		} else {
			tabs = append(tabs, InactiveTabStyle.Render(label))
		}
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	// Fill the rest of the line with the tab bar background
	gap := m.width - lipgloss.Width(row)
	if gap > 0 {
		row += TabBarStyle.Render(strings.Repeat(" ", gap))
	}
	return row
}

func (m *Model) filteredEntries() []types.DiffEntry {
	var filtered []types.DiffEntry
	for _, e := range m.entries {
		// Filter by tab
		if m.activeTab != 0 {
			repoName := m.tabs[m.activeTab]
			if e.Repo != repoName {
				continue
			}
		}
		// We always show file entries now, diff visibility is handled separately
		filtered = append(filtered, e)
	}
	return filtered
}

func (m *Model) renderEntries() string {
	entries := m.filteredEntries()
	if len(entries) == 0 {
		if m.activeTab == 0 {
			return ContextLineStyle.Render("\n  Waiting for file changes...")
		}
		return ContextLineStyle.Render(fmt.Sprintf("\n  No changes in %s", m.tabs[m.activeTab]))
	}

	var b strings.Builder
	for i, entry := range entries {
		if i > 0 {
			sep := SeparatorStyle.Render(strings.Repeat("─", m.width))
			b.WriteString(sep + "\n")
		}
		b.WriteString(renderEntry(entry, m.width, m))
	}
	return b.String()
}

func renderEntry(e types.DiffEntry, width int, m *Model) string {
	var b strings.Builder

	// Repo tag + file path + timestamp
	ts := TimestampStyle.Render(e.Timestamp.Format("15:04:05"))
	fp := FilePathStyle.Render(e.FilePath)

	// Add hidden indicator if file is hidden
	hiddenIndicator := ""
	if m != nil && !m.isFileVisible(e.FilePath) {
		hiddenIndicator = HiddenFileStyle.Render(" [HIDDEN]")
	}

	// Add active indicator if this is the selected file
	activeIndicator := ""
	if m != nil && m.selectedFilePath == e.FilePath {
		activeIndicator = ActiveFileStyle.Render(" ▲ ACTIVE ")
	}

	if e.Repo != "" {
		repo := RepoTagStyle.Render(e.Repo)
		b.WriteString(activeIndicator + repo + " " + fp + "  " + ts + hiddenIndicator + "\n")
	} else {
		b.WriteString(activeIndicator + fp + "  " + ts + hiddenIndicator + "\n")
	}

	if e.Error != "" {
		b.WriteString(ErrorStyle.Render("  error: "+e.Error) + "\n")
		return b.String()
	}

	if e.Diff == "" {
		b.WriteString(ContextLineStyle.Render("  (no diff)") + "\n")
		return b.String()
	}

	// Check if diff should be visible
	if m != nil && !m.isDiffVisible(e.FilePath) {
		b.WriteString(HiddenFileStyle.Render("  [DIFF HIDDEN - press t to show]") + "\n")
		return b.String()
	}

	lines := strings.Split(e.Diff, "\n")
	rendered := 0
	for _, line := range lines {
		// Skip git diff header lines — we already show the file path
		if strings.HasPrefix(line, "diff --git") ||
			strings.HasPrefix(line, "index ") ||
			strings.HasPrefix(line, "--- ") ||
			strings.HasPrefix(line, "+++ ") ||
			strings.HasPrefix(line, "new file") ||
			strings.HasPrefix(line, "old mode") ||
			strings.HasPrefix(line, "new mode") {
			continue
		}

		if rendered >= maxDiffLines {
			b.WriteString(ErrorStyle.Render("  ... (truncated)") + "\n")
			break
		}

		switch {
		case strings.HasPrefix(line, "@@"):
			b.WriteString(HunkHeaderStyle.Render(line) + "\n")
		case strings.HasPrefix(line, "+"):
			b.WriteString(AddedLineStyle.Render(line) + "\n")
		case strings.HasPrefix(line, "-"):
			b.WriteString(RemovedLineStyle.Render(line) + "\n")
		default:
			b.WriteString(ContextLineStyle.Render(line) + "\n")
		}
		rendered++
	}

	return b.String()
}

func removeEntriesForFile(entries []types.DiffEntry, filePath string) []types.DiffEntry {
	result := make([]types.DiffEntry, 0, len(entries))
	for _, e := range entries {
		if e.FilePath != filePath {
			result = append(result, e)
		}
	}
	return result
}

// isFileVisible checks if a file should be visible based on the toggle state
func (m *Model) isFileVisible(filePath string) bool {
	m.visibleFilesMu.Lock()
	defer m.visibleFilesMu.Unlock()
	if visible, exists := m.visibleFiles[filePath]; exists {
		return visible
	}
	// Default to visible if not in the map
	return true
}

// isDiffVisible checks if the diff content should be shown for a file
func (m *Model) isDiffVisible(filePath string) bool {
	m.visibleFilesMu.Lock()
	defer m.visibleFilesMu.Unlock()
	if visible, exists := m.visibleFiles[filePath]; exists {
		return visible
	}
	// Default to visible if not in the map
	return true
}

// toggleFileVisibility toggles the visibility of a file and returns a command to update the view
func (m *Model) toggleFileVisibility(filePath string) tea.Cmd {
	return func() tea.Msg {
		return ToggleFileMsg(filePath)
	}
}

// untoggleFileVisibility shows a hidden file and returns a command to update the view
func (m *Model) untoggleFileVisibility(filePath string) tea.Cmd {
	return func() tea.Msg {
		return UntoggleFileMsg(filePath)
	}
}

// navigateFiles moves the selection up or down by the specified delta
func (m *Model) navigateFiles(delta int) (*Model, tea.Cmd) {
	filtered := m.filteredEntries()
	if len(filtered) == 0 {
		return m, nil
	}

	// Initialize selection if not properly set (index is 0 but path is empty)
	if m.selectedFilePath == "" {
		m.selectedFileIndex = 0
		m.selectedFilePath = filtered[0].FilePath
		m.ensureSelectedFileVisible()
		return m, nil
	}

	// Calculate new index with proper wrapping
	newIndex := m.selectedFileIndex + delta

	// Handle wrapping behavior
	if delta > 0 {
		// Moving down - wrap to first file when going past the end
		if newIndex >= len(filtered) {
			newIndex = 0
		}
	} else if delta < 0 {
		// Moving up - wrap to last file when going before the beginning
		if newIndex < 0 {
			newIndex = len(filtered) - 1
		}
	}

	m.selectedFileIndex = newIndex
	m.selectedFilePath = filtered[newIndex].FilePath

	// Ensure the selected file is visible in the viewport
	m.ensureSelectedFileVisible()

	return m, nil
}

// ensureSelectedFileVisible scrolls the viewport to make sure the selected file is visible
func (m *Model) ensureSelectedFileVisible() {
	m.viewport.SetContent(m.renderEntries())

	// For simple navigation, just ensure the selected file is somewhere in the viewport
	// The exact positioning will be handled by the viewport's natural scrolling
	// This is much more efficient than trying to calculate exact line positions

	// If we're near the beginning, go to top
	if m.selectedFileIndex < 5 {
		m.viewport.GotoTop()
		return
	}

	// If we're near the end, go to bottom
	filtered := m.filteredEntries()
	if m.selectedFileIndex >= len(filtered)-5 {
		m.viewport.GotoBottom()
		return
	}

	// For middle positions, the viewport will naturally show the right area
	// as the user navigates, so we don't need to do precise scrolling
}

func loadInitialEntries(d differ.Differ) tea.Cmd {
	return func() tea.Msg {
		entries, err := d.DirtyFiles()
		if err != nil || len(entries) == 0 {
			return InitialEntriesMsg(nil)
		}
		return InitialEntriesMsg(entries)
	}
}

func waitForChange(ch <-chan string, d differ.Differ) tea.Cmd {
	return func() tea.Msg {
		path, ok := <-ch
		if !ok {
			return nil
		}

		// Debug: Log that we received a change from the channel
		logMessage(fmt.Sprintf("MODEL: Received change from channel: %s", path))

		entry, err := d.Diff(path)
		if err != nil {
			logMessage(fmt.Sprintf("MODEL: Error getting diff for %s: %v", path, err))
			entry = types.DiffEntry{
				FilePath:  path,
				Timestamp: time.Now(),
				Error:     err.Error(),
			}
		} else {
			logMessage(fmt.Sprintf("MODEL: Successfully got diff for %s", path))
		}
		return FileChangedMsg(entry)
	}
}

// processPendingChanges processes all pending changes from the channel
func processPendingChanges(ch <-chan string, d differ.Differ) tea.Cmd {
	return func() tea.Msg {
		// First, try to receive one change immediately
		select {
		case path, ok := <-ch:
			if !ok {
				return nil
			}
			entry, err := d.Diff(path)
			if err != nil {
				entry = types.DiffEntry{
					FilePath:  path,
					Timestamp: time.Now(),
					Error:     err.Error(),
				}
			}
			return FileChangedMsg(entry)
		default:
			// No changes available immediately, return nil
			// This will cause the model to re-schedule this command
			return nil
		}
	}
}
