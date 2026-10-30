package main

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime/debug"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/zeotyn/spec-kanban/internal/cache"
	"github.com/zeotyn/spec-kanban/internal/config"
	"github.com/zeotyn/spec-kanban/internal/herdr"
	"github.com/zeotyn/spec-kanban/internal/source/openspec"
	"github.com/zeotyn/spec-kanban/internal/source/speckit"
	"github.com/zeotyn/spec-kanban/internal/work"
)

func TestOpen(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	proj := filepath.Join(root, "proj")
	if err := os.MkdirAll(filepath.Join(proj, "openspec/changes"), 0o755); err != nil {
		t.Fatal(err)
	}
	base := []string{"plugin", "pane", "open", "--plugin", "spec-kanban", "--entrypoint", "board", "--focus"}
	withProject := append(append([]string{}, base...), "--cwd", proj, "--env", "KANBAN_PROJECT="+proj)

	tests := []struct {
		name, context string
		want          []string
	}{
		{"focused pane in project", `{"focused_pane_cwd":"` + proj + `/openspec","workspace_cwd":"` + root + `"}`, withProject},
		{"workspace cwd fallback", `{"workspace_cwd":"` + proj + `"}`, withProject},
		{"no project", `{"focused_pane_cwd":"` + root + `"}`, base},
		{"no context", ``, base},
		{"focused pane", `{"focused_pane_id":"w2:p1J","focused_pane_cwd":"` + root + `"}`, append(append([]string{}, base...), "--env", "KANBAN_MAIN_PANE=w2:p1J")},
		{"focused pane in project", `{"focused_pane_id":"w2:p1J","focused_pane_cwd":"` + proj + `"}`, append(append([]string{}, withProject...), "--env", "KANBAN_MAIN_PANE=w2:p1J")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			run := func(args ...string) ([]byte, error) { got = args; return nil, nil }
			if err := open(tt.context, run, openspec.New(nil).Detect); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("herdr args\n got %q\nwant %q", got, tt.want)
			}
		})
	}
}

func TestOpenHerdrError(t *testing.T) {
	run := func(args ...string) ([]byte, error) { return nil, errors.New("ui_busy") }
	err := open(`{}`, run, openspec.New(nil).Detect)
	if err == nil || !strings.Contains(err.Error(), "ui_busy") {
		t.Fatalf("want herdr error, got %v", err)
	}
}

func TestLoader(t *testing.T) {
	proj, dir := t.TempDir(), filepath.Join(t.TempDir(), "cache")
	if err := os.MkdirAll(filepath.Join(proj, "openspec/changes/a"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(proj, "openspec/changes/a/proposal.md"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	var (
		mu    sync.Mutex
		calls []string
	)
	fail := false
	run := func(args ...string) ([]byte, error) {
		mu.Lock()
		calls = append(calls, strings.Join(args, " "))
		mu.Unlock()
		if fail {
			return nil, errors.New("boom")
		}
		if args[0] == "list" {
			return []byte(`{"changes":[{"name":"a","totalTasks":2,"lastModified":"2026-10-28T10:00:00Z"}]}`), nil
		}
		return []byte(`{"isPlanningComplete":true,"artifacts":[{"id":"proposal","status":"done"}]}`), nil
	}
	load := loader(dir, proj, openspec.New(func(string) herdr.Runner { return run }), noGit)

	if _, err := load(); err != nil || len(cache.Read(dir, proj)) != 1 {
		t.Fatalf("first load must succeed and write the cache: %v, %v", err, cache.Read(dir, proj))
	}
	calls = nil
	if _, err := load(); err != nil || !reflect.DeepEqual(sorted(calls), []string{"--version", "list --json"}) {
		t.Errorf("second load must reuse the cache and call only version and list, got %q, %v", calls, err)
	}
	before, _ := os.ReadFile(filepath.Join(dir, mustOnly(t, dir)))
	fail = true
	if _, err := load(); err == nil {
		t.Fatal("want the openspec error")
	}
	if after, _ := os.ReadFile(filepath.Join(dir, mustOnly(t, dir))); string(after) != string(before) {
		t.Error("a failed load must leave the cache unchanged")
	}
}

func TestLoaderUnwritableCache(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores permissions")
	}
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o700) })
	run := func(...string) ([]byte, error) {
		return []byte(`{"changes":[{"name":"a","completedTasks":1,"totalTasks":2}]}`), nil
	}
	if changes, err := loader(dir, t.TempDir(), openspec.New(func(string) herdr.Runner { return run }), noGit)(); err != nil || len(changes) != 1 {
		t.Errorf("an unwritable cache must not break loading: %v, %v", changes, err)
	}
}

// mustOnly returns the single file name in dir.
func mustOnly(t *testing.T, dir string) string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("want exactly one cache file in %s: %v %v", dir, entries, err)
	}
	return entries[0].Name()
}

// noGit fails like git outside a repository.
func noGit(...string) ([]byte, error) { return nil, errors.New("fatal: not a git repository") }

func sorted(s []string) []string { return slices.Sorted(slices.Values(s)) }

func TestSyncLoadWritesBoardCache(t *testing.T) {
	proj, dir := t.TempDir(), t.TempDir()
	var ranIn string
	cli := func(p string) herdr.Runner {
		ranIn = p
		return func(...string) ([]byte, error) {
			return []byte(`{"changes":[{"name":"a","completedTasks":1,"totalTasks":2}]}`), nil
		}
	}
	changes, err := syncLoad(dir, func(string) work.Source { return openspec.New(cli) }, func(string) herdr.Runner { return noGit })(proj)
	if err != nil || len(changes) != 1 || ranIn != proj {
		t.Fatalf("want one Change read in %s, got %v, %v (ran in %q)", proj, changes, err, ranIn)
	}
	if got := cache.Read(dir, proj); !reflect.DeepEqual(got, changes) {
		t.Errorf("the Board's cache must hold the sync's load\n got %v\nwant %v", got, changes)
	}
}

func TestLaunchNotifiesOnFailure(t *testing.T) {
	t.Setenv("KANBAN_PROJECT", "/x/proj") // a failed Launch is not watched
	var calls [][]string
	run := func(args ...string) ([]byte, error) {
		calls = append(calls, args)
		if args[0] == "agent" && args[1] == "start" {
			return nil, errors.New("timeout")
		}
		return nil, nil
	}
	err := launch("m1-board-view", "w2:p1J", "w2:tG", "codex", "/opsx-apply m1-board-view", run)
	if err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("want the Launch error, got %v", err)
	}
	want := []string{"notification", "show", "spec-kanban: start failed", "--body", err.Error()}
	if len(calls) != 3 || !reflect.DeepEqual(calls[2], want) {
		t.Errorf("herdr calls %q, want tab focus, agent start then %q", calls, want)
	}
}

func TestLaunchSucceedsQuietly(t *testing.T) {
	t.Setenv("KANBAN_PROJECT", "") // not from the Board: not watched
	var calls [][]string
	run := func(args ...string) ([]byte, error) { calls = append(calls, args); return nil, nil }
	if err := launch("m1-board-view", "w2:p1J", "w2:tG", "codex", "/opsx-apply m1-board-view with TDD", run); err != nil {
		t.Fatal(err)
	}
	want := [][]string{
		{"tab", "focus", "w2:tG"},
		{"agent", "start", "m1-board-view", "--kind", "codex", "--pane", "w2:p1J", "--timeout", "60000"},
		{"agent", "prompt", "m1-board-view", "/opsx-apply m1-board-view with TDD"},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Errorf("want kind and prompt passed through and no notification\n got %q\nwant %q", calls, want)
	}
}

func TestLaunchWatches(t *testing.T) {
	t.Setenv("KANBAN_PROJECT", t.TempDir())
	t.Setenv("HERDR_PLUGIN_CONFIG_DIR", t.TempDir())
	var calls []string
	run := func(args ...string) ([]byte, error) {
		calls = append(calls, strings.Join(args, " "))
		if args[0] == "agent" && args[1] == "get" {
			return nil, errors.New("agent_not_found") // the agent went away: watching ends
		}
		return nil, nil
	}
	if err := launch("m1-board-view", "w2:p1J", "w2:tG", "claude", "/opsx:apply m1-board-view", run); err != nil {
		t.Fatal(err)
	}
	if want := "agent get m1-board-view"; len(calls) != 4 || calls[3] != want {
		t.Errorf("want Launch's three calls, then %q; got %q", want, calls)
	}
}

func TestResumeNotifiesOnFailure(t *testing.T) {
	t.Setenv("KANBAN_PROJECT", "/x/proj")
	t.Setenv("HERDR_PLUGIN_STATE_DIR", t.TempDir()) // no record
	var calls [][]string
	run := func(args ...string) ([]byte, error) { calls = append(calls, args); return nil, nil }
	err := resumeChange("m1-board-view", "w2:p1J", "w2:tG", run)
	if err == nil {
		t.Fatal("want an error without a record")
	}
	want := [][]string{{"notification", "show", "spec-kanban: resume failed", "--body", err.Error()}}
	if !reflect.DeepEqual(calls, want) {
		t.Errorf("herdr calls %q, want %q", calls, want)
	}
}

func TestResumeRecordedSession(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("KANBAN_PROJECT", "/x/proj")
	t.Setenv("HERDR_PLUGIN_STATE_DIR", dir)
	if err := cache.WriteSession(dir, "/x/proj", "m1-board-view", herdr.Session{Agent: "claude", Kind: "id", Value: "abc"}); err != nil {
		t.Fatal(err)
	}
	var calls []string
	run := func(args ...string) ([]byte, error) { calls = append(calls, strings.Join(args, " ")); return nil, nil }
	if err := resumeChange("m1-board-view", "w2:p1J", "w2:tG", run); err != nil {
		t.Fatal(err)
	}
	want := []string{"tab focus w2:tG", "agent start m1-board-view --kind claude --pane w2:p1J --timeout 60000 -- --resume abc"}
	if !reflect.DeepEqual(calls, want) {
		t.Errorf("herdr calls %q, want %q", calls, want)
	}
}

func TestChildEnvDropsPopupIDs(t *testing.T) {
	got := childEnv([]string{"HOME=/h", "HERDR_PANE_ID=w2:p9", "KANBAN_PROJECT=/p", "HERDR_SOCKET=/s", "HERDR_TAB_ID=w2:t9", "KANBAN_MAIN_PANE=w2:p1"})
	if want := []string{"HOME=/h", "KANBAN_PROJECT=/p", "HERDR_SOCKET=/s", "KANBAN_MAIN_PANE=w2:p1"}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStartChangeWarnsAboutBrokenConfig(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{"prompt": "go on"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HERDR_PLUGIN_CONFIG_DIR", dir)
	t.Chdir(t.TempDir()) // no project: startChange stops before herdr
	var stderr strings.Builder
	if err := startChange("m1-board-view", &stderr, nil); err == nil || !strings.Contains(err.Error(), "no OpenSpec or Spec Kit project at or above") {
		t.Fatalf("want the no-project error, got %v", err)
	}
	if got := stderr.String(); !strings.Contains(got, filepath.Join(dir, "config.json")) || !strings.Contains(got, "{change}") {
		t.Errorf("stderr %q lacks the warning naming the file and the problem", got)
	}
}

func TestVersionString(t *testing.T) {
	old := version
	t.Cleanup(func() { version = old })

	version = "9.9.9"
	if got := versionString(); got != "9.9.9" {
		t.Errorf("stamped: got %q, want 9.9.9", got)
	}
	version = ""
	info, _ := debug.ReadBuildInfo()
	if got := versionString(); got != info.Main.Version {
		t.Errorf("unstamped: got %q, want the build info's %q", got, info.Main.Version)
	}
}

func TestLoaderReadsOtherCheckouts(t *testing.T) {
	proj, wt, dir := t.TempDir(), t.TempDir(), t.TempDir()
	if err := os.MkdirAll(filepath.Join(wt, "openspec", "changes", "a"), 0o755); err != nil {
		t.Fatal(err)
	}
	git := func(args ...string) ([]byte, error) {
		if args[0] == "rev-parse" {
			return []byte("\n"), nil
		}
		return []byte("worktree " + proj + "\nbranch refs/heads/main\n\nworktree " + wt + "\nbranch refs/heads/a\n"), nil
	}
	cli := func(p string) herdr.Runner {
		done := map[string]string{proj: "1", wt: "2"}[p]
		return func(...string) ([]byte, error) {
			return []byte(`{"changes":[{"name":"a","completedTasks":` + done + `,"totalTasks":2}]}`), nil
		}
	}
	changes, err := loader(dir, proj, openspec.New(cli), git)()
	if err != nil || len(changes) != 1 || changes[0].Done != 2 || changes[0].Dir != wt {
		t.Errorf("want a read in its worktree %s, got %+v, %v", wt, changes, err)
	}
}

// mkdirs creates dirs below root.
func mkdirs(t *testing.T, root string, dirs ...string) {
	t.Helper()
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSources(t *testing.T) {
	root := t.TempDir()
	mkdirs(t, root, "openspec/changes", "speckit/.specify", "speckit/specs", "both/openspec/changes", "both/.specify", "both/specs", "none/.specify")
	tests := []struct {
		dir, prefer string
		detect      bool
		speckit     bool // sourceFor picks Spec Kit
	}{
		{"", "", true, false},
		{"speckit", "", true, true},
		{"both", "", true, false},
		{"both", "openspec", true, false},
		{"both", "speckit", true, true},
		{"", "speckit", true, false},
		{"none", "", false, false},
	}
	for _, tt := range tests {
		detect, sourceFor := sources(config.Config{Source: tt.prefer})
		dir := filepath.Join(root, tt.dir)
		if got := detect(dir); got != tt.detect {
			t.Errorf("detect(%q) = %v", tt.dir, got)
		}
		if _, got := sourceFor(dir).(*speckit.Source); got != tt.speckit {
			t.Errorf("sourceFor(%q) with source %q: Spec Kit %v, want %v", tt.dir, tt.prefer, got, tt.speckit)
		}
	}
}

func TestStartChangeKeepsPromptOverInstructions(t *testing.T) {
	cfgDir, proj := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(`{"prompt": "go {change}", "instructions": "with ponytail", "worktree": false}`), 0o644); err != nil {
		t.Fatal(err)
	}
	mkdirs(t, proj, ".specify", "specs/001-a")
	if err := os.WriteFile(filepath.Join(proj, "specs/001-a/spec.md"), []byte("# A\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	tab, err := os.ReadFile(filepath.Join("..", "..", "testdata", "herdr", "tab-create.json"))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("HERDR_PLUGIN_CONFIG_DIR", cfgDir)
	t.Setenv("HERDR_PANE_ID", "")
	t.Chdir(proj)
	var calls []string
	run := func(args ...string) ([]byte, error) {
		calls = append(calls, strings.Join(args, " "))
		switch args[0] + " " + args[1] {
		case "agent get":
			return nil, errors.New("agent_not_found")
		case "tab create":
			return tab, nil
		}
		return nil, nil
	}
	var stderr strings.Builder
	if err := startChange("001-a", &stderr, run); err != nil {
		t.Fatal(err)
	}
	if got := stderr.String(); !strings.Contains(got, "instructions is ignored because prompt is set") {
		t.Errorf("stderr %q lacks the warning", got)
	}
	if last := calls[len(calls)-1]; last != "agent prompt 001-a go 001-a" {
		t.Errorf("last herdr call %q, want the configured prompt", last)
	}
}
