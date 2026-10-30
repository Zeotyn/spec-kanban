package board

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/zeotyn/spec-kanban/internal/theme"
	"github.com/zeotyn/spec-kanban/internal/work"
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

var (
	bold  = lipgloss.NewStyle().Bold(true)
	faint = lipgloss.NewStyle().Faint(true)
)

// noHighColours reports 256-colour and true-colour sequences, which would override the terminal theme.
func noHighColours(t *testing.T, v string) {
	t.Helper()
	for _, s := range []string{"38;5;", "48;5;", "38;2;"} {
		if strings.Contains(v, s) {
			t.Errorf("view has %q: only the 16 ANSI colours may be used:\n%q", s, v)
		}
	}
}

func TestCardTypography(t *testing.T) {
	coloured(t)
	running := agentsMsg{running: map[string]bool{"m1-board-view": true}}
	for _, h := range []int{30, 12} { // bordered and fallback Cards
		v := keys(send(board(t, 120, h, fixture), running), "2").View()
		contains(t, v, fg("2").Render("●")+" "+bold.Render("m1-board-view"), fg("5").Render("M1"),
			faint.Render("P S D T  40m"), faint.Render("3h"))
		noHighColours(t, v)
	}
	v := keys(board(t, 120, 30, fixture), "5").View()
	contains(t, v, faint.Render("archived 2026-10-29"),
		fg("4").Render("┏"+strings.Repeat("━", 21)+"┓"), fg("8").Render("╭"+strings.Repeat("─", 21)+"╮"))
}

func TestFooterStyle(t *testing.T) {
	coloured(t)
	v := board(t, 200, 30, fixture).View()
	contains(t, v, "h/l "+faint.Render("column")+" · j/k "+faint.Render("card"), "q "+faint.Render("close"))
	noHighColours(t, v)
	contains(t, board(t, 120, 30, fixture).View(), "\nh/l j/k ⏎ e s / gg/G ^d/^u 1-5 · r · ? · q "+faint.Render("close"))
	m := keys(board(t, 120, 30, fixture), "2")
	starter(&m, nil)
	for name, m := range map[string]Model{"prompt": keys(m, "/", "x"), "notice": keys(m, "1", "s"), "starting": keys(m, "s")} {
		v := m.View()
		if footer := v[strings.LastIndex(v, "\n")+1:]; strings.Contains(footer, "\x1b") {
			t.Errorf("%s footer must carry no style: %q", name, footer)
		}
	}
}

func TestHelpStyle(t *testing.T) {
	top := strings.SplitN(keys(board(t, 120, 40, fixture), "?").View(), "\n", 2)[0]
	coloured(t)
	v := keys(board(t, 120, 40, fixture), "?").View()
	contains(t, v, theme.AccentBold.Render("Board"), theme.AccentBold.Render("Filter prompt"), theme.AccentBold.Render("Details"),
		"  gg/G         "+faint.Render("first/last card"), "spec-kanban test · any key closes")
	if !strings.HasPrefix(v, fg("8").Render(top)+"\n") {
		t.Errorf("the frame must be dim:\n%q", v)
	}
	noHighColours(t, v)
}
