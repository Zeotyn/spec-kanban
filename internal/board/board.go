// Package board is the Board UI: a Project's Changes in Phase columns with vim-style navigation.
package board

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/zeotyn/openspec-kanban/internal/openspec"
	"github.com/zeotyn/openspec-kanban/internal/phase"
)

const (
	minColWidth = 20 // below this per column the Board shows one column at a time
	cardLines   = 4  // three lines of content plus a blank line
	chromeLines = 6  // header, blank, column title, rule, footer, blank above footer
)

type loadedMsg struct {
	changes []openspec.Change
	err     error
}

// Model shows the Board of one Project; an empty project means none was found.
type Model struct {
	project       string
	load          func() ([]openspec.Change, error)
	now           func() time.Time
	loading       bool
	loaded        bool // a load has succeeded at least once
	err           error
	cols          [columns][]openspec.Change
	nav           nav
	width, height int
}

// New builds the Board. A non-nil snapshot (the cached last load) is shown at once while the fresh
// load from Init runs; nil means show a loading notice until then.
func New(project string, load func() ([]openspec.Change, error), now func() time.Time, snapshot []openspec.Change) Model {
	m := Model{project: project, load: load, now: now, loading: project != "", width: 120, height: 30}
	if snapshot != nil {
		m.cols, m.loaded = phase.Group(snapshot), true
		m.nav.keep([columns][]string{}, names(m.cols), m.visible())
	}
	return m
}

func (m Model) Init() tea.Cmd {
	if m.project == "" {
		return nil
	}
	return m.reload
}

func (m Model) reload() tea.Msg {
	changes, err := m.load()
	return loadedMsg{changes, err}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		for c := range columns {
			m.nav.fit(c, len(m.cols[c]), m.visible())
		}
	case loadedMsg:
		m.loading, m.err = false, msg.err
		if msg.err == nil {
			old := m.cols
			m.cols, m.loaded = phase.Group(msg.changes), true
			m.nav.keep(names(old), names(m.cols), m.visible())
		}
	case tea.KeyMsg:
		switch k := msg.String(); {
		case k == "q" || k == "esc" || k == "ctrl+c":
			return m, tea.Quit
		case k == "r" && m.project != "":
			m.nav.pendingG = false
			if m.loading { // one load at a time, so an older result cannot overwrite a newer one
				return m, nil
			}
			m.loading = true
			return m, m.reload
		default:
			m.nav.key(k, m.counts(), m.visible())
		}
	}
	return m, nil
}

func names(cols [columns][]openspec.Change) [columns][]string {
	var n [columns][]string
	for i, col := range cols {
		for _, c := range col {
			n[i] = append(n[i], c.Name)
		}
	}
	return n
}

func (m Model) counts() [columns]int {
	var n [columns]int
	for i, col := range m.cols {
		n[i] = len(col)
	}
	return n
}

// visible is how many Cards fit in a column, leaving one line each for the scroll markers.
func (m Model) visible() int {
	return max(1, (m.height-chromeLines-2)/cardLines)
}

func (m Model) View() string {
	if m.project == "" {
		return "No OpenSpec project here.\n\n" +
			"Open the board from a pane inside a project (a directory containing openspec/changes).\n\n" +
			"q close\n"
	}
	var body string
	switch {
	case m.err != nil:
		// openspec errors can be long or multi-line; keep them inside the view
		body = lipgloss.NewStyle().Width(m.width).MaxHeight(max(1, m.height-6)).Render(m.err.Error()) + "\n\nr retry · q close"
	case !m.loaded:
		body = "Loading changes…"
	default:
		body = m.columns()
	}
	footer := "h/l column · j/k card · gg/G first/last · ctrl+d/u half page · 1-5 column · r reload · q close"
	if lipgloss.Width(footer) > m.width {
		footer = "h/l j/k gg/G ^d/^u 1-5 · r reload · q close"
	}
	return m.header() + "\n\n" + body + "\n\n" + footer // no trailing newline: the view fills the height exactly
}

func (m Model) header() string {
	title := "OpenSpec Kanban · " + filepath.Base(m.project)
	if m.loading && m.loaded { // before the first load the body says so
		title += " · refreshing…"
	}
	var done, total int
	for _, col := range m.cols[:phase.Archived] {
		for _, c := range col {
			done, total = done+c.Done, total+max(c.Total, 0)
		}
	}
	if total == 0 || m.err != nil {
		return title
	}
	progress := fmt.Sprintf("%d/%d %d%%", done, total, (done*200+total)/(2*total)) // rounded
	gap := max(1, m.width-lipgloss.Width(title)-lipgloss.Width(progress))
	return title + strings.Repeat(" ", gap) + progress
}

func (m Model) columns() string {
	if m.width < columns*minColWidth+columns-1 {
		return m.column(m.nav.col, m.width, fmt.Sprintf("◀ %d/%d %s %d ▶", m.nav.col+1, columns, phase.Phase(m.nav.col), len(m.cols[m.nav.col])))
	}
	w := (m.width - (columns - 1)) / columns
	parts := make([]string, 0, 2*columns-1)
	for i := range columns {
		if i > 0 {
			parts = append(parts, " ")
		}
		parts = append(parts, m.column(i, w, fmt.Sprintf("%s %d", phase.Phase(i), len(m.cols[i]))))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

// column renders column i at width w under title.
func (m Model) column(i, w int, title string) string {
	lines := []string{title, strings.Repeat("─", w)}
	cards, off, vis := m.cols[i], m.nav.off[i], m.visible()
	if len(cards) == 0 {
		lines = append(lines, "nothing here")
	}
	if off > 0 {
		lines = append(lines, fmt.Sprintf("↑ %d more", off))
	}
	end := min(len(cards), off+vis)
	for j := off; j < end; j++ {
		lines = append(lines, m.card(cards[j], phase.Phase(i), w, i == m.nav.col && j == m.nav.sel[i])...)
	}
	if end < len(cards) {
		lines = append(lines, fmt.Sprintf("↓ %d more", len(cards)-end))
	}
	return lipgloss.NewStyle().Width(w).Render(strings.Join(lines, "\n"))
}

// card renders three lines and a blank one; the selected Card is marked with ▌ and bold.
func (m Model) card(c openspec.Change, p phase.Phase, w int, selected bool) []string {
	inner := w - 1
	badge := milestone(c.Name)
	nameW := inner
	if badge != "" {
		nameW = inner - len(badge) - 1
	}
	name := fitName(c.Name, nameW)
	first := name
	if badge != "" {
		first += strings.Repeat(" ", max(1, inner-lipgloss.Width(name)-len(badge))) + badge
	}
	var second, third string
	switch {
	case c.Archived && c.ArchivedOn != "":
		second = "archived " + c.ArchivedOn
	case c.Archived:
		second = "archived"
	default:
		second = progressBar(c.Done, c.Total, max(3, min(10, inner-8)))
		third = age(m.now().Sub(c.Modified))
		if p == phase.Draft || p == phase.Ready {
			third = planMarks(c.Artifacts) + "  " + third
		}
	}
	mark, style := " ", lipgloss.NewStyle()
	if selected {
		mark, style = "▌", style.Bold(true)
	}
	out := make([]string, 0, cardLines)
	for _, l := range []string{first, second, third} {
		out = append(out, mark+style.Render(l))
	}
	return append(out, "")
}
