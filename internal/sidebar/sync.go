package sidebar

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/zeotyn/openspec-kanban/internal/herdr"
	"github.com/zeotyn/openspec-kanban/internal/project"
	"github.com/zeotyn/openspec-kanban/internal/work"
)

// maxLoads bounds concurrent Project loads; with cold caches each may start 8 openspec processes.
const maxLoads = 4

// Sync runs one pass: it reports the summary of every herdr workspace's Project, clears it for
// workspaces without one and leaves workspaces alone whose Project fails to load. It returns
// herdr's error when it cannot list workspaces or panes, else the joined load and report errors.
// detect tells a Project directory (project.Find).
func Sync(run herdr.Runner, detect func(dir string) bool, load func(project string) ([]work.Item, error)) error {
	var workspaces struct {
		Result struct {
			Workspaces []struct {
				ID string `json:"workspace_id"`
			} `json:"workspaces"`
		} `json:"result"`
	}
	var panes struct {
		Result struct {
			Panes []struct {
				Workspace     string `json:"workspace_id"`
				Cwd           string `json:"cwd"`
				ForegroundCwd string `json:"foreground_cwd"`
				Focused       bool   `json:"focused"`
			} `json:"panes"`
		} `json:"result"`
	}
	for _, l := range []struct {
		args []string
		v    any
	}{{[]string{"workspace", "list"}, &workspaces}, {[]string{"pane", "list"}, &panes}} {
		out, err := run(l.args...)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(out, l.v); err != nil {
			return fmt.Errorf("herdr %s %s: %w", l.args[0], l.args[1], err)
		}
	}

	shown := map[string]string{} // workspace → Project
	for _, p := range panes.Result.Panes {
		dir := p.ForegroundCwd
		if dir == "" {
			dir = p.Cwd
		}
		if proj := project.Find(dir, detect); proj != "" && (shown[p.Workspace] == "" || p.Focused) {
			shown[p.Workspace] = proj
		}
	}

	type result struct {
		changes []work.Item
		err     error
	}
	results := map[string]*result{}
	var wg sync.WaitGroup
	slots := make(chan struct{}, maxLoads)
	for _, proj := range shown {
		if results[proj] != nil {
			continue
		}
		r := &result{}
		results[proj] = r
		wg.Add(1)
		go func() {
			defer wg.Done()
			slots <- struct{}{}
			defer func() { <-slots }()
			r.changes, r.err = load(proj)
		}()
	}
	wg.Wait()

	var errs []error
	for proj, r := range results {
		if r.err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", proj, r.err))
		}
	}
	for _, w := range workspaces.Result.Workspaces {
		args := []string{"workspace", "report-metadata", w.ID, "--source", "openspec-kanban"}
		if proj := shown[w.ID]; proj == "" {
			args = append(args, "--clear-token", "kanban_change", "--clear-token", "kanban")
		} else if r := results[proj]; r.err != nil {
			continue // keep the last summary through a transient failure
		} else if change, phases := Summary(r.changes); change == "" {
			args = append(args, "--clear-token", "kanban_change", "--token", "kanban="+phases)
		} else {
			args = append(args, "--token", "kanban_change="+change, "--token", "kanban="+phases)
		}
		if _, err := run(args...); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
