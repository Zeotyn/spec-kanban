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
)

// fakeCLI answers `list --json` with list and `status --change <name> --json` from status;
// a missing status entry fails like the real CLI does for an unknown change.
type fakeCLI struct {
	list   string
	status map[string]string
	mu     sync.Mutex
	calls  []string
}

func (f *fakeCLI) run(args ...string) ([]byte, error) {
	call := strings.Join(args, " ")
	f.mu.Lock()
	f.calls = append(f.calls, call)
	f.mu.Unlock()
	if call == "list --json" {
		return []byte(f.list), nil
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
	got, err := Load(t.TempDir(), cli.run)
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
	wantCalls := []string{"list --json", "status --change broken --json", "status --change draft --json", "status --change ready --json"}
	if !reflect.DeepEqual(cli.calls, wantCalls) {
		t.Errorf("calls\n got %q\nwant %q", cli.calls, wantCalls)
	}
}

func TestLoadNoChanges(t *testing.T) {
	got, err := Load(t.TempDir(), (&fakeCLI{list: `{"changes":[]}`}).run)
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
			if _, err := Load(t.TempDir(), run); err == nil || !strings.Contains(err.Error(), tt.want) {
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
	got, err := Load(project, (&fakeCLI{list: `{"changes":[]}`}).run)
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
