// Package openspec reads a Project's Changes through the openspec CLI and its archive directory.
package openspec

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
	Name        string
	Done, Total int
	Modified    time.Time
	Planning    Planning
	Artifacts   map[string]string // artifact id → done|ready|blocked|skipped, only when Planning was read
	Archived    bool
	ArchivedOn  string // "2026-10-29"; "" when open or the archive folder has no date
}

// maxParallel bounds concurrent `openspec status` calls (each starts a Node process).
const maxParallel = 8

var dated = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})-(.+)$`)

// Load returns the open Changes, read by run (the openspec CLI inside project), in `openspec list`
// order, then the archived ones by folder name.
// Planning state is read only for open Changes without completed tasks.
func Load(project string, run herdr.Runner) ([]Change, error) {
	changes, err := list(run)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	slots := make(chan struct{}, maxParallel)
	for i := range changes {
		if changes[i].Done > 0 {
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
