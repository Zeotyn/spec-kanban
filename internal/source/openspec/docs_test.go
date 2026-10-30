package openspec

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/zeotyn/spec-kanban/internal/work"
)

func write(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestChangeDir(t *testing.T) {
	for _, tc := range []struct {
		c    work.Item
		want string
	}{
		{work.Item{Name: "m1-board"}, "/p/openspec/changes/m1-board"},
		{work.Item{Name: "m0-skel", Phase: work.Archived, ArchivedOn: "2026-10-29"}, "/p/openspec/changes/archive/2026-10-29-m0-skel"},
		{work.Item{Name: "m0-skel", Phase: work.Archived}, "/p/openspec/changes/archive/m0-skel"},
		{work.Item{Name: "m1-board", Dir: "/wt/m1-board"}, "/wt/m1-board/openspec/changes/m1-board"},
	} {
		if got := changeDir("/p", tc.c); got != tc.want {
			t.Errorf("changeDir(%+v) = %q, want %q", tc.c, got, tc.want)
		}
	}
}

func TestReadDirAll(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, map[string]string{
		"tasks.md":            "- [x] 1.1 Done\n- [ ] 1.2 Open\n",
		"proposal.md":         "# Proposal\n",
		"design.md":           "# Design\n",
		"specs/board/spec.md": "board spec\n",
		"specs/agent-control/session-wait/spec.md": "wait spec\n",
		"specs/agent-control/spec.md":              "control spec\n",
		"specs/board/notes.md":                     "not a spec\n",
	})
	got := readDir(dir)
	want := changeDocs{
		Tasks:    "- [x] 1.1 Done\n- [ ] 1.2 Open\n",
		Proposal: "# Proposal\n",
		Design:   "# Design\n",
		Specs: "# agent-control/session-wait\n\nwait spec\n\n\n" +
			"# agent-control\n\ncontrol spec\n\n\n" +
			"# board\n\nboard spec\n",
		SpecFiles: []string{
			filepath.Join(dir, "specs/agent-control/session-wait/spec.md"),
			filepath.Join(dir, "specs/agent-control/spec.md"),
			filepath.Join(dir, "specs/board/spec.md"),
		},
		Next: "next ▸ 1.2 Open",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("readDir =\n%#v\nwant\n%#v", got, want)
	}
}

func TestReadDirMissing(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, map[string]string{"proposal.md": "p"})
	if got, want := readDir(dir), (changeDocs{Proposal: "p"}); !reflect.DeepEqual(got, want) {
		t.Errorf("readDir = %#v, want %#v", got, want)
	}
	if got := readDir(filepath.Join(dir, "nope")); !reflect.DeepEqual(got, changeDocs{}) {
		t.Errorf("readDir(missing dir) = %#v, want zero", got)
	}
}

func TestNext(t *testing.T) {
	for _, tc := range []struct{ name, tasks, want string }{
		{"first open", "## 1\n- [x] 1.1 A\n- [x] 1.2 B\n\n## 2\n- [x] 2.1 C\n- [ ] 2.3 Implement the navigation state  \n- [ ] 2.4 D\n", "next ▸ 2.3 Implement the navigation state"},
		{"all done", "- [x] 1.1 A\n- [x] 1.2 B\n", "all tasks done"},
		{"no tasks", "# Tasks\n\nnothing yet\n", ""},
		{"uppercase X", "- [X] 1.1 A\n- [x] 1.2 B\n", "all tasks done"},
		{"indented and star", "- [x] 1 A\n  * [x] 1.1 B\n    - [ ] 1.2 C\n", "next ▸ 1.2 C"},
		{"star open", "* [ ] 3.1 E\n", "next ▸ 3.1 E"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			write(t, dir, map[string]string{"tasks.md": tc.tasks})
			if got := readDir(dir).Next; got != tc.want {
				t.Errorf("Next = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestReadDocs(t *testing.T) {
	project := t.TempDir()
	dir := filepath.Join(project, "openspec", "changes", "m1-board")
	write(t, dir, map[string]string{
		"tasks.md":            "- [ ] 1.1 Open\n",
		"proposal.md":         "# Proposal\n",
		"specs/board/spec.md": "board spec\n",
	})
	got := New(nil).Docs(project, work.Item{Name: "m1-board"})
	want := work.Docs{
		Tabs: []work.Doc{
			{Title: "Tasks", Body: "- [ ] 1.1 Open\n", Files: []string{filepath.Join(dir, "tasks.md")}},
			{Title: "Proposal", Body: "# Proposal\n", Files: []string{filepath.Join(dir, "proposal.md")}},
			{Title: "Design", Files: []string{filepath.Join(dir, "design.md")}},
			{Title: "Specs", Body: "# board\n\nboard spec\n", Files: []string{filepath.Join(dir, "specs/board/spec.md")}},
		},
		Tasks: 0,
		First: 1,
		Next:  "next ▸ 1.1 Open",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ReadDocs =\n%#v\nwant\n%#v", got, want)
	}
}
