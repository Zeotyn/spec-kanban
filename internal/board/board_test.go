package board

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/zeotyn/spec-kanban/internal/detail"
	"github.com/zeotyn/spec-kanban/internal/editor"
	"github.com/zeotyn/spec-kanban/internal/source/openspec"
	"github.com/zeotyn/spec-kanban/internal/work"
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
var fixture = []work.Item{
	{Name: "add-dark-mode", Modified: ago(48 * time.Hour), Phase: work.Draft,
		Artifacts: map[string]string{"proposal": "done", "specs": "ready", "design": "ready", "tasks": "blocked"}},
	{Name: "m2-change-detail", Modified: ago(5 * time.Hour), Phase: work.Draft,
		Artifacts: map[string]string{"proposal": "done", "specs": "ready", "design": "ready", "tasks": "blocked"}},
	{Name: "m1-phase-derivation", Total: 6, Modified: ago(40 * time.Minute), Phase: work.Ready, Artifacts: planned},
	{Name: "m5-release-pipeline-with-a-very-long-name", Total: 4, Modified: ago(20 * 24 * time.Hour), Phase: work.Ready,
		Artifacts: map[string]string{"proposal": "done", "specs": "skipped", "design": "done", "tasks": "done"}},
	{Name: "m1-board-view", Done: 5, Total: 11, Phase: work.InProgress, Modified: ago(3 * time.Hour)},
	{Name: "m0-spike-agent-start", Done: 10, Total: 10, Phase: work.Review, Modified: ago(26 * time.Hour)},
	{Name: "m0-plugin-skeleton", Phase: work.Archived, ArchivedOn: "2026-10-29"},
	{Name: "undated-change", Phase: work.Archived},
}

// docs reads Items' documents from the OpenSpec layout below project.
func docs(project string) func(work.Item) work.Docs {
	return func(it work.Item) work.Docs { return openspec.New(nil).Docs(project, it) }
}

// board returns a loaded Board of changes at the given size.
func board(t *testing.T, w, h int, changes []work.Item) Model {
	t.Helper()
	return hiding(t, w, h, changes)
}

// hiding returns a loaded Board of changes at the given size that hides the Phases ps.
func hiding(t *testing.T, w, h int, changes []work.Item, ps ...work.Phase) Model {
	t.Helper()
	var hide [columns]bool
	for _, p := range ps {
		hide[p] = true
	}
	m := New("/x/proj", docs("/x/proj"), func() ([]work.Item, error) { return changes, nil }, func() time.Time { return now }, nil, nil, nil, nil, nil, hide, "", "test")
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
		case "enter":
			msg = tea.KeyMsg{Type: tea.KeyEnter}
		case "esc":
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		case "backspace":
			msg = tea.KeyMsg{Type: tea.KeyBackspace}
		case " ":
			msg = tea.KeyMsg{Type: tea.KeySpace, Runes: []rune(" ")}
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
	m := New("/x/proj", docs("/x/proj"), func() ([]work.Item, error) { return fixture, nil }, func() time.Time { return now }, nil, nil, nil, nil, nil, [columns]bool{}, "", "test")
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
	m := New("/x/proj", docs("/x/proj"), func() ([]work.Item, error) { calls++; return changes, nil }, func() time.Time { return now }, nil, nil, nil, nil, nil, [columns]bool{}, "", "test")
	m = send(m, m.Init()())
	m = keys(m, "3") // In Progress: m1-board-view
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	m = next.(Model)
	if cmd == nil || !strings.Contains(m.View(), "refreshing…") {
		t.Fatal("r must reload and show it in the header")
	}
	changes = append([]work.Item{{Name: "m3-new", Done: 1, Total: 2, Phase: work.InProgress, Modified: now}}, fixture...)
	m = send(m, cmd())
	if calls != 2 {
		t.Errorf("loader called %d times, want 2", calls)
	}
	if got := m.cols[2][m.nav.sel[2]].Name; got != "m1-board-view" {
		t.Errorf("selection after reload %q, want m1-board-view", got)
	}
}

func TestLoadError(t *testing.T) {
	m := New("/x/proj", docs("/x/proj"), func() ([]work.Item, error) {
		return nil, errors.New("openspec CLI not found: install it with npm i -g @fission-ai/openspec")
	}, func() time.Time { return now }, nil, nil, nil, nil, nil, [columns]bool{}, "", "test")
	m = send(m, tea.WindowSizeMsg{Width: 120, Height: 30})
	m = send(m, m.Init()())
	golden(t, "error", m.View())
	if _, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}); cmd == nil || cmd() != tea.Quit() {
		t.Error("q must close the error view")
	}
}

func TestViews(t *testing.T) {
	golden(t, "no-project", New("", nil, nil, time.Now, nil, nil, nil, nil, nil, [columns]bool{}, "", "test").View())
	golden(t, "loading", New("/x/proj", docs("/x/proj"), nil, time.Now, nil, nil, nil, nil, nil, [columns]bool{}, "", "test").View())
	golden(t, "board-200", board(t, 200, 30, fixture).View())
	golden(t, "board-120", keys(board(t, 120, 30, fixture), "l", "j").View())
	golden(t, "board-60", keys(board(t, 60, 30, fixture), "l", "l").View())
	golden(t, "board-empty", board(t, 120, 30, nil).View())
	golden(t, "board-scrolled", keys(board(t, 120, 14, fixture), "j").View()) // room for one Card per column
	running := agentsMsg{running: map[string]bool{"m1-board-view": true, "m5-release-pipeline-with-a-very-long-name": true}}
	golden(t, "board-running", keys(send(board(t, 120, 30, fixture), running), "3").View())
}

// tall is a Ready column of n Cards.
func tall(n int) []work.Item {
	cs := make([]work.Item, n)
	for i := range cs {
		cs[i] = work.Item{Name: fmt.Sprintf("change-%02d", i), Total: 2, Phase: work.Ready, Modified: ago(time.Duration(i) * time.Hour)}
	}
	return cs
}

func TestViewFitsHeight(t *testing.T) {
	for _, w := range []int{60, 120} {
		for h := 10; h <= 40; h++ {
			m := keys(board(t, w, h, tall(12)), "2", "j", "j", "j", "j", "j")
			if lines := strings.Count(m.View(), "\n") + 1; lines > h {
				t.Errorf("%dx%d: view has %d lines", w, h, lines)
			}
		}
	}
	m := board(t, 60, 14, fixture)
	m.load = func() ([]work.Item, error) {
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
	if !strings.Contains(m.View(), "┃change-09") {
		t.Errorf("selected Card not visible after shrinking:\n%s", m.View())
	}
	m = send(m, tea.WindowSizeMsg{Width: 120, Height: 12})
	if !strings.Contains(m.View(), "▌change-09") {
		t.Errorf("selected Card not visible below 13 rows:\n%s", m.View())
	}
}

func TestBorderedCards(t *testing.T) {
	m := keys(board(t, 120, 30, tall(12)), "2")
	lines := strings.Split(m.column(1, 23, "Ready 12"), "\n")
	if len(lines) != 2+4*5+1 || lines[22] != pad("↓ 8 more", 23) {
		t.Fatalf("want four 5-line Cards and a scroll marker:\n%s", strings.Join(lines, "\n"))
	}
	for i, top := range []string{"┏", "╭", "╭", "╭"} { // the boxes follow each other without a blank line
		if !strings.HasPrefix(lines[2+5*i], top) || !strings.HasPrefix(lines[6+5*i], map[string]string{"┏": "┗", "╭": "╰"}[top]) {
			t.Errorf("Card %d is not a closed box:\n%s", i, strings.Join(lines, "\n"))
		}
	}
	for _, l := range lines {
		if lipgloss.Width(l) != 23 {
			t.Errorf("line %q is %d cells, want the column width 23", l, lipgloss.Width(l))
		}
	}
	if v := m.View(); !strings.Contains(v, "┃change-00") || strings.Contains(v, "▌") {
		t.Errorf("the selected Card has a thick border and no ▌:\n%s", v)
	}
	if v := keys(board(t, 120, 13, tall(3)), "2").View(); !strings.Contains(v, "┃change-00") {
		t.Errorf("13 rows: want a bordered Card:\n%s", v)
	}
	if v := keys(board(t, 60, 30, tall(3)), "2").View(); !strings.Contains(v, "┃change-00") || !strings.Contains(v, "╭") {
		t.Errorf("single column: want bordered Cards:\n%s", v)
	}
	low := keys(board(t, 120, 12, tall(3)), "2")
	if v := low.View(); strings.ContainsAny(v, "╭┏") || !strings.Contains(v, "▌change-00") {
		t.Errorf("12 rows: want borderless Cards with ▌:\n%s", v)
	}
	if c := low.card(tall(1)[0], work.Ready, 23, false); len(c) != 4 || c[3] != "" {
		t.Errorf("12 rows: a Card is three lines and a blank one: %q", c)
	}
}

func TestBoardFitsWidth(t *testing.T) {
	for _, w := range []int{60, 104, 120} {
		v := board(t, w, 30, fixture).View()
		for _, l := range strings.Split(v, "\n") {
			if lipgloss.Width(l) > w {
				t.Errorf("%d: line %q is %d cells", w, l, lipgloss.Width(l))
			}
		}
		if w > 60 && !strings.Contains(v, "2026-10-29") {
			t.Errorf("%d: the archive date must not be cut:\n%s", w, v)
		}
		if w == 104 && strings.Count(v, "\n") != strings.Count(board(t, 120, 30, fixture).View(), "\n") {
			t.Errorf("104: a long line must not wrap a Card:\n%s", v)
		}
	}
}

// pad fills s with spaces to w cells.
func pad(s string, w int) string { return s + strings.Repeat(" ", w-lipgloss.Width(s)) }

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
	m := board(t, 120, 30, []work.Item{{Name: "a", Done: 120, Total: 287, Phase: work.InProgress, Modified: now}})
	if !strings.Contains(m.View(), "120/287 42%") {
		t.Errorf("want rounded 120/287 42%% in header:\n%s", m.View())
	}
	m.load = func() ([]work.Item, error) { return nil, errors.New("boom") }
	m = send(m, m.reload())
	if strings.Contains(m.View(), "%") {
		t.Errorf("error view must not show stale progress:\n%s", m.View())
	}
}

func TestSnapshot(t *testing.T) {
	fresh := append([]work.Item{{Name: "m3-new", Done: 1, Total: 2, Phase: work.InProgress, Modified: now}}, fixture...)
	m := New("/x/proj", docs("/x/proj"), func() ([]work.Item, error) { return fresh, nil }, func() time.Time { return now }, fixture, nil, nil, nil, nil, [columns]bool{}, "", "test")
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
	m := New("/x/proj", docs("/x/proj"), nil, func() time.Time { return now }, []work.Item{{Name: "doing", Done: 1, Total: 2, Phase: work.InProgress, Modified: now}}, nil, nil, nil, nil, [columns]bool{}, "", "test")
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

// fakeEditor records the editor's arguments and runs edit (if any) before reporting success.
func fakeEditor(t *testing.T, edit func()) *[]string {
	t.Helper()
	var args []string
	old := editor.Exec
	editor.Exec = func(c *exec.Cmd, done tea.ExecCallback) tea.Cmd {
		args = c.Args
		return func() tea.Msg {
			if edit != nil {
				edit()
			}
			return done(nil)
		}
	}
	t.Cleanup(func() { editor.Exec = old })
	t.Setenv("EDITOR", "ed")
	return &args
}

func TestEditTasksFromBoard(t *testing.T) {
	proj := t.TempDir()
	tasks := filepath.Join(proj, "openspec", "changes", "m1-board-view", "tasks.md")
	if err := os.MkdirAll(filepath.Dir(tasks), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tasks, []byte("- [ ] 1.1 a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	changes := fixture
	loads := 0
	m := New(proj, docs(proj), func() ([]work.Item, error) { loads++; return changes, nil }, func() time.Time { return now }, nil, nil, nil, nil, nil, [columns]bool{}, "", "test")
	m = send(m, tea.WindowSizeMsg{Width: 120, Height: 30})
	m = keys(send(m, m.Init()()), "3") // In Progress: m1-board-view 5/11
	args := fakeEditor(t, func() {
		changes = append([]work.Item{{Name: "m1-board-view", Done: 6, Total: 11, Phase: work.InProgress, Modified: now}}, fixture[:4]...)
	})
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	if cmd == nil {
		t.Fatal("e must start the editor")
	}
	m = next.(Model)
	next, cmd = m.Update(cmd()) // editor exits → reload
	if cmd == nil || !slices.Equal(*args, []string{"ed", tasks}) {
		t.Fatalf("editor %q, want ed %s; and a reload must follow", *args, tasks)
	}
	m = send(next.(Model), cmd())
	if loads != 2 || !strings.Contains(m.View(), "6/11") || m.cols[2][m.nav.sel[2]].Name != "m1-board-view" {
		t.Errorf("after editing: %d loads, view:\n%s", loads, m.View())
	}
}

func TestEditRefusedOnBoard(t *testing.T) {
	args := fakeEditor(t, nil)
	m := keys(board(t, 120, 30, fixture), "1") // Draft without tasks.md on disk
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	if cmd != nil || *args != nil || !strings.Contains(next.(Model).View(), "no tasks to edit") {
		t.Errorf("Draft without tasks: cmd %v, view:\n%s", cmd != nil, next.(Model).View())
	}
	m = keys(board(t, 120, 30, fixture), "5")
	next, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	if cmd != nil || !strings.Contains(next.(Model).View(), "archived changes are read-only") {
		t.Errorf("archived: cmd %v, view:\n%s", cmd != nil, next.(Model).View())
	}
}

func TestEditDuringLoadReloadsAgain(t *testing.T) {
	m := board(t, 120, 30, fixture)
	m, _ = m.refresh() // a load is running
	next, cmd := m.Update(editedMsg{})
	if cmd != nil {
		t.Fatal("no second load while one runs")
	}
	next, cmd = next.(Model).Update(loadedMsg{changes: fixture})
	if cmd == nil {
		t.Error("the load that raced the edit must be followed by another one")
	}
}

func TestDetailEditedAfterCloseStillReloads(t *testing.T) {
	m := board(t, 120, 30, fixture)
	if _, cmd := m.Update(detail.EditedMsg{}); cmd == nil {
		t.Error("an edit reported after the detail closed must still reload the Board")
	}
}

func TestNoticeFitsWidth(t *testing.T) {
	m := send(board(t, 60, 30, fixture), editedMsg{err: errors.New(`exec: "nvimx": executable file not found in $PATH`)})
	for _, l := range strings.Split(m.View(), "\n") {
		if lipgloss.Width(l) > 60 {
			t.Errorf("line wider than 60: %q", l)
		}
	}
}

func TestDetailHeaderFollowsReload(t *testing.T) {
	m, _ := enter(keys(board(t, 120, 30, fixture), "3")) // m1-board-view 5/11
	fresh := append([]work.Item{{Name: "m1-board-view", Done: 6, Total: 11, Phase: work.InProgress, Modified: now}}, fixture[:4]...)
	m = send(m, loadedMsg{changes: fresh})
	if !strings.HasPrefix(m.View(), "m1-board-view  M1  6/11  In Progress") {
		t.Errorf("detail header after reload:\n%s", m.View())
	}
}

// run executes cmd and sends its messages to m, the parts of a batch in order.
func run(m Model, cmd tea.Cmd) Model {
	if cmd == nil {
		return m
	}
	msg := cmd()
	if b, ok := msg.(tea.BatchMsg); ok {
		for _, c := range b {
			m = run(m, c)
		}
		return m
	}
	return send(m, msg)
}

// agents is a lookup reporting the given Changes as running and counting its calls.
func agents(calls *int, names ...string) func() (map[string]bool, error) {
	return func() (map[string]bool, error) {
		*calls++
		running := map[string]bool{}
		for _, n := range names {
			running[n] = true
		}
		return running, nil
	}
}

func TestRunningBadge(t *testing.T) {
	calls := 0
	m := New("/x/proj", docs("/x/proj"), func() ([]work.Item, error) { return fixture, nil }, func() time.Time { return now }, nil,
		agents(&calls, "m1-board-view", "m0-plugin-skeleton", "not-a-change"), nil, nil, nil, [columns]bool{}, "", "test")
	m = run(send(m, tea.WindowSizeMsg{Width: 200, Height: 30}), m.Init())
	v := m.View()
	if calls != 1 || !strings.Contains(v, "m1-board-view") {
		t.Fatalf("Init must load Changes and agents: %d lookups, view:\n%s", calls, v)
	}
	if !strings.Contains(v, "● m1-board-view") || strings.Count(v, "●") != 1 {
		t.Errorf("want one badge, on m1-board-view and not on the Archived Card:\n%s", v)
	}
}

func TestRunningBadgeLookupError(t *testing.T) {
	m := board(t, 120, 30, fixture)
	want := m.View()
	m = send(m, agentsMsg{err: errors.New("herdr not running")})
	if got := m.View(); got != want {
		t.Errorf("a failed lookup must change nothing:\n%s", got)
	}
}

func TestRunningBadgeSurvivesReloads(t *testing.T) {
	calls := 0
	fresh := append([]work.Item{{Name: "m3-new", Done: 1, Total: 2, Phase: work.InProgress, Modified: now}}, fixture...)
	m := New("/x/proj", docs("/x/proj"), func() ([]work.Item, error) { return fresh, nil }, func() time.Time { return now }, fixture,
		agents(&calls, "m1-board-view"), nil, nil, nil, [columns]bool{}, "", "test")
	m = send(m, tea.WindowSizeMsg{Width: 120, Height: 30})
	if v := m.View(); !strings.Contains(v, "m1-board-view") || strings.Contains(v, "●") {
		t.Fatalf("the snapshot must show before the lookup answers, without badges:\n%s", v)
	}
	cmds := m.Init()().(tea.BatchMsg) // load, lookup
	m = send(m, cmds[1]())
	m = send(m, cmds[0]()) // cached → fresh
	if v := m.View(); !strings.Contains(v, "m3-new") || !strings.Contains(v, "● m1-board-view") {
		t.Errorf("badge lost in the swap to the fresh load:\n%s", v)
	}
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	m = run(next.(Model), cmd)
	if calls != 2 || !strings.Contains(m.View(), "● m1-board-view") {
		t.Errorf("r must look agents up again (%d lookups) and keep the badge:\n%s", calls, m.View())
	}
}

// starter fakes the start func: it records the Changes it is called with and returns err.
func starter(m *Model, err error) *[]string {
	var started []string
	m.start = func(change string) error {
		started = append(started, change)
		return err
	}
	return &started
}

func pressKey(m Model, k string) (Model, tea.Cmd) {
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)})
	return next.(Model), cmd
}

func TestStart(t *testing.T) {
	for col, name := range map[string]string{"2": "m1-phase-derivation", "3": "m1-board-view"} {
		m := keys(board(t, 120, 30, fixture), col)
		started := starter(&m, nil)
		m, cmd := pressKey(m, "s")
		if cmd == nil || len(*started) != 0 || !strings.Contains(m.View(), "starting "+name+"…") {
			t.Fatalf("s must start %s in a command and say so: started %q, view:\n%s", name, *started, m.View())
		}
		if _, cmd = m.Update(cmd()); !slices.Equal(*started, []string{name}) || cmd == nil || cmd() != tea.Quit() {
			t.Errorf("start %q, want %s, then quit", *started, name)
		}
	}
}

func TestStartFails(t *testing.T) {
	m := keys(board(t, 120, 30, fixture), "2")
	starter(&m, errors.New("claude did not become ready in tab w1:t3"))
	m, cmd := pressKey(m, "s")
	next, quit := m.Update(cmd())
	m = next.(Model)
	if quit != nil || !strings.Contains(m.View(), "start failed: claude did not become ready in tab w1:t3") {
		t.Fatalf("a failed Start must keep the Board and show the error:\n%s", m.View())
	}
	if m = keys(m, "l"); m.nav.col != 2 {
		t.Error("keys must work again after a failed Start")
	}
}

func TestStartFailsLooksUpAgents(t *testing.T) {
	m := keys(board(t, 120, 30, fixture), "2")
	starter(&m, errors.New("prompt not sent"))
	m.agents = func() (map[string]bool, error) { return map[string]bool{"m1-board-view": true}, nil }
	m, cmd := pressKey(m, "s")
	_, cmd = m.Update(cmd())
	if cmd == nil {
		t.Fatal("a failed Start may leave an agent behind: its badge must be looked up again")
	}
	if _, ok := cmd().(agentsMsg); !ok {
		t.Error("want an agent lookup after a failed Start")
	}
}

// resumer fakes the resumable and resume funcs: a record of the given kind, reopenable or not.
func resumer(m *Model, kind string, ok bool, err error) *[]string {
	var resumed []string
	m.resumable = func(string) (string, bool) { return kind, ok }
	m.resume = func(change string) error {
		resumed = append(resumed, change)
		return err
	}
	return &resumed
}

func TestResume(t *testing.T) {
	for _, kind := range []string{"claude", "codex"} {
		m := keys(board(t, 120, 30, fixture), "4")
		started := starter(&m, nil)
		resumed := resumer(&m, kind, true, nil)
		m, cmd := pressKey(m, "s")
		if cmd == nil || len(*resumed) != 0 || !strings.Contains(m.View(), "starting m0-spike-agent-start…") {
			t.Fatalf("%s: s on Review must resume in a command and say so: view:\n%s", kind, m.View())
		}
		if _, cmd = m.Update(cmd()); !slices.Equal(*resumed, []string{"m0-spike-agent-start"}) || len(*started) != 0 || cmd == nil || cmd() != tea.Quit() {
			t.Errorf("%s: resumed %q, started %q, want resume then quit", kind, *resumed, *started)
		}
	}
}

func TestResumeRunning(t *testing.T) {
	m := keys(board(t, 120, 30, fixture), "4")
	m.running = map[string]bool{"m0-spike-agent-start": true}
	resumed := resumer(&m, "", false, nil)
	m, cmd := pressKey(m, "s")
	if cmd == nil {
		t.Fatalf("a Running badge must resume even without a record:\n%s", m.View())
	}
	m.Update(cmd())
	if !slices.Equal(*resumed, []string{"m0-spike-agent-start"}) {
		t.Errorf("resumed %q", *resumed)
	}
}

func TestResumeFails(t *testing.T) {
	m := keys(board(t, 120, 30, fixture), "4")
	resumer(&m, "claude", true, errors.New("no session"))
	m, cmd := pressKey(m, "s")
	next, quit := m.Update(cmd())
	if m = next.(Model); quit != nil || !strings.Contains(m.View(), "start failed: no session") {
		t.Errorf("a failed resume must keep the Board and show the error:\n%s", m.View())
	}
}

func TestResumeRefused(t *testing.T) {
	for kind, notice := range map[string]string{
		"":   "all tasks of m0-spike-agent-start are done",
		"pi": "resume not supported for pi",
	} {
		m := keys(board(t, 120, 30, fixture), "4")
		resumed := resumer(&m, kind, false, nil)
		if m, cmd := pressKey(m, "s"); cmd != nil || len(*resumed) != 0 || !strings.Contains(m.View(), notice) {
			t.Errorf("kind %q: cmd %v, resumed %q, view:\n%s", kind, cmd != nil, *resumed, m.View())
		}
	}
}

func TestResumeOnlyInReview(t *testing.T) {
	for col, notice := range map[string]string{"1": "is a draft", "5": "archived changes cannot be started"} {
		m := keys(board(t, 120, 30, fixture), col)
		resumed := resumer(&m, "claude", true, nil)
		if m, cmd := pressKey(m, "s"); cmd != nil || len(*resumed) != 0 || !strings.Contains(m.View(), notice) {
			t.Errorf("column %s: cmd %v, resumed %q, view:\n%s", col, cmd != nil, *resumed, m.View())
		}
	}
}

func TestStartWithoutStarter(t *testing.T) {
	m := keys(board(t, 120, 30, fixture), "2")
	if m, cmd := pressKey(m, "s"); cmd != nil || m.starting != "" {
		t.Error("a Board without a start func must ignore s")
	}
}

func TestStartRefused(t *testing.T) {
	for col, notice := range map[string]string{
		"1": "m2-change-detail is a draft: plan it first",
		"4": "all tasks of m0-spike-agent-start are done",
		"5": "archived changes cannot be started",
	} {
		m := keys(board(t, 120, 30, fixture), col)
		started := starter(&m, nil)
		m, cmd := pressKey(m, "s")
		if cmd != nil || len(*started) != 0 || !strings.Contains(m.View(), notice) {
			t.Errorf("column %s: cmd %v, started %q, view:\n%s", col, cmd != nil, *started, m.View())
		}
	}
	m := board(t, 120, 30, nil)
	started := starter(&m, nil)
	if m, cmd := pressKey(m, "s"); cmd != nil || len(*started) != 0 || m.starting != "" || m.notice != "" {
		t.Error("s in an empty column must do nothing")
	}
}

func TestKeysWhileStarting(t *testing.T) {
	m := keys(board(t, 120, 30, tall(3)), "2")
	started := starter(&m, nil)
	m, _ = pressKey(m, "s")
	for _, k := range []string{"q", "j", "s"} {
		var cmd tea.Cmd
		if m, cmd = pressKey(m, k); cmd != nil || m.nav.sel[1] != 0 || m.starting != "change-00" {
			t.Errorf("%s while starting: cmd %v, sel %d, starting %q", k, cmd != nil, m.nav.sel[1], m.starting)
		}
	}
	if _, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC}); cmd == nil || cmd() != tea.Quit() || len(*started) != 0 {
		t.Error("ctrl+c must quit while starting")
	}
}

func TestStartNotInDetail(t *testing.T) {
	m, _ := enter(keys(board(t, 120, 30, fixture), "3"))
	started := starter(&m, nil)
	if m, cmd := pressKey(m, "s"); cmd != nil || m.starting != "" || len(*started) != 0 {
		t.Error("s in the detail must do nothing")
	}
}

// shown lists the names of the Cards on the Board, column by column.
func shown(m Model) []string {
	var got []string
	for _, col := range m.cols {
		for _, c := range col {
			got = append(got, c.Name)
		}
	}
	return got
}

func TestFilterPrompt(t *testing.T) {
	m := board(t, 120, 30, fixture)
	progress := "15/31 48%"
	if !strings.Contains(m.View(), progress) {
		t.Fatalf("fixture Progress changed:\n%s", m.View())
	}
	m = keys(m, "/")
	if !m.prompting || !strings.HasSuffix(m.View(), "\n/█") {
		t.Fatalf("/ must open the prompt in the footer:\n%s", m.View())
	}
	m = keys(m, "b", "o", "a", "r", "d")
	v := m.View()
	if !slices.Equal(shown(m), []string{"m1-board-view"}) || !strings.Contains(v, "In Progress 1") || !strings.Contains(v, "Draft 0") {
		t.Errorf("typing must filter columns and counts at once: %q\n%s", shown(m), v)
	}
	if !strings.Contains(v, "spec-kanban · proj · /board") || !strings.Contains(v, progress) || !strings.HasSuffix(v, "\n/board█") {
		t.Errorf("header must show the filter and keep the Progress of all Changes:\n%s", v)
	}
	if m = keys(m, "esc"); m.prompting || m.filter != "" || len(shown(m)) != len(fixture) || strings.Contains(m.View(), "· /") {
		t.Errorf("Esc in the prompt must close it and clear the filter: %q\n%s", m.filter, m.View())
	}
}

func TestFilterCaseAndMilestone(t *testing.T) {
	if got := shown(keys(board(t, 120, 30, fixture), "/", "B", "O", "A", "R", "D")); !slices.Equal(got, []string{"m1-board-view"}) {
		t.Errorf("/BOARD shows %q", got)
	}
	if got := shown(keys(board(t, 120, 30, fixture), "/", "m", "1")); !slices.Equal(got, []string{"m1-phase-derivation", "m1-board-view"}) {
		t.Errorf("/m1 shows %q", got)
	}
}

func TestPromptTypesBoardKeys(t *testing.T) {
	m := keys(board(t, 120, 30, fixture), "2", "/")
	started := starter(&m, nil)
	args := fakeEditor(t, nil)
	for _, k := range []string{"j", "q", "n", "N", "s", "e", "r", "g", "G", "h", "l", "1", "/", " "} {
		var cmd tea.Cmd
		if m, cmd = pressKey(m, k); cmd != nil {
			t.Errorf("%q in the prompt returned a command", k)
		}
	}
	if m.filter != "jqnNsergGhl1/ " || !m.prompting || len(*started) != 0 || *args != nil || m.loading {
		t.Errorf("keys in the prompt must be text: filter %q, prompting %v", m.filter, m.prompting)
	}
}

func TestPromptPasteKeepsPrintable(t *testing.T) {
	m := keys(board(t, 120, 30, fixture), "/")
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m1\n\x03"), Paste: true})
	if m = next.(Model); m.filter != "m1" {
		t.Errorf("a paste must only add printable characters: %q", m.filter)
	}
}

func TestHeaderFitsLongFilter(t *testing.T) {
	m := keys(board(t, 60, 30, fixture), "/", "enter")
	m.filter = strings.Repeat("x", 80)
	header := strings.Split(m.View(), "\n")[0]
	if lipgloss.Width(header) > 60 || !strings.HasSuffix(header, "15/31 48%") {
		t.Errorf("header must fit the width and keep the Progress: %q", header)
	}
}

func TestPromptBackspace(t *testing.T) {
	m := keys(board(t, 120, 30, fixture), "/", "m", "2")
	if m = keys(m, "backspace"); m.filter != "m" || !m.prompting {
		t.Fatalf("backspace must delete a character: %q", m.filter)
	}
	if m = keys(m, "backspace"); m.filter != "" || !m.prompting || len(shown(m)) != len(fixture) {
		t.Fatalf("backspace to empty: %q, prompting %v", m.filter, m.prompting)
	}
	if m = keys(m, "backspace"); m.prompting {
		t.Error("backspace on an empty prompt must close it")
	}
}

func TestPromptEnterKeepsFilter(t *testing.T) {
	m := keys(board(t, 120, 30, fixture), "/", "m", "1", "enter")
	if m.prompting || m.filter != "m1" || m.detail != nil {
		t.Fatalf("Enter must close the prompt, keep the filter and open no detail: prompting %v, filter %q", m.prompting, m.filter)
	}
	if v := m.View(); !strings.Contains(v, "· /m1") || strings.Contains(v, "/m1█") {
		t.Errorf("header keeps the filter, footer shows the keys again:\n%s", v)
	}
	if m = keys(m, "l", "enter"); m.detail == nil || !strings.HasPrefix(m.View(), "m1-board-view") {
		t.Errorf("Enter on the filtered Board opens the selected Card:\n%s", m.View())
	}
}

func TestPromptCtrlCQuits(t *testing.T) {
	m := keys(board(t, 120, 30, fixture), "/", "x")
	if _, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC}); cmd == nil || cmd() != tea.Quit() {
		t.Error("ctrl+c must quit from the prompt")
	}
}

func TestFilterMovesFocus(t *testing.T) {
	m := keys(board(t, 120, 30, fixture), "/", "b", "o", "a", "r", "d")
	if m.nav.col != 2 || m.cols[2][m.nav.sel[2]].Name != "m1-board-view" {
		t.Errorf("focus %d, want In Progress with m1-board-view", m.nav.col)
	}
	m = keys(m, "x")
	if v := m.View(); !strings.Contains(v, "· /boardx") || strings.Contains(v, "▌") || strings.Count(v, "nothing here") != columns {
		t.Errorf("nothing matches: every column empty, no selection, filter in the header:\n%s", v)
	}
}

func TestFilterKeepsSelectionByName(t *testing.T) {
	m := keys(board(t, 120, 30, tall(6)), "2", "j", "j", "j", "/", "3")
	if m.nav.sel[1] != 0 || m.cols[1][0].Name != "change-03" {
		t.Fatalf("filtered: sel %d", m.nav.sel[1])
	}
	if m = keys(m, "backspace"); m.cols[1][m.nav.sel[1]].Name != "change-03" {
		t.Errorf("after clearing: selected %q, want change-03", m.cols[1][m.nav.sel[1]].Name)
	}
}

func TestFilterSurvivesReload(t *testing.T) {
	fresh := append([]work.Item{{Name: "m1-new", Done: 1, Total: 2, Phase: work.InProgress, Modified: now}}, fixture...)
	m := New("/x/proj", docs("/x/proj"), func() ([]work.Item, error) { return fresh, nil }, func() time.Time { return now }, fixture, nil, nil, nil, nil, [columns]bool{}, "", "test")
	m = keys(send(m, tea.WindowSizeMsg{Width: 120, Height: 30}), "/", "m", "1", "enter")
	m = send(m, m.Init()()) // cached → fresh
	want := []string{"m1-phase-derivation", "m1-new", "m1-board-view"}
	if got := shown(m); !slices.Equal(got, want) {
		t.Fatalf("after the swap %q, want %q", got, want)
	}
	m, cmd := pressKey(m, "r")
	if m = send(m, cmd()); !slices.Equal(shown(m), want) || !strings.Contains(m.View(), "· /m1") {
		t.Errorf("after r %q:\n%s", shown(m), m.View())
	}
}

func TestFilteredDetailHeaderFollowsReload(t *testing.T) {
	m := keys(board(t, 120, 30, fixture), "/", "b", "o", "a", "r", "d", "enter", "enter")
	fresh := append([]work.Item{{Name: "m1-board-view", Done: 6, Total: 11, Phase: work.InProgress, Modified: now}}, fixture[:4]...)
	m = send(m, loadedMsg{changes: fresh})
	if !strings.HasPrefix(m.View(), "m1-board-view  M1  6/11  In Progress") {
		t.Errorf("detail header after reload:\n%s", m.View())
	}
}

func TestEscClearsFilter(t *testing.T) {
	m := keys(board(t, 120, 30, fixture), "/", "m", "1", "enter")
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m = next.(Model); cmd != nil || m.filter != "" || len(shown(m)) != len(fixture) || strings.Contains(m.View(), "· /") {
		t.Fatalf("Esc with a filter must clear it and keep the Board open:\n%s", m.View())
	}
	if _, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEsc}); cmd == nil || cmd() != tea.Quit() {
		t.Error("Esc without a filter must close the Board")
	}
}

func TestMatchKeys(t *testing.T) {
	m := keys(board(t, 120, 30, fixture), "2")
	if m = keys(m, "n"); m.nav.col != 1 || m.nav.sel[1] != 0 {
		t.Errorf("n without a filter moved to col %d sel %d", m.nav.col, m.nav.sel[1])
	}
	if m = keys(m, "N"); m.nav.col != 1 || m.nav.sel[1] != 0 {
		t.Errorf("N without a filter moved to col %d sel %d", m.nav.col, m.nav.sel[1])
	}
	m = keys(m, "/", "m", "1", "enter") // Ready: m1-phase-derivation, In Progress: m1-board-view
	if m = keys(m, "n"); m.nav.col != 2 {
		t.Errorf("n: col %d, want In Progress", m.nav.col)
	}
	if m = keys(m, "n"); m.nav.col != 1 {
		t.Errorf("n past the last match: col %d, want Ready", m.nav.col)
	}
	if m = keys(m, "N"); m.nav.col != 2 {
		t.Errorf("N before the first match: col %d, want In Progress", m.nav.col)
	}
}

func TestFilterNarrow(t *testing.T) {
	m := keys(board(t, 60, 30, fixture), "/", "m", "0", "enter")
	v := m.View()
	if !strings.Contains(v, "◀ 4/5 Review 1 ▶") || !strings.HasSuffix(v, "q close") {
		t.Errorf("narrow filtered Board:\n%s", v)
	}
	for _, l := range strings.Split(v, "\n") {
		if lipgloss.Width(l) > 60 {
			t.Errorf("line wider than 60: %q", l)
		}
	}
}

func TestHiddenArchived(t *testing.T) {
	m := hiding(t, 120, 30, fixture, work.Archived)
	golden(t, "board-hidden", m.View())
	if m = keys(m, "5"); m.nav.col != 0 {
		t.Errorf("5 on hidden Archived moved focus to %d", m.nav.col)
	}
	if m = keys(m, "4", "l"); m.nav.col != 3 {
		t.Errorf("l from Review moved focus to %d", m.nav.col)
	}
}

func TestHiddenDraftKeepsProgress(t *testing.T) {
	changes := []work.Item{
		{Name: "draft", Total: 4, Modified: now, Phase: work.Draft}, // a Draft with tasks
		{Name: "ready", Total: 2, Modified: now, Phase: work.Ready},
	}
	m := hiding(t, 120, 30, changes, work.Draft)
	if v := m.View(); strings.Contains(v, "Draft") || !strings.Contains(v, "0/6 0%") {
		t.Errorf("hidden Draft must not be drawn but count in the Progress:\n%s", v)
	}
	if m = keys(m, "h"); m.nav.col != 1 {
		t.Errorf("h from Ready moved focus to %d", m.nav.col)
	}
}

func TestHiddenFocusFallback(t *testing.T) {
	// only Draft and Archived have Cards; Draft is hidden, so focus goes to Archived, not Draft
	changes := []work.Item{{Name: "draft", Modified: now}, {Name: "old", Phase: work.Archived}}
	m := hiding(t, 120, 30, changes, work.Draft)
	if m.nav.col != 4 {
		t.Errorf("focus %d after load, want Archived (4)", m.nav.col)
	}
	m = keys(m, "/", "o", "enter")
	if len(m.cols[0]) != 0 || m.nav.col != 4 {
		t.Errorf("filter must keep the hidden Draft empty: %q, focus %d", shown(m), m.nav.col)
	}
	m, cmd := pressKey(keys(m, "esc"), "r")
	if m = send(m, cmd()); len(m.cols[0]) != 0 || m.nav.col != 4 {
		t.Errorf("reload must keep the hidden Draft empty: %q, focus %d", shown(m), m.nav.col)
	}
}

func TestHiddenNarrow(t *testing.T) {
	m := hiding(t, 60, 30, fixture, work.Draft, work.Archived)
	for _, k := range []string{"h", "1", "l", "l", "l", "5"} {
		if m = keys(m, k); strings.Contains(m.View(), "Draft") || strings.Contains(m.View(), "Archived") {
			t.Errorf("after %s a hidden column is shown:\n%s", k, m.View())
		}
	}
	// nothing shown has Cards: focus falls back to Draft, the view shows the first shown column
	m = keys(hiding(t, 60, 30, nil, work.Draft), "/", "x")
	if v := m.View(); strings.Contains(v, "Draft") || !strings.Contains(v, "◀ 2/5 Ready 0 ▶") {
		t.Errorf("empty Board with hidden Draft:\n%s", v)
	}
}

func TestInitialNotice(t *testing.T) {
	m := New("/x/proj", docs("/x/proj"), func() ([]work.Item, error) { return fixture, nil }, func() time.Time { return now }, nil, nil, nil, nil, nil,
		[columns]bool{}, "config.json: unknown key promt", "test")
	m = send(send(m, tea.WindowSizeMsg{Width: 120, Height: 30}), m.Init()())
	if !strings.HasSuffix(m.View(), "\nconfig.json: unknown key promt") {
		t.Fatalf("the initial notice must replace the footer:\n%s", m.View())
	}
	if m = keys(m, "j"); strings.Contains(m.View(), "promt") {
		t.Error("the next key must clear the initial notice")
	}
}

const helpTitle = "spec-kanban test · any key closes"

func TestHelp(t *testing.T) {
	framed := keys(board(t, 120, 30, fixture), "3", "?").View()
	golden(t, "help-120", framed)
	lines := strings.Split(framed, "\n")
	if len(lines) > 30 || !strings.HasPrefix(lines[0], "╭") || !strings.HasPrefix(lines[1], "│"+helpTitle) || !strings.HasPrefix(lines[len(lines)-1], "╰") {
		t.Errorf("help in a 120x30 popup must be framed and fit:\n%s", framed)
	}
	for _, g := range []string{"│Board", "│Filter prompt", "│Details"} {
		if !strings.Contains(framed, g) {
			t.Errorf("help lacks group %q", g)
		}
	}
	low := keys(board(t, 120, 20, fixture), "?").View()
	golden(t, "help-low", low)
	if lines := strings.Count(low, "\n") + 1; lines > 20 || !strings.HasPrefix(low, helpTitle) || strings.Contains(low, "╭") {
		t.Errorf("help in a 120x20 popup: unframed and cut, %d lines:\n%s", lines, low)
	}
	short := keys(board(t, 60, 14, fixture), "?").View()
	if lines := strings.Count(short, "\n") + 1; lines > 14 || !strings.HasPrefix(short, helpTitle) {
		t.Errorf("help in a 60x14 popup: %d lines:\n%s", lines, short)
	}
	for _, l := range strings.Split(short, "\n") {
		if lipgloss.Width(l) > 60 {
			t.Errorf("line wider than 60: %q", l)
		}
	}
}

func TestHelpClosesWithoutEffect(t *testing.T) {
	want := keys(board(t, 120, 30, tall(3)), "2").View()
	for _, k := range []tea.KeyMsg{{Type: tea.KeyRunes, Runes: []rune("q")}, {Type: tea.KeyRunes, Runes: []rune("j")},
		{Type: tea.KeyRunes, Runes: []rune("s")}, {Type: tea.KeyEsc}} {
		m := keys(board(t, 120, 30, tall(3)), "2", "?")
		started := starter(&m, nil)
		next, cmd := m.Update(k)
		if m = next.(Model); cmd != nil || len(*started) != 0 || m.View() != want {
			t.Errorf("%s must only close the help: cmd %v, started %q, view:\n%s", k, cmd != nil, *started, m.View())
		}
	}
	m := keys(board(t, 120, 30, fixture), "?")
	if _, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC}); cmd == nil || cmd() != tea.Quit() {
		t.Error("ctrl+c must quit from the help")
	}
}

func TestHelpInDetail(t *testing.T) {
	proj := t.TempDir()
	proposal := filepath.Join(proj, "openspec", "changes", "m1-board-view", "proposal.md")
	if err := os.MkdirAll(filepath.Dir(proposal), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(proposal, []byte(strings.Repeat("line\n\n", 40)), 0o644); err != nil {
		t.Fatal(err)
	}
	m := New(proj, docs(proj), func() ([]work.Item, error) { return fixture, nil }, func() time.Time { return now }, nil, nil, nil, nil, nil, [columns]bool{}, "", "test")
	m = send(m, tea.WindowSizeMsg{Width: 120, Height: 30})
	m, _ = enter(keys(send(m, m.Init()()), "3"))
	top := keys(m, "2").View()
	m = keys(m, "2", "j")
	want := m.View()
	if want == top {
		t.Fatal("the proposal must scroll")
	}
	if m = keys(m, "?"); m.View() != keys(board(t, 120, 30, fixture), "?").View() {
		t.Fatalf("? in the detail must show the help:\n%s", m.View())
	}
	if m = keys(m, "j"); m.View() != want {
		t.Errorf("the next key must return to the detail at the same position:\n%s", m.View())
	}
	if m = keys(m, "g", "?", "x", "g"); m.View() != want {
		t.Errorf("g, ?, x, g must not jump: the help cancels the pending g:\n%s", m.View())
	}
}

func TestHelpKeyInPromptAndWhileStarting(t *testing.T) {
	if m := keys(board(t, 120, 30, fixture), "/", "?"); m.filter != "?" || strings.Contains(m.View(), helpTitle) {
		t.Errorf("? in the prompt must be typed: filter %q", m.filter)
	}
	m := keys(board(t, 120, 30, fixture), "2")
	starter(&m, nil)
	m, _ = pressKey(m, "s")
	if m = keys(m, "?"); strings.Contains(m.View(), helpTitle) || m.starting == "" {
		t.Errorf("? while starting must be ignored:\n%s", m.View())
	}
}
