package project

import (
	"os"
	"path/filepath"
	"testing"
)

// newProject creates <tmp>/proj/openspec/changes and <tmp>/proj/internal/ui and returns the resolved project path.
func newProject(t *testing.T) string {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	proj := filepath.Join(root, "proj")
	for _, dir := range []string{"openspec/changes", "internal/ui"} {
		if err := os.MkdirAll(filepath.Join(proj, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return proj
}

func TestFind(t *testing.T) {
	proj := newProject(t)
	link := filepath.Join(filepath.Dir(proj), "link")
	if err := os.Symlink(filepath.Join(proj, "internal"), link); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name, start, want string
	}{
		{"inside", filepath.Join(proj, "internal/ui"), proj},
		{"root", proj, proj},
		{"symlinked cwd", filepath.Join(link, "ui"), proj},
		{"no project", filepath.Dir(proj), ""},
		{"missing dir", filepath.Join(proj, "gone"), ""},
		{"empty", "", ""},
		{"relative", "internal", ""},
	}
	t.Chdir(proj) // relative and empty starts must not resolve against the process cwd
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Find(tt.start); got != tt.want {
				t.Errorf("Find(%q) = %q, want %q", tt.start, got, tt.want)
			}
		})
	}
}
