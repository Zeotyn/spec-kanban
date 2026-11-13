package board

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"

	"github.com/zeotyn/spec-kanban/internal/detail"
	"github.com/zeotyn/spec-kanban/internal/work"
)

// at returns the cell of the first text in the drawn view, searching from the bottom line when last is set.
func at(t *testing.T, m Model, text string, last bool) (x, y int) {
	t.Helper()
	lines := strings.Split(ansi.Strip(m.View()), "\n")
	for i := range lines {
		if last {
			i = len(lines) - 1 - i
		}
		if j := strings.Index(lines[i], text); j >= 0 {
			return ansi.StringWidth(lines[i][:j]), i
		}
	}
	t.Fatalf("%q is not drawn:\n%s", text, m.View())
	return 0, 0
}

func mouse(m Model, x, y int, b tea.MouseButton) (Model, tea.Cmd) {
	next, cmd := m.Update(tea.MouseMsg{X: x, Y: y, Button: b, Action: tea.MouseActionPress})
	return next.(Model), cmd
}

// click presses button b on the first text drawn in the view.
func click(t *testing.T, m Model, text string, b tea.MouseButton) (Model, tea.Cmd) {
	t.Helper()
	x, y := at(t, m, text, false)
	return mouse(m, x, y, b)
}

// clickFooter left-clicks the footer, the view's last line, dx cells right of where text starts.
func clickFooter(t *testing.T, m Model, text string, dx int) (Model, tea.Cmd) {
	t.Helper()
	x, y := at(t, m, text, true)
	if y != strings.Count(m.View(), "\n") {
		t.Fatalf("%q is not in the footer:\n%s", text, m.View())
	}
	return mouse(m, x+dx, y, tea.MouseButtonLeft)
}

func left(t *testing.T, m Model, text string) Model {
	t.Helper()
	m, _ = click(t, m, text, tea.MouseButtonLeft)
	return m
}

func TestClickSelectsCard(t *testing.T) {
	for _, h := range []int{30, 12} { // bordered and borderless
		m := keys(board(t, 120, h, fixture), "2", "j") // Ready: m5-release-…
		m = left(t, m, "m1-board-view")
		if m.nav.col != 2 || m.nav.sel[2] != 0 || m.nav.sel[1] != 1 || m.detail != nil {
			t.Errorf("%d rows: col %d sel %v detail %v", h, m.nav.col, m.nav.sel, m.detail != nil)
		}
		m = left(t, m, "10/10") // a line of the Review Card below its name
		if m.nav.col != 3 || m.detail != nil {
			t.Errorf("%d rows: a click on a Card's second line must select it: col %d", h, m.nav.col)
		}
	}
}

func TestSecondClickOpensDetail(t *testing.T) {
	for _, h := range []int{30, 12} {
		m := left(t, keys(board(t, 120, h, fixture), "3"), "m1-board-view")
		if m.detail == nil || !strings.HasPrefix(m.View(), "m1-board-view  M1  5/11  In Progress") {
			t.Errorf("%d rows: a click on the selected Card must open its details:\n%s", h, m.View())
		}
	}
}

func TestClickOnEmptySpace(t *testing.T) {
	m := keys(board(t, 120, 30, append(tall(3), fixture[4])), "3") // Draft is empty
	want := m.nav
	_, y := at(t, m, "m1-board-view", false)
	x, _ := at(t, m, "In Progress", false)
	for _, c := range [][2]int{{0, 0}, {x, y + 5}, {x, y + 12}, {x, 1}} {
		if got, cmd := mouse(m, c[0], c[1], tea.MouseButtonLeft); got.nav != want || got.detail != nil || cmd != nil {
			t.Errorf("click at %v changed the Board: %+v", c, got.nav)
		}
	}
	if got := left(t, m, "nothing here"); got.nav != want {
		t.Errorf("click on an empty column changed the Board: %+v", got.nav)
	}
	got := left(t, m, "Ready 3") // the column title
	if got.nav != want {
		t.Errorf("click on a column title changed the Board: %+v", got.nav)
	}
}

func TestMouseIgnoresReleaseAndMotion(t *testing.T) {
	m := board(t, 120, 30, fixture)
	x, y := at(t, m, "m1-board-view", false)
	for _, a := range []tea.MouseAction{tea.MouseActionRelease, tea.MouseActionMotion} {
		next, _ := m.Update(tea.MouseMsg{X: x, Y: y, Button: tea.MouseButtonLeft, Action: a})
		if next.(Model).nav != m.nav {
			t.Errorf("action %v changed the Board", a)
		}
	}
}

// inProgress is one Ready Card and n In Progress Cards.
func inProgress(n int) []work.Item {
	cs := tall(n + 1)
	for i := range cs[1:] {
		cs[i+1].Phase = work.InProgress
	}
	return cs
}

func TestWheelOverColumn(t *testing.T) {
	m := keys(board(t, 120, 40, inProgress(10)), "2") // room for 6 Cards
	if m.visible() != 6 {
		t.Fatalf("visible %d, want 6", m.visible())
	}
	x, y := at(t, m, "In Progress 10", false)
	for range 7 {
		m, _ = mouse(m, x, y+10, tea.MouseButtonWheelDown)
	}
	if m.nav.col != 2 || m.nav.sel[2] != 7 || !strings.Contains(m.View(), "change-08") || !strings.Contains(m.View(), "↑ 2 more") {
		t.Errorf("col %d sel %d:\n%s", m.nav.col, m.nav.sel[2], m.View())
	}
	m, _ = mouse(m, x, y, tea.MouseButtonWheelUp)
	if m.nav.sel[2] != 6 {
		t.Errorf("wheel up: sel %d, want 6", m.nav.sel[2])
	}
	for range 10 {
		m, _ = mouse(m, x, y, tea.MouseButtonWheelUp)
	}
	if m.nav.sel[2] != 0 {
		t.Errorf("wheel up stops at the first Card: sel %d", m.nav.sel[2])
	}
}

func TestWheelOverEmptyColumn(t *testing.T) {
	m := keys(board(t, 120, 30, fixture[:5]), "2") // no Review Cards
	x, y := at(t, m, "Review 0", false)
	if got, _ := mouse(m, x, y+3, tea.MouseButtonWheelDown); got.nav != m.nav {
		t.Errorf("wheel over an empty column changed the Board: %+v", got.nav)
	}
}

func TestArrowsInSingleColumn(t *testing.T) {
	m := keys(board(t, 60, 30, fixture), "2")
	at(t, m, "◀ 2/5 Ready 2 ▶", false)
	m = left(t, m, "▶")
	if m.nav.col != 2 {
		t.Fatalf("▶ shows column %d, want In Progress", m.nav.col)
	}
	m = left(t, m, "◀")
	if m.nav.col != 1 {
		t.Errorf("◀ shows column %d, want Ready", m.nav.col)
	}
	m, _ = mouse(m, 59, 10, tea.MouseButtonWheelDown)
	if m.nav.col != 1 || m.nav.sel[1] != 1 {
		t.Errorf("the wheel must act on the shown column: col %d sel %d", m.nav.col, m.nav.sel[1])
	}
}

func TestMouseWhilePromptingOrStarting(t *testing.T) {
	m := keys(board(t, 120, 30, fixture), "/")
	if got := left(t, m, "m1-board-view"); got.nav != m.nav || !got.prompting {
		t.Error("a click while prompting must do nothing")
	}
	m = keys(board(t, 120, 30, fixture), "2")
	started := starter(&m, nil)
	m, _ = pressKey(m, "s")
	got, cmd := click(t, m, "m1-board-view", tea.MouseButtonLeft)
	got, cmd2 := mouse(got, 0, 10, tea.MouseButtonWheelDown)
	if got.nav != m.nav || cmd != nil || cmd2 != nil || got.starting == "" || len(*started) != 0 {
		t.Error("mouse while starting must do nothing")
	}
}

func TestClickClosesHelp(t *testing.T) {
	m := board(t, 120, 30, fixture)
	_, y := at(t, m, "m1-board-view", false)
	x, _ := at(t, m, "In Progress", false)
	m = keys(m, "?")
	got, cmd := mouse(m, x, y, tea.MouseButtonLeft)
	if got.help || got.nav != m.nav || cmd != nil {
		t.Errorf("a click must only close the help: help %v nav %+v", got.help, got.nav)
	}
}

func TestClickFooter(t *testing.T) {
	type spot struct {
		text string
		dx   int
	}
	for w, at := range map[int]map[string]spot{ // where the actions are in the full and the short footer
		200: {"⏎": {"⏎ details", 2}, "e": {"e edit tasks", 7}, "s": {"s start", 0}, "/": {"/ filter", 3}, "r": {"r reload", 2},
			"?": {"? help", 0}, "q": {"q close", 4}, "j": {"j/k card", 5}},
		60: {"⏎": {"⏎ e s /", 0}, "e": {"⏎ e s /", 2}, "s": {"⏎ e s /", 4}, "/": {"⏎ e s /", 6}, "r": {"· r ·", 2},
			"?": {"· ? ·", 2}, "q": {"q close", 4}, "j": {"j/k", 0}},
	} {
		m := board(t, w, 30, fixture)
		press := func(k string) (Model, tea.Cmd) { return clickFooter(t, m, at[k].text, at[k].dx) }
		if _, cmd := press("q"); cmd == nil || cmd() != tea.Quit() {
			t.Errorf("%d: q close must close the Board", w)
		}
		if got, _ := press("/"); !got.prompting {
			t.Errorf("%d: / filter must open the prompt", w)
		}
		if got, _ := press("?"); !got.help {
			t.Errorf("%d: ? help must show the help", w)
		}
		if got, _ := press("⏎"); got.detail == nil {
			t.Errorf("%d: ⏎ details must open the details", w)
		}
		if got, cmd := press("s"); cmd != nil || !strings.Contains(got.notice, "plan it first") { // Draft is focused
			t.Errorf("%d: s start on a Draft Card must explain why not: %q", w, got.notice)
		}
		if got, cmd := press("r"); cmd == nil || !got.loading {
			t.Errorf("%d: r reload must reload", w)
		}
		if got, _ := press("e"); got.notice != "no tasks to edit" {
			t.Errorf("%d: e edit tasks must act like e: notice %q", w, got.notice)
		}
		if got, cmd := press("j"); got.nav != m.nav || cmd != nil {
			t.Errorf("%d: navigation keys in the footer must not react", w)
		}
		if got, cmd := clickFooter(t, m, at["⏎"].text, -1); got.detail != nil || cmd != nil {
			t.Errorf("%d: the space before ⏎ must not react", w)
		}
	}
}

func TestClickFooterWithNotice(t *testing.T) {
	m := keys(board(t, 120, 30, fixture), "s") // Draft: a notice
	if m.notice == "" {
		t.Fatal("want a notice")
	}
	x, y := at(t, m, m.notice, true)
	if got, cmd := mouse(m, x, y, tea.MouseButtonLeft); cmd != nil || got.notice != m.notice {
		t.Error("a click on a notice must do nothing")
	}
}

func TestFooterTextUnchanged(t *testing.T) {
	short := ansi.Strip(board(t, 60, 30, fixture).View())
	if !strings.HasSuffix(short, "\nh/l j/k ⏎ e s / gg/G ^d/^u 1-5 · r · ? · q close") {
		t.Errorf("short footer changed:\n%s", short)
	}
}

// --- Card menu ---

func TestMenuStart(t *testing.T) {
	m := board(t, 200, 30, fixture)
	started := starter(&m, nil)
	m, _ = click(t, m, "m1-phase-derivation", tea.MouseButtonRight)
	if m.menu == nil || m.nav.col != 1 || m.nav.sel[1] != 0 || !strings.Contains(m.View(), "▶ Start  s") {
		t.Fatalf("a right click must select the Card and open the menu:\n%s", m.View())
	}
	m, cmd := click(t, m, "Start", tea.MouseButtonLeft)
	if m.menu != nil || cmd == nil || !strings.Contains(m.View(), "starting m1-phase-derivation…") {
		t.Fatalf("a click on Start must close the menu and start:\n%s", m.View())
	}
	if run(m, cmd); len(*started) != 1 || (*started)[0] != "m1-phase-derivation" {
		t.Errorf("started %q", *started)
	}
}

func TestMenuResume(t *testing.T) {
	m := board(t, 200, 30, fixture)
	resumed := resumer(&m, "claude", true, nil)
	m, _ = click(t, m, "m0-spike-agent-start", tea.MouseButtonRight)
	if !strings.Contains(m.View(), "↻ Resume  s") {
		t.Fatalf("Review Card: want Resume:\n%s", m.View())
	}
	m, cmd := click(t, m, "Resume", tea.MouseButtonLeft)
	if run(m, cmd); len(*resumed) != 1 {
		t.Errorf("resumed %q", *resumed)
	}
}

func TestMenuOnDraft(t *testing.T) {
	m := board(t, 200, 30, fixture)
	started := starter(&m, nil)
	m, _ = click(t, m, "add-dark-mode", tea.MouseButtonRight)
	m, cmd := click(t, m, "Start", tea.MouseButtonLeft)
	if m.menu != nil || cmd != nil || len(*started) != 0 || !strings.Contains(m.notice, "plan it first") {
		t.Errorf("Draft: menu %v notice %q", m.menu != nil, m.notice)
	}
}

func TestMenuEntryStyle(t *testing.T) {
	lipgloss.SetColorProfile(termenv.ANSI)
	defer lipgloss.SetColorProfile(termenv.Ascii)
	faint := func(name string) bool {
		m, _ := click(t, board(t, 200, 30, fixture), name, tea.MouseButtonRight)
		return strings.Contains(m.View(), "\x1b[2m▶ Start  s")
	}
	if !faint("add-dark-mode") || !faint("m0-plugin-skeleton") || faint("m1-board-view") {
		t.Error("the entry is faint on Draft and Archived Cards only")
	}
}

func TestMenuCloses(t *testing.T) {
	open := func() Model {
		m, _ := click(t, keys(board(t, 200, 30, fixture), "2"), "m1-board-view", tea.MouseButtonRight)
		return m
	}
	m := open()
	want := m.nav
	for _, k := range []string{"j", "esc", "q"} {
		got, cmd := pressKey(m, k)
		if k == "esc" {
			next, c := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
			got, cmd = next.(Model), c
		}
		if got.menu != nil || got.nav != want || cmd != nil {
			t.Errorf("%s must only close the menu: menu %v nav %+v cmd %v", k, got.menu != nil, got.nav, cmd != nil)
		}
	}
	got := left(t, m, "m0-spike-agent-start")
	if got.menu != nil || got.nav != want {
		t.Errorf("a left click outside must only close the menu: %+v", got.nav)
	}
	got, _ = click(t, m, "m0-spike-agent-start", tea.MouseButtonRight)
	if got.menu == nil || got.menu.card.Name != "m0-spike-agent-start" || got.nav.col != 3 {
		t.Error("a right click on another Card must open the menu there")
	}
	if got, _ = mouse(m, 0, 0, tea.MouseButtonRight); got.menu != nil {
		t.Error("a right click outside a Card must close the menu")
	}
}

func TestMenuStartKey(t *testing.T) {
	m := open(t, "m1-board-view")
	started := starter(&m, nil)
	m, cmd := pressKey(m, "s")
	if run(m, cmd); m.menu != nil || len(*started) != 1 {
		t.Errorf("s in the menu must start: %q", *started)
	}
}

func open(t *testing.T, name string) Model {
	t.Helper()
	m, _ := click(t, board(t, 200, 30, fixture), name, tea.MouseButtonRight)
	return m
}

// box returns the cell of the drawn menu's top-left corner and its size, found by its entry.
func box(t *testing.T, m Model) (x, y, w, h int) {
	t.Helper()
	x, y = at(t, m, "▶ Start  s", false)
	lines := strings.Split(ansi.Strip(m.View()), "\n")
	top := ansi.Cut(lines[y-1], x-2, x+12)
	if top != "╭────────────╮" || ansi.Cut(lines[y+1], x-2, x+12) != "╰────────────╯" {
		t.Fatalf("the menu at %d,%d is not a whole box:\n%s", x-2, y-1, m.View())
	}
	return x - 2, y - 1, 14, 3
}

func TestMenuAtEdges(t *testing.T) {
	m := board(t, 120, 30, fixture)
	m, _ = mouse(m, 118, 6, tea.MouseButtonRight) // the archived Card at the right edge
	if m.menu == nil {
		t.Fatalf("no menu:\n%s", m.View())
	}
	if x, y, w, _ := box(t, m); x+w > 120 || x >= 118 || y != 6 {
		t.Errorf("menu at %d,%d must end inside the popup, left of the pointer:\n%s", x, y, m.View())
	}
	for _, l := range strings.Split(m.View(), "\n") {
		if ansi.StringWidth(l) > 120 {
			t.Errorf("line %q is wider than the popup", l)
		}
	}
	m, _ = enter(keys(board(t, 120, 30, fixture), "3")) // the details fill the popup
	m, _ = mouse(m, 0, 29, tea.MouseButtonRight)
	if _, y, _, h := box(t, m); y+h > 30 || strings.Count(m.View(), "\n") != 29 {
		t.Errorf("menu at row %d must end inside 30 rows:\n%s", y, m.View())
	}
}

func TestMenuOverWideCharacters(t *testing.T) {
	lipgloss.SetColorProfile(termenv.ANSI)
	defer lipgloss.SetColorProfile(termenv.Ascii)
	cs := []work.Item{{Name: "日本語の変更日本語の変更", Total: 2, Phase: work.Ready, Modified: ago(time.Hour)}}
	m := board(t, 120, 30, cs)
	before := strings.Split(m.View(), "\n")
	nx, ny := at(t, m, "日本", false)
	for dx := 0; dx < 4; dx++ { // the box's edges fall inside a wide character or between two
		got, _ := mouse(m, nx+dx, ny, tea.MouseButtonRight)
		lines := strings.Split(got.View(), "\n")
		for i, l := range lines {
			if i < len(before) && ansi.StringWidth(l) != ansi.StringWidth(before[i]) && ansi.StringWidth(l) > 120 {
				t.Errorf("dx %d: line %d is %d cells", dx, i, ansi.StringWidth(l))
			}
		}
		if !strings.Contains(ansi.Strip(got.View()), "│ ▶ Start  s │") {
			t.Errorf("dx %d: the entry is not drawn whole:\n%s", dx, got.View())
		}
		if box, _, _ := got.menu.box(120, 30); !strings.Contains(lines[ny+1], ansi.ResetStyle+box[1]) {
			t.Errorf("dx %d: the style before the box must be reset: %q", dx, lines[ny+1])
		}
	}
}

func TestMenuInDetail(t *testing.T) {
	m, _ := enter(keys(board(t, 120, 30, fixture), "3"))
	started := starter(&m, nil)
	m, _ = mouse(m, 40, 15, tea.MouseButtonRight)
	if m.menu == nil || !strings.Contains(m.View(), "▶ Start  s") || m.detail == nil {
		t.Fatalf("a right click in the details must open the menu:\n%s", m.View())
	}
	m, cmd := pressKey(m, "s")
	if run(m, cmd); len(*started) != 1 || (*started)[0] != "m1-board-view" {
		t.Errorf("s in the menu must start the shown Card: %q", *started)
	}
}

func TestMenuVerifyArchive(t *testing.T) {
	m := board(t, 200, 30, fixture)
	acted := actor(&m, nil, "")
	m, _ = click(t, m, "m0-spike-agent-start", tea.MouseButtonRight)
	if v := ansi.Strip(m.View()); !strings.Contains(v, "│ ↻ Resume  s  │") || !strings.Contains(v, "│ ✓ Verify  v  │") || !strings.Contains(v, "│ ⇥ Archive  a │") {
		t.Fatalf("Review Card: want Resume, Verify and Archive:\n%s", v)
	}
	m, cmd := click(t, m, "✓ Verify", tea.MouseButtonLeft)
	if m.menu != nil || !strings.HasSuffix(m.View(), "verifying m0-spike-agent-start…") {
		t.Fatalf("a click on Verify must close the menu and verify:\n%s", m.View())
	}
	if run(m, cmd); !slices.Equal(*acted, []string{"verify m0-spike-agent-start"}) {
		t.Errorf("acted %q", *acted)
	}
	m = board(t, 200, 30, fixture)
	acted = actor(&m, nil, "")
	m, _ = click(t, m, "m0-spike-agent-start", tea.MouseButtonRight)
	m, cmd = pressKey(m, "a")
	if run(m, cmd); m.menu != nil || !slices.Equal(*acted, []string{"archive m0-spike-agent-start"}) {
		t.Errorf("a in the menu must archive: %q", *acted)
	}
}

func TestMenuVerifyFaintOutsideReview(t *testing.T) {
	lipgloss.SetColorProfile(termenv.ANSI)
	defer lipgloss.SetColorProfile(termenv.Ascii)
	m := board(t, 200, 30, fixture)
	acted := actor(&m, nil, "")
	m, _ = click(t, m, "m1-board-view", tea.MouseButtonRight)
	if v := m.View(); !strings.Contains(v, "\x1b[2m✓ Verify  v") || !strings.Contains(v, "\x1b[2m⇥ Archive  a") || strings.Contains(v, "\x1b[2m▶ Start") {
		t.Fatalf("In Progress: Verify and Archive faint, Start not:\n%q", v)
	}
	m, cmd := click(t, m, "⇥ Archive", tea.MouseButtonLeft)
	if m.menu != nil || cmd != nil || len(*acted) != 0 || m.notice != "m1-board-view is not in Review" {
		t.Errorf("a faint entry shows its key's notice: menu %v notice %q", m.menu != nil, m.notice)
	}
}

func TestMenuSpecKit(t *testing.T) {
	m := board(t, 200, 30, fixture)
	acted := actor(&m, nil, "Spec Kit has no verify")
	m, _ = click(t, m, "m0-spike-agent-start", tea.MouseButtonRight)
	if v := ansi.Strip(m.View()); !strings.Contains(v, "│ ↻ Resume  s │") || strings.Contains(v, "✓ Verify") || strings.Contains(v, "⇥ Archive") {
		t.Fatalf("a Source without the commands: Resume only:\n%s", v)
	}
	if m, cmd := pressKey(m, "v"); m.menu != nil || cmd != nil || len(*acted) != 0 || m.notice != "" {
		t.Errorf("v is no entry here: it only closes the menu, notice %q", m.notice)
	}
}

func TestMenuOfThreeAtEdges(t *testing.T) {
	m := board(t, 120, 30, fixture)
	actor(&m, nil, "")
	m, _ = mouse(m, 118, 6, tea.MouseButtonRight) // the archived Card at the right edge
	x, y := at(t, m, "⇥ Archive  a", false)
	if lines := strings.Split(ansi.Strip(m.View()), "\n"); x+14 > 120 || x >= 118 || y != 9 || ansi.Cut(lines[y+1], x-2, x+14) != "╰──────────────╯" {
		t.Errorf("menu entry at %d,%d must end inside the popup, left of the pointer:\n%s", x, y, m.View())
	}
	m, _ = enter(keys(board(t, 120, 30, fixture), "4"))
	actor(&m, nil, "")
	m, _ = mouse(m, 0, 29, tea.MouseButtonRight)
	if _, y := at(t, m, "⇥ Archive  a", false); y != 28 || strings.Count(m.View(), "\n") != 29 {
		t.Errorf("menu ending at row %d must end inside 30 rows:\n%s", y+1, m.View())
	}
}

// settle feeds the messages of cmd back into the Board, as bubbletea does, until there is none.
func settle(m Model, cmd tea.Cmd) Model {
	for cmd != nil {
		msg := cmd()
		if _, ok := msg.(tea.KeyMsg); !ok {
			if _, ok := msg.(detail.CloseMsg); !ok {
				return m
			}
		}
		next, c := m.Update(msg)
		m, cmd = next.(Model), c
	}
	return m
}

func TestClickDetailFooter(t *testing.T) {
	for _, w := range []int{120, 40} { // full and short footer
		m, _ := enter(keys(board(t, w, 30, fixture), "3"))
		text, dx := "? help", 0
		if w == 40 {
			text, dx = "· ? ·", 2
		}
		got, cmd := clickFooter(t, m, text, dx)
		if got = settle(got, cmd); !got.help {
			t.Errorf("%d: ? help in the details must show the help", w)
		}
		got, cmd = clickFooter(t, m, "q back", 0)
		if got = settle(got, cmd); got.detail != nil || got.nav.col != 2 {
			t.Errorf("%d: q back must return to the Board", w)
		}
	}
	m, _ := enter(keys(board(t, 120, 30, fixture), "3"))
	if got := left(t, m, "Proposal"); !strings.Contains(got.View(), "▌") || got.detail == nil {
		t.Error("a click on a tab must reach the details")
	}
}

func TestMenuInDetailShowsNotices(t *testing.T) {
	m, _ := enter(board(t, 120, 30, fixture)) // Draft: m2-change-detail
	m, _ = mouse(m, 40, 15, tea.MouseButtonRight)
	m, _ = pressKey(m, "s")
	if m.detail != nil || !strings.HasSuffix(m.View(), "m2-change-detail is a draft: plan it first") || m.nav.col != 0 || m.nav.sel[0] != 0 {
		t.Errorf("the menu's action returns to the Board, where the notice shows:\n%s", m.View())
	}
}

func TestStartFromDetail(t *testing.T) {
	m, _ := enter(keys(board(t, 120, 30, fixture), "3"))
	starter(&m, nil)
	m, _ = mouse(m, 40, 15, tea.MouseButtonRight)
	m, _ = click(t, m, "Start", tea.MouseButtonLeft)
	if v := m.View(); m.detail != nil || !strings.HasSuffix(v, "starting m1-board-view…") {
		t.Fatalf("the Board must say what is starting:\n%s", v)
	}
	for _, k := range []string{"q", "j", "enter"} {
		if got, cmd := pressKey(m, k); cmd != nil || got.detail != nil || got.View() != m.View() {
			t.Errorf("%s while starting must do nothing", k)
		}
	}
	m = send(m, startedMsg{errors.New("no tab")})
	if !strings.HasSuffix(m.View(), "start failed: no tab") {
		t.Errorf("a failed Start shows in the footer:\n%s", m.View())
	}
}

func TestDetailMenuFollowsReload(t *testing.T) {
	m, _ := enter(keys(board(t, 120, 30, fixture), "2")) // Ready: m1-phase-derivation
	moved := append([]work.Item(nil), fixture...)
	moved[2].Phase, moved[2].Done = work.Review, 6
	m = send(m, loadedMsg{changes: moved})
	if m, _ = mouse(m, 40, 15, tea.MouseButtonRight); !strings.Contains(m.View(), "↻ Resume  s") {
		t.Errorf("the menu must show the Change's current Phase:\n%s", m.View())
	}
}

func TestHelpAndTheWheel(t *testing.T) {
	m := keys(board(t, 120, 30, fixture), "?")
	if got, _ := mouse(m, 40, 10, tea.MouseButtonWheelDown); !got.help || got.nav != m.nav {
		t.Error("the wheel is no click: the help stays and nothing moves")
	}
	none := keys(New("", nil, nil, time.Now, nil, nil, nil, nil, nil, nil, nil, [columns]bool{}, "", "test", Wave{}), "?")
	if got, _ := mouse(none, 0, 0, tea.MouseButtonLeft); got.help {
		t.Error("a click closes the help without a project too")
	}
}

func TestWheelOutsideColumns(t *testing.T) {
	m := keys(board(t, 120, 30, fixture), "2")
	x, _ := at(t, m, "Ready 2", false)
	for _, y := range []int{0, 1, strings.Count(m.View(), "\n")} { // the header, the blank line, the footer
		if got, _ := mouse(m, x, y, tea.MouseButtonWheelDown); got.nav != m.nav {
			t.Errorf("wheel on row %d changed the Board", y)
		}
	}
}

func TestMenuFollowsReload(t *testing.T) {
	m := open(t, "m1-phase-derivation") // Ready
	started := starter(&m, nil)
	moved := append([]work.Item(nil), fixture...)
	moved[2].Phase, moved[2].Done = work.Review, 6
	m = send(m, loadedMsg{changes: moved})
	if m, _ = pressKey(m, "s"); m.starting != "" || len(*started) != 0 || !strings.Contains(m.notice, "are done") {
		t.Errorf("the menu must act on the Card's current Phase: starting %q notice %q", m.starting, m.notice)
	}
}
