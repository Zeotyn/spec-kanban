// Package openspec reads a Project's Changes through the openspec CLI and its archive directory, and their documents from disk.
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
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/zeotyn/openspec-kanban/internal/herdr"
	"github.com/zeotyn/openspec-kanban/internal/work"
	"github.com/zeotyn/openspec-kanban/internal/worktree"
)

// maxParallel bounds concurrent `openspec status` calls (each starts a Node process).
const maxParallel = 8

var dated = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})-(.+)$`)

// load returns the open Changes as Items, read by cli(project) (the openspec CLI inside a directory), in
// `openspec list` order, then the archived ones by folder name.
// A checkout in others (the Project's own skipped) whose branch names a Change it holds supplies that Change, read there with Dir
// set: it replaces the Project's open Change of that name or is appended, unless the Project archived
// it; a failed read leaves the Project's Changes.
// Planning state is read only for open Changes without completed tasks, in their Dir, and decides
// their Phase; Phase and Artifacts are copied from known instead when known holds the Change with the
// same Dir, the same non-zero Modified, the same Stamp (file fingerprint) and non-nil Artifacts.
func load(project string, cli func(dir string) herdr.Runner, others []worktree.Checkout, known map[string]work.Item) ([]work.Item, error) {
	run := cli(project)
	version := make(chan string, 1) // in parallel with list, so it costs no extra time
	go func() {
		out, _ := run("--version")
		version <- strings.TrimSpace(string(out))
	}()
	archived, err := archive(project)
	if err != nil {
		return nil, err
	}
	isArchived := func(name string) bool { return isIn(archived, name) }
	found := make([]*work.Item, len(others)) // in parallel with the Project's list too
	var lists sync.WaitGroup
	for i, o := range others {
		if filepath.Clean(o.Path) == filepath.Clean(project) || isArchived(o.Branch) {
			continue
		}
		if fi, err := os.Stat(filepath.Join(o.Path, "openspec", "changes", o.Branch)); err != nil || !fi.IsDir() {
			continue
		}
		lists.Add(1)
		go func() {
			defer lists.Done()
			changes, _ := list(cli(o.Path))
			for _, c := range changes {
				if c.Name == o.Branch {
					c.Dir = o.Path
					found[i] = &c
				}
			}
		}()
	}
	changes, err := list(run)
	lists.Wait()
	if err != nil {
		return nil, err
	}

	for _, f := range found {
		if f == nil {
			continue
		}
		if i := slices.IndexFunc(changes, func(c work.Item) bool { return c.Name == f.Name }); i >= 0 {
			changes[i] = *f
		} else {
			changes = append(changes, *f)
		}
	}

	v := <-version
	base := setup(project, v)
	var wg sync.WaitGroup
	slots := make(chan struct{}, maxParallel)
	for i := range changes {
		if changes[i].Done > 0 {
			changes[i].Phase = phaseOf(changes[i].Done, changes[i].Total, false)
			continue
		}
		dir, stamp := project, base
		if changes[i].Dir != "" {
			dir, stamp = changes[i].Dir, setup(changes[i].Dir, v)
		}
		if v != "" { // without a version an upgrade would go unnoticed, so leave the fingerprint empty
			changes[i].Stamp = files(filepath.Join(dir, "openspec", "changes", changes[i].Name), stamp)
		}
		if k, ok := known[changes[i].Name]; ok && k.Dir == changes[i].Dir && k.Artifacts != nil && !k.Modified.IsZero() &&
			k.Modified.Equal(changes[i].Modified) && k.Stamp != "" && k.Stamp == changes[i].Stamp {
			changes[i].Phase, changes[i].Artifacts = k.Phase, k.Artifacts
			continue
		}
		wg.Add(1)
		go func(c *work.Item) {
			defer wg.Done()
			slots <- struct{}{}
			defer func() { <-slots }()
			var planned bool
			c.Artifacts, planned = status(cli(dir), c.Name)
			c.Phase = phaseOf(c.Done, c.Total, planned)
		}(&changes[i])
	}
	wg.Wait()
	return append(changes, archived...), nil
}

// archive returns the Changes archived in dir's openspec/changes/archive, by folder name; none when
// it does not exist.
func archive(dir string) ([]work.Item, error) {
	entries, err := os.ReadDir(filepath.Join(dir, "openspec", "changes", "archive"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	var archived []work.Item
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		c := work.Item{Name: e.Name(), Phase: work.Archived}
		if m := dated.FindStringSubmatch(e.Name()); m != nil {
			c.ArchivedOn, c.Name = m[1], m[2]
		}
		archived = append(archived, c)
	}
	return archived, nil
}

func isIn(changes []work.Item, name string) bool {
	return slices.ContainsFunc(changes, func(c work.Item) bool { return c.Name == name })
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

func list(run herdr.Runner) ([]work.Item, error) {
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
	changes := make([]work.Item, len(got.Changes))
	for i, c := range got.Changes {
		changes[i] = work.Item{Name: c.Name, Done: c.CompletedTasks, Total: c.TotalTasks, Modified: c.LastModified}
	}
	return changes, nil
}

// status reads one Change's planning artifacts and whether its planning is complete; any failure
// yields no artifacts and incomplete planning, per the change-source spec.
func status(run herdr.Runner, name string) (map[string]string, bool) {
	out, err := run("status", "--change", name, "--json")
	var got struct {
		IsPlanningComplete bool `json:"isPlanningComplete"`
		Artifacts          []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"artifacts"`
	}
	if err != nil || json.Unmarshal(out, &got) != nil || len(got.Artifacts) == 0 {
		return nil, false
	}
	artifacts := make(map[string]string, len(got.Artifacts))
	for _, a := range got.Artifacts {
		artifacts[a.ID] = a.Status
	}
	return artifacts, got.IsPlanningComplete
}
