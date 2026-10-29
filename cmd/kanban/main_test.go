package main

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/zeotyn/openspec-kanban/internal/cache"
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
