package speckit

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/zeotyn/spec-kanban/internal/herdr"
	"github.com/zeotyn/spec-kanban/internal/work"
	"github.com/zeotyn/spec-kanban/internal/worktree"
)

// fakeGit records git calls (dir, then args) and answers each by the longest matching prefix of
// the joined args: output, or an error when the answer is nil. Unknown calls fail.
type fakeGit struct {
	calls   []string
	answers map[string][]byte
}

func (g *fakeGit) cli(dir string) herdr.Runner {
	return func(args ...string) ([]byte, error) {
		call := strings.Join(args, " ")
		g.calls = append(g.calls, dir+": "+call)
		best, found := "", false
		for prefix := range g.answers {
			if strings.HasPrefix(call, prefix) && len(prefix) >= len(best) {
				best, found = prefix, true
			}
		}
		if !found || g.answers[best] == nil {
			return nil, errors.New("git failed")
		}
		return g.answers[best], nil
	}
}

const fixture = "testdata/project"

// Planning marks as board.planMarks renders them.
var (
	marksS   = map[string]string{"proposal": "skipped", "specs": "done"}                                    // - S · ·
	marksSD  = map[string]string{"proposal": "skipped", "specs": "done", "design": "done"}                  // - S D ·
	marksSDT = map[string]string{"proposal": "skipped", "specs": "done", "design": "done", "tasks": "done"} // - S D T
	marksST  = map[string]string{"proposal": "skipped", "specs": "done", "tasks": "done"}                   // - S · T
)

func TestLoadFixture(t *testing.T) {
	git := &fakeGit{answers: map[string][]byte{"branch --merged": []byte("main\n")}}
	items, err := New(git.cli, "main").Load(fixture, nil, map[string]work.Item{"001-user-login-email": {Name: "001-user-login-email", Phase: work.Archived, Stamp: "s"}})
	if err != nil {
		t.Fatal(err)
	}
	type want struct {
		done, total int
		phase       work.Phase
		marks       map[string]string
	}
	wants := map[string]want{
		"001-user-login-email":       {0, 0, work.Draft, marksS},
		"002-export":                 {0, 0, work.Draft, marksSD},
		"003-search":                 {0, 9, work.Ready, marksSDT},
		"004-reports":                {5, 12, work.InProgress, marksSDT},
		"005-audit":                  {9, 9, work.Review, marksSDT},
		"007-status":                 {3, 3, work.Review, marksST},
		"20261030-142501-user-login": {0, 0, work.Draft, marksS},
	}
	var names []string
	for _, it := range items {
		names = append(names, it.Name)
		w, ok := wants[it.Name]
		if !ok {
			continue
		}
		got := want{it.Done, it.Total, it.Phase, it.Artifacts}
		if !reflect.DeepEqual(got, w) {
			t.Errorf("%s = %+v, want %+v", it.Name, got, w)
		}
		if it.Stamp != "" || it.Dir != "" || it.ArchivedOn != "" || it.Modified.IsZero() {
			t.Errorf("%s: Stamp %q, Dir %q, ArchivedOn %q, Modified %v", it.Name, it.Stamp, it.Dir, it.ArchivedOn, it.Modified)
		}
	}
	if want := []string{"001-user-login-email", "002-export", "003-search", "004-reports", "005-audit", "007-status", "20261030-142501-user-login"}; !slices.Equal(names, want) {
		t.Errorf("names %q, want %q", names, want)
	}
	open, err := New(nil, "").Open(fixture)
	if err != nil || !slices.Equal(open, names) {
		t.Errorf("Open = %q, %v; want %q", open, err, names)
	}
}

func TestLoadModifiedAtAnyDepth(t *testing.T) {
	p := t.TempDir()
	write(t, p, map[string]string{"specs/001-a/spec.md": "x", "specs/001-a/contracts/api/openapi.yaml": "y"})
	old, newest := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	for _, f := range []string{"specs/001-a/spec.md", "specs/001-a/contracts/api", "specs/001-a/contracts", "specs/001-a"} {
		if err := os.Chtimes(filepath.Join(p, f), old, old); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Chtimes(filepath.Join(p, "specs/001-a/contracts/api/openapi.yaml"), newest, newest); err != nil {
		t.Fatal(err)
	}
	items, err := New(nil, "").Load(p, nil, nil)
	if err != nil || len(items) != 1 || !items[0].Modified.Equal(newest) {
		t.Errorf("Load = %+v, %v; want Modified %v", items, err, newest)
	}
}

func TestLoadUnreadableSpecs(t *testing.T) {
	p := t.TempDir()
	if _, err := New(nil, "").Load(p, nil, nil); err == nil {
		t.Error("Load without specs: want an error")
	}
	if _, err := New(nil, "").Open(p); err == nil {
		t.Error("Open without specs: want an error")
	}
}

func TestLoadOtherCheckouts(t *testing.T) {
	p, wt, added, plain := t.TempDir(), t.TempDir(), t.TempDir(), t.TempDir()
	tasks := "## Phase 1: Setup\n\n- [ ] T001 A\n- [ ] T002 B\n- [ ] T003 C\n"
	write(t, p, map[string]string{"specs/001-a/spec.md": "x", "specs/001-a/tasks.md": tasks})
	write(t, wt, map[string]string{"specs/001-a/spec.md": "x", "specs/001-a/tasks.md": strings.Replace(tasks, "[ ] T001", "[x] T001", 1)})
	write(t, added, map[string]string{"specs/004-reports/spec.md": "x"})
	write(t, plain, map[string]string{"specs/001-a/spec.md": "x", "specs/001-a/tasks.md": "- [x] T001 A\n"})
	others := []worktree.Checkout{
		{Path: p + "/", Branch: "001-a"},     // the Project itself
		{Path: wt, Branch: "001-a"},          // replaces
		{Path: added, Branch: "004-reports"}, // adds
		{Path: plain, Branch: "main"},        // not a Feature branch
		{Path: plain, Branch: "002-none"},    // no such folder
	}
	items, err := New(nil, "").Load(p, others, nil)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, it := range items {
		got = append(got, it.Name+" "+it.Phase.String()+" "+it.Dir)
		if it.Name == "001-a" && (it.Done != 1 || it.Total != 3) {
			t.Errorf("001-a progress %d/%d, want 1/3 from the worktree", it.Done, it.Total)
		}
	}
	if want := []string{"001-a In Progress " + wt, "004-reports Draft " + added}; !slices.Equal(got, want) {
		t.Errorf("items %q, want %q", got, want)
	}
}

// merged writes Features "001-review" (all done), "002-ready" (none done) and "003-wip" below a new dir.
func merged(t *testing.T) string {
	p := t.TempDir()
	write(t, p, map[string]string{
		"specs/001-review/spec.md": "x", "specs/001-review/tasks.md": "- [x] T001 A\n",
		"specs/002-ready/spec.md": "x", "specs/002-ready/tasks.md": "- [ ] T001 A\n",
		"specs/003-wip/spec.md": "x", "specs/003-wip/tasks.md": "- [x] T001 A\n- [ ] T002 B\n",
	})
	return p
}

func phases(items []work.Item) map[string]work.Phase {
	m := map[string]work.Phase{}
	for _, it := range items {
		m[it.Name] = it.Phase
		if it.ArchivedOn != "" {
			m[it.Name+" dated"] = it.Phase
		}
	}
	return m
}

func TestMergedRule(t *testing.T) {
	const listed = "main\n001-review\n002-ready\n003-wip\n"
	p := merged(t)
	tests := []struct {
		name, base string
		answers    map[string][]byte
		want       work.Phase
		calls      []string
	}{
		{"merged", "main", map[string][]byte{"branch --merged main --format=%(refname:short)": []byte(listed)}, work.Archived,
			[]string{p + ": branch --merged main --format=%(refname:short)"}},
		{"not merged", "release/2.x", map[string][]byte{"branch --merged": []byte("main\n001-review-old\n")}, work.Review,
			[]string{p + ": branch --merged release/2.x --format=%(refname:short)"}},
		{"git fails", "main", nil, work.Review, []string{p + ": branch --merged main --format=%(refname:short)"}},
		{"detected base", "", map[string][]byte{"rev-parse --verify --quiet refs/heads/main": {}, "branch --merged main": []byte(listed)}, work.Archived,
			[]string{p + ": symbolic-ref --short refs/remotes/origin/HEAD", p + ": rev-parse --verify --quiet refs/heads/main", p + ": branch --merged main --format=%(refname:short)"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			git := &fakeGit{answers: tt.answers}
			items, err := New(git.cli, tt.base).Load(p, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			want := map[string]work.Phase{"001-review": tt.want, "002-ready": work.Ready, "003-wip": work.InProgress}
			if got := phases(items); !reflect.DeepEqual(got, want) {
				t.Errorf("phases %v, want %v", got, want)
			}
			if !slices.Equal(git.calls, tt.calls) {
				t.Errorf("git calls\n got %q\nwant %q", git.calls, tt.calls)
			}
		})
	}
}

func TestMergedRuleNeedsReview(t *testing.T) {
	p := t.TempDir()
	write(t, p, map[string]string{"specs/002-ready/spec.md": "x", "specs/002-ready/tasks.md": "- [ ] T001 A\n", "specs/005-draft/spec.md": "x"})
	git := &fakeGit{}
	if _, err := New(git.cli, "").Load(p, nil, nil); err != nil {
		t.Fatal(err)
	}
	if len(git.calls) != 0 {
		t.Errorf("no Review Feature: git calls %q, want none", git.calls)
	}
}

// TestMergedRuleRealGit runs the merged rule against a real repository.
func TestMergedRuleRealGit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	p := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@example.com", "-c", "commit.gpgsign=false"}, args...)...)
		cmd.Dir = p
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %q: %v\n%s", args, err, out)
		}
	}
	commit := func(files map[string]string, msg string) {
		t.Helper()
		write(t, p, files)
		run("add", "-A")
		run("commit", "-q", "-m", msg)
	}
	run("init", "-q", "-b", "main")
	open3 := "- [ ] T001 A\n- [ ] T002 B\n- [ ] T003 C\n"
	commit(map[string]string{
		"specs/001-merged/spec.md": "x", "specs/001-merged/tasks.md": open3,
		"specs/002-fresh/spec.md": "x", "specs/002-fresh/tasks.md": open3,
		"specs/003-unmerged/spec.md": "x", "specs/003-unmerged/tasks.md": "- [x] T001 A\n",
	}, "specs")
	run("checkout", "-q", "-b", "001-merged")
	commit(map[string]string{"specs/001-merged/tasks.md": strings.ReplaceAll(open3, "[ ]", "[x]")}, "done")
	run("checkout", "-q", "main")
	run("merge", "-q", "--no-ff", "-m", "merge", "001-merged")
	run("branch", "-q", "002-fresh")
	run("checkout", "-q", "-b", "003-unmerged")
	commit(map[string]string{"src.go": "package x\n"}, "work")
	run("checkout", "-q", "main")

	git := func(dir string) herdr.Runner {
		return func(args ...string) ([]byte, error) { return herdr.Run("git", dir, args...) }
	}
	items, err := New(git, "").Load(p, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]work.Phase{"001-merged": work.Archived, "002-fresh": work.Ready, "003-unmerged": work.Review}
	if got := phases(items); !reflect.DeepEqual(got, want) {
		t.Errorf("phases %v, want %v", got, want)
	}
}
