// Package detail is the detail view of one Change: its documents in tabs, rendered and scrollable.
package detail

import (
	_ "embed"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/zeotyn/spec-kanban/internal/editor"
	"github.com/zeotyn/spec-kanban/internal/theme"
	"github.com/zeotyn/spec-kanban/internal/work"
)

// Style names a glamour standard style; "" uses terminal.json, which colours only with the 16 ANSI
// colours so the terminal's own theme decides how things look. Tests use "notty".
// Not glamour's "auto": that queries the terminal, which can stall in a popup that does not answer.
var Style = ""

//go:embed terminal.json
var terminalStyle []byte

// CloseMsg asks the Board to close the detail view.
type CloseMsg struct{}

// EditedMsg reports that the editor started from the detail view has exited.
type EditedMsg struct{ Err error }

// maxTabs is the most tabs a Source may return: one per digit key.
const maxTabs = 9

// chrome is header, tab bar, two rules and the footer.
const chrome = 5

type Model struct {
	header        string
	read          func() work.Docs
	archived      bool
	notice        string // one-line message in place of the footer, cleared by the next key
	docs          work.Docs
	tab           int
	off           [maxTabs]int
	lines         [maxTabs][]string
	width, height int
	pendingG      bool
}

// New opens the details of c, whose documents read returns (again after the editor), under header
// (built by the Board, which owns Card formatting).
func New(read func() work.Docs, c work.Item, header string, width, height int) Model {
	m := Model{header: header, read: read, archived: c.Phase == work.Archived, docs: read(), width: width, height: height}
	m.tab = m.docs.First
	if c.Total > 0 {
		m.tab = m.docs.Tasks
	}
	m.render()
	return m
}

func (m Model) tabs() int { return len(m.docs.Tabs) }

func (m Model) doc(tab int) string { return m.docs.Tabs[tab].Body }

func (m Model) name(tab int) string { return m.docs.Tabs[tab].Title }

// WithHeader replaces the header, e.g. after the Board reloaded the Change.
func (m Model) WithHeader(header string) Model {
	m.header = header
	return m
}

// content returns the markdown of tab, or a note when the Change has no such document.
func (m Model) content(tab int) string {
	doc := m.doc(tab)
	if doc == "" {
		return "_This change has no " + strings.ToLower(m.name(tab)) + "._"
	}
	if tab == m.docs.Tasks && m.docs.Next != "" {
		return m.docs.Next + "\n\n" + doc
	}
	return doc
}

func (m *Model) render() {
	style := glamour.WithStylesFromJSONBytes(terminalStyle)
	if Style != "" {
		style = glamour.WithStandardStyle(Style)
	}
	r, err := glamour.NewTermRenderer(style, glamour.WithWordWrap(max(20, m.width-4)))
	for t := range m.tabs() {
		out := m.content(t)
		if err == nil {
			if s, err := r.Render(out); err == nil {
				out = s
			}
		}
		m.lines[t] = strings.Split(strings.Trim(out, "\n"), "\n")
		m.off[t] = min(m.off[t], m.maxOff(t))
	}
}

func (m Model) body() int { return max(1, m.height-chrome) }

func (m Model) maxOff(t int) int { return max(0, len(m.lines[t])-m.body()) }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.render()
	case EditedMsg:
		m.docs = m.read()
		m.render()
		if msg.Err != nil {
			m.notice = "editor failed: " + msg.Err.Error()
		}
	case tea.MouseMsg:
		if msg.Action != tea.MouseActionPress {
			break
		}
		switch {
		case msg.Button == tea.MouseButtonLeft && msg.Y == 1:
			_, hits := m.tabBar()
			for _, h := range hits {
				if msg.X >= h.x0 && msg.X < h.x1 {
					m.notice = ""
					m.tab = h.tab
				}
			}
		case msg.Button == tea.MouseButtonLeft && msg.Y == m.body()+chrome-1 && m.notice == "": // header, tab bar, rule, body, rule, footer
			_, hits := m.footer()
			for _, h := range hits {
				if msg.X >= h.x0 && msg.X < h.x1 {
					// as a message, so the Board sees the key too: ? is the Board's
					return m, func() tea.Msg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(h.key)} }
				}
			}
		case msg.Button == tea.MouseButtonWheelDown, msg.Button == tea.MouseButtonWheelUp:
			m.notice = ""
			step := 3
			if msg.Button == tea.MouseButtonWheelUp {
				step = -3
			}
			m.off[m.tab] = max(0, min(m.off[m.tab]+step, m.maxOff(m.tab)))
		}
	case tea.KeyMsg:
		m.notice = ""
		g := m.pendingG
		m.pendingG = false
		off := &m.off[m.tab]
		switch k := msg.String(); k {
		case "e":
			return m.edit()
		case "q", "esc":
			return m, func() tea.Msg { return CloseMsg{} }
		case "ctrl+c":
			return m, tea.Quit
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			if t := int(k[0] - '1'); t < m.tabs() {
				m.tab = t
			}
		case "tab":
			m.tab = (m.tab + 1) % m.tabs()
		case "shift+tab":
			m.tab = (m.tab + m.tabs() - 1) % m.tabs()
		case "j":
			*off++
		case "k":
			*off--
		case "ctrl+d":
			*off += max(1, m.body()/2)
		case "ctrl+u":
			*off -= max(1, m.body()/2)
		case "g":
			m.pendingG = !g
			if g {
				*off = 0
			}
		case "G":
			*off = m.maxOff(m.tab)
		}
		m.off[m.tab] = max(0, min(m.off[m.tab], m.maxOff(m.tab))) // only the shown tab; others keep theirs
	}
	return m, nil
}

// edit opens the current tab's Files in the editor, or explains why not.
func (m Model) edit() (Model, tea.Cmd) {
	switch {
	case m.archived:
		m.notice = "archived changes are read-only"
	case m.doc(m.tab) == "":
		m.notice = "no " + strings.ToLower(m.name(m.tab)) + " to edit"
	default:
		return m, editor.Open(func(err error) tea.Msg { return EditedMsg{err} }, m.docs.Tabs[m.tab].Files...)
	}
	return m, nil
}

// span is the cells [x0, x1) of a row that a click acts on: it shows tab, or presses key.
type span struct {
	x0, x1 int
	tab    int
	key    string
}

// tabBar returns the tab bar and where each tab is on it, the short form when the full one is too wide.
func (m Model) tabBar() (string, []span) {
	bar := func(short bool) (string, []span) {
		var b strings.Builder
		var hits []span
		for t := range m.tabs() {
			mark, name := " ", m.name(t)
			switch {
			case t == m.tab:
				mark = "▌"
			case m.doc(t) == "":
				mark = "·"
			}
			if short {
				name = name[:1]
			}
			style := theme.Faint
			if t == m.tab {
				style = theme.AccentBold
			}
			if t > 0 {
				b.WriteString(" │ ")
			}
			x := lipgloss.Width(b.String())
			part := style.Render(mark + string(rune('1'+t)) + " " + name)
			hits = append(hits, span{x0: x, x1: x + lipgloss.Width(part), tab: t})
			b.WriteString(part)
		}
		return b.String(), hits
	}
	if s, hits := bar(false); lipgloss.Width(s) <= m.width { // keep the tab bar on one line
		return s, hits
	}
	return bar(true)
}

// footer returns the footer and where its clickable actions are, the short form when the full one is too wide.
func (m Model) footer() (string, []span) {
	type piece struct{ text, key string }
	build := func(pieces ...piece) (string, []span) {
		var b strings.Builder
		var hits []span
		for _, p := range pieces {
			x := lipgloss.Width(b.String())
			if p.key != "" {
				hits = append(hits, span{x0: x, x1: x + lipgloss.Width(p.text), key: p.key})
			}
			b.WriteString(p.text)
		}
		return b.String(), hits
	}
	s, hits := build(
		piece{text: fmt.Sprintf("1-%d/tab tab · j/k scroll · ctrl+d/u half page · gg/G top/end · ", m.tabs())},
		piece{"e edit", "e"}, piece{text: " · "}, piece{"? help", "?"}, piece{text: " · "}, piece{"q back", "q"})
	if lipgloss.Width(s) <= m.width {
		return s, hits
	}
	return build(
		piece{text: fmt.Sprintf("1-%d j/k ^d/^u gg/G ", m.tabs())},
		piece{"e", "e"}, piece{text: " · "}, piece{"?", "?"}, piece{text: " · "}, piece{"q back", "q"})
}

func (m Model) View() string {
	tabBar, _ := m.tabBar()
	rule := strings.Repeat("─", m.width)
	lines := m.lines[m.tab][m.off[m.tab]:min(len(m.lines[m.tab]), m.off[m.tab]+m.body())]
	// glamour does not wrap code, tables or long URLs: cut them, or the terminal wraps them and the view grows
	fit := lipgloss.NewStyle().MaxWidth(m.width)
	body := fit.Height(m.body()).MaxHeight(m.body()).Render(strings.Join(lines, "\n"))
	footer, _ := m.footer()
	if m.notice != "" {
		footer = m.notice
	}
	return fit.Render(m.header) + "\n" + fit.Render(tabBar) + "\n" + rule + "\n" + body + "\n" + rule + "\n" + fit.Render(footer)
}
