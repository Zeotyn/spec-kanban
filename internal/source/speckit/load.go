package speckit

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/zeotyn/spec-kanban/internal/work"
	"github.com/zeotyn/spec-kanban/internal/worktree"
)

// Load returns project's Features as Items, by folder name. A checkout in others (the Project's own
// skipped) whose branch names a Feature it holds supplies that Feature, read there with Dir set: it
// replaces the Project's Feature of that name or is appended. known is not used: reading the files
// costs less than validating a cache. Review Features whose branch is merged into the base are
// Archived.
func (s *Source) Load(project string, others []worktree.Checkout, _ map[string]work.Item) ([]work.Item, error) {
	items, err := features(project)
	if err != nil {
		return nil, err
	}
	for _, o := range others {
		if filepath.Clean(o.Path) == filepath.Clean(project) {
			continue
		}
		it, ok := feature(o.Path, o.Branch)
		if !ok {
			continue
		}
		it.Dir = o.Path
		if i := slices.IndexFunc(items, func(c work.Item) bool { return c.Name == it.Name }); i >= 0 {
			items[i] = it
		} else {
			items = append(items, it)
		}
	}
	s.archive(project, items)
	return items, nil
}

// Open returns the names of the Features in project.
func (*Source) Open(project string) ([]string, error) {
	items, err := features(project)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(items))
	for i, it := range items {
		names[i] = it.Name
	}
	return names, nil
}

// features reads every Feature in project's specs directory.
func features(project string) ([]work.Item, error) {
	entries, err := os.ReadDir(filepath.Join(project, "specs"))
	if err != nil {
		return nil, err
	}
	var items []work.Item
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if it, ok := feature(project, e.Name()); ok {
			items = append(items, it)
		}
	}
	return items, nil
}

// feature reads the Feature name below the checkout dir; false when specs/<name>/spec.md is missing.
func feature(dir, name string) (work.Item, bool) {
	fd := filepath.Join(dir, "specs", name)
	exists := func(f string) bool { _, err := os.Stat(filepath.Join(fd, f)); return err == nil }
	if !exists("spec.md") {
		return work.Item{}, false
	}
	it := work.Item{Name: name, Artifacts: map[string]string{"proposal": "skipped", "specs": "done"}}
	tasks, err := os.ReadFile(filepath.Join(fd, "tasks.md"))
	if err == nil {
		it.Artifacts["tasks"] = "done"
	}
	if exists("plan.md") {
		it.Artifacts["design"] = "done"
	}
	it.Done, it.Total, _ = scan(string(tasks))
	it.Phase = phaseOf(it.Done, it.Total)
	filepath.WalkDir(fd, func(_ string, e fs.DirEntry, err error) error {
		if err == nil && !e.IsDir() {
			if fi, err := e.Info(); err == nil && fi.ModTime().After(it.Modified) {
				it.Modified = fi.ModTime()
			}
		}
		return nil
	})
	return it, true
}

var task = regexp.MustCompile(`(?m)^[ \t]*[-*] \[( |x|X)\] (.*)$`)

// scan counts the checkbox lines of tasks, the checked ones as done; next is "next ▸ <task text>" for
// the first open one, "all tasks done" when there are tasks and none is open, "" without tasks.
func scan(tasks string) (done, total int, next string) {
	for _, m := range task.FindAllStringSubmatch(tasks, -1) {
		total++
		if m[1] != " " {
			done++
		} else if next == "" {
			next = "next ▸ " + strings.TrimSpace(m[2])
		}
	}
	if total > 0 && done == total {
		next = "all tasks done"
	}
	return done, total, next
}

// phaseOf gives an open Feature's Phase from its Progress.
func phaseOf(done, total int) work.Phase {
	switch {
	case total == 0:
		return work.Draft
	case done == total:
		return work.Review
	case done > 0:
		return work.InProgress
	}
	return work.Ready
}

// archive moves the Review Features whose branch is merged into the base to Archived, with one git
// call after detecting the base, and none when no Feature is in Review; a git failure archives none.
func (s *Source) archive(project string, items []work.Item) {
	if !slices.ContainsFunc(items, func(it work.Item) bool { return it.Phase == work.Review }) {
		return
	}
	git := s.git(project)
	out, err := git("branch", "--merged", worktree.Base(s.base, git), "--format=%(refname:short)")
	if err != nil {
		return
	}
	merged := strings.Fields(string(out))
	for i := range items {
		if items[i].Phase == work.Review && slices.Contains(merged, items[i].Name) {
			items[i].Phase = work.Archived
		}
	}
}
