package detail

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/zeotyn/spec-kanban/internal/work"
)

// at finds text in the rendered view and returns its first cell.
func at(t *testing.T, view, text string) (x, y int) {
	t.Helper()
	for y, line := range strings.Split(ansi.Strip(view), "\n") {
		if i := strings.Index(line, text); i >= 0 {
			return ansi.StringWidth(line[:i]), y
		}
	}
	t.Fatalf("%q not drawn:\n%s", text, ansi.Strip(view))
	return 0, 0
}

func mouse(m Model, x, y int, b tea.MouseButton, a tea.MouseAction) (Model, tea.Cmd) {
	return m.Update(tea.MouseMsg{X: x, Y: y, Button: b, Action: a})
}

// click presses the left button on text as drawn.
func click(t *testing.T, m Model, text string) (Model, tea.Cmd) {
	t.Helper()
	x, y := at(t, m.View(), text)
	return mouse(m, x, y, tea.MouseButtonLeft, tea.MouseActionPress)
}

func wheel(m Model, b tea.MouseButton, n int) Model {
	for range n {
		m, _ = mouse(m, 5, 5, b, tea.MouseActionPress)
	}
	return m
}

func TestClickTab(t *testing.T) {
	for _, tt := range []struct {
		name  string
		width int
		text  string
	}{{"full bar", 80, "Design"}, {"short bar", 40, "3 D"}} {
		m := open(t, change, tt.width, 24)
		if m.tab != tasks {
			t.Fatalf("%s: start tab %d", tt.name, m.tab)
		}
		m, _ = click(t, m, tt.text)
		if m.tab != design {
			t.Errorf("%s: tab %d, want design", tt.name, m.tab)
		}
	}
}

func TestClickTabSeparatorDoesNothing(t *testing.T) {
	m := open(t, change, 80, 24)
	x, y := at(t, m.View(), " │ ")
	for _, dx := range []int{0, 1, 2} {
		if got, _ := mouse(m, x+dx, y, tea.MouseButtonLeft, tea.MouseActionPress); got.tab != m.tab {
			t.Errorf("separator cell +%d changed the tab to %d", dx, got.tab)
		}
	}
}

func TestWheel(t *testing.T) {
	m := press(open(t, change, 80, 24), "2")
	if m = wheel(m, tea.MouseButtonWheelDown, 2); m.off[proposal] != 6 {
		t.Errorf("2 steps down: offset %d, want 6", m.off[proposal])
	}
	if m = wheel(m, tea.MouseButtonWheelDown, 100); m.off[proposal] != m.maxOff(proposal) {
		t.Errorf("down at the end: offset %d, want %d", m.off[proposal], m.maxOff(proposal))
	}
	if m = wheel(m, tea.MouseButtonWheelUp, 100); m.off[proposal] != 0 {
		t.Errorf("up at the top: offset %d", m.off[proposal])
	}
	m = wheel(m, tea.MouseButtonWheelDown, 1)
	if m = press(m, "3"); wheel(m, tea.MouseButtonWheelDown, 1).off[proposal] != 3 || m.off[design] != 0 {
		t.Error("each tab keeps its own offset")
	}
}

// follow feeds the key a footer click sends back into m, as bubbletea does through the Board.
func follow(t *testing.T, m Model, cmd tea.Cmd) (Model, tea.Cmd) {
	t.Helper()
	if cmd == nil {
		t.Fatal("the click sends no key")
	}
	k, ok := cmd().(tea.KeyMsg)
	if !ok {
		t.Fatal("the click must send its key as a message")
	}
	return m.Update(k)
}

func TestClickBack(t *testing.T) {
	for _, width := range []int{120, 40} {
		m, cmd := click(t, open(t, change, width, 24), "q back")
		if _, cmd = follow(t, m, cmd); cmd == nil {
			t.Fatalf("width %d: no command", width)
		}
		if _, ok := cmd().(CloseMsg); !ok {
			t.Errorf("width %d: want CloseMsg", width)
		}
	}
}

func TestClickHelp(t *testing.T) {
	if _, cmd := click(t, open(t, change, 120, 24), "? help"); cmd == nil || cmd().(tea.KeyMsg).String() != "?" {
		t.Error("? help must send ? for the Board, which shows the help")
	}
}

func TestClickEdit(t *testing.T) {
	archived := work.Item{Name: "m1-board-view", Phase: work.Archived}
	for _, width := range []int{120, 40} {
		m := New(docsIn(t.TempDir(), archived), archived, "h", width, 24)
		text := "e edit"
		if width < 60 {
			text = "gg/G e"
		}
		x, y := at(t, m.View(), text)
		m, cmd := mouse(m, x+len(text)-1, y, tea.MouseButtonLeft, tea.MouseActionPress)
		if m, _ = follow(t, m, cmd); !strings.Contains(m.View(), "archived changes are read-only") {
			t.Errorf("width %d: no read-only notice:\n%s", width, m.View())
		}
	}
}

func TestFooterClickWithNoticeDoesNothing(t *testing.T) {
	archived := work.Item{Name: "m1-board-view", Phase: work.Archived}
	m := press(New(docsIn(t.TempDir(), archived), archived, "h", 120, 24), "e")
	x, y := 5, 24-1 // the notice covers the footer, so aim at where "q back" was
	if m.notice == "" {
		t.Fatal("no notice")
	}
	got, cmd := mouse(m, x+len("archived changes are read-only")-1, y, tea.MouseButtonLeft, tea.MouseActionPress)
	if cmd != nil || got.notice != m.notice {
		t.Errorf("footer click with a notice: cmd %v, notice %q", cmd != nil, got.notice)
	}
	m = press(open(t, change, 120, 24), "e")
	m.notice = "x"
	x, _ = at(t, open(t, change, 120, 24).View(), "q back")
	if _, cmd := mouse(m, x, y, tea.MouseButtonLeft, tea.MouseActionPress); cmd != nil {
		t.Error("q back must do nothing while a notice shows")
	}
}

func TestOtherMouseEventsDoNothing(t *testing.T) {
	m := open(t, change, 80, 24)
	tx, ty := at(t, m.View(), "Design")
	qx, qy := at(t, m.View(), "q back")
	for _, e := range []tea.MouseMsg{
		{X: tx, Y: ty, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease},
		{X: tx, Y: ty, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion},
		{X: tx, Y: ty, Button: tea.MouseButtonRight, Action: tea.MouseActionPress},
		{X: tx, Y: ty, Button: tea.MouseButtonMiddle, Action: tea.MouseActionPress},
		{X: qx, Y: qy, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease},
		{X: qx, Y: qy, Button: tea.MouseButtonRight, Action: tea.MouseActionPress},
		{X: 5, Y: 5, Button: tea.MouseButtonWheelDown, Action: tea.MouseActionRelease},
		{X: 0, Y: 0, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress},
		{X: 79, Y: 10, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress},
	} {
		got, cmd := m.Update(e)
		if cmd != nil || got.tab != m.tab || got.off != m.off {
			t.Errorf("%+v changed something: tab %d cmd %v", e, got.tab, cmd != nil)
		}
	}
}
