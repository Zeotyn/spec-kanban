// Package theme is the palette of the Board and the detail view: only the 16 ANSI colours, bold and
// faint, so the terminal's own theme decides how things look.
package theme

import "github.com/charmbracelet/lipgloss"

var (
	Accent    = lipgloss.Color("4") // selected Card border, active tab, help group titles
	Dim       = lipgloss.Color("8") // Card and help borders, Draft and Archived
	Milestone = lipgloss.Color("5") // M<N> badge
	Running   = lipgloss.Color("2") // ● Running agent badge
)

var (
	Faint      = lipgloss.NewStyle().Faint(true)
	AccentBold = lipgloss.NewStyle().Foreground(Accent).Bold(true)
)
