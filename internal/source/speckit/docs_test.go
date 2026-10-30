package speckit

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/zeotyn/spec-kanban/internal/work"
)

func TestDocsAll(t *testing.T) {
	p := t.TempDir()
	dir := filepath.Join(p, "specs", "001-a")
	tasks := "## Phase 1: Setup\n\n- [x] T001 A\n- [X] T002 B\n- [x] T003 C\n- [ ] T004 [P] [US1] Create model  \n- [ ] T005 D\n"
	write(t, dir, map[string]string{"tasks.md": tasks, "spec.md": "# Spec\n", "plan.md": "# Plan\n", "research.md": "# Research\n"})
	file := func(name string) []string { return []string{filepath.Join(dir, name)} }
	want := work.Docs{
		Tabs: []work.Doc{
			{Title: "Tasks", Body: tasks, Files: file("tasks.md")},
			{Title: "Spec", Body: "# Spec\n", Files: file("spec.md")},
			{Title: "Plan", Body: "# Plan\n", Files: file("plan.md")},
			{Title: "Research", Body: "# Research\n", Files: file("research.md")},
		},
		Tasks: 0,
		First: 1,
		Next:  "next ▸ T004 [P] [US1] Create model",
	}
	if got := New(nil, "").Docs(p, work.Item{Name: "001-a"}); !reflect.DeepEqual(got, want) {
		t.Errorf("Docs =\n%#v\nwant\n%#v", got, want)
	}
	write(t, dir, map[string]string{"tasks.md": strings.ReplaceAll(tasks, "[ ]", "[x]")})
	if got := New(nil, "").Docs(p, work.Item{Name: "001-a"}).Next; got != "all tasks done" {
		t.Errorf("all done: Next = %q", got)
	}
}

func TestDocsSpecOnly(t *testing.T) {
	wt := t.TempDir()
	dir := filepath.Join(wt, "specs", "001-a")
	write(t, dir, map[string]string{"spec.md": "# Spec\n"})
	want := work.Docs{
		Tabs: []work.Doc{
			{Title: "Tasks", Files: []string{filepath.Join(dir, "tasks.md")}},
			{Title: "Spec", Body: "# Spec\n", Files: []string{filepath.Join(dir, "spec.md")}},
		},
		Tasks: 0,
		First: 1,
	}
	if got := New(nil, "").Docs("/elsewhere", work.Item{Name: "001-a", Dir: wt}); !reflect.DeepEqual(got, want) {
		t.Errorf("Docs from the checkout =\n%#v\nwant\n%#v", got, want)
	}
}

func TestDocsPlanWithoutResearch(t *testing.T) {
	p := t.TempDir()
	write(t, p, map[string]string{"specs/001-a/spec.md": "s", "specs/001-a/plan.md": ""})
	var titles []string
	for _, d := range New(nil, "").Docs(p, work.Item{Name: "001-a"}).Tabs {
		titles = append(titles, d.Title)
	}
	if want := []string{"Tasks", "Spec", "Plan"}; !reflect.DeepEqual(titles, want) {
		t.Errorf("tabs %q, want %q", titles, want)
	}
}
