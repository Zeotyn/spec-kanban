// Package detail is the detail view of one Change: its documents in tabs, rendered and scrollable.
package detail

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/zeotyn/openspec-kanban/internal/docs"
	"github.com/zeotyn/openspec-kanban/internal/openspec"
)

// Style is the glamour standard style. Not "auto": that queries the terminal, which can stall for
// seconds when it does not answer and races bubbletea for input. A light option comes with M5 config.
var Style = "dark"

// CloseMsg asks the Board to close the detail view.
type CloseMsg struct{}

const (
	tasks = iota
	proposal
	design
	specs
	tabs
)

var tabNames = [tabs]string{"Tasks", "Proposal", "Design", "Specs"}

// chrome is header, tab bar, two rules and the footer.
const chrome = 5

type Model struct {
	header        string
	docs          docs.Docs
	tab           int
	off           [tabs]int
	lines         [tabs][]string
	width, height int
	pendingG      bool
}

// New opens the details of c in project under header (built by the Board, which owns Card formatting).
func New(project string, c openspec.Change, header string, width, height int) Model {
	m := Model{header: header, docs: docs.Read(docs.Dir(project, c)), width: width, height: height, tab: proposal}
	if c.Total > 0 {
		m.tab = tasks
	}
	m.render()
	return m
}

func (m Model) doc(tab int) string {
	return [tabs]string{m.docs.Tasks, m.docs.Proposal, m.docs.Design, m.docs.Specs}[tab]
}

// content returns the markdown of tab, or a note when the Change has no such document.
func (m Model) content(tab int) string {
	doc := m.doc(tab)
	if doc == "" {
		return "_This change has no " + strings.ToLower(tabNames[tab]) + "._"
	}
	if tab == tasks && m.docs.Next != "" {
		return m.docs.Next + "\n\n" + doc
	}
	return doc
}

func (m *Model) render() {
	r, err := glamour.NewTermRenderer(glamour.WithStandardStyle(Style), glamour.WithWordWrap(max(20, m.width-4)))
	for t := range tabs {
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
	case tea.KeyMsg:
		g := m.pendingG
		m.pendingG = false
		off := &m.off[m.tab]
		switch k := msg.String(); k {
		case "q", "esc":
			return m, func() tea.Msg { return CloseMsg{} }
		case "ctrl+c":
			return m, tea.Quit
		case "1", "2", "3", "4":
			m.tab = int(k[0] - '1')
		case "tab":
			m.tab = (m.tab + 1) % tabs
		case "shift+tab":
			m.tab = (m.tab + tabs - 1) % tabs
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

func (m Model) View() string {
	bar := func(short bool) string {
		parts := make([]string, tabs)
		for t := range tabs {
			mark, name := " ", tabNames[t]
			switch {
			case t == m.tab:
				mark = "▌"
			case m.doc(t) == "":
				mark = "·"
			}
			if short {
				name = name[:1]
			}
			parts[t] = mark + string(rune('1'+t)) + " " + name
		}
		return strings.Join(parts, " │ ")
	}
	tabBar := bar(false)
	if lipgloss.Width(tabBar) > m.width { // keep the tab bar on one line
		tabBar = bar(true)
	}
	rule := strings.Repeat("─", m.width)
	lines := m.lines[m.tab][m.off[m.tab]:min(len(m.lines[m.tab]), m.off[m.tab]+m.body())]
	// glamour does not wrap code, tables or long URLs: cut them, or the terminal wraps them and the view grows
	fit := lipgloss.NewStyle().MaxWidth(m.width)
	body := fit.Height(m.body()).MaxHeight(m.body()).Render(strings.Join(lines, "\n"))
	footer := "1-4/tab tab · j/k scroll · ctrl+d/u half page · gg/G top/end · q back"
	if lipgloss.Width(footer) > m.width {
		footer = "1-4 j/k ^d/^u gg/G · q back"
	}
	return fit.Render(m.header) + "\n" + fit.Render(tabBar) + "\n" + rule + "\n" + body + "\n" + rule + "\n" + fit.Render(footer)
}
