package openspec

import (
	"os"
	"path/filepath"

	"github.com/zeotyn/spec-kanban/internal/herdr"
	"github.com/zeotyn/spec-kanban/internal/work"
	"github.com/zeotyn/spec-kanban/internal/worktree"
)

// Source is the OpenSpec work.Source; cli runs the openspec CLI inside a directory.
type Source struct{ cli func(dir string) herdr.Runner }

// New returns the Source that runs the openspec CLI through cli.
func New(cli func(dir string) herdr.Runner) *Source { return &Source{cli} }

// Detect reports whether dir holds openspec/changes.
func (*Source) Detect(dir string) bool {
	fi, err := os.Stat(filepath.Join(dir, "openspec", "changes"))
	return err == nil && fi.IsDir()
}

// Load returns project's Changes as Items; see load for others and known.
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

// Rel is where an open change lives below a checkout.
func (*Source) Rel(change string) string { return rel(change) }

func rel(change string) string { return "openspec/changes/" + change }

// Prompt is the default Start prompt.
func (*Source) Prompt() string { return "/opsx:apply {change}" }

// Finish is the Handoff instruction after review.
func (*Source) Finish(change string) string { return "archive it with /opsx:archive " + change }
