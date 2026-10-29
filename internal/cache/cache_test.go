package cache

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/zeotyn/openspec-kanban/internal/openspec"
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
	if got, want := Dir(), filepath.Join(base, "openspec-kanban"); got != want {
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
		{"invalid JSON", `{"version":1,`},
		{"other version", `{"version":2,"project":"/abs/project","changes":[]}`},
		{"other project", `{"version":1,"project":"/other","changes":[]}`},
		{"empty name", `{"version":1,"project":"/abs/project","changes":[{"name":"a"},{"name":""}]}`},
		{"unknown planning", `{"version":1,"project":"/abs/project","changes":[{"name":"a","planning":7}]}`},
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
	writeFile(t, dir, `{"version":1,"project":"/abs/project","changes":[]}`)
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
	want := []openspec.Change{
		{Name: "a", Total: 3, Modified: time.Date(2026, 10, 29, 12, 0, 0, 123e6, time.UTC),
			Planning: openspec.Complete, Artifacts: map[string]string{"proposal": "done"}},
		{Name: "b", Done: 1, Total: 2, Modified: time.Date(2026, 10, 28, 8, 30, 0, 0, time.UTC)},
		{Name: "c", Archived: true, ArchivedOn: "2026-10-01"},
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
