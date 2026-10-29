// Package board is the Board UI.
package board

import (
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

// Model shows the Board of one Project; an empty project means none was found.
type Model struct {
	project string
}

func New(project string) Model { return Model{project: project} }

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && (key.String() == "q" || key.Type == tea.KeyEsc || key.Type == tea.KeyCtrlC) {
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) View() string {
	if m.project == "" {
		return "No OpenSpec project here.\n\n" +
			"Open the board from a pane inside a project (a directory containing openspec/changes).\n\n" +
			"q close\n"
	}
	return "OpenSpec Kanban · " + filepath.Base(m.project) + "\n\nq close\n"
}
