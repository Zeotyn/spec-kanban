package detail

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/zeotyn/spec-kanban/internal/source/speckit"
	"github.com/zeotyn/spec-kanban/internal/theme"
	"github.com/zeotyn/spec-kanban/internal/work"
)

// three is a Source's Docs with three tabs, tasks last and the second one missing.
var three = work.Docs{
	Tabs: []work.Doc{
		{Title: "Spec", Body: "# Spec\n\nWhat to build.\n", Files: []string{"/w/spec.md"}},
		{Title: "Plan", Files: []string{"/w/plan.md"}},
		{Title: "Tasks", Body: "- [ ] T1 Build it\n", Files: []string{"/w/tasks.md", "/w/more.md"}},
	},
	Tasks: 2,
	First: 0,
	Next:  "next ▸ T1 Build it",
}

func openThree(it work.Item) Model {
	return New(func() work.Docs { return three }, it, "h", 120, 24)
}

func TestThreeTabs(t *testing.T) {
	m := openThree(work.Item{Name: "x", Done: 0, Total: 1})
	if m.tab != 2 {
		t.Errorf("with tasks: tab %d, want the tasks tab 2", m.tab)
	}
	v := m.View()
	for _, s := range []string{" 1 Spec │ ·2 Plan │ ▌3 Tasks", "1-3/tab", "next ▸ T1 Build it"} {
		if !strings.Contains(v, s) {
			t.Errorf("view lacks %q:\n%s", s, v)
		}
	}
	if m = press(m, "4"); m.tab != 2 {
		t.Errorf("4 with three tabs: tab %d, want 2", m.tab)
	}
	if m = press(m, "tab"); m.tab != 0 {
		t.Errorf("tab after the last: %d, want 0", m.tab)
	}
	if m := openThree(work.Item{Name: "x"}); m.tab != 0 || strings.Contains(m.View(), "next ▸") {
		t.Errorf("without tasks: tab %d, want First 0 without the next line:\n%s", m.tab, m.View())
	}
}

func TestThreeTabsEdit(t *testing.T) {
	args := fakeEditor(t, nil, nil)
	m := openThree(work.Item{Name: "x", Total: 1})
	if _, ran := edit(m); !ran || !slices.Equal(*args, []string{"ed", "/w/tasks.md", "/w/more.md"}) {
		t.Errorf("e on the tasks tab: editor %q", *args)
	}
	*args = nil
	m = press(m, "2")
	if !strings.Contains(m.View(), "This change has no plan.") {
		t.Errorf("empty tab needs its notice:\n%s", m.View())
	}
	if m, ran := edit(m); ran || *args != nil || !strings.Contains(m.View(), "no plan to edit") {
		t.Errorf("e on an empty tab: ran %v, view:\n%s", ran, m.View())
	}
}

func TestTabBarStyle(t *testing.T) {
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })
	m := openThree(work.Item{Name: "x", Total: 1})
	for _, c := range []struct {
		width int
		want  []string
	}{
		{120, []string{theme.AccentBold.Render("▌3 Tasks"), theme.Faint.Render(" 1 Spec"), theme.Faint.Render("·2 Plan")}},
		{20, []string{theme.AccentBold.Render("▌3 T"), theme.Faint.Render(" 1 S"), theme.Faint.Render("·2 P")}}, // short form
	} {
		m, _ = m.Update(tea.WindowSizeMsg{Width: c.width, Height: 24})
		v := m.View()
		for _, s := range c.want {
			if !strings.Contains(v, s) {
				t.Errorf("%d wide: tab bar lacks %q:\n%q", c.width, s, v)
			}
		}
		for _, s := range []string{"38;5;", "48;5;", "38;2;"} {
			if strings.Contains(v, s) {
				t.Errorf("%d wide: only the 16 ANSI colours may be used, found %q", c.width, s)
			}
		}
	}
}

func TestSpecKitSpecOnly(t *testing.T) {
	p := t.TempDir()
	if err := os.MkdirAll(filepath.Join(p, "specs", "001-a"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p, "specs", "001-a", "spec.md"), []byte("# Feature Specification\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	it := work.Item{Name: "001-a"}
	m := New(func() work.Docs { return speckit.New(nil, "").Docs(p, it) }, it, "h", 120, 24)
	if m.tabs() != 2 || m.tab != 1 {
		t.Errorf("tabs %d, shown %d; want 2 tabs, Spec shown", m.tabs(), m.tab)
	}
	if v := m.View(); !strings.Contains(v, "·1 Tasks │ ▌2 Spec") || !strings.Contains(v, "Feature Specification") {
		t.Errorf("view:\n%s", v)
	}
	if m = press(m, "3"); m.tab != 1 {
		t.Errorf("3 with two tabs: tab %d, want 1", m.tab)
	}
	if m = press(m, "1"); !strings.Contains(m.View(), "This change has no tasks.") {
		t.Errorf("Tasks tab needs the no-tasks notice:\n%s", m.View())
	}
}
