// Package start hands a Change to a Claude agent session in its own herdr tab.
package start

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/zeotyn/openspec-kanban/internal/herdr"
)

// Start runs /opsx:apply <change> in a herdr agent named after the Change: it focuses the agent if it
// is running, restarts it in its tab if it exited, and opens a new tab otherwise. openspec runs the
// openspec CLI inside project.
func Start(project, change string, run, openspec herdr.Runner) error {
	if err := checkOpen(project, change, openspec); err != nil {
		return err
	}

	var pane string
	reused := false
	out, err := run("agent", "get", change)
	switch {
	case err == nil:
		var got struct {
			Result struct {
				Agent struct {
					PaneID string `json:"pane_id"`
				} `json:"agent"`
			} `json:"result"`
		}
		if err := json.Unmarshal(out, &got); err != nil {
			return fmt.Errorf("agent get: %w", err)
		}
		pane, reused = got.Result.Agent.PaneID, true
		running, err := agentRunning(run, pane)
		if err != nil {
			return err
		}
		if running {
			return step(run, "agent", "focus", change)
		}
	case !strings.Contains(err.Error(), "agent_not_found"):
		return fmt.Errorf("agent get: %w", err)
	default:
		out, err := run("tab", "create", "--cwd", project, "--label", change, "--focus")
		if err != nil {
			return fmt.Errorf("tab create: %w", err)
		}
		var tab struct {
			Result struct {
				RootPane struct {
					PaneID string `json:"pane_id"`
				} `json:"root_pane"`
			} `json:"result"`
		}
		if err := json.Unmarshal(out, &tab); err != nil || tab.Result.RootPane.PaneID == "" {
			return fmt.Errorf("tab create: no pane id in %s", out)
		}
		pane = tab.Result.RootPane.PaneID
	}

	if _, err := run("agent", "start", change, "--kind", "claude", "--pane", pane, "--timeout", "60000"); err != nil {
		return fmt.Errorf("claude did not become ready in tab %q, the prompt was not sent: %w", change, err)
	}
	if _, err := run("agent", "prompt", change, "/opsx:apply "+change); err != nil {
		return fmt.Errorf("claude runs in tab %q but the prompt was not sent: %w", change, err)
	}
	if reused {
		return step(run, "agent", "focus", change) // tab create focuses new tabs only
	}
	return nil
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

func checkOpen(project, change string, openspec herdr.Runner) error {
	out, err := openspec("list", "--json")
	if err != nil {
		return fmt.Errorf("openspec list: %w", err)
	}
	var list struct {
		Changes []struct {
			Name string `json:"name"`
		} `json:"changes"`
	}
	if err := json.Unmarshal(out, &list); err != nil {
		return fmt.Errorf("openspec list: %w", err)
	}
	for _, c := range list.Changes {
		if c.Name == change {
			return nil
		}
	}
	return fmt.Errorf("%q is not an open change of %s", change, filepath.Base(project))
}

func step(run herdr.Runner, args ...string) error {
	if _, err := run(args...); err != nil {
		return fmt.Errorf("%s %s: %w", args[0], args[1], err)
	}
	return nil
}
