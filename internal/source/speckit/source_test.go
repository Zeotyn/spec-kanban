package speckit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zeotyn/spec-kanban/internal/work"
)

var _ work.Source = (*Source)(nil)

func write(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestDetect(t *testing.T) {
	src := New(nil, "")
	dir := t.TempDir()
	if src.Detect(dir) {
		t.Error("empty dir: want false")
	}
	write(t, dir, map[string]string{".specify/init-options.json": "{}"})
	if src.Detect(dir) {
		t.Error(".specify alone: want false")
	}
	write(t, dir, map[string]string{"specs": ""})
	if src.Detect(dir) {
		t.Error("specs is a file: want false")
	}
	if err := os.Remove(filepath.Join(dir, "specs")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if !src.Detect(dir) {
		t.Error("with .specify and specs: want true")
	}
	other := t.TempDir()
	write(t, other, map[string]string{".specify": "", "specs/001-a/spec.md": "x"})
	if src.Detect(other) {
		t.Error(".specify is a file: want false")
	}
}

func TestSourceTexts(t *testing.T) {
	src := New(nil, "")
	if got := src.Rel("001-a"); got != "specs/001-a" {
		t.Errorf("Rel = %q", got)
	}
	dir := t.TempDir()
	write(t, dir, map[string]string{".specify/x": "", "specs/001-a/spec.md": "x"})
	if src.ArchivedIn(dir, "001-a") || src.ArchivedIn(dir, "nope") {
		t.Error("ArchivedIn: want always false")
	}
}

func TestPrepare(t *testing.T) {
	dir := t.TempDir()
	src := New(nil, "")
	p := filepath.Join(dir, ".specify", "feature.json")
	for _, name := range []string{"002-export", "001-user-login-email"} { // the second replaces the first
		if err := src.Prepare(dir, name); err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(p)
		if want := `{"feature_directory":"specs/` + name + `"}` + "\n"; err != nil || string(b) != want {
			t.Errorf("feature.json = %q, %v; want %q", b, err, want)
		}
	}
}

func TestPrepareFails(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, map[string]string{".specify": ""}) // a file where the directory belongs
	err := New(nil, "").Prepare(dir, "001-a")
	if p := filepath.Join(dir, ".specify", "feature.json"); err == nil || !strings.Contains(err.Error(), p) {
		t.Errorf("Prepare = %v, want an error naming %s", err, p)
	}
}

func TestPrompt(t *testing.T) {
	spike, err := os.ReadFile(filepath.Join(fixture, ".specify", "integration.json"))
	if err != nil {
		t.Fatal(err)
	}
	const multi = `{"installed_integrations": ["claude", "codex", "gemini"], "integration": "gemini",
		"integration_settings": {"claude": {"invoke_separator": "-"}, "gemini": {"invoke_separator": "."}}}`
	tests := []struct{ name, file, kind, want string }{
		{"spike, claude", string(spike), "claude", "/speckit-implement"},
		{"codex installed", multi, "codex", "$speckit-implement"},
		{"claude installed", multi, "claude", "/speckit-implement"},
		{"kind not installed", multi, "cursor", "/speckit.implement"},
		{"zcode default", `{"integration": "zcode"}`, "claude", "$speckit-implement"},
		{"no file", "", "claude", "/speckit-implement"},
		{"broken JSON", `{"integration": `, "claude", "/speckit-implement"},
		{"no separator", `{"integration": "gemini", "integration_settings": {"gemini": {}}}`, "gemini", "/speckit-implement"},
		{"no integration", `{"integration_settings": {"gemini": {"invoke_separator": "."}}}`, "claude", "/speckit-implement"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := t.TempDir()
			if tt.file != "" {
				write(t, p, map[string]string{".specify/integration.json": tt.file})
			}
			if got := New(nil, "").Prompt(p, tt.kind); got != tt.want {
				t.Errorf("Prompt(%q) = %q, want %q", tt.kind, got, tt.want)
			}
		})
	}
}

func TestFinish(t *testing.T) {
	src := New(nil, "")
	if got := src.Finish("001-a", ""); got != "" {
		t.Errorf("Finish without a worktree = %q, want none", got)
	}
	if got := src.Finish("001-a", "/wt"); got != "delete /wt/.specify/feature.json unless git tracks it" {
		t.Errorf("Finish in a worktree = %q", got)
	}
}
