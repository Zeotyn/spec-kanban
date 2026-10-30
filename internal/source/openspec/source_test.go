package openspec

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zeotyn/spec-kanban/internal/work"
)

var _ work.Source = (*Source)(nil)

func TestDetect(t *testing.T) {
	dir := t.TempDir()
	src := New(nil)
	if src.Detect(dir) {
		t.Error("empty dir: want false")
	}
	if err := os.MkdirAll(filepath.Join(dir, "openspec"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "openspec", "changes"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if src.Detect(dir) {
		t.Error("openspec/changes is a file: want false")
	}
	if err := os.Remove(filepath.Join(dir, "openspec", "changes")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "openspec", "changes"), 0o755); err != nil {
		t.Fatal(err)
	}
	if !src.Detect(dir) {
		t.Error("with openspec/changes: want true")
	}
}

func TestSourceTexts(t *testing.T) {
	src := New(nil)
	if got := src.Rel("x"); got != "openspec/changes/x" {
		t.Errorf("Rel = %q", got)
	}
	if got := src.Prompt("/p", "codex"); got != "/opsx:apply {change}" {
		t.Errorf("Prompt = %q", got)
	}
	for _, checkout := range []string{"", "/wt"} {
		if got := src.Finish("x", checkout); got != "archive it with /opsx:archive x" {
			t.Errorf("Finish(x, %q) = %q", checkout, got)
		}
	}
}

func TestPrepareWritesNothing(t *testing.T) {
	dir := t.TempDir()
	if err := New(nil).Prepare(dir, "x"); err != nil {
		t.Fatal(err)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Errorf("Prepare wrote %v", entries)
	}
}
