package config

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/zeotyn/spec-kanban/internal/source/openspec"
	"github.com/zeotyn/spec-kanban/internal/work"
)

func TestParse(t *testing.T) {
	def := Default()
	tests := []struct {
		name  string
		json  string
		want  Config
		wants string // substring of the error, "" for none
	}{
		{"empty object gives defaults", `{}`, Config{Worktree: true}, ""},
		{"worktree off", `{"worktree": false}`, Config{Prompt: def.Prompt}, ""},
		{"base kept", `{"base": "develop"}`, Config{Prompt: def.Prompt, Worktree: true, Base: "develop"}, ""},
		{"worktree as a string", `{"worktree": "no"}`, def, "worktree"},
		{"base empty", `{"base": ""}`, def, "base"},
		{"base as a number", `{"base": 1}`, def, "base"},
		{"agent kept", `{"agent": "codex"}`, Config{Agent: "codex", Prompt: def.Prompt, Worktree: true}, ""},
		{"prompt alone", `{"prompt": "/opsx:apply {change} with TDD"}`, Config{Prompt: "/opsx:apply {change} with TDD", Worktree: true}, ""},
		{"hide alone", `{"hide": ["Archived"]}`, Config{Prompt: def.Prompt, Hide: [5]bool{work.Archived: true}, Worktree: true}, ""},
		{"phase names ignore case", `{"hide": ["archived", "IN PROGRESS"]}`, Config{Prompt: def.Prompt, Hide: [5]bool{work.Archived: true, work.InProgress: true}, Worktree: true}, ""},
		{"agent with capital", `{"agent": "Codex"}`, def, `"Codex"`},
		{"agent unknown", `{"agent": "bash"}`, def, `"bash"`},
		{"agent empty", `{"agent": ""}`, def, `agent ""`},
		{"invalid JSON", `{"prompt":`, def, "unexpected EOF"},
		{"trailing data", `{} {"prompt": "x"}`, def, "after"},
		{"trailing brace", `{}}`, def, "after"},
		{"unknown key", `{"promt": "/opsx:apply {change}"}`, def, `"promt"`},
		{"wrong type", `{"hide": "Archived"}`, def, "hide"},
		{"prompt without change", `{"prompt": "go on"}`, def, "{change}"},
		{"empty prompt", `{"prompt": ""}`, def, "{change}"},
		{"unknown phase", `{"hide": ["Done"]}`, def, `"Done"`},
		{"all phases hidden", `{"hide": ["Draft", "Ready", "In Progress", "Review", "archived"]}`, def, "all"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse([]byte(tt.json))
			if tt.wants == "" && err != nil || tt.wants != "" && (err == nil || !strings.Contains(err.Error(), tt.wants)) {
				t.Fatalf("want error with %q, got %v", tt.wants, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
			if p := got.PromptFor("m1", openspec.New(nil).Prompt()); tt.wants != "" && p != "/opsx:apply m1" {
				t.Errorf("an invalid file must send the Source's prompt, got %q", p)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	write := func(t *testing.T, content string) string {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return dir
	}
	t.Run("valid file", func(t *testing.T) {
		cfg, warn := Load(write(t, `{"prompt": "go {change}"}`))
		if warn != nil || cfg.Prompt != "go {change}" {
			t.Errorf("got %+v, %v", cfg, warn)
		}
	})
	t.Run("missing file: defaults, no warning", func(t *testing.T) {
		if cfg, warn := Load(t.TempDir()); warn != nil || !reflect.DeepEqual(cfg, Default()) {
			t.Errorf("got %+v, %v", cfg, warn)
		}
	})
	t.Run("no dir: defaults, no warning", func(t *testing.T) {
		if cfg, warn := Load(""); warn != nil || !reflect.DeepEqual(cfg, Default()) {
			t.Errorf("got %+v, %v", cfg, warn)
		}
	})
	t.Run("invalid content: defaults, warning naming the file", func(t *testing.T) {
		dir := write(t, `{"promt": "x"}`)
		cfg, warn := Load(dir)
		if !reflect.DeepEqual(cfg, Default()) || warn == nil || !strings.Contains(warn.Error(), filepath.Join(dir, "config.json")) || !strings.Contains(warn.Error(), "promt") {
			t.Errorf("got %+v, %v", cfg, warn)
		}
	})
	t.Run("unreadable file: defaults, warning naming the file", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Mkdir(filepath.Join(dir, "config.json"), 0o755); err != nil { // a directory cannot be read as a file
			t.Fatal(err)
		}
		cfg, warn := Load(dir)
		if !reflect.DeepEqual(cfg, Default()) || warn == nil || !strings.Contains(warn.Error(), filepath.Join(dir, "config.json")) {
			t.Errorf("got %+v, %v", cfg, warn)
		}
	})
}

func TestPromptFor(t *testing.T) {
	src := openspec.New(nil)
	if got := Default().PromptFor("m1-board-view", src.Prompt()); got != "/opsx:apply m1-board-view" {
		t.Errorf("default prompt %q", got)
	}
	cfg := Config{Prompt: "{change}: /opsx:apply {change}"}
	if got := cfg.PromptFor("m1", src.Prompt()); got != "m1: /opsx:apply m1" {
		t.Errorf("custom prompt %q", got)
	}
}

func TestDir(t *testing.T) {
	var calls []string
	tests := []struct {
		name, env, out string
		err            error
		want           string
		wantCalls      []string
	}{
		{"env wins", "/cfg", "/other\n", nil, "/cfg", nil},
		{"herdr fallback, trimmed", "", "/home/u/.config/herdr/plugins/config/spec-kanban\n", nil, "/home/u/.config/herdr/plugins/config/spec-kanban", []string{"plugin config-dir spec-kanban"}},
		{"herdr fails", "", "", errors.New("herdr: not found"), "", []string{"plugin config-dir spec-kanban"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HERDR_PLUGIN_CONFIG_DIR", tt.env)
			calls = nil
			run := func(args ...string) ([]byte, error) {
				calls = append(calls, strings.Join(args, " "))
				return []byte(tt.out), tt.err
			}
			if got := Dir(run); got != tt.want {
				t.Errorf("Dir = %q, want %q", got, tt.want)
			}
			if !reflect.DeepEqual(calls, tt.wantCalls) {
				t.Errorf("herdr calls %q, want %q", calls, tt.wantCalls)
			}
		})
	}
}
