// Package openspec reads a Project's Changes through the openspec CLI and its archive directory.
package openspec

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/zeotyn/openspec-kanban/internal/herdr"
)

// Planning tells whether a Change's planning artifacts are complete; Unknown when not read or unreadable.
type Planning int

const (
	Unknown Planning = iota
	Incomplete
	Complete
)

type Change struct {
	Name       string            `json:"name"`
	Done       int               `json:"done"`
	Total      int               `json:"total"`
	Modified   time.Time         `json:"modified"`
	Planning   Planning          `json:"planning"`
	Artifacts  map[string]string `json:"artifacts,omitempty"` // artifact id → done|ready|blocked|skipped, only when Planning was read
	Archived   bool              `json:"archived,omitempty"`
	ArchivedOn string            `json:"archivedOn,omitempty"` // "2026-10-29"; "" when open or the archive folder has no date
	Files      string            `json:"files,omitempty"`      // fingerprint of the file paths, only for open Changes without completed tasks
}

// maxParallel bounds concurrent `openspec status` calls (each starts a Node process).
const maxParallel = 8

var dated = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})-(.+)$`)

// Load returns the open Changes, read by run (the openspec CLI inside project), in `openspec list`
// order, then the archived ones by folder name.
// Planning state is read only for open Changes without completed tasks; it is copied from known
// instead when known holds the Change with the same non-zero Modified, the same file fingerprint
// and a Planning other than Unknown.
func Load(project string, run herdr.Runner, known map[string]Change) ([]Change, error) {
	version := make(chan string, 1) // in parallel with list, so it costs no extra time
	go func() {
		out, _ := run("--version")
		version <- strings.TrimSpace(string(out))
	}()
	changes, err := list(run)
	if err != nil {
		return nil, err
	}

	v := <-version
	base := setup(project, v)
	var wg sync.WaitGroup
	slots := make(chan struct{}, maxParallel)
	for i := range changes {
		if changes[i].Done > 0 {
			continue
		}
		if v != "" { // without a version an upgrade would go unnoticed, so leave the fingerprint empty
			changes[i].Files = files(filepath.Join(project, "openspec", "changes", changes[i].Name), base)
		}
		if k, ok := known[changes[i].Name]; ok && k.Planning != Unknown && !k.Modified.IsZero() &&
			k.Modified.Equal(changes[i].Modified) && k.Files != "" && k.Files == changes[i].Files {
			changes[i].Planning, changes[i].Artifacts = k.Planning, k.Artifacts
			continue
		}
		wg.Add(1)
		go func(c *Change) {
			defer wg.Done()
			slots <- struct{}{}
			defer func() { <-slots }()
			c.Planning, c.Artifacts = status(run, c.Name)
		}(&changes[i])
	}
	wg.Wait()

	entries, err := os.ReadDir(filepath.Join(project, "openspec", "changes", "archive"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		c := Change{Name: e.Name(), Archived: true}
		if m := dated.FindStringSubmatch(e.Name()); m != nil {
			c.ArchivedOn, c.Name = m[1], m[2]
		}
		changes = append(changes, c)
	}
	return changes, nil
}

// files fingerprints the OpenSpec setup (see setup) and the relative file paths under dir. openspec's
// lastModified is the newest file time, which misses an older file being deleted or renamed; the
// path list catches that. "" on error, which never matches.
func files(dir, setup string) string {
	h := sha256.New()
	io.WriteString(h, setup)
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			rel, _ := filepath.Rel(dir, p)
			io.WriteString(h, rel+"\n")
		}
		return err
	})
	if err != nil {
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// setup stamps what `openspec status` depends on outside a Change: the Project's config.yaml, its
// project schemas, and the openspec CLI version.
func setup(project, version string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "openspec %s\n", version)
	stamp := func(p string) {
		if fi, err := os.Stat(p); err == nil {
			fmt.Fprintf(&b, "%s %d %d\n", p, fi.Size(), fi.ModTime().UnixNano())
		}
	}
	stamp(filepath.Join(project, "openspec", "config.yaml"))
	filepath.WalkDir(filepath.Join(project, "openspec", "schemas"), func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			stamp(p)
		}
		return nil
	})
	return b.String()
}

// OpenNames returns the names of the open Changes.
func OpenNames(run herdr.Runner) ([]string, error) {
	changes, err := list(run)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(changes))
	for i, c := range changes {
		names[i] = c.Name
	}
	return names, nil
}

func list(run herdr.Runner) ([]Change, error) {
	out, err := run("list", "--json")
	if errors.Is(err, exec.ErrNotFound) {
		return nil, errors.New("openspec CLI not found: install it with npm i -g @fission-ai/openspec")
	}
	if err != nil {
		return nil, fmt.Errorf("openspec list: %w", err)
	}
	var got struct {
		Changes []struct {
			Name           string    `json:"name"`
			CompletedTasks int       `json:"completedTasks"`
			TotalTasks     int       `json:"totalTasks"`
			LastModified   time.Time `json:"lastModified"`
		} `json:"changes"`
	}
	if err := json.Unmarshal(out, &got); err != nil {
		return nil, fmt.Errorf("openspec list: %w", err)
	}
	changes := make([]Change, len(got.Changes))
	for i, c := range got.Changes {
		changes[i] = Change{Name: c.Name, Done: c.CompletedTasks, Total: c.TotalTasks, Modified: c.LastModified}
	}
	return changes, nil
}

// status reads one Change's planning state; any failure yields Unknown, per the change-source spec.
func status(run herdr.Runner, name string) (Planning, map[string]string) {
	out, err := run("status", "--change", name, "--json")
	var got struct {
		IsPlanningComplete bool `json:"isPlanningComplete"`
		Artifacts          []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"artifacts"`
	}
	if err != nil || json.Unmarshal(out, &got) != nil || len(got.Artifacts) == 0 {
		return Unknown, nil
	}
	artifacts := make(map[string]string, len(got.Artifacts))
	for _, a := range got.Artifacts {
		artifacts[a.ID] = a.Status
	}
	if got.IsPlanningComplete {
		return Complete, artifacts
	}
	return Incomplete, artifacts
}
