// Package docs reads a Change's documents from disk.
package docs

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/zeotyn/openspec-kanban/internal/openspec"
)

// Dir returns where a Change's files live: openspec/changes/<name> when open, openspec/changes/archive/<ArchivedOn>-<name>
// when archived with a date, openspec/changes/archive/<name> when archived without one.
func Dir(project string, c openspec.Change) string {
	base := filepath.Join(project, "openspec", "changes")
	switch {
	case !c.Archived:
		return filepath.Join(base, c.Name)
	case c.ArchivedOn != "":
		return filepath.Join(base, "archive", c.ArchivedOn+"-"+c.Name)
	}
	return filepath.Join(base, "archive", c.Name)
}

// Docs holds a Change's documents as markdown; a field is "" when the document does not exist.
type Docs struct {
	Tasks, Proposal, Design string
	Specs                   string   // every specs/**/spec.md in lexical path order, each as "# <capability path>\n\n<content>", joined by "\n\n"
	SpecFiles               []string // paths of those spec.md files, same order, for the editor
	Next                    string   // "next ▸ <task text>" for the first "- [ ] " task, "all tasks done" when tasks exist and none is open, "" without tasks
}

var task = regexp.MustCompile(`(?m)^[ \t]*[-*] \[( |x|X)\] (.*)$`)

// Read reads the documents of the Change in dir; missing files and a missing dir give empty fields.
func Read(dir string) Docs {
	read := func(name string) string {
		b, _ := os.ReadFile(filepath.Join(dir, name))
		return string(b)
	}
	d := Docs{Tasks: read("tasks.md"), Proposal: read("proposal.md"), Design: read("design.md")}

	specs := filepath.Join(dir, "specs")
	var parts []string
	filepath.WalkDir(specs, func(path string, e fs.DirEntry, err error) error {
		if err != nil || e.IsDir() || e.Name() != "spec.md" {
			return nil // skip what cannot be read; one bad folder must not hide the other specs
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(specs, filepath.Dir(path))
		parts = append(parts, "# "+filepath.ToSlash(rel)+"\n\n"+string(b))
		d.SpecFiles = append(d.SpecFiles, path)
		return nil
	})
	d.Specs = strings.Join(parts, "\n\n")

	for _, m := range task.FindAllStringSubmatch(d.Tasks, -1) {
		if m[1] == " " {
			d.Next = "next ▸ " + strings.TrimSpace(m[2])
			return d
		}
		d.Next = "all tasks done"
	}
	return d
}
