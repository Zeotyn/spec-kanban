package openspec

import (
	"os"
	"path/filepath"

	"github.com/zeotyn/openspec-kanban/internal/herdr"
	"github.com/zeotyn/openspec-kanban/internal/work"
	"github.com/zeotyn/openspec-kanban/internal/worktree"
)

// Source is the OpenSpec work.Source; cli runs the openspec CLI inside a directory.
type Source struct{ cli func(dir string) herdr.Runner }

func New(cli func(dir string) herdr.Runner) *Source { return &Source{cli} }

// Detect reports whether dir holds openspec/changes.
func (*Source) Detect(dir string) bool {
	fi, err := os.Stat(filepath.Join(dir, "openspec", "changes"))
	return err == nil && fi.IsDir()
}

func (s *Source) Load(project string, others []worktree.Checkout, known map[string]work.Item) ([]work.Item, error) {
	return load(project, s.cli, others, known)
}

// Open returns the names of the open Changes.
func (s *Source) Open(project string) ([]string, error) {
	changes, err := list(s.cli(project))
	if err != nil {
		return nil, err
	}
	names := make([]string, len(changes))
	for i, c := range changes {
		names[i] = c.Name
	}
	return names, nil
}

// ArchivedIn reports whether the archive of the checkout dir holds change, dated or not.
func (*Source) ArchivedIn(dir, change string) bool {
	archived, _ := archive(dir)
	return isIn(archived, change)
}

func (*Source) Rel(change string) string { return "openspec/changes/" + change }

func (*Source) Prompt() string { return "/opsx:apply {change}" }

func (*Source) Finish(change string) string { return "archive it with /opsx:archive " + change }
