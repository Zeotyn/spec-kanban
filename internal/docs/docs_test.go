package docs

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/zeotyn/openspec-kanban/internal/openspec"
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

func TestDir(t *testing.T) {
	for _, tc := range []struct {
		c    openspec.Change
		want string
	}{
		{openspec.Change{Name: "m1-board"}, "/p/openspec/changes/m1-board"},
		{openspec.Change{Name: "m0-skel", Archived: true, ArchivedOn: "2026-10-29"}, "/p/openspec/changes/archive/2026-10-29-m0-skel"},
		{openspec.Change{Name: "m0-skel", Archived: true}, "/p/openspec/changes/archive/m0-skel"},
		{openspec.Change{Name: "m1-board", Dir: "/wt/m1-board"}, "/wt/m1-board/openspec/changes/m1-board"},
	} {
		if got := Dir("/p", tc.c); got != tc.want {
			t.Errorf("Dir(%+v) = %q, want %q", tc.c, got, tc.want)
		}
	}
}

func TestReadAll(t *testing.T) {
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
	got := Read(dir)
	want := Docs{
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
		t.Errorf("Read =\n%#v\nwant\n%#v", got, want)
	}
}

func TestReadMissing(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, map[string]string{"proposal.md": "p"})
	if got, want := Read(dir), (Docs{Proposal: "p"}); !reflect.DeepEqual(got, want) {
		t.Errorf("Read = %#v, want %#v", got, want)
	}
	if got := Read(filepath.Join(dir, "nope")); !reflect.DeepEqual(got, Docs{}) {
		t.Errorf("Read(missing dir) = %#v, want zero", got)
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
			if got := Read(dir).Next; got != tc.want {
				t.Errorf("Next = %q, want %q", got, tc.want)
			}
		})
	}
}
