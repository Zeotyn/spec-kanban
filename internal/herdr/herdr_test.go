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

func TestAgentName(t *testing.T) {
	for item, want := range map[string]string{
		"m1-board-view":                             "m1-board-view",
		"f-x":                                       "f-x",
		"001-user-login-email":                      "f-001-user-login-email",
		"20261030-142501-user-login":                "f-20261030-142501-user-login",
		"add-a-change-of-thirty-two-chars":          "add-a-change-of-thirty-two-chars",
		"update-a-change-from-its-base-branch":      "update-a-change-from-its-302ef9",
		"001-a-feature-name-far-too-long-for-herdr": "f-001-a-feature-name-far-8b5dbb",
	} {
		got := AgentName(item)
		if got != want {
			t.Errorf("AgentName(%q) = %q, want %q", item, got, want)
		}
		if len(got) > 32 || AgentName(item) != got {
			t.Errorf("AgentName(%q) = %q: over 32 characters or not stable", item, got)
		}
	}
}
