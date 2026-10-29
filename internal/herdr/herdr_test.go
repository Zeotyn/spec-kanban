package herdr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecReportsStderr(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "herdr")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\necho ui_busy >&2\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HERDR_BIN_PATH", bin)
	_, err := Exec("plugin", "pane", "open")
	if err == nil || !strings.Contains(err.Error(), "ui_busy") {
		t.Fatalf("want error with herdr's stderr, got %v", err)
	}
}
