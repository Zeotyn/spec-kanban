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

	"github.com/zeotyn/openspec-kanban/internal/cache"
	"github.com/zeotyn/openspec-kanban/internal/herdr"
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
	base := []string{"plugin", "pane", "open", "--plugin", "openspec-kanban", "--entrypoint", "board", "--focus"}
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
			if err := open(tt.context, run); err != nil {
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
	err := open(`{}`, run)
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
	load := loader(dir, proj, run)

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
	if changes, err := loader(dir, t.TempDir(), run)(); err != nil || len(changes) != 1 {
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
	changes, err := syncLoad(dir, cli)(proj)
	if err != nil || len(changes) != 1 || ranIn != proj {
		t.Fatalf("want one Change read in %s, got %v, %v (ran in %q)", proj, changes, err, ranIn)
	}
	if got := cache.Read(dir, proj); !reflect.DeepEqual(got, changes) {
		t.Errorf("the Board's cache must hold the sync's load\n got %v\nwant %v", got, changes)
	}
}

func TestLaunchNotifiesOnFailure(t *testing.T) {
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
	want := []string{"notification", "show", "openspec-kanban: start failed", "--body", err.Error()}
	if len(calls) != 3 || !reflect.DeepEqual(calls[2], want) {
		t.Errorf("herdr calls %q, want tab focus, agent start then %q", calls, want)
	}
}

func TestLaunchSucceedsQuietly(t *testing.T) {
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

func TestChildEnvDropsPopupIDs(t *testing.T) {
	got := childEnv([]string{"HOME=/h", "HERDR_PANE_ID=w2:p9", "HERDR_SOCKET=/s", "HERDR_TAB_ID=w2:t9"})
	if want := []string{"HOME=/h", "HERDR_SOCKET=/s"}; !reflect.DeepEqual(got, want) {
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
	if err := startChange("m1-board-view", &stderr); err == nil || !strings.Contains(err.Error(), "no OpenSpec project") {
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
