package board

import (
	"errors"
	"slices"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/zeotyn/spec-kanban/internal/herdr"
	"github.com/zeotyn/spec-kanban/internal/work"
)

// texter records what start gets: the Change and S's line, "<nil>" for s.
func texter(m *Model, err error) *[]string {
	var got []string
	m.start = func(change string, text *string) error {
		t := "<nil>"
		if text != nil {
			t = *text
		}
		got = append(got, change+": "+t)
		return err
	}
	return &got
}

func withText(m Model) Model {
	return m.WithConfig(Config{Text: func(change string) string { return "with TDD" }})
}

func TestStartWithAdjustedText(t *testing.T) {
	m := withText(keys(board(t, 120, 30, fixture), "2")) // Ready: m1-phase-derivation
	got := texter(&m, nil)
	m = keys(m, "S")
	if !strings.Contains(m.View(), "start> with TDD█") {
		t.Fatalf("S must open the line pre-filled:\n%s", m.View())
	}
	m = keys(m, "backspace", "backspace", "backspace", "tests", " ", "first")
	if !strings.Contains(m.View(), "start> with tests first█") {
		t.Fatalf("the line must take typing and backspace:\n%s", m.View())
	}
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if cmd == nil || !strings.Contains(m.View(), "starting m1-phase-derivation…") {
		t.Fatalf("enter must start in a command and say so:\n%s", m.View())
	}
	if _, cmd = m.Update(cmd()); !slices.Equal(*got, []string{"m1-phase-derivation: with tests first"}) || cmd() != tea.Quit() {
		t.Errorf("started %q, then quit", *got)
	}
}

func TestStartWithEmptyLine(t *testing.T) {
	m := withText(keys(board(t, 120, 30, fixture), "3")) // In Progress: m1-board-view
	got := texter(&m, nil)
	m = keys(m, "S")
	for range "with TDD" {
		m = keys(m, "backspace")
	}
	m = keys(m, "backspace") // an empty line stays open
	if !strings.Contains(m.View(), "start> █") {
		t.Fatalf("backspace on an empty line must keep it open:\n%s", m.View())
	}
	m, cmd := pressEnter(m)
	m.Update(cmd())
	if !slices.Equal(*got, []string{"m1-board-view: "}) {
		t.Errorf("started %q, want the empty line", *got)
	}
}

func pressEnter(m Model) (Model, tea.Cmd) {
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	return next.(Model), cmd
}

func TestStartWithCancel(t *testing.T) {
	m := withText(keys(board(t, 120, 30, fixture), "2"))
	got := texter(&m, nil)
	before := m.View()
	m = keys(m, "S", "x", "esc")
	if m.View() != before || len(*got) != 0 {
		t.Errorf("esc must close the line and start nothing: started %q, view:\n%s", *got, m.View())
	}
	if m = keys(m, "l"); m.nav.col != 2 {
		t.Error("keys must work again after esc")
	}
}

func TestStartWithNotices(t *testing.T) {
	tests := []struct {
		name string
		keys []string
		want string
	}{
		{"Draft", []string{"1"}, "S starts a Ready or In Progress change"},
		{"Review", []string{"4"}, "S starts a Ready or In Progress change"},
		{"marks", []string{"2", " ", "j", " "}, "S starts one change: clear the marks first"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := withText(keys(board(t, 120, 30, fixture), tt.keys...))
			got := texter(&m, nil)
			m, cmd := pressKey(m, "S")
			if cmd != nil || len(*got) != 0 || strings.Contains(m.View(), "start>") || !strings.Contains(m.View(), tt.want) {
				t.Errorf("want notice %q and no line, view:\n%s", tt.want, m.View())
			}
		})
	}
}

func TestStartWithRunningAgentFocuses(t *testing.T) {
	m := withText(keys(board(t, 120, 30, fixture), "3"))
	got := texter(&m, nil)
	m.running = map[string]bool{herdr.AgentName("m1-board-view"): true}
	m, cmd := pressKey(m, "S")
	if cmd == nil || strings.Contains(m.View(), "start>") {
		t.Fatalf("S on a running agent must start as s does, view:\n%s", m.View())
	}
	m.Update(cmd())
	if !slices.Equal(*got, []string{"m1-board-view: <nil>"}) {
		t.Errorf("started %q, want s's Start", *got)
	}
}

func TestStartUsesNoText(t *testing.T) {
	m := withText(keys(board(t, 120, 30, fixture), "2"))
	got := texter(&m, nil)
	m, cmd := pressKey(m, "s")
	m.Update(cmd())
	if !slices.Equal(*got, []string{"m1-phase-derivation: <nil>"}) {
		t.Errorf("s must send the configured text, started %q", *got)
	}
}

func TestStartWithFromMenu(t *testing.T) {
	m := withText(keys(board(t, 120, 30, fixture), "2"))
	c := m.cols[work.Ready][0]
	if k := m.menuFor(c, 0, 0).keys; !slices.Contains(k, "S") {
		t.Fatalf("a Ready Card's menu must offer S, keys %q", k)
	}
	m.menu = m.menuFor(c, 0, 0)
	if !strings.Contains(m.View(), "Start with…  S") {
		t.Fatalf("the menu must show the entry:\n%s", m.View())
	}
	if m = keys(m, "S"); !strings.Contains(m.View(), "start> with TDD█") {
		t.Errorf("the entry must open the line:\n%s", m.View())
	}
	for _, p := range []work.Phase{work.Draft, work.Review, work.Archived} {
		if k := m.menuFor(work.Item{Name: "x", Phase: p}, 0, 0).keys; slices.Contains(k, "S") {
			t.Errorf("%s: no S in the menu, keys %q", p, k)
		}
	}
	m.running = map[string]bool{herdr.AgentName(c.Name): true}
	if k := m.menuFor(c, 0, 0).keys; slices.Contains(k, "S") {
		t.Errorf("a running agent: no S in the menu, keys %q", k)
	}
}

func TestEditConfig(t *testing.T) {
	m := keys(board(t, 120, 30, fixture), "/", "m1", "enter", "3")
	args := fakeEditor(t, nil)
	reloads := 0
	m = m.WithConfig(Config{
		Path: func() (string, error) { return "/cfg/config.json", nil },
		Reload: func() ([columns]bool, Wave, string) {
			reloads++
			return [columns]bool{work.Archived: true}, Wave{Size: 3}, "/cfg/config.json: invalid character, using defaults"
		},
	})
	sel := m.cols[m.nav.col][m.nav.sel[m.nav.col]].Name
	m, cmd := pressKey(m, "c")
	if cmd == nil {
		t.Fatal("c must open the editor")
	}
	m = send(m, cmd())
	if !slices.Equal(*args, []string{"ed", "/cfg/config.json"}) || reloads != 1 {
		t.Fatalf("editor %q, reloads %d", *args, reloads)
	}
	v := m.View()
	if strings.Contains(v, "Archived") || !strings.Contains(v, "/cfg/config.json: invalid character, using defaults") || m.wave.Size != 3 {
		t.Errorf("the reload must hide Archived, set the wave size and show the warning:\n%s", v)
	}
	if m.filter != "m1" || m.cols[m.nav.col][m.nav.sel[m.nav.col]].Name != sel {
		t.Errorf("filter %q and selection must stay", m.filter)
	}
}

func TestEditConfigWithoutDir(t *testing.T) {
	m := board(t, 120, 30, fixture).WithConfig(Config{Path: func() (string, error) { return "", errors.New("no config directory") }})
	m, cmd := pressKey(m, "c")
	if cmd != nil || !strings.Contains(m.View(), "no config directory") {
		t.Errorf("want a notice and no editor:\n%s", m.View())
	}
}
