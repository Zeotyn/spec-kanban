// Package board is the Board UI: a Project's Changes in Phase columns with vim-style navigation.
package board

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/zeotyn/openspec-kanban/internal/detail"
	"github.com/zeotyn/openspec-kanban/internal/docs"
	"github.com/zeotyn/openspec-kanban/internal/editor"
	"github.com/zeotyn/openspec-kanban/internal/openspec"
	"github.com/zeotyn/openspec-kanban/internal/phase"
)

const (
	minColWidth = 20 // below this per column the Board shows one column at a time
	cardLines   = 4  // three lines of content plus a blank line
	chromeLines = 6  // header, blank, column title, rule, footer, blank above footer
)

// editedMsg reports that the editor started from the Board has exited.
type editedMsg struct{ err error }

type loadedMsg struct {
	changes []openspec.Change
	err     error
}

// agentsMsg carries the names of the Changes with a Running agent.
type agentsMsg struct {
	running map[string]bool
	err     error
}

// startedMsg reports that Start has returned.
type startedMsg struct{ err error }

// helpText lists every key by view; README's keys table says the same.
const helpText = `Keys

Board
  h/l j/k      column, card
  gg/G         first/last card
  ctrl+d/u     half page down/up
  1-5          column by number
  enter        details
  e            edit tasks
  s            start the agent
  /            filter by name
  n/N          next/previous match
  esc          clear the filter, else close
  r            reload
  ?            this help
  q            close

Filter prompt
  enter/esc    keep/clear the filter
  backspace    delete, close when empty

Details
  1-4, tab     switch document
  shift+tab    previous document
  j/k gg/G     scroll, top/end
  ctrl+d/u     half page down/up
  e            edit the document
  q/esc        back to the Board

any key closes`

var runningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // terminal green

// Model shows the Board of one Project; an empty project means none was found.
type Model struct {
	project       string
	load          func() ([]openspec.Change, error)
	agents        func() (map[string]bool, error) // names of the Changes with a Running agent
	start         func(change string) error
	now           func() time.Time
	loading       bool
	loaded        bool // a load has succeeded at least once
	err           error
	all           [columns][]openspec.Change // every Change by Phase
	cols          [columns][]openspec.Change // the Changes of all that match filter
	filter        string
	prompting     bool            // keys go to the filter prompt
	running       map[string]bool // Changes with a Running agent, kept across loads
	starting      string          // the Change being started; keys other than ctrl+c are ignored meanwhile
	nav           nav
	detail        *detail.Model   // the open detail view, nil on the Board
	detailOf      openspec.Change // the Change the detail shows, matched by name and archive state
	stale         bool            // something changed while a load ran; load again when it ends
	notice        string          // one-line message in place of the footer, cleared by the next key
	hide          [columns]bool   // Phases whose columns are not shown
	help          bool            // the key help is shown; the next key closes it
	width, height int
}

// New builds the Board. A non-nil snapshot (the cached last load) is shown while Init loads afresh.
// agents and start may be nil (no badges, s does nothing). hide marks Phases not shown as columns;
// notice replaces the footer until the first key.
func New(project string, load func() ([]openspec.Change, error), now func() time.Time, snapshot []openspec.Change,
	agents func() (map[string]bool, error), start func(change string) error, hide [columns]bool, notice string) Model {
	m := Model{project: project, load: load, agents: agents, start: start, now: now, hide: hide, notice: notice,
		loading: project != "", width: 120, height: 30}
	if snapshot != nil {
		m.all, m.loaded = phase.Group(snapshot), true
		m.refilter()
	}
	return m
}

func (m Model) Init() tea.Cmd {
	if m.project == "" {
		return nil
	}
	return tea.Batch(m.reload, m.lookup())
}

// lookup asks for the Running agents, or returns nil without a lookup.
func (m Model) lookup() tea.Cmd {
	if m.agents == nil {
		return nil
	}
	return func() tea.Msg {
		running, err := m.agents()
		return agentsMsg{running, err}
	}
}

func (m Model) reload() tea.Msg {
	changes, err := m.load()
	return loadedMsg{changes, err}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok && m.starting == "" && !m.prompting {
		switch {
		case m.help && k.String() == "ctrl+c":
			return m, tea.Quit
		case m.help:
			m.help = false
			return m, nil
		case k.String() == "?":
			m.help, m.notice, m.nav.pendingG = true, "", false
			if m.detail != nil { // ? is no detail key: it only clears the detail's pending g and notice
				d, _ := m.detail.Update(k)
				m.detail = &d
			}
			return m, nil
		}
	}
	if m.detail != nil {
		switch msg.(type) {
		case detail.CloseMsg:
			m.detail = nil
			return m, nil
		case detail.EditedMsg:
			d, cmd := m.detail.Update(msg)
			m.detail = &d
			m, reload := m.refresh() // the edit may have changed the Card too
			return m, tea.Batch(cmd, reload)
		case tea.KeyMsg:
			d, cmd := m.detail.Update(msg)
			m.detail = &d
			return m, cmd
		case tea.WindowSizeMsg:
			d, _ := m.detail.Update(msg)
			m.detail = &d // and fall through so the Board resizes too
		}
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		for c := range columns {
			m.nav.fit(c, len(m.cols[c]), m.visible())
		}
	case loadedMsg:
		m.loading, m.err = false, msg.err
		if msg.err == nil {
			m.all, m.loaded = phase.Group(msg.changes), true
			m.refilter()
			m.updateDetailHeader()
		}
		if m.stale {
			m.stale = false
			return m.refresh()
		}
	case agentsMsg:
		if msg.err == nil { // outside herdr there are simply no badges
			m.running = msg.running
		}
	case startedMsg:
		if msg.err == nil {
			return m, tea.Quit
		}
		m.starting, m.notice = "", "start failed: "+msg.err.Error()
		return m, m.lookup() // a half-done Start may have left an agent behind
	case detail.EditedMsg: // the detail closed before the editor reported back
		return m.refresh()
	case editedMsg:
		if msg.err != nil {
			m.notice = "editor failed: " + msg.err.Error()
		}
		return m.refresh()
	case tea.KeyMsg:
		if m.starting != "" { // quitting now would leave a half-started agent behind
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			return m, nil
		}
		m.notice = ""
		if m.prompting {
			return m.prompt(msg)
		}
		switch k := msg.String(); {
		case k == "esc" && m.filter != "":
			m.nav.pendingG, m.filter = false, ""
			m.refilter()
		case k == "q" || k == "esc" || k == "ctrl+c":
			return m, tea.Quit
		case k == "/":
			m.nav.pendingG, m.prompting = false, true
		case k == "n" && m.filter != "":
			m.nav.match(1, m.counts(), m.visible())
		case k == "N" && m.filter != "":
			m.nav.match(-1, m.counts(), m.visible())
		case k == "enter":
			m.nav.pendingG = false
			if col := m.cols[m.nav.col]; len(col) > 0 {
				c := col[m.nav.sel[m.nav.col]]
				d := detail.New(m.project, c, detailHeader(c, phase.Phase(m.nav.col)), m.width, m.height)
				m.detail, m.detailOf = &d, c
			}
		case k == "e":
			m.nav.pendingG = false
			if col := m.cols[m.nav.col]; len(col) > 0 {
				return m.edit(col[m.nav.sel[m.nav.col]])
			}
		case k == "s":
			m.nav.pendingG = false
			if col := m.cols[m.nav.col]; len(col) > 0 {
				return m.startCard(col[m.nav.sel[m.nav.col]], phase.Phase(m.nav.col))
			}
		case k == "r" && m.project != "":
			m.nav.pendingG = false
			m, cmd := m.refresh()
			return m, tea.Batch(cmd, m.lookup())
		default:
			m.nav.key(k, m.counts(), m.visible())
		}
	}
	return m, nil
}

// prompt applies a key to the filter prompt: text is typed, Backspace deletes or closes the empty
// prompt, Enter closes it keeping the filter, Esc closes it clearing the filter.
func (m Model) prompt(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEnter:
		m.prompting = false
	case tea.KeyEsc:
		m.prompting, m.filter = false, ""
	case tea.KeyBackspace:
		if r := []rune(m.filter); len(r) > 0 {
			m.filter = string(r[:len(r)-1])
		} else {
			m.prompting = false
		}
	case tea.KeyRunes, tea.KeySpace:
		for _, r := range msg.Runes { // a paste may carry newlines and control characters
			if unicode.IsPrint(r) {
				m.filter += string(r)
			}
		}
	}
	m.refilter()
	return m, nil
}

// refilter shows the Changes matching the filter in the shown columns, keeping each column's selection by name.
func (m *Model) refilter() {
	old := m.cols
	m.cols = [columns][]openspec.Change{}
	for p, col := range m.all {
		for _, c := range col {
			if !m.hide[p] && matches(c.Name, m.filter) {
				m.cols[p] = append(m.cols[p], c)
			}
		}
	}
	m.nav.keep(names(old), names(m.cols), m.visible())
}

// refresh starts a load unless one is running, so an older result cannot overwrite a newer one.
func (m Model) refresh() (Model, tea.Cmd) {
	if m.loading {
		m.stale = true
		return m, nil
	}
	if m.project == "" {
		return m, nil
	}
	m.loading = true
	return m, m.reload
}

// updateDetailHeader refreshes the open detail's header from the freshly loaded Change.
func (m *Model) updateDetailHeader() {
	if m.detail == nil {
		return
	}
	for p, col := range m.all {
		for _, c := range col {
			if c.Name == m.detailOf.Name && c.Archived == m.detailOf.Archived {
				d := m.detail.WithHeader(detailHeader(c, phase.Phase(p)))
				m.detail = &d
				return
			}
		}
	}
}

// edit opens c's tasks.md in the editor, or explains why not.
func (m Model) edit(c openspec.Change) (Model, tea.Cmd) {
	tasks := filepath.Join(docs.Dir(m.project, c), "tasks.md")
	if c.Archived {
		m.notice = "archived changes are read-only"
		return m, nil
	}
	if _, err := os.Stat(tasks); err != nil {
		m.notice = "no tasks to edit"
		return m, nil
	}
	return m, editor.Open(func(err error) tea.Msg { return editedMsg{err} }, tasks)
}

// startCard Starts c in a command, or explains why its Phase p cannot be started.
func (m Model) startCard(c openspec.Change, p phase.Phase) (Model, tea.Cmd) {
	switch p {
	case phase.Draft:
		m.notice = c.Name + " is a draft: plan it first"
	case phase.Review:
		m.notice = "all tasks of " + c.Name + " are done"
	case phase.Archived:
		m.notice = "archived changes cannot be started"
	default:
		if m.start == nil {
			return m, nil
		}
		m.starting = c.Name
		return m, func() tea.Msg { return startedMsg{m.start(c.Name)} }
	}
	return m, nil
}

// detailHeader sums up a Change for the detail view: name, milestone, Progress and Phase or archive date.
func detailHeader(c openspec.Change, p phase.Phase) string {
	parts := []string{c.Name}
	if b := milestone(c.Name); b != "" {
		parts = append(parts, b)
	}
	switch {
	case c.Archived:
		parts = append(parts, strings.TrimSpace("archived "+c.ArchivedOn))
	case c.Total > 0:
		parts = append(parts, fmt.Sprintf("%d/%d", c.Done, c.Total), p.String())
	default:
		parts = append(parts, "no tasks", p.String())
	}
	return strings.Join(parts, "  ")
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
	if m.help {
		return lipgloss.NewStyle().MaxWidth(m.width).MaxHeight(m.height).Render(helpText)
	}
	if m.detail != nil {
		return m.detail.View()
	}
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
	footer := "h/l column · j/k card · ⏎ details · e edit tasks · s start · / filter · gg/G first/last · ctrl+d/u half page · 1-5 column · r reload · ? help · q close"
	if lipgloss.Width(footer) > m.width {
		footer = "h/l j/k ⏎ e s / gg/G ^d/^u 1-5 · r · ? · q close"
	}
	if m.prompting {
		footer = lipgloss.NewStyle().MaxWidth(m.width).Render("/" + m.filter + "█")
	}
	if m.notice != "" {
		footer = lipgloss.NewStyle().MaxWidth(m.width).Render(m.notice)
	}
	if m.starting != "" {
		footer = lipgloss.NewStyle().MaxWidth(m.width).Render("starting " + m.starting + "…")
	}
	return m.header() + "\n\n" + body + "\n\n" + footer // no trailing newline: the view fills the height exactly
}

func (m Model) header() string {
	title := "OpenSpec Kanban · " + filepath.Base(m.project)
	if m.filter != "" {
		title += " · /" + m.filter
	}
	if m.loading && m.loaded { // before the first load the body says so
		title += " · refreshing…"
	}
	var done, total int
	for _, col := range m.all[:phase.Archived] {
		for _, c := range col {
			done, total = done+c.Done, total+max(c.Total, 0)
		}
	}
	if total == 0 || m.err != nil {
		return title
	}
	progress := fmt.Sprintf("%d/%d %d%%", done, total, (done*200+total)/(2*total)) // rounded
	// a long filter must not push Progress off
	title = lipgloss.NewStyle().MaxWidth(max(0, m.width-lipgloss.Width(progress)-1)).Render(title)
	gap := max(1, m.width-lipgloss.Width(title)-lipgloss.Width(progress))
	return title + strings.Repeat(" ", gap) + progress
}

func (m Model) columns() string {
	n := columns
	for _, h := range m.hide {
		if h {
			n--
		}
	}
	if m.width < n*minColWidth+n-1 {
		c := m.nav.col
		for m.hide[c] { // only when nothing shown has Cards and focus fell back to Draft
			c++
		}
		return m.column(c, m.width, fmt.Sprintf("◀ %d/%d %s %d ▶", c+1, columns, phase.Phase(c), len(m.cols[c])))
	}
	w := (m.width - (n - 1)) / n
	parts := make([]string, 0, 2*n-1)
	for i := range columns {
		if m.hide[i] {
			continue
		}
		if len(parts) > 0 {
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

// card renders three lines and a blank one; the selected Card is marked with ▌ and bold, an open
// Card with a Running agent with a green ● before its name.
func (m Model) card(c openspec.Change, p phase.Phase, w int, selected bool) []string {
	inner := w - 1
	running := !c.Archived && m.running[c.Name]
	firstW := inner
	if running {
		firstW -= 2 // "● "
	}
	badge := milestone(c.Name)
	nameW := firstW
	if badge != "" {
		nameW = firstW - len(badge) - 1
	}
	name := FitName(c.Name, nameW)
	first := name
	if badge != "" {
		first += strings.Repeat(" ", max(1, firstW-lipgloss.Width(name)-len(badge))) + badge
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
	if running { // styled apart so the badge's colour reset does not end the bold name
		out[0] = mark + runningStyle.Render("●") + style.Render(" "+first)
	}
	return append(out, "")
}
