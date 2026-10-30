// Package speckit reads a Project's Spec Kit Features from its specs directory, and their merge state
// through git.
package speckit

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/zeotyn/spec-kanban/internal/herdr"
)

// Source is the Spec Kit work.Source; git runs git inside a directory, base is the configured base
// ("" to detect it).
type Source struct {
	git  func(dir string) herdr.Runner
	base string
}

// New returns the Spec Kit Source that reads merge state through git against base.
func New(git func(dir string) herdr.Runner, base string) *Source { return &Source{git, base} }

// Detect reports whether dir holds both a .specify and a specs directory.
func (*Source) Detect(dir string) bool {
	for _, d := range []string{".specify", "specs"} {
		if fi, err := os.Stat(filepath.Join(dir, d)); err != nil || !fi.IsDir() {
			return false
		}
	}
	return true
}

// ArchivedIn is always false: Spec Kit has no archive.
func (*Source) ArchivedIn(string, string) bool { return false }

// Rel is where a Feature lives below a checkout.
func (*Source) Rel(name string) string { return "specs/" + name }

// Prepare writes .specify/feature.json in the checkout dir, so Spec Kit's commands find the Feature
// name without a branch or environment variable.
func (*Source) Prepare(dir, name string) error {
	p := filepath.Join(dir, ".specify", "feature.json")
	b, _ := json.Marshal(map[string]string{"feature_directory": "specs/" + name})
	err := os.MkdirAll(filepath.Dir(p), 0o755)
	if err == nil {
		err = os.WriteFile(p, append(b, '\n'), 0o644)
	}
	if pe := (*fs.PathError)(nil); errors.As(err, &pe) {
		err = pe.Err
	}
	if err != nil {
		return fmt.Errorf("write %s: %w", p, err)
	}
	return nil
}

// Prompt is Spec Kit's implement command for an agent of kind, as the integration of that name (else
// the default one) in project's .specify/integration.json invokes it; /speckit-implement when the file
// does not say.
func (*Source) Prompt(project, kind string) string {
	var cfg struct {
		Integration string   `json:"integration"`
		Installed   []string `json:"installed_integrations"`
		Settings    map[string]struct {
			Separator string `json:"invoke_separator"`
		} `json:"integration_settings"`
	}
	b, err := os.ReadFile(filepath.Join(project, ".specify", "integration.json"))
	if err != nil || json.Unmarshal(b, &cfg) != nil {
		return "/speckit-implement"
	}
	name := cfg.Integration
	if slices.Contains(cfg.Installed, kind) {
		name = kind
	}
	switch sep := cfg.Settings[name].Separator; {
	case name == "codex" || name == "zcode":
		return "$speckit-implement"
	case sep != "":
		return "/speckit" + sep + "implement"
	}
	return "/speckit-implement"
}

// Finish asks to delete the feature.json Prepare wrote in a worktree, which would block its removal;
// nothing without one.
func (*Source) Finish(_, checkout string) string {
	if checkout == "" {
		return ""
	}
	return "delete " + filepath.Join(checkout, ".specify", "feature.json") + " unless git tracks it"
}
