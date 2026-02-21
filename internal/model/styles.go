// styles.go defines the visual styling for vibewatch's UI components.
// This file contains lipgloss styles for all UI elements including:
// - Header and status bars
// - File paths, timestamps, and diff lines
// - Tabs, repo tags, and branch indicators
// - Various text styles (added, removed, context, errors, etc.)
//
// The color palette uses a purple/blue theme with hex codes that match
// Test change to trigger watcher detection - attempt 2
// the vibewatch aesthetic and provide good contrast and readability.
package model

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Debug file for model logging
	modelDebugFile  *os.File
	modelDebugMutex sync.Mutex
)

var (
	// Header bar
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	// File path for each diff entry
	FilePathStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF79C6")).
			MarginTop(1)

	// Timestamp
	TimestampStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")).
			Italic(true)

	// Diff lines
	AddedLineStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B"))
	RemovedLineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555"))
	HunkHeaderStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD"))
	ContextLineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#BFBFBF"))
	ErrorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB86C"))

	// Status bar
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#44475A")).
			Padding(0, 1)

	PausedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")).
			Bold(true)

	// Repo tag (shown in multi-repo mode)
	RepoTagStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#282A36")).
			Background(lipgloss.Color("#BD93F9")).
			Padding(0, 1)

	// Tabs
	ActiveTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F8F8F2")).
			Background(lipgloss.Color("#BD93F9")).
			Padding(0, 1)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6272A4")).
				Background(lipgloss.Color("#282A36")).
				Padding(0, 1)

	TabWithChangesStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F8F8F2")).
				Background(lipgloss.Color("#44475A")).
				Padding(0, 1)

	TabBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#282A36"))

	// Branch name in header (single-repo)
	BranchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B")).
			Bold(true)

	// Branch name in tabs (multi-repo)
	BranchLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#8BE9FD")).
				Italic(true)

	// Separator
	SeparatorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4"))

	// Hidden file indicator
	HiddenFileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")).
			Italic(true)

	// Active file indicator
	ActiveFileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B")).
			Bold(true).
			Background(lipgloss.Color("#282A36")).
			Padding(0, 1)

	// Debug console
	DebugConsoleStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#282A36")).
				Foreground(lipgloss.Color("#F8F8F2")).
				Padding(0, 1).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("#BD93F9"))

	// Test change in styles.go with restored filtering
	// Fourth test with simplified filtering
)

// logMessage writes a debug message to the model debug file
func logMessage(message string) {
	modelDebugMutex.Lock()
	defer modelDebugMutex.Unlock()

	if modelDebugFile != nil {
		timestamp := time.Now().Format("15:04:05")
		modelDebugFile.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, message))
		modelDebugFile.Sync()
	}
}

// initModelDebugLogging initializes the debug log file for the model
func initModelDebugLogging(root string) {
	modelDebugMutex.Lock()
	defer modelDebugMutex.Unlock()

	if modelDebugFile != nil {
		modelDebugFile.Close()
	}

	debugPath := fmt.Sprintf("%s/model.log", root)
	var err error
	modelDebugFile, err = os.Create(debugPath)
	if err != nil {
		modelDebugFile = nil
		return
	}

	modelDebugFile.WriteString("=== Model Log ===\n")
	modelDebugFile.WriteString(fmt.Sprintf("Started: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	modelDebugFile.Sync()
}
