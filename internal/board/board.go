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
	"github.com/zeotyn/openspec-kanban/internal/editor"
	"github.com/zeotyn/openspec-kanban/internal/theme"
	"github.com/zeotyn/openspec-kanban/internal/work"
)

const (
	minColWidth     = 20 // below this per column the Board shows one column at a time
	chromeLines     = 6  // header, blank, column title, rule, footer, blank above footer
	borderMinHeight = 13 // chromeLines, two scroll markers and one bordered Card
)

// editedMsg reports that the editor started from the Board has exited.
type editedMsg struct{ err error }

type loadedMsg struct {
	changes []work.Item
	err     error
}

// agentsMsg carries the names of the Changes with a Running agent.
type agentsMsg struct {
	running map[string]bool
	err     error
}

// startedMsg reports that Start has returned.
type startedMsg struct{ err error }

// key is a key and what it does, for the footer and the help.
type key struct{ key, desc string }

var footerKeys = []key{{"h/l", "column"}, {"j/k", "card"}, {"⏎", "details"}, {"e", "edit tasks"}, {"s", "start"}, {"/", "filter"},
	{"gg/G", "first/last"}, {"ctrl+d/u", "half page"}, {"1-5", "column"}, {"r", "reload"}, {"?", "help"}, {"q", "close"}}

// keyList renders keys as "key description" pairs separated by " · ", the descriptions faint.
func keyList(keys []key) string {
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = k.key + " " + theme.Faint.Render(k.desc)
	}
	return strings.Join(parts, " · ")
}

// helpGroups lists every key by view; README's keys table says the same.
var helpGroups = []struct {
	title string
	keys  []key
}{
	{"Board", []key{{"h/l j/k", "column, card"}, {"gg/G", "first/last card"}, {"ctrl+d/u", "half page down/up"}, {"1-5", "column by number"},
		{"enter", "details"}, {"e", "edit tasks"}, {"s", "start the agent"}, {"/", "filter by name"}, {"n/N", "next/previous match"},
		{"esc", "clear the filter, else close"}, {"r", "reload"}, {"?", "this help"}, {"q", "close"}}},
	{"Filter prompt", []key{{"enter/esc", "keep/clear the filter"}, {"backspace", "delete, close when empty"}}},
	{"Details", []key{{"1-4, tab", "switch document"}, {"shift+tab", "previous document"}, {"j/k gg/G", "scroll, top/end"},
		{"ctrl+d/u", "half page down/up"}, {"e", "edit the document"}, {"q/esc", "back to the Board"}}},
}

// Model shows the Board of one Project; an empty project means none was found.
type Model struct {
	project       string
	docs          func(work.Item) work.Docs
	load          func() ([]work.Item, error)
	agents        func() (map[string]bool, error) // names of the Changes with a Running agent
	start         func(change string) error
	resumable     func(change string) (kind string, ok bool) // the recorded agent kind ("" if none) and whether it can be reopened
	resume        func(change string) error
	now           func() time.Time
	loading       bool
	loaded        bool // a load has succeeded at least once
	err           error
	all           [columns][]work.Item // every Change by Phase
	cols          [columns][]work.Item // the Changes of all that match filter
	filter        string
	prompting     bool            // keys go to the filter prompt
	running       map[string]bool // Changes with a Running agent, kept across loads
	starting      string          // the Change being started; keys other than ctrl+c are ignored meanwhile
	nav           nav
	detail        *detail.Model // the open detail view, nil on the Board
	detailOf      work.Item     // the Change the detail shows, matched by name and archive state
	stale         bool          // something changed while a load ran; load again when it ends
	notice        string        // one-line message in place of the footer, cleared by the next key
	hide          [columns]bool // Phases whose columns are not shown
	help          bool          // the key help is shown; the next key closes it
	version       string        // shown in the help
	width, height int
}

// New builds the Board; docs reads an Item's documents. A non-nil snapshot (the cached last load) is shown while Init loads afresh.
// agents, start, resumable and resume may be nil (no badges, s does nothing, no record to resume).
// hide marks Phases not shown as columns; notice replaces the footer until the first key; version is shown in the help.
func New(project string, docs func(work.Item) work.Docs, load func() ([]work.Item, error), now func() time.Time, snapshot []work.Item,
	agents func() (map[string]bool, error), start func(change string) error,
	resumable func(change string) (string, bool), resume func(change string) error, hide [columns]bool, notice, version string) Model {
	m := Model{project: project, docs: docs, load: load, agents: agents, start: start, resumable: resumable, resume: resume, now: now, hide: hide, notice: notice,
		version: version, loading: project != "", width: 120, height: 30}
	if snapshot != nil {
		m.all, m.loaded = work.Group(snapshot), true
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
			m.all, m.loaded = work.Group(msg.changes), true
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
				d := detail.New(func() work.Docs { return m.docs(c) }, c, detailHeader(c, work.Phase(m.nav.col)), m.width, m.height)
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
				return m.startCard(col[m.nav.sel[m.nav.col]], work.Phase(m.nav.col))
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
	m.cols = [columns][]work.Item{}
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
			if c.Name == m.detailOf.Name && (c.Phase == work.Archived) == (m.detailOf.Phase == work.Archived) {
				d := m.detail.WithHeader(detailHeader(c, work.Phase(p)))
				m.detail = &d
				return
			}
		}
	}
}

// edit opens c's tasks file in the editor, or explains why not.
func (m Model) edit(c work.Item) (Model, tea.Cmd) {
	if c.Phase == work.Archived {
		m.notice = "archived changes are read-only"
		return m, nil
	}
	docs := m.docs(c)
	tasks := docs.Tabs[docs.Tasks].Files[0]
	if _, err := os.Stat(tasks); err != nil {
		m.notice = "no tasks to edit"
		return m, nil
	}
	return m, editor.Open(func(err error) tea.Msg { return editedMsg{err} }, tasks)
}

// startCard Starts c in a command, or explains why its Phase p cannot be started.
func (m Model) startCard(c work.Item, p work.Phase) (Model, tea.Cmd) {
	switch p {
	case work.Draft:
		m.notice = c.Name + " is a draft: plan it first"
	case work.Review:
		kind, ok := "", false
		if m.resumable != nil && m.resume != nil {
			kind, ok = m.resumable(c.Name)
		}
		switch {
		case ok || (m.running[c.Name] && m.resume != nil):
			m.starting = c.Name
			return m, func() tea.Msg { return startedMsg{m.resume(c.Name)} }
		case kind != "":
			m.notice = "resume not supported for " + kind
		default:
			m.notice = "all tasks of " + c.Name + " are done"
		}
	case work.Archived:
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
func detailHeader(c work.Item, p work.Phase) string {
	parts := []string{c.Name}
	if b := milestone(c.Name); b != "" {
		parts = append(parts, b)
	}
	switch {
	case c.Phase == work.Archived:
		parts = append(parts, strings.TrimSpace("archived "+c.ArchivedOn))
	case c.Total > 0:
		parts = append(parts, fmt.Sprintf("%d/%d", c.Done, c.Total), p.String())
	default:
		parts = append(parts, "no tasks", p.String())
	}
	return strings.Join(parts, "  ")
}

func names(cols [columns][]work.Item) [columns][]string {
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

// bordered reports whether the Board is high enough to draw Cards as boxes.
func (m Model) bordered() bool { return m.height >= borderMinHeight }

// cardLines is a Card's height: a box around three lines, or three lines and a blank one.
func (m Model) cardLines() int {
	if m.bordered() {
		return 5
	}
	return 4
}

// visible is how many Cards fit in a column, leaving one line each for the scroll markers.
func (m Model) visible() int {
	return max(1, (m.height-chromeLines-2)/m.cardLines())
}

func (m Model) View() string {
	if m.help {
		return m.helpView()
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
		body = lipgloss.NewStyle().MaxHeight(max(1, m.height-4)).Render(m.columns()) // too low for one Card: cut it
	}
	footer := keyList(footerKeys)
	if lipgloss.Width(footer) > m.width {
		footer = "h/l j/k ⏎ e s / gg/G ^d/^u 1-5 · r · ? · q " + theme.Faint.Render("close")
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

// helpView renders the key help, framed when the frame fits, else cut to the view.
func (m Model) helpView() string {
	lines := []string{"openspec-kanban " + m.version + " · any key closes"}
	for _, g := range helpGroups {
		lines = append(lines, "", theme.AccentBold.Render(g.title))
		for _, k := range g.keys {
			lines = append(lines, fmt.Sprintf("  %-13s", k.key)+theme.Faint.Render(k.desc))
		}
	}
	text := strings.Join(lines, "\n")
	if lipgloss.Height(text)+2 <= m.height && lipgloss.Width(text)+2 <= m.width {
		return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(theme.Dim).Render(text)
	}
	return lipgloss.NewStyle().MaxWidth(m.width).MaxHeight(m.height).Render(text)
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
	for _, col := range m.all[:work.Archived] {
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
		return m.column(c, m.width, fmt.Sprintf("◀ %d/%d %s %d ▶", c+1, columns, work.Phase(c), len(m.cols[c])))
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
		parts = append(parts, m.column(i, w, fmt.Sprintf("%s %d", work.Phase(i), len(m.cols[i]))))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

// column renders column i at width w under title.
func (m Model) column(i, w int, title string) string {
	colour := lipgloss.NewStyle().Foreground(phaseColor(work.Phase(i)))
	lines := []string{colour.Render(title), colour.Render(strings.Repeat("─", w))}
	cards, off, vis := m.cols[i], m.nav.off[i], m.visible()
	if len(cards) == 0 {
		lines = append(lines, "nothing here")
	}
	if off > 0 {
		lines = append(lines, fmt.Sprintf("↑ %d more", off))
	}
	end := min(len(cards), off+vis)
	for j := off; j < end; j++ {
		lines = append(lines, m.card(cards[j], work.Phase(i), w, i == m.nav.col && j == m.nav.sel[i])...)
	}
	if end < len(cards) {
		lines = append(lines, fmt.Sprintf("↓ %d more", len(cards)-end))
	}
	return lipgloss.NewStyle().Width(w).Render(strings.Join(lines, "\n"))
}

// card renders a box around three lines, the selected one thick in the accent colour; below
// borderMinHeight three lines and a blank one, the selected Card marked with ▌. An open
// Card with a Running agent has a green ● before its name.
func (m Model) card(c work.Item, p work.Phase, w int, selected bool) []string {
	inner := w - 1
	if m.bordered() {
		inner = w - 2
	}
	running := c.Phase != work.Archived && m.running[c.Name]
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
	// each segment is styled apart so one's reset does not end another's style
	first := lipgloss.NewStyle().Bold(true).Render(name)
	if badge != "" {
		first += strings.Repeat(" ", max(1, firstW-lipgloss.Width(name)-len(badge))) + lipgloss.NewStyle().Foreground(theme.Milestone).Render(badge)
	}
	if running {
		first = lipgloss.NewStyle().Foreground(theme.Running).Render("●") + " " + first
	}
	var second, third string
	switch {
	case c.Phase == work.Archived && c.ArchivedOn != "":
		second = "archived " + c.ArchivedOn
		if len(second) > inner { // a 20-cell column has no room for both
			second = c.ArchivedOn
		}
		second = theme.Faint.Render(second)
	case c.Phase == work.Archived:
		second = theme.Faint.Render("archived")
	default:
		second = progressBar(c.Done, c.Total, max(3, min(10, inner-8)), phaseColor(p))
		third = age(m.now().Sub(c.Modified))
		if p == work.Draft || p == work.Ready {
			third = planMarks(c.Artifacts) + "  " + third
		}
		third = theme.Faint.Render(third)
	}
	if m.bordered() {
		border, colour := lipgloss.RoundedBorder(), theme.Dim
		if selected {
			border, colour = lipgloss.ThickBorder(), theme.Accent
		}
		// cut first: Width would wrap a long line into a taller box
		content := lipgloss.NewStyle().MaxWidth(inner).Render(strings.Join([]string{first, second, third}, "\n"))
		box := lipgloss.NewStyle().Border(border).BorderForeground(colour).Width(inner).Render(content)
		return strings.Split(box, "\n")
	}
	mark := " "
	if selected {
		mark = "▌"
	}
	return []string{mark + first, mark + second, mark + third, ""}
}
