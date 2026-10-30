// Package config reads the user settings from config.json in herdr's config dir for the plugin.
package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/zeotyn/openspec-kanban/internal/herdr"
	"github.com/zeotyn/openspec-kanban/internal/work"
)

// Kinds: update from `herdr agent start --help`.
var Kinds = []string{"pi", "claude", "codex", "gemini", "cursor", "devin", "agy", "cline", "omp", "mastracode", "opencode", "copilot", "kimi", "kiro", "droid", "amp", "grok", "hermes", "kilo", "qodercli", "qwen", "letta", "maki", "muse"}

// Config holds the settings; an empty Agent means detect the kind from the invoking pane, Hide is
// indexed by Phase, an empty Base means detect the base of a new worktree.
type Config struct {
	Agent    string
	Prompt   string
	Hide     [5]bool
	Worktree bool
	Base     string
}

// Default is the config without a file; an empty Prompt means the Source's.
func Default() Config { return Config{Worktree: true} }

// PromptFor is the prompt Start sends for change; def (the Source's Prompt) when none is configured.
func (c Config) PromptFor(change, def string) string {
	p := c.Prompt
	if p == "" {
		p = def
	}
	return strings.ReplaceAll(p, "{change}", change)
}

// Parse decodes and validates a config file; on any problem it returns Default and the problem.
func Parse(b []byte) (Config, error) {
	var raw struct {
		Agent    *string  `json:"agent"`
		Prompt   *string  `json:"prompt"`
		Hide     []string `json:"hide"`
		Worktree *bool    `json:"worktree"`
		Base     *string  `json:"base"`
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&raw); err != nil {
		return Default(), err
	}
	if dec.Decode(new(any)) != io.EOF {
		return Default(), errors.New("data after the JSON object")
	}
	cfg := Default()
	if raw.Agent != nil {
		if !slices.Contains(Kinds, *raw.Agent) {
			return Default(), fmt.Errorf("agent %q is not a herdr agent kind", *raw.Agent)
		}
		cfg.Agent = *raw.Agent
	}
	if raw.Prompt != nil {
		if !strings.Contains(*raw.Prompt, "{change}") {
			return Default(), fmt.Errorf("prompt %q lacks {change}", *raw.Prompt)
		}
		cfg.Prompt = *raw.Prompt
	}
	for _, name := range raw.Hide {
		p := work.Draft
		for p <= work.Archived && !strings.EqualFold(p.String(), name) {
			p++
		}
		if p > work.Archived {
			return Default(), fmt.Errorf("hide: unknown phase %q", name)
		}
		cfg.Hide[p] = true
	}
	if !slices.Contains(cfg.Hide[:], false) {
		return Default(), errors.New("hide: all phases hidden")
	}
	if raw.Worktree != nil {
		cfg.Worktree = *raw.Worktree
	}
	if raw.Base != nil {
		if *raw.Base == "" {
			return Default(), errors.New("base is empty")
		}
		cfg.Base = *raw.Base
	}
	return cfg, nil
}

// Load reads config.json in dir. A missing file or an empty dir gives Default and no warning, any
// other problem Default and a warning naming the file.
func Load(dir string) (Config, error) {
	if dir == "" {
		return Default(), nil
	}
	path := filepath.Join(dir, "config.json")
	b, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return Default(), nil
	}
	if err == nil {
		var cfg Config
		if cfg, err = Parse(b); err == nil {
			return cfg, nil
		}
	}
	return Default(), fmt.Errorf("%s: %w, using defaults", path, err)
}

// Dir is HERDR_PLUGIN_CONFIG_DIR, or what herdr prints for the plugin outside plugin commands; "" when
// herdr fails.
func Dir(run herdr.Runner) string {
	if dir := os.Getenv("HERDR_PLUGIN_CONFIG_DIR"); dir != "" {
		return dir
	}
	out, err := run("plugin", "config-dir", "openspec-kanban")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
