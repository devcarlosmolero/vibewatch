// help.go provides the help text and styling for vibewatch's interactive help system.
// This includes keybindings documentation and visual styling for the help overlay.
// Testing file change detection after disabling git ignore filter.
// Another test in help.go with simplified filtering.
package model

import "github.com/charmbracelet/lipgloss"

// helpStyle defines the visual appearance of the help overlay using a
// purple-themed color scheme that matches vibewatch's aesthetic.
// Colors use hex codes: background (#282A36), foreground (#F8F8F2), border (#BD93F9)
var helpStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#F8F8F2")).
	Background(lipgloss.Color("#282A36")).
	Padding(1, 2).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#BD93F9"))

func renderHelp() string {
	content := `  vibewatch - Real-time Code Change Viewer

  Keybindings:
  ───────────────────────────────
  Tab            Next repo tab
  Shift+Tab      Previous repo tab
  1-9            Jump to tab by number
  j / ↓          Scroll down
  k / ↑          Scroll up
  d / PgDn       Page down
  u / PgUp       Page up
  g / Home       Go to top
  G / End        Go to bottom
  p              Pause / Resume
  c              Clear all entries
  t              Toggle file visibility
  U              Untoggle file visibility
  ↑/↓/j/k        Navigate between files
  ?              Toggle this help
  q / Ctrl+C     Quit`

	return helpStyle.Render(content)
}
