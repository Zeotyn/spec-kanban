package detail

import (
	"slices"
	"strings"
	"testing"

	"github.com/zeotyn/openspec-kanban/internal/work"
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
