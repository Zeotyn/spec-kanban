package cache

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/zeotyn/spec-kanban/internal/herdr"
	"github.com/zeotyn/spec-kanban/internal/work"
)

const project = "/abs/project"

func TestDirUsesStateDir(t *testing.T) {
	t.Setenv("HERDR_PLUGIN_STATE_DIR", "/state")
	if got := Dir(); got != "/state" {
		t.Errorf("Dir() = %q, want /state", got)
	}
}

func TestDirFallsBackToUserCacheDir(t *testing.T) {
	t.Setenv("HERDR_PLUGIN_STATE_DIR", "")
	base, err := os.UserCacheDir()
	if err != nil {
		t.Skip("no user cache dir:", err)
	}
	if got, want := Dir(), filepath.Join(base, "spec-kanban"); got != want {
		t.Errorf("Dir() = %q, want %q", got, want)
	}
}

func writeFile(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name(project)), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestReadInvalid(t *testing.T) {
	for _, tc := range []struct{ name, content string }{
		{"invalid JSON", `{"version":2,`},
		{"version 1", `{"version":1,"project":"/abs/project","changes":[]}`},
		{"other version", `{"version":3,"project":"/abs/project","changes":[]}`},
		{"other project", `{"version":2,"project":"/other","changes":[]}`},
		{"empty name", `{"version":2,"project":"/abs/project","changes":[{"name":"a"},{"name":""}]}`},
		{"phase after Archived", `{"version":2,"project":"/abs/project","changes":[{"name":"a","phase":5}]}`},
		{"negative phase", `{"version":2,"project":"/abs/project","changes":[{"name":"a","phase":-1}]}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFile(t, dir, tc.content)
			if got := Read(dir, project); got != nil {
				t.Errorf("Read = %v, want nil", got)
			}
		})
	}
}

func TestReadMissing(t *testing.T) {
	if got := Read(t.TempDir(), project); got != nil {
		t.Errorf("Read = %v, want nil", got)
	}
}

func TestReadEmptyDir(t *testing.T) {
	if got := Read("", project); got != nil {
		t.Errorf("Read = %v, want nil", got)
	}
}

func TestReadValidEmpty(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, `{"version":2,"project":"/abs/project","changes":[]}`)
	got := Read(dir, project)
	if got == nil || len(got) != 0 {
		t.Errorf("Read = %#v, want non-nil empty", got)
	}
}

func TestWriteNilReadsEmpty(t *testing.T) {
	dir := t.TempDir()
	if err := Write(dir, project, nil); err != nil {
		t.Fatal(err)
	}
	if got := Read(dir, project); got == nil || len(got) != 0 {
		t.Errorf("Read = %#v, want non-nil empty", got)
	}
}

func TestRoundTrip(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "sub")
	want := []work.Item{
		{Name: "a", Total: 3, Modified: time.Date(2026, 10, 29, 12, 0, 0, 123e6, time.UTC),
			Phase: work.Ready, Artifacts: map[string]string{"proposal": "done"}, Stamp: "0123abcd", Dir: "/abs/wt"},
		{Name: "b", Done: 1, Total: 2, Modified: time.Date(2026, 10, 28, 8, 30, 0, 0, time.UTC), Phase: work.InProgress},
		{Name: "c", Phase: work.Archived, ArchivedOn: "2026-10-01"},
	}
	if err := Write(dir, project, want); err != nil {
		t.Fatal(err)
	}
	if got := Read(dir, project); !reflect.DeepEqual(got, want) {
		t.Errorf("Read = %#v, want %#v", got, want)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != name(project) {
		t.Errorf("dir holds %v, want only %s", entries, name(project))
	}
}

func TestWriteUnwritable(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores permissions")
	}
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o700) })
	if err := Write(dir, project, nil); err == nil {
		t.Error("Write into read-only dir: want error")
	}
}

func TestWriteEmptyDir(t *testing.T) {
	if err := Write("", project, nil); err == nil {
		t.Error(`Write(""): want error`)
	}
}

func TestSession(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "state")
	s := herdr.Session{Agent: "claude", Kind: "id", Value: "abc"}
	if got := ReadSession(dir, project, "m1"); got != (herdr.Session{}) {
		t.Errorf("missing: got %v, want zero", got)
	}
	if err := WriteSession(dir, project, "m1", s); err != nil {
		t.Fatal(err)
	}
	if got := ReadSession(dir, project, "m1"); got != s {
		t.Errorf("round trip: got %v, want %v", got, s)
	}
	if got := ReadSession(dir, "/other", "m1"); got != (herdr.Session{}) {
		t.Errorf("other project: got %v, want zero", got)
	}
	newer := herdr.Session{Agent: "codex", Kind: "id", Value: "def"}
	if err := WriteSession(dir, "/other", "m1", newer); err != nil {
		t.Fatal(err)
	}
	if err := WriteSession(dir, project, "m1", newer); err != nil {
		t.Fatal(err)
	}
	if got := ReadSession(dir, project, "m1"); got != newer {
		t.Errorf("overwrite: got %v, want %v", got, newer)
	}
}

func TestSessionBroken(t *testing.T) {
	dir := t.TempDir()
	if err := WriteSession(dir, project, "m1", herdr.Session{Agent: "claude"}); err != nil {
		t.Fatal(err)
	}
	files, _ := filepath.Glob(filepath.Join(dir, "sessions", "*", "m1"))
	if len(files) != 1 {
		t.Fatalf("want one session file, got %v", files)
	}
	if err := os.WriteFile(files[0], []byte(`{"agent":`), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := ReadSession(dir, project, "m1"); got != (herdr.Session{}) {
		t.Errorf("broken: got %v, want zero", got)
	}
}

func TestSessionEmptyDir(t *testing.T) {
	if err := WriteSession("", project, "m1", herdr.Session{Agent: "claude"}); err == nil {
		t.Error(`WriteSession(""): want error`)
	}
	if got := ReadSession("", project, "m1"); got != (herdr.Session{}) {
		t.Errorf(`ReadSession(""): got %v, want zero`, got)
	}
}
