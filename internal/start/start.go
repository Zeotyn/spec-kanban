// Package start hands a Change to an agent session in its own herdr tab.
package start

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/zeotyn/openspec-kanban/internal/config"
	"github.com/zeotyn/openspec-kanban/internal/herdr"
	"github.com/zeotyn/openspec-kanban/internal/openspec"
	"github.com/zeotyn/openspec-kanban/internal/worktree"
)

// Worktree holds the worktree settings: whether Start works in the Change's checkout, and the
// configured base of a new worktree ("" to detect it).
type Worktree struct {
	On   bool
	Base string
}

// Start sends prompt to a herdr agent of kind named after the Change: Prepare, then Launch when the
// agent has to be started. openspecRun runs the openspec CLI inside project, git runs git there.
func Start(project, change, kind, prompt string, wt Worktree, run, openspecRun, git herdr.Runner) error {
	if err := checkOpen(project, change, openspecRun); err != nil {
		return err
	}
	pane, tab, err := Prepare(project, change, wt, run, git)
	if err != nil || pane == "" {
		return err
	}
	return Launch(change, kind, prompt, pane, tab, run)
}

// Kind is the agent kind Start launches: configured when set, else the kind herdr reports for the
// agent in pane, else claude. A reported kind herdr cannot start counts as none.
func Kind(configured, pane string, run herdr.Runner) string {
	if configured != "" {
		return configured
	}
	if pane == "" {
		return "claude"
	}
	out, err := run("agent", "get", pane)
	var got struct {
		Result struct {
			Agent struct {
				Agent string `json:"agent"`
			} `json:"agent"`
		} `json:"result"`
	}
	if err != nil || json.Unmarshal(out, &got) != nil || !slices.Contains(config.Kinds, got.Result.Agent.Agent) {
		return "claude"
	}
	return got.Result.Agent.Agent
}

// Prepare returns the pane and tab to launch the agent in, "" for none: it focuses a busy agent tab,
// returns an idle one or opens a new one unfocused (focus moves only in Launch). With wt.On the new
// tab is in the Change's checkout, the one on the branch named like the Change, which herdr creates
// as a worktree when there is none. An agent of that name outside project and its checkouts is
// refused. git runs git inside project.
func Prepare(project, change string, wt Worktree, run, git herdr.Runner) (pane, tab string, err error) {
	var checkouts []worktree.Checkout
	if wt.On {
		checkouts = worktree.List(git)
	}
	dirs := []string{project}
	for _, c := range checkouts {
		dirs = append(dirs, c.Path)
	}
	out, err := run("agent", "get", change)
	switch {
	case err == nil:
		var got struct {
			Result struct {
				Agent struct {
					PaneID string `json:"pane_id"`
					TabID  string `json:"tab_id"`
					Cwd    string `json:"cwd"`
				} `json:"agent"`
			} `json:"result"`
		}
		if err := json.Unmarshal(out, &got); err != nil {
			return "", "", fmt.Errorf("agent get: %w", err)
		}
		if cwd := got.Result.Agent.Cwd; cwd != "" && !insideAny(cwd, dirs) { // no cwd: nothing to check against
			return "", "", fmt.Errorf("agent %s already runs in %s", change, cwd)
		}
		a := got.Result.Agent
		running, err := agentRunning(run, a.PaneID)
		if err != nil {
			return "", "", err
		}
		if running {
			return "", "", step(run, "agent", "focus", change)
		}
		return a.PaneID, a.TabID, nil
	case !strings.Contains(err.Error(), "agent_not_found"):
		return "", "", fmt.Errorf("agent get: %w", err)
	}
	i := slices.IndexFunc(checkouts, func(c worktree.Checkout) bool { return c.Branch == change })
	switch {
	case len(checkouts) == 0 || i >= 0 && filepath.Clean(checkouts[i].Path) == filepath.Clean(project):
		return newTab(run, "--cwd", project, "--label", change, "--no-focus")
	case i >= 0:
		path := checkouts[i].Path
		out, err := run("worktree", "open", "--cwd", project, "--path", path, "--label", change, "--no-focus")
		if err != nil {
			return "", "", fmt.Errorf("worktree open: %w", err)
		}
		var opened struct {
			Result struct {
				AlreadyOpen bool `json:"already_open"`
				Workspace   struct {
					WorkspaceID string `json:"workspace_id"`
				} `json:"workspace"`
			} `json:"result"`
		}
		if err := json.Unmarshal(out, &opened); err != nil {
			return "", "", fmt.Errorf("worktree open: %w", err)
		}
		if opened.Result.AlreadyOpen { // its root pane may be in use
			return newTab(run, "--workspace", opened.Result.Workspace.WorkspaceID, "--cwd", path, "--label", change, "--no-focus")
		}
		return created("worktree open", out)
	}
	args := []string{"worktree", "create", "--cwd", project, "--branch", change}
	ref := change
	if _, err := git("rev-parse", "--verify", "--quiet", "refs/heads/"+change); err != nil {
		ref = worktree.Base(wt.Base, git)
		args = append(args, "--base", ref)
	}
	if _, err := git("cat-file", "-e", ref+":openspec/changes/"+change); err != nil {
		return "", "", fmt.Errorf("%s is not on %s: commit it there or set base", change, ref)
	}
	out, err = run(append(args, "--label", change, "--no-focus")...)
	if err != nil {
		return "", "", fmt.Errorf("worktree create: %w", err)
	}
	return created("worktree create", out)
}

// newTab runs tab create with args and returns its root pane and tab.
func newTab(run herdr.Runner, args ...string) (pane, tab string, err error) {
	out, err := run(append([]string{"tab", "create"}, args...)...)
	if err != nil {
		return "", "", fmt.Errorf("tab create: %w", err)
	}
	return created("tab create", out)
}

// created reads the root pane and tab from what tab create, worktree create or worktree open printed.
func created(cmd string, out []byte) (pane, tab string, err error) {
	var got struct {
		Result struct {
			RootPane struct {
				PaneID string `json:"pane_id"`
				TabID  string `json:"tab_id"`
			} `json:"root_pane"`
			Tab struct {
				TabID string `json:"tab_id"`
			} `json:"tab"`
		} `json:"result"`
	}
	err = json.Unmarshal(out, &got)
	pane, tab = got.Result.RootPane.PaneID, got.Result.Tab.TabID
	if tab == "" {
		tab = got.Result.RootPane.TabID
	}
	if err != nil || pane == "" || tab == "" {
		return "", "", fmt.Errorf("%s: no pane or tab id in %s", cmd, out)
	}
	return pane, tab, nil
}

// Launch focuses tab, starts an agent of kind named <change> in pane, waits until it is ready and
// sends prompt.
func Launch(change, kind, prompt, pane, tab string, run herdr.Runner) error {
	if _, err := run("tab", "focus", tab); err != nil {
		return fmt.Errorf("tab focus: %w", err)
	}
	if _, err := run("agent", "start", change, "--kind", kind, "--pane", pane, "--timeout", "60000"); err != nil {
		return fmt.Errorf("%s did not become ready in tab %q, the prompt was not sent: %w", kind, change, err)
	}
	if _, err := run("agent", "prompt", change, prompt); err != nil {
		return fmt.Errorf("%s runs in tab %q but the prompt was not sent: %w", kind, change, err)
	}
	return nil
}

// Running returns the names of the herdr agents that work in one of dirs (the Project and its
// checkouts) or below it. Only agents started with a name count, and Start names them after their
// Change; every agent_status counts.
func Running(dirs []string, run herdr.Runner) (map[string]bool, error) {
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
		if a.Name != "" && insideAny(a.Cwd, dirs) {
			names[a.Name] = true
		}
	}
	return names, nil
}

// insideAny reports whether dir is one of dirs or lies below one.
func insideAny(dir string, dirs []string) bool {
	dir = filepath.Clean(dir)
	return slices.ContainsFunc(dirs, func(d string) bool {
		d = filepath.Clean(d)
		return dir == d || strings.HasPrefix(dir, d+string(filepath.Separator))
	})
}

// agentRunning reports whether anything but the shell holds the foreground of pane: the agent, or a
// program the user started there, which is focused rather than overwritten. Its argv0 says nothing,
// since node-based agents show up as node.
func agentRunning(run herdr.Runner, pane string) (bool, error) {
	out, err := run("pane", "process-info", "--pane", pane)
	if err != nil {
		return false, fmt.Errorf("pane process-info: %w", err)
	}
	var info struct {
		Result struct {
			ProcessInfo struct {
				ShellPID int `json:"shell_pid"`
				Group    int `json:"foreground_process_group_id"`
			} `json:"process_info"`
		} `json:"result"`
	}
	if err := json.Unmarshal(out, &info); err != nil || info.Result.ProcessInfo.ShellPID == 0 {
		return false, fmt.Errorf("pane process-info: unexpected output %s", out)
	}
	return info.Result.ProcessInfo.Group != info.Result.ProcessInfo.ShellPID, nil
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
