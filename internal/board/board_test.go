package board

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/zeotyn/openspec-kanban/internal/detail"
	"github.com/zeotyn/openspec-kanban/internal/openspec"
)

var update = flag.Bool("update", false, "rewrite golden files")

func TestMain(m *testing.M) {
	lipgloss.SetColorProfile(termenv.Ascii) // goldens without escape codes
	detail.Style = "notty"
	os.Exit(m.Run())
}

func golden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", "golden", name+".txt")
	if *update {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v (run go test -update)", err)
	}
	if got != string(want) {
		t.Errorf("view differs from %s\n got:\n%s\nwant:\n%s", path, got, want)
	}
}

var now = time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC)

func ago(d time.Duration) time.Time { return now.Add(-d) }

var planned = map[string]string{"proposal": "done", "specs": "done", "design": "done", "tasks": "done"}

// fixture is a Board with every Phase, a long name, milestone badges and an undated archive.
var fixture = []openspec.Change{
	{Name: "add-dark-mode", Modified: ago(48 * time.Hour), Planning: openspec.Incomplete,
		Artifacts: map[string]string{"proposal": "done", "specs": "ready", "design": "ready", "tasks": "blocked"}},
	{Name: "m2-change-detail", Modified: ago(5 * time.Hour), Planning: openspec.Incomplete,
		Artifacts: map[string]string{"proposal": "done", "specs": "ready", "design": "ready", "tasks": "blocked"}},
	{Name: "m1-phase-derivation", Total: 6, Modified: ago(40 * time.Minute), Planning: openspec.Complete, Artifacts: planned},
	{Name: "m5-release-pipeline-with-a-very-long-name", Total: 4, Modified: ago(20 * 24 * time.Hour), Planning: openspec.Complete,
		Artifacts: map[string]string{"proposal": "done", "specs": "skipped", "design": "done", "tasks": "done"}},
	{Name: "m1-board-view", Done: 5, Total: 11, Modified: ago(3 * time.Hour)},
	{Name: "m0-spike-agent-start", Done: 10, Total: 10, Modified: ago(26 * time.Hour)},
	{Name: "m0-plugin-skeleton", Archived: true, ArchivedOn: "2026-10-29"},
	{Name: "undated-change", Archived: true},
}

// board returns a loaded Board of the fixture at the given size.
func board(t *testing.T, w, h int, changes []openspec.Change) Model {
	t.Helper()
	m := New("/x/proj", func() ([]openspec.Change, error) { return changes, nil }, func() time.Time { return now }, nil)
	m = send(m, tea.WindowSizeMsg{Width: w, Height: h})
	return send(m, m.Init()())
}

func send(m Model, msg tea.Msg) Model {
	next, _ := m.Update(msg)
	return next.(Model)
}

func keys(m Model, ks ...string) Model {
	for _, k := range ks {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
		switch k {
		case "ctrl+d":
			msg = tea.KeyMsg{Type: tea.KeyCtrlD}
		case "ctrl+u":
			msg = tea.KeyMsg{Type: tea.KeyCtrlU}
		}
		m = send(m, msg)
	}
	return m
}

func TestQuitKeys(t *testing.T) {
	for _, key := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune("q")},
		{Type: tea.KeyEsc},
		{Type: tea.KeyCtrlC},
	} {
		_, cmd := board(t, 120, 30, fixture).Update(key)
		if cmd == nil || cmd() != tea.Quit() {
			t.Errorf("%q does not quit", key)
		}
	}
}

func TestLoading(t *testing.T) {
	m := New("/x/proj", func() ([]openspec.Change, error) { return fixture, nil }, func() time.Time { return now }, nil)
	if !strings.Contains(m.View(), "Loading changes…") {
		t.Errorf("no loading notice before data:\n%s", m.View())
	}
	m = send(m, m.Init()())
	if strings.Contains(m.View(), "Loading") || !strings.Contains(m.View(), "m1-board-view") {
		t.Errorf("columns not shown after load:\n%s", m.View())
	}
}

func TestReload(t *testing.T) {
	calls := 0
	changes := fixture
	m := New("/x/proj", func() ([]openspec.Change, error) { calls++; return changes, nil }, func() time.Time { return now }, nil)
	m = send(m, m.Init()())
	m = keys(m, "3") // In Progress: m1-board-view
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	m = next.(Model)
	if cmd == nil || !strings.Contains(m.View(), "refreshing…") {
		t.Fatal("r must reload and show it in the header")
	}
	changes = append([]openspec.Change{{Name: "m3-new", Done: 1, Total: 2, Modified: now}}, fixture...)
	m = send(m, cmd())
	if calls != 2 {
		t.Errorf("loader called %d times, want 2", calls)
	}
	if got := m.cols[2][m.nav.sel[2]].Name; got != "m1-board-view" {
		t.Errorf("selection after reload %q, want m1-board-view", got)
	}
}

func TestLoadError(t *testing.T) {
	m := New("/x/proj", func() ([]openspec.Change, error) {
		return nil, errors.New("openspec CLI not found: install it with npm i -g @fission-ai/openspec")
	}, func() time.Time { return now }, nil)
	m = send(m, tea.WindowSizeMsg{Width: 120, Height: 30})
	m = send(m, m.Init()())
	golden(t, "error", m.View())
	if _, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}); cmd == nil || cmd() != tea.Quit() {
		t.Error("q must close the error view")
	}
}

func TestViews(t *testing.T) {
	golden(t, "no-project", New("", nil, time.Now, nil).View())
	golden(t, "loading", New("/x/proj", nil, time.Now, nil).View())
	golden(t, "board-200", board(t, 200, 30, fixture).View())
	golden(t, "board-120", keys(board(t, 120, 30, fixture), "l", "j").View())
	golden(t, "board-60", keys(board(t, 60, 30, fixture), "l", "l").View())
	golden(t, "board-empty", board(t, 120, 30, nil).View())
	golden(t, "board-scrolled", keys(board(t, 120, 14, fixture), "j").View()) // room for one Card per column
}

// tall is a Ready column of n Cards.
func tall(n int) []openspec.Change {
	cs := make([]openspec.Change, n)
	for i := range cs {
		cs[i] = openspec.Change{Name: fmt.Sprintf("change-%02d", i), Total: 2, Planning: openspec.Complete, Modified: ago(time.Duration(i) * time.Hour)}
	}
	return cs
}

func TestViewFitsHeight(t *testing.T) {
	for _, w := range []int{60, 120} {
		for h := 14; h <= 40; h++ {
			m := keys(board(t, w, h, tall(12)), "2", "j", "j", "j", "j", "j")
			if lines := strings.Count(m.View(), "\n") + 1; lines > h {
				t.Errorf("%dx%d: view has %d lines", w, h, lines)
			}
		}
	}
	m := board(t, 60, 14, fixture)
	m.load = func() ([]openspec.Change, error) {
		return nil, errors.New("openspec list: " + strings.Repeat("a long and\nmulti-line error ", 20))
	}
	m = send(m, m.reload())
	if lines := strings.Count(m.View(), "\n") + 1; lines > 14 {
		t.Errorf("long error: view has %d lines in 14:\n%s", lines, m.View())
	}
}

func TestResizeKeepsSelectionVisible(t *testing.T) {
	m := keys(board(t, 120, 60, tall(10)), "2", "G")
	m = send(m, tea.WindowSizeMsg{Width: 120, Height: 16})
	if !strings.Contains(m.View(), "▌change-09") {
		t.Errorf("selected Card not visible after shrinking:\n%s", m.View())
	}
}

func TestReloadKeys(t *testing.T) {
	m := keys(board(t, 120, 30, tall(5)), "2", "G", "g")
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	m = next.(Model)
	if _, again := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")}); again != nil {
		t.Error("r while loading must not start a second load")
	}
	m = keys(send(m, cmd()), "g")
	if m.nav.sel[1] != 4 {
		t.Errorf("g, r, g jumped to Card %d; r must cancel the pending g", m.nav.sel[1])
	}
}

func TestHeaderProgress(t *testing.T) {
	m := board(t, 120, 30, []openspec.Change{{Name: "a", Done: 120, Total: 287, Modified: now}})
	if !strings.Contains(m.View(), "120/287 42%") {
		t.Errorf("want rounded 120/287 42%% in header:\n%s", m.View())
	}
	m.load = func() ([]openspec.Change, error) { return nil, errors.New("boom") }
	m = send(m, m.reload())
	if strings.Contains(m.View(), "%") {
		t.Errorf("error view must not show stale progress:\n%s", m.View())
	}
}

func TestSnapshot(t *testing.T) {
	fresh := append([]openspec.Change{{Name: "m3-new", Done: 1, Total: 2, Modified: now}}, fixture...)
	m := New("/x/proj", func() ([]openspec.Change, error) { return fresh, nil }, func() time.Time { return now }, fixture)
	m = send(m, tea.WindowSizeMsg{Width: 120, Height: 30})
	if v := m.View(); strings.Contains(v, "Loading changes…") || !strings.Contains(v, "refreshing…") || !strings.Contains(v, "m1-board-view") {
		t.Fatalf("snapshot must show at once, marked refreshing:\n%s", v)
	}
	golden(t, "refreshing", m.View())
	m = keys(m, "3") // In Progress: m1-board-view
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("a snapshot must still trigger a fresh load")
	}
	m = send(m, cmd())
	if v := m.View(); strings.Contains(v, "refreshing…") || !strings.Contains(v, "m3-new") {
		t.Errorf("fresh load must replace the snapshot and clear the marker:\n%s", v)
	}
	if got := m.cols[2][m.nav.sel[2]].Name; got != "m1-board-view" {
		t.Errorf("selection after the swap %q, want m1-board-view", got)
	}
}

func TestSnapshotMovesFocusOffEmptyDraft(t *testing.T) {
	m := New("/x/proj", nil, func() time.Time { return now }, []openspec.Change{{Name: "doing", Done: 1, Total: 2, Modified: now}})
	if m.nav.col != 2 {
		t.Errorf("focus %d, want In Progress (2)", m.nav.col)
	}
}

func enter(m Model) (Model, tea.Cmd) {
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	return next.(Model), cmd
}

func TestEnterOpensDetail(t *testing.T) {
	m, _ := enter(keys(board(t, 120, 30, fixture), "3")) // In Progress: m1-board-view
	if m.detail == nil || !strings.HasPrefix(m.View(), "m1-board-view  M1  5/11  In Progress") {
		t.Fatalf("Enter must open the selected Card's details:\n%s", m.View())
	}
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	m = send(next.(Model), cmd())
	if m.detail != nil || m.nav.col != 2 || m.cols[2][m.nav.sel[2]].Name != "m1-board-view" {
		t.Errorf("q must return to the Board with the same selection:\n%s", m.View())
	}
}

func TestEnterOnArchivedCard(t *testing.T) {
	m, _ := enter(keys(board(t, 120, 30, fixture), "5"))
	if !strings.HasPrefix(m.View(), "m0-plugin-skeleton  M0  archived 2026-10-29") {
		t.Errorf("archived header:\n%s", m.View())
	}
}

func TestEnterOnEmptyBoard(t *testing.T) {
	if m, _ := enter(board(t, 120, 30, nil)); m.detail != nil {
		t.Error("Enter without Cards must do nothing")
	}
}

func TestDetailKeysAndQuit(t *testing.T) {
	m, _ := enter(keys(board(t, 120, 30, fixture), "3"))
	m = keys(m, "j", "l") // detail keys; l is not a detail key and must not move the Board
	if m.nav.col != 2 {
		t.Errorf("keys in the detail must not reach the Board, col %d", m.nav.col)
	}
	if _, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC}); cmd == nil || cmd() != tea.Quit() {
		t.Error("ctrl+c must quit from the detail")
	}
}

func TestDetailStaysOpenAcrossReloadAndResize(t *testing.T) {
	m, _ := enter(keys(board(t, 120, 30, fixture), "3"))
	m = send(m, loadedMsg{changes: fixture})
	m = send(m, tea.WindowSizeMsg{Width: 60, Height: 20})
	if m.detail == nil || !strings.HasPrefix(m.View(), "m1-board-view") {
		t.Fatalf("a reload or resize must not close the detail:\n%s", m.View())
	}
	if lines := strings.Count(m.View(), "\n") + 1; lines > 20 || m.width != 60 {
		t.Errorf("resize must reach detail and Board: %d lines, board width %d", lines, m.width)
	}
}

func TestEnterCancelsPendingG(t *testing.T) {
	m := keys(board(t, 120, 30, tall(5)), "2", "G", "g")
	m, _ = enter(m)
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	m = keys(send(next.(Model), cmd()), "g")
	if m.nav.sel[1] != 4 {
		t.Errorf("g, Enter, q, g jumped to Card %d; Enter must cancel the pending g", m.nav.sel[1])
	}
}
