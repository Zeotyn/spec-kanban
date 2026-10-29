package main

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
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
