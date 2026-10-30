package openspec

import (
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/zeotyn/openspec-kanban/internal/herdr"
)

// TestContract runs the real openspec CLI against testdata/project, so a change in its JSON output
// fails here rather than on the Board.
func TestContract(t *testing.T) {
	if _, err := exec.LookPath("openspec"); err != nil {
		t.Skip("openspec CLI not on PATH")
	}
	project, err := filepath.Abs(filepath.Join("..", "..", "testdata", "project"))
	if err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) ([]byte, error) { return herdr.Run("openspec", project, args...) } // as main wires it
	changes, err := Load(project, func(string) herdr.Runner { return run }, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	type want struct {
		done, total int
		planning    Planning
		specs       string // status of the specs artifact; "" when planning is not read
		archivedOn  string
	}
	wants := map[string]want{
		"proposal-only":      {0, 0, Incomplete, "ready", ""},
		"skip-specs-planned": {0, 3, Complete, "skipped", ""},
		"ready":              {0, 3, Complete, "done", ""},
		"nested-specs":       {0, 3, Complete, "done", ""},
		"no-tasks-file":      {0, 0, Incomplete, "done", ""},
		"in-progress":        {2, 4, Unknown, "", ""},
		"all-done":           {2, 2, Unknown, "", ""},
		"old-change":         {0, 0, Unknown, "", "2026-10-01"},
		"undated-change":     {0, 0, Unknown, "", ""},
	}
	artifacts := map[string]map[string]string{
		"proposal-only":      {"proposal": "done", "specs": "ready", "design": "ready", "tasks": "blocked"},
		"skip-specs-planned": {"proposal": "done", "specs": "skipped", "design": "done", "tasks": "done"},
	}
	if len(changes) != len(wants) {
		t.Errorf("got %d changes, want %d: %+v", len(changes), len(wants), changes)
	}
	for _, c := range changes {
		w, ok := wants[c.Name]
		if !ok {
			t.Errorf("unexpected change %q", c.Name)
			continue
		}
		got := want{c.Done, c.Total, c.Planning, c.Artifacts["specs"], c.ArchivedOn}
		if got != w {
			t.Errorf("%s: got %+v, want %+v", c.Name, got, w)
		}
		if a, ok := artifacts[c.Name]; ok && !reflect.DeepEqual(c.Artifacts, a) {
			t.Errorf("%s: artifacts %v, want %v", c.Name, c.Artifacts, a)
		}
		if !c.Archived && c.Modified.IsZero() {
			t.Errorf("%s: no modification time", c.Name)
		}
	}
}

// TestMissingCLI checks that a missing binary reaches Load through the real runner as "not found".
func TestMissingCLI(t *testing.T) {
	dir := t.TempDir() // once: the --version goroutine may still call run after the test ended
	run := func(args ...string) ([]byte, error) {
		return herdr.Run("openspec-not-installed-xyz", dir, args...)
	}
	if _, err := Load(t.TempDir(), func(string) herdr.Runner { return run }, nil, nil); err == nil || !strings.Contains(err.Error(), "openspec CLI not found") {
		t.Fatalf("want not-found error, got %v", err)
	}
}
