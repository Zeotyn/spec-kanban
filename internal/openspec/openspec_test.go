package openspec

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/zeotyn/openspec-kanban/internal/herdr"
	"github.com/zeotyn/openspec-kanban/internal/worktree"
)

// one runs every directory's CLI through run.
func one(run herdr.Runner) func(string) herdr.Runner {
	return func(string) herdr.Runner { return run }
}

// fakeCLI answers `list --json` with list and `status --change <name> --json` from status;
// a missing status entry fails like the real CLI does for an unknown change.
type fakeCLI struct {
	list      string
	noVersion bool // --version fails
	status    map[string]string
	mu        sync.Mutex
	calls     []string
}

func (f *fakeCLI) run(args ...string) ([]byte, error) {
	call := strings.Join(args, " ")
	f.mu.Lock()
	f.calls = append(f.calls, call)
	f.mu.Unlock()
	if call == "list --json" {
		return []byte(f.list), nil
	}
	if call == "--version" && !f.noVersion {
		return []byte("1.8.0\n"), nil
	}
	if len(args) == 4 && args[0] == "status" {
		if out, ok := f.status[args[2]]; ok {
			return []byte(out), nil
		}
		return nil, errors.New("exit status 1: change_error")
	}
	return nil, fmt.Errorf("unexpected call %q", call)
}

const planned = `{"isPlanningComplete":true,"artifacts":[{"id":"proposal","status":"done"},{"id":"specs","status":"done"},{"id":"design","status":"done"},{"id":"tasks","status":"done"}]}`
const proposalOnly = `{"isPlanningComplete":false,"artifacts":[{"id":"proposal","status":"done"},{"id":"specs","status":"ready"},{"id":"design","status":"ready"},{"id":"tasks","status":"blocked"}]}`

func TestLoadOpenChanges(t *testing.T) {
	cli := &fakeCLI{
		list: `{"changes":[
			{"name":"started","completedTasks":3,"totalTasks":5,"lastModified":"2026-10-29T08:56:18.587Z","status":"in-progress"},
			{"name":"draft","completedTasks":0,"totalTasks":0,"lastModified":"2026-10-28T10:00:00Z","status":"no-tasks"},
			{"name":"ready","completedTasks":0,"totalTasks":4,"lastModified":"2026-10-27T10:00:00Z","status":"in-progress"},
			{"name":"broken","completedTasks":0,"totalTasks":2,"lastModified":"2026-10-26T10:00:00Z","status":"in-progress"}]}`,
		status: map[string]string{"draft": proposalOnly, "ready": planned},
	}
	got, err := Load(t.TempDir(), one(cli.run), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := []Change{
		{Name: "started", Done: 3, Total: 5, Modified: time.Date(2026, 10, 29, 8, 56, 18, 587e6, time.UTC)},
		{Name: "draft", Modified: time.Date(2026, 10, 28, 10, 0, 0, 0, time.UTC), Planning: Incomplete,
			Artifacts: map[string]string{"proposal": "done", "specs": "ready", "design": "ready", "tasks": "blocked"}},
		{Name: "ready", Total: 4, Modified: time.Date(2026, 10, 27, 10, 0, 0, 0, time.UTC), Planning: Complete,
			Artifacts: map[string]string{"proposal": "done", "specs": "done", "design": "done", "tasks": "done"}},
		{Name: "broken", Total: 2, Modified: time.Date(2026, 10, 26, 10, 0, 0, 0, time.UTC)},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Load\n got %+v\nwant %+v", got, want)
	}
	slices.Sort(cli.calls)
	wantCalls := []string{"--version", "list --json", "status --change broken --json", "status --change draft --json", "status --change ready --json"}
	if !reflect.DeepEqual(cli.calls, wantCalls) {
		t.Errorf("calls\n got %q\nwant %q", cli.calls, wantCalls)
	}
}

// withChanges creates a project whose Changes each hold proposal.md and design.md.
func withChanges(t *testing.T, names ...string) string {
	t.Helper()
	project := t.TempDir()
	for _, n := range names {
		dir := filepath.Join(project, "openspec", "changes", n)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		for _, f := range []string{"proposal.md", "design.md"} {
			if err := os.WriteFile(filepath.Join(dir, f), nil, 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	return project
}

func fingerprint(project, name string) string {
	return files(filepath.Join(project, "openspec", "changes", name), setup(project, "1.8.0"))
}

func TestLoadReusesKnownPlanning(t *testing.T) {
	cli := &fakeCLI{
		list: `{"changes":[
			{"name":"same","completedTasks":0,"totalTasks":4,"lastModified":"2026-10-28T10:00:00.123Z"},
			{"name":"edited","completedTasks":0,"totalTasks":0,"lastModified":"2026-10-28T11:00:00Z"},
			{"name":"failed","completedTasks":0,"totalTasks":4,"lastModified":"2026-10-28T12:00:00Z"},
			{"name":"started","completedTasks":1,"totalTasks":4,"lastModified":"2026-10-28T13:00:00Z"}]}`,
		status: map[string]string{"edited": proposalOnly, "failed": planned},
	}
	project := withChanges(t, "same", "edited", "failed", "started")
	fp := func(n string) string { return fingerprint(project, n) }
	done := map[string]string{"proposal": "done", "specs": "done", "design": "done", "tasks": "done"}
	known := map[string]Change{
		"same":    {Name: "same", Modified: time.Date(2026, 10, 28, 10, 0, 0, 123e6, time.UTC), Planning: Complete, Artifacts: done, Files: fp("same")},
		"edited":  {Name: "edited", Modified: time.Date(2026, 10, 27, 11, 0, 0, 0, time.UTC), Planning: Complete, Artifacts: done, Files: fp("edited")},
		"failed":  {Name: "failed", Modified: time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC), Files: fp("failed")},
		"started": {Name: "started", Modified: time.Date(2026, 10, 28, 13, 0, 0, 0, time.UTC), Planning: Complete, Artifacts: done},
	}
	got, err := Load(project, one(cli.run), nil, known)
	if err != nil {
		t.Fatal(err)
	}
	want := []Change{
		{Name: "same", Total: 4, Modified: time.Date(2026, 10, 28, 10, 0, 0, 123e6, time.UTC), Planning: Complete, Artifacts: done, Files: fp("same")},
		{Name: "edited", Modified: time.Date(2026, 10, 28, 11, 0, 0, 0, time.UTC), Planning: Incomplete,
			Artifacts: map[string]string{"proposal": "done", "specs": "ready", "design": "ready", "tasks": "blocked"}, Files: fp("edited")},
		{Name: "failed", Total: 4, Modified: time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC), Planning: Complete, Artifacts: done, Files: fp("failed")},
		{Name: "started", Done: 1, Total: 4, Modified: time.Date(2026, 10, 28, 13, 0, 0, 0, time.UTC)},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Load\n got %+v\nwant %+v", got, want)
	}
	slices.Sort(cli.calls)
	wantCalls := []string{"--version", "list --json", "status --change edited --json", "status --change failed --json"}
	if !reflect.DeepEqual(cli.calls, wantCalls) {
		t.Errorf("calls\n got %q\nwant %q", cli.calls, wantCalls)
	}
}

func TestLoadUnchangedCallsOnlyList(t *testing.T) {
	cli := &fakeCLI{list: `{"changes":[{"name":"a","lastModified":"2026-10-28T10:00:00Z"},{"name":"b","lastModified":"2026-10-28T11:00:00Z"}]}`}
	project := withChanges(t, "a", "b")
	known := map[string]Change{
		"a": {Modified: time.Date(2026, 10, 28, 10, 0, 0, 0, time.UTC), Planning: Incomplete, Artifacts: map[string]string{"proposal": "ready"}, Files: fingerprint(project, "a")},
		"b": {Modified: time.Date(2026, 10, 28, 11, 0, 0, 0, time.UTC), Planning: Complete, Artifacts: map[string]string{"proposal": "done"}, Files: fingerprint(project, "b")},
	}
	got, err := Load(project, one(cli.run), nil, known)
	if err != nil {
		t.Fatal(err)
	}
	if slices.Sort(cli.calls); !reflect.DeepEqual(cli.calls, []string{"--version", "list --json"}) {
		t.Errorf("calls %q, want only version and list", cli.calls)
	}
	for _, c := range got {
		if c.Planning != known[c.Name].Planning || !reflect.DeepEqual(c.Artifacts, known[c.Name].Artifacts) {
			t.Errorf("%s: got %v %v, want %v %v", c.Name, c.Planning, c.Artifacts, known[c.Name].Planning, known[c.Name].Artifacts)
		}
	}
}

func TestLoadDoesNotReuse(t *testing.T) {
	tests := []struct {
		name   string
		list   string
		modify func(project, dir string) error
		known  func(fp string) Change
	}{
		{"older artifact deleted, newest time unchanged", `{"changes":[{"name":"a","lastModified":"2026-10-28T10:00:00Z"}]}`,
			func(_, dir string) error { return os.Remove(filepath.Join(dir, "design.md")) },
			func(fp string) Change {
				return Change{Modified: time.Date(2026, 10, 28, 10, 0, 0, 0, time.UTC), Planning: Complete, Files: fp}
			}},
		{"artifact renamed", `{"changes":[{"name":"a","lastModified":"2026-10-28T10:00:00Z"}]}`,
			func(_, dir string) error {
				return os.Rename(filepath.Join(dir, "design.md"), filepath.Join(dir, "design-old.md"))
			},
			func(fp string) Change {
				return Change{Modified: time.Date(2026, 10, 28, 10, 0, 0, 0, time.UTC), Planning: Complete, Files: fp}
			}},
		{"config edited", `{"changes":[{"name":"a","lastModified":"2026-10-28T10:00:00Z"}]}`,
			func(project, _ string) error {
				return os.WriteFile(filepath.Join(project, "openspec", "config.yaml"), []byte("schema: custom\n"), 0o644)
			},
			func(fp string) Change {
				return Change{Modified: time.Date(2026, 10, 28, 10, 0, 0, 0, time.UTC), Planning: Complete, Files: fp}
			}},
		{"no modification time", `{"changes":[{"name":"a"}]}`,
			func(string, string) error { return nil },
			func(fp string) Change { return Change{Planning: Complete, Files: fp} }},
		{"no fingerprint", `{"changes":[{"name":"a","lastModified":"2026-10-28T10:00:00Z"}]}`,
			func(string, string) error { return nil },
			func(string) Change {
				return Change{Modified: time.Date(2026, 10, 28, 10, 0, 0, 0, time.UTC), Planning: Complete}
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project := withChanges(t, "a")
			known := map[string]Change{"a": tt.known(fingerprint(project, "a"))}
			if err := tt.modify(project, filepath.Join(project, "openspec", "changes", "a")); err != nil {
				t.Fatal(err)
			}
			cli := &fakeCLI{list: tt.list, status: map[string]string{"a": proposalOnly}}
			got, err := Load(project, one(cli.run), nil, known)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Contains(cli.calls, "status --change a --json") || got[0].Planning != Incomplete {
				t.Errorf("want a fresh status call, got calls %q, planning %v", cli.calls, got[0].Planning)
			}
		})
	}
}

func TestLoadNoChanges(t *testing.T) {
	got, err := Load(t.TempDir(), one((&fakeCLI{list: `{"changes":[]}`}).run), nil, nil)
	if err != nil || len(got) != 0 {
		t.Fatalf("want no changes and no error, got %v, %v", got, err)
	}
}

func TestLoadFailures(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"CLI not installed", fmt.Errorf("openspec list --json: %w", exec.ErrNotFound), "openspec CLI not found"},
		{"list fails", errors.New("exit status 1: no openspec root"), "no openspec root"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run := func(...string) ([]byte, error) { return nil, tt.err }
			if _, err := Load(t.TempDir(), one(run), nil, nil); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("want error with %q, got %v", tt.want, err)
			}
		})
	}
}

func TestLoadArchived(t *testing.T) {
	project := t.TempDir()
	archive := filepath.Join(project, "openspec", "changes", "archive")
	for _, dir := range []string{"2026-10-29-m0-plugin-skeleton", "undated-change"} {
		if err := os.MkdirAll(filepath.Join(archive, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(archive, ".DS_Store"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Load(project, one((&fakeCLI{list: `{"changes":[]}`}).run), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := []Change{
		{Name: "m0-plugin-skeleton", Archived: true, ArchivedOn: "2026-10-29"},
		{Name: "undated-change", Archived: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("archived\n got %+v\nwant %+v", got, want)
	}
}

func TestOpenNames(t *testing.T) {
	cli := &fakeCLI{list: `{"changes":[{"name":"a"},{"name":"b"}]}`}
	got, err := OpenNames(cli.run)
	if err != nil || !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestLoadAfterUpgrade(t *testing.T) {
	project := withChanges(t, "a")
	old := files(filepath.Join(project, "openspec", "changes", "a"), setup(project, "1.7.0"))
	known := map[string]Change{"a": {Modified: time.Date(2026, 10, 28, 10, 0, 0, 0, time.UTC), Planning: Complete, Files: old}}
	cli := &fakeCLI{list: `{"changes":[{"name":"a","lastModified":"2026-10-28T10:00:00Z"}]}`, status: map[string]string{"a": proposalOnly}}
	if _, err := Load(project, one(cli.run), nil, known); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(cli.calls, "status --change a --json") {
		t.Errorf("an openspec upgrade must re-read planning, calls %q", cli.calls)
	}
}

func TestLoadWithoutVersion(t *testing.T) {
	project := withChanges(t, "a")
	known := map[string]Change{"a": {Modified: time.Date(2026, 10, 28, 10, 0, 0, 0, time.UTC), Planning: Complete,
		Files: files(filepath.Join(project, "openspec", "changes", "a"), setup(project, ""))}}
	cli := &fakeCLI{list: `{"changes":[{"name":"a","lastModified":"2026-10-28T10:00:00Z"}]}`, status: map[string]string{"a": proposalOnly}, noVersion: true}
	got, err := Load(project, one(cli.run), nil, known)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(cli.calls, "status --change a --json") || got[0].Files != "" {
		t.Errorf("without a CLI version nothing may be reused or fingerprinted, calls %q, files %q", cli.calls, got[0].Files)
	}
}

func TestLoadOtherCheckouts(t *testing.T) {
	project := withChanges(t, "a", "b")
	if err := os.MkdirAll(filepath.Join(project, "openspec", "changes", "archive", "2026-10-01-old"), 0o755); err != nil {
		t.Fatal(err)
	}
	wtA, wtN, wtOld := withChanges(t, "a", "x"), withChanges(t, "n"), withChanges(t, "old")
	noChange, broken, lacking := withChanges(t), withChanges(t, "f"), withChanges(t, "g")
	clis := map[string]*fakeCLI{
		project: {list: `{"changes":[
			{"name":"a","completedTasks":1,"totalTasks":4,"lastModified":"2026-10-28T10:00:00Z"},
			{"name":"b","completedTasks":2,"totalTasks":2,"lastModified":"2026-10-28T11:00:00Z"}]}`},
		wtA: {list: `{"changes":[
			{"name":"a","completedTasks":3,"totalTasks":4,"lastModified":"2026-10-30T10:00:00Z"},
			{"name":"x","completedTasks":1,"totalTasks":1,"lastModified":"2026-10-30T10:00:00Z"}]}`},
		wtN:     {list: `{"changes":[{"name":"n","totalTasks":2,"lastModified":"2026-10-30T12:00:00Z"}]}`, status: map[string]string{"n": planned}},
		wtOld:   {list: `{"changes":[{"name":"old","completedTasks":1,"totalTasks":2}]}`},
		lacking: {list: `{"changes":[]}`},
	}
	cli := func(dir string) herdr.Runner {
		if dir == broken {
			return func(...string) ([]byte, error) { return nil, errors.New("exit status 1") }
		}
		if c, ok := clis[dir]; ok {
			return c.run
		}
		return func(args ...string) ([]byte, error) { return nil, fmt.Errorf("unexpected call %q in %s", args, dir) }
	}
	others := []worktree.Checkout{{Path: wtA, Branch: "a"}, {Path: wtN, Branch: "n"}, {Path: wtOld, Branch: "old"},
		{Path: noChange, Branch: "fix-typo"}, {Path: broken, Branch: "f"}, {Path: lacking, Branch: "g"}, {Path: project, Branch: "b"}}
	got, err := Load(project, cli, others, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := []Change{
		{Name: "a", Done: 3, Total: 4, Modified: time.Date(2026, 10, 30, 10, 0, 0, 0, time.UTC), Dir: wtA},
		{Name: "b", Done: 2, Total: 2, Modified: time.Date(2026, 10, 28, 11, 0, 0, 0, time.UTC)},
		{Name: "n", Total: 2, Modified: time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC), Dir: wtN, Planning: Complete,
			Artifacts: map[string]string{"proposal": "done", "specs": "done", "design": "done", "tasks": "done"},
			Files:     files(filepath.Join(wtN, "openspec", "changes", "n"), setup(wtN, "1.8.0"))},
		{Name: "old", Archived: true, ArchivedOn: "2026-10-01"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Load\n got %+v\nwant %+v", got, want)
	}
	lists := slices.DeleteFunc(slices.Clone(clis[project].calls), func(c string) bool { return c != "list --json" })
	if len(clis[wtOld].calls) != 0 || len(lists) != 1 {
		t.Errorf("a checkout of an archived Change and the Project's own checkout must not be listed, calls %q, %q", clis[wtOld].calls, clis[project].calls)
	}
	if want := []string{"status --change n --json"}; !slices.Contains(clis[wtN].calls, want[0]) || slices.Contains(clis[project].calls, want[0]) {
		t.Errorf("status of n must run in its checkout, project calls %q, worktree calls %q", clis[project].calls, clis[wtN].calls)
	}
}

func TestLoadDoesNotReuseAcrossCheckouts(t *testing.T) {
	project, wt := withChanges(t), withChanges(t, "n")
	c := &fakeCLI{list: `{"changes":[{"name":"n","lastModified":"2026-10-30T12:00:00Z"}]}`, status: map[string]string{"n": proposalOnly}}
	empty := &fakeCLI{list: `{"changes":[]}`}
	cli := func(dir string) herdr.Runner {
		if dir == wt {
			return c.run
		}
		return empty.run
	}
	fp := files(filepath.Join(wt, "openspec", "changes", "n"), setup(wt, "1.8.0"))
	known := map[string]Change{"n": {Name: "n", Modified: time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC), Planning: Complete, Files: fp}}
	got, err := Load(project, cli, []worktree.Checkout{{Path: wt, Branch: "n"}}, known)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Planning != Incomplete {
		t.Errorf("planning cached for the Project must not be reused for the checkout, got %+v", got)
	}
	known["n"] = got[0]
	c.calls = nil
	if _, err := Load(project, cli, []worktree.Checkout{{Path: wt, Branch: "n"}}, known); err != nil || slices.Contains(c.calls, "status --change n --json") {
		t.Errorf("planning cached for the same checkout must be reused, calls %q, %v", c.calls, err)
	}
}
