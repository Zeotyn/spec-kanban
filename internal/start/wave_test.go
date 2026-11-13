package start

import (
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/zeotyn/spec-kanban/internal/work"
)

func TestCoordinatorPrompt(t *testing.T) {
	got := CoordinatorPrompt("/opt/kanban/bin/kanban", 1, "/wt/wave-1", "main", []string{"f5-bar", "c4-foo"}, src)
	for _, want := range []string{
		"wave 1", "branch wave-1", "worktree /wt/wave-1", "base main",
		"f5-bar (/wt/wave-1/openspec/changes/f5-bar/proposal.md)", "c4-foo (/wt/wave-1/openspec/changes/c4-foo/proposal.md)",
		"/opt/kanban/bin/kanban wave start 1 <change>", "/opt/kanban/bin/kanban wave drop 1 <change>",
		`/opt/kanban/bin/kanban wave escalate 1 "<what you need>"`, "/opt/kanban/bin/kanban wave finish 1",
		"merged into wave-1", "right away", "resolve every conflict yourself", "run the Project's tests", "/opsx:archive <change>",
		"never with --force", "keep its branch", "Drop a member", "final report", "only when you cannot decide",
		"archived on wave-1 or dropped",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("prompt lacks %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "\n") {
		t.Errorf("prompt spans lines:\n%s", got)
	}
	if q := CoordinatorPrompt("/my apps/kanban", 2, "/wt", "main", []string{"a", "b"}, src); !strings.Contains(q, `'/my apps/kanban' wave finish 2`) {
		t.Errorf("binary with a space is not quoted:\n%s", q)
	}
}

func TestWave(t *testing.T) {
	items := []work.Item{{Name: "f5-bar", Phase: work.Ready}, {Name: "c4-foo", Phase: work.Ready}, {Name: "m9-x", Phase: work.InProgress}}
	members := []string{"f5-bar", "c4-foo"}
	create := "worktree create --cwd " + proj + " --branch wave-1 --base main --label wave-1 --no-focus"
	path := "/Users/U758014/.herdr/worktrees/spec-kanban/m1-board-view" // worktree-create.json
	coord := "agent start wave-1 --kind claude --pane w3:p1 --timeout 60000"
	text := CoordinatorPrompt("/opt/kanban", 1, path, "main", members, src)
	committed := map[string]string{"cat-file -e main:openspec/changes/f5-bar": "", "cat-file -e main:openspec/changes/c4-foo": ""}
	t.Run("started", func(t *testing.T) {
		h := &fakeHerdr{t: t, out: map[string]string{"worktree create": "worktree-create.json"}}
		var recorded string
		forgot := false
		err := Wave(proj, 1, members, items, "", "claude", "/opt/kanban", h.run, src, gitWith(committed, proj, "main").run,
			func(pane, path, base string) error {
				recorded = pane + " " + path + " " + base + " after " + h.calls[len(h.calls)-1]
				return nil
			}, func() { forgot = true })
		if err != nil {
			t.Fatal(err)
		}
		want := []string{create, `pane run w3:p1  echo kanban-rea""dy-n`, "pane wait-output w3:p1 --match kanban-ready-n --timeout 2000",
			coord, "agent prompt wave-1 " + text, "tab focus w3:t1"}
		if !reflect.DeepEqual(h.calls, want) {
			t.Errorf("herdr calls\n got %q\nwant %q", h.calls, want)
		}
		if want := "w3:p1 " + path + " main after " + create; recorded != want || forgot {
			t.Errorf("recorded %q (forgot %v), want %q: before the coordinator can run kanban wave", recorded, forgot, want)
		}
	})
	t.Run("member not in Ready", func(t *testing.T) {
		h := &fakeHerdr{t: t}
		err := Wave(proj, 1, []string{"f5-bar", "m9-x"}, items, "", "claude", "/opt/kanban", h.run, src, gitWith(committed, proj, "main").run,
			func(string, string, string) error { t.Error("recorded"); return nil }, func() {})
		if err == nil || err.Error() != "m9-x is not in Ready" || len(h.calls) != 0 {
			t.Errorf("got %v, calls %q", err, h.calls)
		}
	})
	t.Run("member not open", func(t *testing.T) {
		err := Wave(proj, 1, []string{"f5-bar", "gone"}, items, "", "claude", "/opt/kanban", (&fakeHerdr{t: t}).run, src, gitWith(committed, proj, "main").run,
			func(string, string, string) error { return nil }, func() {})
		if err == nil || err.Error() != `"gone" is not an open change of `+filepath.Base(proj) {
			t.Errorf("got %v", err)
		}
	})
	t.Run("worktree fails", func(t *testing.T) {
		h := &fakeHerdr{t: t, fail: map[string]error{"worktree create": errors.New("branch exists")}}
		err := Wave(proj, 1, members, items, "", "claude", "/opt/kanban", h.run, src, gitWith(committed, proj, "main").run,
			func(string, string, string) error { t.Error("recorded"); return nil }, func() {})
		if err == nil || err.Error() != "worktree create: branch exists" {
			t.Errorf("got %v", err)
		}
	})
	t.Run("coordinator fails: record forgotten", func(t *testing.T) {
		h := &fakeHerdr{t: t, out: map[string]string{"worktree create": "worktree-create.json", "pane read": "pane-read.txt"}, fail: map[string]error{"agent start": errors.New("timeout")}}
		forgot := false
		err := Wave(proj, 1, members, items, "", "claude", "/opt/kanban", h.run, src, gitWith(committed, proj, "main").run,
			func(string, string, string) error { return nil }, func() { forgot = true })
		if err == nil || !strings.Contains(err.Error(), `claude did not become ready in tab "wave-1"`) || !forgot {
			t.Errorf("got %v, forgot %v", err, forgot)
		}
	})
	t.Run("member not on the base", func(t *testing.T) {
		h := &fakeHerdr{t: t}
		err := Wave(proj, 1, members, items, "", "claude", "/opt/kanban", h.run, src, gitWith(map[string]string{"cat-file -e main:openspec/changes/f5-bar": ""}, proj, "main").run,
			func(string, string, string) error { t.Error("recorded"); return nil }, func() {})
		if err == nil || err.Error() != "c4-foo is not on main: commit it there or set base" || len(h.calls) != 0 {
			t.Errorf("got %v, calls %q", err, h.calls)
		}
	})
	t.Run("record fails", func(t *testing.T) {
		h := &fakeHerdr{t: t, out: map[string]string{"worktree create": "worktree-create.json"}}
		err := Wave(proj, 1, members, items, "", "claude", "/opt/kanban", h.run, src, gitWith(committed, proj, "main").run,
			func(string, string, string) error { return errors.New("disk full") }, func() {})
		if err == nil || err.Error() != "record wave 1: disk full" || len(h.calls) != 1 {
			t.Errorf("got %v, calls %q", err, h.calls)
		}
	})
}
