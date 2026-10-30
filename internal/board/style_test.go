package board

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/zeotyn/openspec-kanban/internal/work"
)

// coloured renders with 256 colours until the test ends.
func coloured(t *testing.T) {
	t.Helper()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })
}

func fg(c string) lipgloss.Style { return lipgloss.NewStyle().Foreground(lipgloss.Color(c)) }

// contains reports the fragments missing from v.
func contains(t *testing.T, v string, fragments ...string) {
	t.Helper()
	for _, f := range fragments {
		if !strings.Contains(v, f) {
			t.Errorf("view lacks %q:\n%q", f, v)
		}
	}
}

func TestPhaseColor(t *testing.T) {
	for p, want := range []lipgloss.Color{"8", "4", "3", "2", "8"} {
		if got := phaseColor(work.Phase(p)); got != want {
			t.Errorf("%s: %q, want %q", work.Phase(p), got, want)
		}
	}
}

func TestPhaseColours(t *testing.T) {
	coloured(t)
	contains(t, board(t, 120, 30, fixture).View(),
		fg("4").Render("Ready 2"), fg("4").Render(strings.Repeat("─", 23)),
		fg("3").Render("█████"), lipgloss.NewStyle().Faint(true).Render("░░░░░"))
	contains(t, keys(board(t, 60, 30, fixture), "l", "l").View(), fg("3").Render("◀ 3/5 In Progress 1 ▶"))
}
