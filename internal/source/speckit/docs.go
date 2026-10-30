package speckit

import (
	"os"
	"path/filepath"

	"github.com/zeotyn/spec-kanban/internal/work"
)

// Docs returns the documents of it, read in its checkout, as the detail view's tabs: Tasks, Spec,
// then Plan and Research when they exist.
func (*Source) Docs(project string, it work.Item) work.Docs {
	if it.Dir != "" {
		project = it.Dir
	}
	dir := filepath.Join(project, "specs", it.Name)
	read := func(title, name string) (work.Doc, bool) {
		p := filepath.Join(dir, name)
		b, err := os.ReadFile(p)
		return work.Doc{Title: title, Body: string(b), Files: []string{p}}, err == nil
	}
	tasks, _ := read("Tasks", "tasks.md")
	spec, _ := read("Spec", "spec.md")
	d := work.Docs{Tabs: []work.Doc{tasks, spec}, Tasks: 0, First: 1}
	for _, f := range [][2]string{{"Plan", "plan.md"}, {"Research", "research.md"}} {
		if doc, ok := read(f[0], f[1]); ok {
			d.Tabs = append(d.Tabs, doc)
		}
	}
	_, _, d.Next = scan(tasks.Body)
	return d
}
