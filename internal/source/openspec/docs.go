package openspec

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/zeotyn/openspec-kanban/internal/work"
)

// changeDir returns where a Change's files live: openspec/changes/<name> when open, openspec/changes/archive/<ArchivedOn>-<name>
// when archived with a date, openspec/changes/archive/<name> when archived without one; below c.Dir
// instead of project when set.
func changeDir(project string, c work.Item) string {
	if c.Dir != "" {
		project = c.Dir
	}
	base := filepath.Join(project, "openspec", "changes")
	switch {
	case c.Phase != work.Archived:
		return filepath.Join(base, c.Name)
	case c.ArchivedOn != "":
		return filepath.Join(base, "archive", c.ArchivedOn+"-"+c.Name)
	}
	return filepath.Join(base, "archive", c.Name)
}

// changeDocs holds a Change's documents as markdown; a field is "" when the document does not exist.
type changeDocs struct {
	Tasks, Proposal, Design string
	Specs                   string   // every specs/**/spec.md in lexical path order, each as "# <capability path>\n\n<content>", joined by "\n\n"
	SpecFiles               []string // paths of those spec.md files, same order, for the editor
	Next                    string   // "next ▸ <task text>" for the first "- [ ] " task, "all tasks done" when tasks exist and none is open, "" without tasks
}

var task = regexp.MustCompile(`(?m)^[ \t]*[-*] \[( |x|X)\] (.*)$`)

// readDir reads the documents of the Change in dir; missing files and a missing dir give empty fields.
func readDir(dir string) changeDocs {
	read := func(name string) string {
		b, _ := os.ReadFile(filepath.Join(dir, name))
		return string(b)
	}
	d := changeDocs{Tasks: read("tasks.md"), Proposal: read("proposal.md"), Design: read("design.md")}

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

// ReadDocs returns the documents of it as the detail view's tabs: Tasks, Proposal, Design, Specs.
func ReadDocs(project string, it work.Item) work.Docs {
	dir := changeDir(project, it)
	d := readDir(dir)
	file := func(name string) []string { return []string{filepath.Join(dir, name)} }
	return work.Docs{
		Tabs: []work.Doc{
			{Title: "Tasks", Body: d.Tasks, Files: file("tasks.md")},
			{Title: "Proposal", Body: d.Proposal, Files: file("proposal.md")},
			{Title: "Design", Body: d.Design, Files: file("design.md")},
			{Title: "Specs", Body: d.Specs, Files: d.SpecFiles},
		},
		Tasks: 0,
		First: 1,
		Next:  d.Next,
	}
}
