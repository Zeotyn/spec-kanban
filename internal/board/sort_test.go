package board

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/zeotyn/spec-kanban/internal/work"
)

// sorting holds Ready Cards whose default queue order differs from name order, and In Progress
// Cards with and without a creation date.
var sorting = []work.Item{
	{Name: "f2-fix", Total: 2, Phase: work.Ready, Artifacts: planned},
	{Name: "m9-mouse", Total: 2, Phase: work.Ready, Artifacts: planned},
	{Name: "c10-docs", Total: 2, Phase: work.Ready, Artifacts: planned},
	{Name: "a-one", Done: 1, Total: 2, Phase: work.InProgress, Created: "2026-11-10", Modified: ago(1)},
	{Name: "b-two", Done: 1, Total: 2, Phase: work.InProgress, Created: "2026-11-17", Modified: ago(2)},
	{Name: "c-three", Done: 1, Total: 2, Phase: work.InProgress, Modified: ago(3)},
}

func colNames(m Model, p work.Phase) []string {
	var n []string
	for _, c := range m.cols[p] {
		n = append(n, c.Name)
	}
	return n
}

func TestSortKey(t *testing.T) {
	m := board(t, 120, 30, sorting)
	footer := func(m Model) string { f, _ := m.footer(); return f }
	for _, step := range []struct {
		sort         string
		ready, doing []string
	}{
		{"", []string{"f2-fix", "m9-mouse", "c10-docs"}, []string{"a-one", "b-two", "c-three"}},
		{"sort: name", []string{"c10-docs", "f2-fix", "m9-mouse"}, []string{"a-one", "b-two", "c-three"}},
		{"sort: date", []string{"c10-docs", "f2-fix", "m9-mouse"}, []string{"b-two", "a-one", "c-three"}},
		{"", []string{"f2-fix", "m9-mouse", "c10-docs"}, []string{"a-one", "b-two", "c-three"}},
	} {
		if got := colNames(m, work.Ready); !slices.Equal(got, step.ready) {
			t.Errorf("%q: Ready %q, want %q", step.sort, got, step.ready)
		}
		if got := colNames(m, work.InProgress); !slices.Equal(got, step.doing) {
			t.Errorf("%q: In Progress %q, want %q", step.sort, got, step.doing)
		}
		if f := footer(m); step.sort != "" && !strings.HasPrefix(f, step.sort+" · ") || step.sort == "" && strings.Contains(f, "sort:") {
			t.Errorf("%q: footer %q", step.sort, f)
		}
		m = keys(m, "o")
	}
}

func TestSortKeepsSelection(t *testing.T) {
	m := keys(board(t, 120, 30, sorting), "2", "j")
	if m.cols[work.Ready][m.nav.sel[work.Ready]].Name != "m9-mouse" {
		t.Fatalf("selected %q", m.cols[work.Ready][m.nav.sel[work.Ready]].Name)
	}
	m = keys(m, "o")
	if got := m.cols[work.Ready][m.nav.sel[work.Ready]].Name; got != "m9-mouse" || m.nav.sel[work.Ready] != 2 {
		t.Errorf("after o: selected %q at %d, want m9-mouse at 2", got, m.nav.sel[work.Ready])
	}
}

func TestSortWithFilterAndReload(t *testing.T) {
	m := keys(board(t, 120, 30, sorting), "/", "-", "enter", "o")
	if got := colNames(m, work.Ready); !slices.Equal(got, []string{"c10-docs", "f2-fix", "m9-mouse"}) {
		t.Errorf("filtered and sorted Ready %q", got)
	}
	m = send(m, m.reload())
	if got := colNames(m, work.Ready); !slices.Equal(got, []string{"c10-docs", "f2-fix", "m9-mouse"}) || m.filter != "-" {
		t.Errorf("after reload: Ready %q, filter %q", got, m.filter)
	}
}

func TestSortStartsDefault(t *testing.T) {
	if m := board(t, 120, 30, sorting); m.sort != work.Default {
		t.Errorf("new Board sort %v", m.sort)
	}
}

func TestHelpListsSort(t *testing.T) {
	if v := keys(board(t, 120, 40, sorting), "?").View(); !strings.Contains(v, "sort by name, date, default") {
		t.Errorf("help lacks o:\n%s", v)
	}
}

func TestCardCreated(t *testing.T) {
	m := board(t, 120, 30, nil)
	third := func(c work.Item, p work.Phase, w int) string { return m.card(c, p, w, false)[3] }
	doing := work.Item{Name: "m1-board-view", Done: 5, Total: 11, Phase: work.InProgress, Modified: ago(3 * time.Hour)}
	for _, tt := range []struct {
		name, created string
		p             work.Phase
		w             int
		want          string
	}{
		{"current year", "2026-10-17", work.InProgress, 24, "10-17 · 3h"},
		{"other year", "2025-11-17", work.InProgress, 24, "2025-11-17 · 3h"},
		{"no date", "", work.InProgress, 24, "│3h "},
		{"with plan marks", "2026-10-17", work.Ready, 24, "P S D T  10-17 · 3h"},
		{"too narrow drops the date", "2026-10-17", work.Ready, 20, "P S D T  3h "},
	} {
		c := doing
		c.Created, c.Phase, c.Artifacts = tt.created, tt.p, planned
		if got := third(c, tt.p, tt.w); !strings.Contains(got, tt.want) {
			t.Errorf("%s: %q, want %q in it", tt.name, got, tt.want)
		}
	}
}
