package board

import (
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/zeotyn/spec-kanban/internal/theme"
	"github.com/zeotyn/spec-kanban/internal/work"
)

// layout is where the Board draws its columns, for drawing and for hit-testing: x is each shown column's
// left edge (-1 when not shown) and w the column width. In single-column mode only column col is shown, under title.
type layout struct {
	single bool
	col    int
	title  string
	x      [columns]int
	w      int
}

// at returns the column drawn at cell x, or -1.
func (l layout) at(x int) int {
	return slices.IndexFunc(l.x[:], func(x0 int) bool { return x0 >= 0 && x >= x0 && x < x0+l.w })
}

func (m Model) layout() layout {
	n := columns
	for _, h := range m.hide {
		if h {
			n--
		}
	}
	l := layout{x: [columns]int{-1, -1, -1, -1, -1}}
	if m.width < n*minColWidth+n-1 {
		c := m.nav.col
		for m.hide[c] { // only when nothing shown has Cards and focus fell back to Draft
			c++
		}
		l.single, l.col, l.w, l.x[c] = true, c, m.width, 0
		l.title = fmt.Sprintf("◀ %d/%d %s %d ▶", c+1, columns, work.Phase(c), len(m.cols[c]))
		return l
	}
	l.w = (m.width - (n - 1)) / n
	x := 0
	for i := range columns {
		if !m.hide[i] {
			l.x[i], x = x, x+l.w+1
		}
	}
	return l
}

// cardAt returns the column and the Card drawn at cell x, y; ok is false where no Card is drawn.
func (m Model) cardAt(x, y int) (col, card int, ok bool) {
	if !m.loaded || m.err != nil || y >= 2+max(1, m.height-4) { // below the body's cut
		return 0, 0, false
	}
	if col = m.layout().at(x); col < 0 {
		return 0, 0, false
	}
	top, off := 4, m.nav.off[col] // below the header, a blank line, the column title and its rule
	if off > 0 {
		top++ // ↑ n more
	}
	card = off + (y-top)/m.cardLines()
	if y < top || card >= min(len(m.cols[col]), off+m.visible()) {
		return 0, 0, false
	}
	return col, card, true
}

// hit is a footer action: a click on the cells [x0, x1) presses key.
type hit struct {
	x0, x1 int
	key    string
}

// footer renders footerKeys in full, or short when the full form is wider than the Board, and where the actions are.
func (m Model) footer() (string, []hit) {
	build := func(short bool) (string, []hit) {
		var b strings.Builder
		var hits []hit
		for i, s := range footerKeys {
			switch {
			case i == 0:
			case !short || i >= shortApart:
				b.WriteString(" · ")
			default:
				b.WriteString(" ")
			}
			text := s.key + " " + theme.Faint.Render(s.desc)
			if short && i < len(footerKeys)-1 { // the last one keeps its description
				text = s.short
			}
			x := lipgloss.Width(b.String())
			if s.action != "" {
				hits = append(hits, hit{x, x + lipgloss.Width(text), s.action})
			}
			b.WriteString(text)
		}
		return b.String(), hits
	}
	if s, hits := build(false); lipgloss.Width(s) <= m.width {
		return s, hits
	}
	return build(true)
}

// menu is the open Card menu: the Card it acts on, the cell it was opened at and the keys of its entries.
type menu struct {
	card work.Item
	x, y int
	keys []string
}

// menuFor opens the Card menu for c at cell x, y: Start or Resume, then the actions the Source names.
func (m Model) menuFor(c work.Item, x, y int) *menu {
	keys := []string{"s"}
	for _, a := range actions {
		if m.act != nil && m.lacking(a.step) == "" {
			keys = append(keys, a.key)
		}
	}
	return &menu{c, x, y, keys}
}

// box renders the menu and the cell of its top-left corner: at the pointer, shifted to stay inside width × height.
func (c menu) box(width, height int) (lines []string, x, y int) {
	entries := []string{"▶ Start  s"}
	switch c.card.Phase {
	case work.Review:
		entries[0] = "↻ Resume  s"
	case work.Draft, work.Archived: // s only explains why not
		entries[0] = theme.Faint.Render(entries[0])
	}
	for _, k := range c.keys[1:] {
		e := actionOf(k).entry
		if c.card.Phase != work.Review { // its key only explains why not
			e = theme.Faint.Render(e)
		}
		entries = append(entries, e)
	}
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(theme.Dim).Padding(0, 1).Render(strings.Join(entries, "\n"))
	lines = strings.Split(box, "\n")
	return lines, max(0, min(c.x, width-lipgloss.Width(box))), max(0, min(c.y, height-len(lines)))
}

// entryAt returns the key of the entry drawn at cell x, y, "" when there is none.
func (c menu) entryAt(x, y, width, height int) string {
	lines, x0, y0 := c.box(width, height)
	if i := y - y0 - 1; i >= 0 && i < len(c.keys) && x >= x0 && x < x0+lipgloss.Width(lines[0]) {
		return c.keys[i]
	}
	return ""
}

// overlay draws box over view with its top-left corner at cell x, y, cutting the lines by display cells so
// styles and wide characters stay whole.
func overlay(view string, box []string, x, y int) string {
	lines := strings.Split(view, "\n")
	for len(lines) < y+len(box) {
		lines = append(lines, "")
	}
	for i, b := range box {
		l, end := lines[y+i], x+ansi.StringWidth(b)
		left := ansi.Truncate(l, x, "")
		left += strings.Repeat(" ", x-ansi.StringWidth(left)) // a wide character across the left edge
		right, rest := "", ansi.StringWidth(l)-end
		if rest > 0 {
			right = ansi.TruncateLeft(l, end, "")
			if ansi.StringWidth(right) > rest { // a wide character across the right edge
				right = ansi.TruncateLeft(l, end+1, " ")
			}
		}
		lines[y+i] = left + ansi.ResetStyle + b + right
	}
	return strings.Join(lines, "\n")
}

// choose does what key k does on the menu's Card, on the Board: what it starts or its notice shows in
// the Board's footer.
func (m Model) choose(c work.Item, k string) (Model, tea.Cmd) {
	m.detail = nil
	if k == "s" {
		return m.startCard(c, c.Phase)
	}
	return m.actCard(c, c.Phase, actionOf(k))
}

// press feeds key k through the key handling, as a click on its footer action does.
func (m Model) press(k string) (Model, tea.Cmd) {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
	if k == "enter" {
		msg = tea.KeyMsg{Type: tea.KeyEnter}
	}
	next, cmd := m.Update(msg)
	return next.(Model), cmd
}

// mouse handles a press of a mouse button; releases and motion are ignored, and so is everything while
// the filter prompt is open or a Start runs.
func (m Model) mouse(msg tea.MouseMsg) (Model, tea.Cmd) {
	if msg.Action != tea.MouseActionPress || m.prompting || m.starting != "" {
		return m, nil
	}
	if m.help { // a click closes it; a trackpad swipe of wheel events would close it and scroll on
		m.help = m.help && tea.MouseEvent(msg).IsWheel()
		return m, nil
	}
	if m.project == "" {
		return m, nil
	}
	if m.menu != nil {
		open := *m.menu
		m.menu = nil
		if k := open.entryAt(msg.X, msg.Y, m.width, m.height); msg.Button == tea.MouseButtonLeft && k != "" {
			return m.choose(open.card, k)
		}
		if msg.Button != tea.MouseButtonRight { // a right click opens the menu again where it is
			return m, nil
		}
	}
	if m.detail != nil {
		if msg.Button == tea.MouseButtonRight {
			m.menu = m.menuFor(m.detailOf, msg.X, msg.Y)
			return m, nil
		}
		d, cmd := m.detail.Update(msg)
		m.detail = &d
		return m, cmd
	}
	l := m.layout()
	switch msg.Button {
	case tea.MouseButtonLeft:
		if msg.Y == lipgloss.Height(m.body())+3 { // the footer, below the header, a blank line, the body and a blank line
			if _, hits := m.footer(); m.notice == "" {
				for _, h := range hits {
					if msg.X >= h.x0 && msg.X < h.x1 {
						return m.press(h.key)
					}
				}
			}
			return m, nil
		}
		if l.single && msg.Y == 2 && m.loaded && m.err == nil {
			switch msg.X {
			case 0:
				return m.press("h")
			case lipgloss.Width(l.title) - 1:
				return m.press("l")
			}
		}
		if c, i, ok := m.cardAt(msg.X, msg.Y); ok {
			if c == m.nav.col && i == m.nav.sel[c] {
				return m.press("enter")
			}
			m.notice, m.nav.pendingG, m.nav.col, m.nav.sel[c] = "", false, c, i
		}
	case tea.MouseButtonRight:
		if c, i, ok := m.cardAt(msg.X, msg.Y); ok {
			m.notice, m.nav.pendingG, m.nav.col, m.nav.sel[c] = "", false, c, i
			m.menu = m.menuFor(m.cols[c][i], msg.X, msg.Y)
		}
	case tea.MouseButtonWheelDown, tea.MouseButtonWheelUp:
		body := msg.Y >= 2 && msg.Y < 2+lipgloss.Height(m.body()) // below the header and a blank line
		if c := l.at(msg.X); c >= 0 && body && len(m.cols[c]) > 0 && m.loaded && m.err == nil {
			k := "j"
			if msg.Button == tea.MouseButtonWheelUp {
				k = "k"
			}
			m.notice, m.nav.col = "", c
			m.nav.key(k, m.counts(), m.visible())
		}
	}
	return m, nil
}
