// help.go provides the help text and styling for vibewatch's interactive help system.
// This includes keybindings documentation and visual styling for the help overlay.
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
	content := "  Keybindings:\n" +
		"  ───────────────────────────────\n" +
		"  Tab            Next repo tab\n" +
		"  Shift+Tab      Previous repo tab\n" +
		"  1-9            Jump to tab by number\n" +
		"  j / ↓          Scroll down\n" +
		"  k / ↑          Scroll up\n" +
		"  d / PgDn       Page down\n" +
		"  u / PgUp       Page up\n" +
		"  g / Home       Go to top\n" +
		"  G / End        Go to bottom\n" +
		"  p              Pause / Resume\n" +
		"  c              Clear all entries\n" +
		"  t              Toggle file visibility\n" +
		"  T              Toggle first visible file\n" +
		"  U              Untoggle file visibility\n" +
		"  ?              Toggle this help\n" +
		"  q / Ctrl+C     Quit"

	return helpStyle.Render(content)
}
