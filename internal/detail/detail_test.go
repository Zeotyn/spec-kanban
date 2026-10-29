package detail

import (
	"flag"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/zeotyn/openspec-kanban/internal/openspec"
)

var update = flag.Bool("update", false, "rewrite golden files")

func TestMain(m *testing.M) {
	Style = "notty"
	lipgloss.SetColorProfile(termenv.Ascii)
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

// project writes a Change "m1-board-view" with tasks, proposal and nested specs, but no design.
func project(t *testing.T) string {
	t.Helper()
	p := t.TempDir()
	files := map[string]string{
		"tasks.md": "## 1. Model\n\n- [x] 1.1 Loading\n- [ ] 1.2 Navigation\n\n## 2. View\n\n- [ ] 2.1 Columns\n",
		"proposal.md": "## Why\n\n" + strings.Repeat("A line of the proposal that says why.\n\n", 30) +
			"See https://example.com/a/very/long/url/that/glamour/does/not/wrap/because/it/has/no/spaces/at/all\n\n" +
			"```go\nfmt.Println(\"a code line that is far longer than any narrow popup is wide, and glamour keeps it\")\n```\n",
		"specs/board/spec.md":                 "## ADDED Requirements\n\n### Requirement: Columns\nThe Board SHALL show columns.\n",
		"specs/agent-control/session/spec.md": "## ADDED Requirements\n\n### Requirement: Wait\nSessions SHALL wait.\n",
	}
	for name, content := range files {
		path := filepath.Join(p, "openspec", "changes", "m1-board-view", name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return p
}

var change = openspec.Change{Name: "m1-board-view", Done: 1, Total: 3}

func open(t *testing.T, c openspec.Change, w, h int) Model {
	return New(project(t), c, "m1-board-view  M1  1/3  In Progress", w, h)
}

func press(m Model, keys ...string) Model {
	for _, k := range keys {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
		switch k {
		case "tab":
			msg = tea.KeyMsg{Type: tea.KeyTab}
		case "shift+tab":
			msg = tea.KeyMsg{Type: tea.KeyShiftTab}
		case "ctrl+d":
			msg = tea.KeyMsg{Type: tea.KeyCtrlD}
		case "esc":
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		}
		m, _ = m.Update(msg)
	}
	return m
}

func TestDefaultTab(t *testing.T) {
	if m := open(t, change, 80, 24); m.tab != tasks {
		t.Errorf("with tasks: tab %d, want Tasks", m.tab)
	}
	if m := open(t, openspec.Change{Name: "m1-board-view"}, 80, 24); m.tab != proposal {
		t.Errorf("without tasks: tab %d, want Proposal", m.tab)
	}
}

func TestTabs(t *testing.T) {
	m := open(t, change, 80, 24)
	var seen []int
	for range 4 {
		m = press(m, "tab")
		seen = append(seen, m.tab)
	}
	if want := []int{proposal, design, specs, tasks}; !slices.Equal(seen, want) {
		t.Errorf("tab cycles %v, want %v", seen, want)
	}
	if m = press(m, "shift+tab"); m.tab != specs {
		t.Errorf("shift+tab from Tasks: %d, want Specs", m.tab)
	}
	if m = press(m, "2"); m.tab != proposal {
		t.Errorf("2: %d, want Proposal", m.tab)
	}
}

func TestScrollKeepsPerTabOffset(t *testing.T) {
	m := press(open(t, change, 80, 24), "2", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j")
	if m.off[proposal] != 10 {
		t.Fatalf("offset %d, want 10", m.off[proposal])
	}
	m = press(m, "3", "2")
	if m.off[proposal] != 10 {
		t.Errorf("offset after switching back %d, want 10", m.off[proposal])
	}
	if m = press(m, "G"); m.off[proposal] != m.maxOff(proposal) || m.maxOff(proposal) == 0 {
		t.Errorf("G: offset %d, max %d", m.off[proposal], m.maxOff(proposal))
	}
	if m = press(m, "j"); m.off[proposal] != m.maxOff(proposal) {
		t.Error("j at the bottom must stop")
	}
	if m = press(m, "g", "g"); m.off[proposal] != 0 {
		t.Errorf("gg: offset %d, want 0", m.off[proposal])
	}
	if m = press(m, "k"); m.off[proposal] != 0 {
		t.Error("k at the top must stop")
	}
}

func TestClose(t *testing.T) {
	for _, k := range []tea.KeyMsg{{Type: tea.KeyRunes, Runes: []rune("q")}, {Type: tea.KeyEsc}} {
		_, cmd := open(t, change, 80, 24).Update(k)
		if cmd == nil {
			t.Fatalf("%s: no command", k)
		}
		if _, ok := cmd().(CloseMsg); !ok {
			t.Errorf("%s: want CloseMsg", k)
		}
	}
	if _, cmd := open(t, change, 80, 24).Update(tea.KeyMsg{Type: tea.KeyCtrlC}); cmd == nil || cmd() != tea.Quit() {
		t.Error("ctrl+c must quit")
	}
}

func TestResizeRerenders(t *testing.T) {
	m := open(t, change, 120, 24)
	wide := len(m.lines[proposal])
	m, _ = m.Update(tea.WindowSizeMsg{Width: 40, Height: 24})
	if len(m.lines[proposal]) <= wide {
		t.Errorf("narrower width must wrap into more lines: %d → %d", wide, len(m.lines[proposal]))
	}
}

func TestViewFitsHeight(t *testing.T) {
	for _, w := range []int{30, 80} {
		for _, h := range []int{10, 24, 40} {
			for _, k := range []string{"1", "2", "3", "4"} {
				m := press(open(t, change, w, h), k, "G") // the bottom holds the unwrappable lines
				if lines := strings.Count(m.View(), "\n") + 1; lines > h {
					t.Errorf("%dx%d tab %s: %d lines", w, h, k, lines)
				}
				for _, l := range strings.Split(m.View(), "\n") {
					if lipgloss.Width(l) > w {
						t.Errorf("%dx%d tab %s: line wider than the view: %q", w, h, k, l)
						break
					}
				}
			}
		}
	}
}

func TestViews(t *testing.T) {
	golden(t, "tasks", open(t, change, 80, 20).View())
	golden(t, "specs", press(open(t, change, 80, 20), "4").View())
	golden(t, "missing-design", press(open(t, change, 80, 20), "3").View())
	golden(t, "narrow", press(open(t, change, 40, 20), "2").View())
}
