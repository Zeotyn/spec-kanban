// Package start hands a Change to a Claude agent session in its own herdr tab.
package start

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/zeotyn/openspec-kanban/internal/herdr"
	"github.com/zeotyn/openspec-kanban/internal/openspec"
)

// Start runs /opsx:apply <change> in a herdr agent named after the Change: Prepare, then Launch when
// Claude has to be started. openspecRun runs the openspec CLI inside project.
func Start(project, change string, run, openspecRun herdr.Runner) error {
	if err := checkOpen(project, change, openspecRun); err != nil {
		return err
	}
	pane, launch, err := Prepare(project, change, run)
	if err != nil || !launch {
		return err
	}
	return Launch(change, pane, run)
}

// Prepare focuses the Change's agent tab and returns the pane Claude has to be launched in. It only
// focuses a running agent (launch false), focuses the tab of an exited one, and opens and focuses a
// new tab otherwise. An agent of that name working outside project is refused.
func Prepare(project, change string, run herdr.Runner) (pane string, launch bool, err error) {
	out, err := run("agent", "get", change)
	switch {
	case err == nil:
		var got struct {
			Result struct {
				Agent struct {
					PaneID string `json:"pane_id"`
					Cwd    string `json:"cwd"`
				} `json:"agent"`
			} `json:"result"`
		}
		if err := json.Unmarshal(out, &got); err != nil {
			return "", false, fmt.Errorf("agent get: %w", err)
		}
		if cwd := got.Result.Agent.Cwd; cwd != "" && !inside(cwd, project) { // no cwd: nothing to check against
			return "", false, fmt.Errorf("agent %s already runs in %s", change, cwd)
		}
		pane = got.Result.Agent.PaneID
		running, err := agentRunning(run, pane)
		if err != nil {
			return "", false, err
		}
		return pane, !running, step(run, "agent", "focus", change)
	case !strings.Contains(err.Error(), "agent_not_found"):
		return "", false, fmt.Errorf("agent get: %w", err)
	}
	out, err = run("tab", "create", "--cwd", project, "--label", change, "--focus")
	if err != nil {
		return "", false, fmt.Errorf("tab create: %w", err)
	}
	var tab struct {
		Result struct {
			RootPane struct {
				PaneID string `json:"pane_id"`
			} `json:"root_pane"`
		} `json:"result"`
	}
	if err := json.Unmarshal(out, &tab); err != nil || tab.Result.RootPane.PaneID == "" {
		return "", false, fmt.Errorf("tab create: no pane id in %s", out)
	}
	return tab.Result.RootPane.PaneID, true, nil
}

// Launch starts Claude as the agent <change> in pane, waits until it is ready and sends /opsx:apply <change>.
func Launch(change, pane string, run herdr.Runner) error {
	if _, err := run("agent", "start", change, "--kind", "claude", "--pane", pane, "--timeout", "60000"); err != nil {
		return fmt.Errorf("claude did not become ready in tab %q, the prompt was not sent: %w", change, err)
	}
	if _, err := run("agent", "prompt", change, "/opsx:apply "+change); err != nil {
		return fmt.Errorf("claude runs in tab %q but the prompt was not sent: %w", change, err)
	}
	return nil
}

// Running returns the names of the herdr agents that work in project or below it. Only agents
// started with a name count, and Start names them after their Change; every agent_status counts.
func Running(project string, run herdr.Runner) (map[string]bool, error) {
	out, err := run("agent", "list")
	if err != nil {
		return nil, fmt.Errorf("agent list: %w", err)
	}
	var list struct {
		Result struct {
			Agents []struct {
				Name string `json:"name"`
				Cwd  string `json:"cwd"`
			} `json:"agents"`
		} `json:"result"`
	}
	if err := json.Unmarshal(out, &list); err != nil {
		return nil, fmt.Errorf("agent list: %w", err)
	}
	names := map[string]bool{}
	for _, a := range list.Result.Agents {
		if a.Name != "" && inside(a.Cwd, project) {
			names[a.Name] = true
		}
	}
	return names, nil
}

// inside reports whether dir is project or lies below it.
func inside(dir, project string) bool {
	dir, project = filepath.Clean(dir), filepath.Clean(project)
	return dir == project || strings.HasPrefix(dir, project+string(filepath.Separator))
}

// agentRunning reports whether Claude runs in the foreground of pane. It is false when only the
// shell is left (Claude exited) and an error when another program holds the pane.
func agentRunning(run herdr.Runner, pane string) (bool, error) {
	out, err := run("pane", "process-info", "--pane", pane)
	if err != nil {
		return false, fmt.Errorf("pane process-info: %w", err)
	}
	var info struct {
		Result struct {
			ProcessInfo struct {
				ShellPID   int `json:"shell_pid"`
				Group      int `json:"foreground_process_group_id"`
				Foreground []struct {
					Argv0 string `json:"argv0"`
				} `json:"foreground_processes"`
			} `json:"process_info"`
		} `json:"result"`
	}
	if err := json.Unmarshal(out, &info); err != nil || info.Result.ProcessInfo.ShellPID == 0 {
		return false, fmt.Errorf("pane process-info: unexpected output %s", out)
	}
	p := info.Result.ProcessInfo
	for _, proc := range p.Foreground {
		if proc.Argv0 == "claude" {
			return true, nil
		}
	}
	if p.Group == p.ShellPID {
		return false, nil
	}
	busy := "another program"
	if len(p.Foreground) > 0 {
		busy = p.Foreground[0].Argv0
	}
	return false, fmt.Errorf("pane %s runs %s, not Claude: quit it and start again", pane, busy)
}

// checkOpen fails unless change is an open Change of project.
func checkOpen(project, change string, run herdr.Runner) error {
	names, err := openspec.OpenNames(run)
	if err != nil {
		return err
	}
	if !slices.Contains(names, change) {
		return fmt.Errorf("%q is not an open change of %s", change, filepath.Base(project))
	}
	return nil
}

func step(run herdr.Runner, args ...string) error {
	if _, err := run(args...); err != nil {
		return fmt.Errorf("%s %s: %w", args[0], args[1], err)
	}
	return nil
}
