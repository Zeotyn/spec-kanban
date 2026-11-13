package board

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/zeotyn/spec-kanban/internal/work"
)

// marksFooter is the footer with the fixture's two Ready Cards marked.
const marksFooter = "2 marked · s wave · esc clear"

// waver fakes the wave's Start: it records the Changes it is called with and returns err.
func waver(m *Model, size int, blocked string, err error) *[][]string {
	var started [][]string
	m.wave = Wave{Size: size, Blocked: blocked, Start: func(changes []string) error {
		started = append(started, changes)
		return err
	}}
	return &started
}

func footerOf(m Model) string {
	v := m.View()
	return v[strings.LastIndex(v, "\n")+1:]
}

func TestMarkCards(t *testing.T) {
	m := keys(board(t, 120, 30, fixture), "2", " ", "j", " ")
	v := m.View()
	contains(t, v, "✓ m1-phase", "✓ m5-relea")
	if f := footerOf(m); f != marksFooter {
		t.Errorf("footer %q, want %q", f, marksFooter)
	}
	m = keys(m, " ")
	if v := m.View(); strings.Contains(v, "✓ m5-relea") || !strings.Contains(v, "✓ m1-phase") {
		t.Errorf("space on a marked Card must unmark it:\n%s", v)
	}
	if f := footerOf(m); f != "1 marked · s wave · esc clear" {
		t.Errorf("footer %q", f)
	}
	if f := footerOf(keys(m, "k", " ")); strings.Contains(f, "marked") {
		t.Errorf("without marks the key footer returns: %q", f)
	}
}

func TestMarkOutsideReady(t *testing.T) {
	for _, col := range []string{"1", "3", "4", "5"} {
		m := keys(board(t, 120, 30, fixture), col, " ")
		if len(m.marks) != 0 || footerOf(m) != "only Ready cards join a wave" {
			t.Errorf("column %s: marks %v, footer %q", col, m.marks, footerOf(m))
		}
	}
}

func TestMarkLimit(t *testing.T) {
	m := keys(board(t, 120, 30, fixture), "2")
	waver(&m, 1, "", nil)
	m = keys(m, " ", "j", " ")
	if len(m.marks) != 1 || footerOf(m) != "a wave holds at most 1 changes" {
		t.Errorf("marks %v, footer %q", m.marks, footerOf(m))
	}
	if m = keys(m, "k", " "); len(m.marks) != 0 {
		t.Error("unmarking must work at the limit")
	}
}

func TestMarkNothing(t *testing.T) {
	m := keys(board(t, 120, 30, nil), " ")
	if len(m.marks) != 0 || m.notice != "" {
		t.Errorf("space in an empty column: marks %v, notice %q", m.marks, m.notice)
	}
	m, _ = enter(keys(board(t, 120, 30, fixture), "2"))
	if m = keys(m, " "); len(m.marks) != 0 || m.detail == nil {
		t.Errorf("space in the detail view: marks %v", m.marks)
	}
}

func TestMarksSurviveFilter(t *testing.T) {
	m := keys(board(t, 120, 30, fixture), "/", "m1", "enter", " ", "/", "backspace", "backspace", "m5", "enter", " ", "esc")
	if !m.marks["m1-phase-derivation"] || !m.marks["m5-release-pipeline-with-a-very-long-name"] || m.filter != "" {
		t.Errorf("marks %v, filter %q", m.marks, m.filter)
	}
}

func TestMarksAcrossReload(t *testing.T) {
	changes := fixture
	m := New("/x/proj", docs("/x/proj"), func() ([]work.Item, error) { return changes, nil }, func() time.Time { return now }, nil,
		nil, nil, nil, nil, nil, nil, [columns]bool{}, "", "test", Wave{})
	m = send(send(m, tea.WindowSizeMsg{Width: 120, Height: 30}), m.Init()())
	m = keys(m, "2", " ", "j", " ")
	changes = slices.Clone(fixture)
	changes[2].Phase, changes[2].Done = work.InProgress, 1 // m1-phase-derivation leaves Ready
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	m = send(next.(Model), cmd())
	if m.marks["m1-phase-derivation"] || !m.marks["m5-release-pipeline-with-a-very-long-name"] {
		t.Errorf("after the reload marks %v, want only m5", m.marks)
	}
}

func TestEscOrder(t *testing.T) {
	m := keys(board(t, 120, 30, fixture), "2", " ", "j", " ", "/", "m5", "enter")
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m = next.(Model); cmd != nil || m.filter != "" || len(m.marks) != 2 {
		t.Fatalf("first esc clears the filter only: filter %q, marks %v", m.filter, m.marks)
	}
	next, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m = next.(Model); cmd != nil || len(m.marks) != 0 {
		t.Fatalf("second esc clears the marks: %v", m.marks)
	}
	if _, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEsc}); cmd == nil || cmd() != tea.Quit() {
		t.Error("third esc closes the Board")
	}
}

func TestWaveStart(t *testing.T) {
	m := keys(board(t, 120, 30, fixture), "2", "j", " ", "k", " ")
	started := waver(&m, 5, "", nil)
	m, cmd := pressKey(m, "s")
	if cmd != nil || footerOf(m) != "start wave of 2? y/n" {
		t.Fatalf("s must ask: footer %q", footerOf(m))
	}
	m, cmd = pressKey(m, "y")
	if cmd == nil || len(*started) != 0 || footerOf(m) != "starting wave…" {
		t.Fatalf("y must start in a command and say so: footer %q", footerOf(m))
	}
	if mm := keys(m, "esc", "j"); mm.marks == nil || mm.nav != m.nav {
		t.Error("keys must be ignored while the wave starts")
	}
	if mm, _ := m.mouse(tea.MouseMsg{X: 30, Y: 5, Action: tea.MouseActionPress, Button: tea.MouseButtonRight}); mm.menu != nil {
		t.Error("clicks must be ignored while the wave starts")
	}
	want := []string{"m1-phase-derivation", "m5-release-pipeline-with-a-very-long-name"} // Board order
	if _, cmd = m.Update(cmd()); len(*started) != 1 || !slices.Equal((*started)[0], want) || cmd == nil || cmd() != tea.Quit() {
		t.Errorf("started %q, want %q, then quit", *started, want)
	}
}

func TestWaveCancel(t *testing.T) {
	m := keys(board(t, 120, 30, fixture), "2", " ", "j", " ")
	started := waver(&m, 5, "", nil)
	m, _ = pressKey(m, "s")
	m, cmd := pressKey(m, "n")
	if cmd != nil || len(*started) != 0 || len(m.marks) != 2 || footerOf(m) != marksFooter {
		t.Errorf("n must cancel and keep the marks: started %q, footer %q", *started, footerOf(m))
	}
}

func TestWaveOneMark(t *testing.T) {
	m := keys(board(t, 120, 30, fixture), "2", " ", "j")
	waves := waver(&m, 5, "", nil)
	started := starter(&m, nil)
	m, cmd := pressKey(m, "s")
	if cmd == nil || !strings.Contains(m.View(), "starting m5-release-pipeline-with-a-very-long-name…") {
		t.Fatalf("s with one mark starts the selected Card:\n%s", m.View())
	}
	m.Update(cmd())
	if len(*waves) != 0 || len(*started) != 1 {
		t.Errorf("waves %q, starts %q", *waves, *started)
	}
}

func TestWaveBlocked(t *testing.T) {
	for _, blocked := range []string{"waves need worktree on", "waves need OpenSpec"} {
		m := keys(board(t, 120, 30, fixture), "2", " ", "j", " ")
		started := waver(&m, 5, blocked, nil)
		m, cmd := pressKey(m, "s")
		if cmd != nil || len(*started) != 0 || footerOf(m) != blocked || len(m.marks) != 2 {
			t.Errorf("%s: started %q, footer %q", blocked, *started, footerOf(m))
		}
	}
}

func TestWaveStartFails(t *testing.T) {
	m := keys(board(t, 120, 30, fixture), "2", " ", "j", " ")
	waver(&m, 5, "", errors.New("worktree create failed"))
	m, _ = pressKey(m, "s")
	m, cmd := pressKey(m, "y")
	next, quit := m.Update(cmd())
	m = next.(Model)
	if quit != nil && quit() == tea.Quit() {
		t.Fatal("a failed wave must keep the Board")
	}
	if footerOf(m) != "wave failed: worktree create failed" || len(m.marks) != 2 {
		t.Errorf("footer %q, marks %v", footerOf(m), m.marks)
	}
	if m = keys(m, "l"); m.nav.col != 2 {
		t.Error("keys must work again after a failed wave")
	}
}

// waved returns a loaded Board of fixture whose wave records make each Change in of a member.
func waved(t *testing.T, of map[string]int) Model {
	t.Helper()
	m := New("/x/proj", docs("/x/proj"), func() ([]work.Item, error) { return fixture, nil }, func() time.Time { return now }, nil,
		nil, nil, nil, nil, nil, nil, [columns]bool{}, "", "test", Wave{Of: func() map[string]int { return of }})
	return send(send(m, tea.WindowSizeMsg{Width: 120, Height: 30}), m.Init()())
}

func TestWaveBadge(t *testing.T) {
	m := waved(t, map[string]int{"m1-board-view": 2, "add-dark-mode": 3, "m0-plugin-skeleton": 1})
	contains(t, m.View(), "m1-board-view   ≋2 M1", "add-dark-mode      ≋3")
	if strings.Contains(m.View(), "≋1") {
		t.Errorf("an Archived Card shows no wave badge:\n%s", m.View())
	}
	if m.cols[work.InProgress][0].Name != "m1-board-view" {
		t.Error("the badge must not move the Card")
	}
	if v := waved(t, map[string]int{}).View(); strings.Contains(v, "≋") {
		t.Errorf("no badge without membership:\n%s", v)
	}
}

func TestWaveBadgeColour(t *testing.T) {
	coloured(t)
	contains(t, waved(t, map[string]int{"m1-board-view": 2}).View(), fg("8").Render("≋2")+" "+fg("5").Render("M1"))
}

func TestWaveViews(t *testing.T) {
	golden(t, "board-marked", keys(board(t, 120, 30, fixture), "2", " ", "j", " ").View())
	golden(t, "board-wave", waved(t, map[string]int{"m1-board-view": 2, "m2-change-detail": 2}).View())
}

func TestHelpListsSpace(t *testing.T) {
	contains(t, keys(board(t, 120, 40, fixture), "?").View(), "space        mark for a wave")
}
