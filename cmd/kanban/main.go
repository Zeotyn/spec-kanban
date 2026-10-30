// Command kanban is the openspec-kanban herdr plugin.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/zeotyn/openspec-kanban/internal/board"
	"github.com/zeotyn/openspec-kanban/internal/cache"
	"github.com/zeotyn/openspec-kanban/internal/herdr"
	"github.com/zeotyn/openspec-kanban/internal/openspec"
	"github.com/zeotyn/openspec-kanban/internal/project"
	"github.com/zeotyn/openspec-kanban/internal/sidebar"
	"github.com/zeotyn/openspec-kanban/internal/start"
)

const version = "0.1.0"

const usage = `kanban — OpenSpec kanban board for herdr

Usage:
  kanban open       Open the Board popup for the invoking pane's project (plugin action)
  kanban board      Show the Board for $KANBAN_PROJECT (set by kanban open)
  kanban start <change>
                    Run /opsx:apply <change> in a Claude tab of the current directory's project
  kanban launch <change> <pane>
                    Start Claude in pane and send /opsx:apply <change> (internal, spawned by the Board)
  kanban sync       Write every herdr workspace's project summary into its sidebar rows (plugin hook)
  kanban --version  Print the version
`

func main() {
	cmd := ""
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}
	var err error
	switch cmd {
	case "open":
		err = open(os.Getenv("HERDR_PLUGIN_CONTEXT_JSON"), herdr.Exec)
	case "start":
		if len(os.Args) != 3 {
			fmt.Fprintf(os.Stderr, "kanban: start needs exactly one change name\n\n%s", usage)
			os.Exit(2)
		}
		err = startChange(os.Args[2])
	case "launch":
		if len(os.Args) != 4 {
			fmt.Fprintf(os.Stderr, "kanban: launch needs a change name and a pane\n\n%s", usage)
			os.Exit(2)
		}
		err = launch(os.Args[2], os.Args[3], herdr.Exec)
	case "board":
		err = showBoard(os.Getenv("KANBAN_PROJECT"))
	case "sync":
		dir := cache.Dir()
		err = sidebar.Coalesce(dir, func() error { return sidebar.Sync(herdr.Exec, syncLoad(dir, openspecCLI)) })
	case "--version":
		fmt.Println(version)
	case "-h", "--help", "":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "kanban: unknown command %q\n\n%s", cmd, usage)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "kanban:", err)
		os.Exit(1)
	}
}

// open opens the board pane for the project of the pane the action was invoked from.
func open(contextJSON string, run herdr.Runner) error {
	var ctx struct {
		FocusedPaneCwd string `json:"focused_pane_cwd"`
		WorkspaceCwd   string `json:"workspace_cwd"`
	}
	_ = json.Unmarshal([]byte(contextJSON), &ctx) // missing or broken context means no project
	cwd := ctx.FocusedPaneCwd
	if cwd == "" {
		cwd = ctx.WorkspaceCwd
	}
	args := []string{"plugin", "pane", "open", "--plugin", "openspec-kanban", "--entrypoint", "board", "--focus"}
	if proj := project.Find(cwd); proj != "" {
		args = append(args, "--cwd", proj, "--env", "KANBAN_PROJECT="+proj)
	}
	out, err := run(args...)
	os.Stdout.Write(out) // shows up in herdr plugin log list
	return err
}

func startChange(change string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	proj := project.Find(cwd)
	if proj == "" {
		return fmt.Errorf("no OpenSpec project at or above %s", cwd)
	}
	return start.Start(proj, change, herdr.Exec, openspecCLI(proj))
}

// showBoard runs the Board UI. No cwd fallback: herdr starts the pane in the plugin root,
// which is itself an OpenSpec project.
func showBoard(proj string) error {
	dir := cache.Dir()
	running := func() (map[string]bool, error) { return start.Running(proj, herdr.Exec) }
	startOne := func(change string) error { // no CheckOpen: the Board only offers open Changes, and it costs ~0.65 s
		pane, launch, err := start.Prepare(proj, change, herdr.Exec)
		if err != nil || !launch {
			return err
		}
		return spawnLaunch(change, pane)
	}
	_, err := tea.NewProgram(board.New(proj, loader(dir, proj, openspecCLI(proj)), time.Now, cache.Read(dir, proj), running, startOne), tea.WithAltScreen()).Run()
	return err
}

// spawnLaunch runs kanban launch detached in its own session, so it outlives the Board's popup.
func spawnLaunch(change, pane string) error {
	self, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(self, "launch", change, pane)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}

// launch runs start.Launch and, since nobody watches its stderr, reports a failure as a herdr notification.
func launch(change, pane string, run herdr.Runner) error {
	err := start.Launch(change, pane, run)
	if err != nil {
		run("notification", "show", "openspec-kanban: start failed", "--body", err.Error())
	}
	return err
}

// loader reads the Project's Changes, reusing cached planning state, and caches every successful load.
func loader(dir, proj string, run herdr.Runner) func() ([]openspec.Change, error) {
	return func() ([]openspec.Change, error) {
		known := map[string]openspec.Change{}
		for _, c := range cache.Read(dir, proj) { // re-read each time, so every refresh reuses the newest load
			if !c.Archived { // an archived Change may share an open one's name
				known[c.Name] = c
			}
		}
		changes, err := openspec.Load(proj, run, known)
		if err == nil {
			_ = cache.Write(dir, proj, changes) // cache problems never break the Board
		}
		return changes, err
	}
}

// syncLoad loads a Project for the sidebar through the Board's loader, so every sync warms the Board's cache.
func syncLoad(dir string, cli func(proj string) herdr.Runner) func(string) ([]openspec.Change, error) {
	return func(proj string) ([]openspec.Change, error) { return loader(dir, proj, cli(proj))() }
}

// openspecCLI runs the openspec CLI inside proj.
func openspecCLI(proj string) herdr.Runner {
	return func(args ...string) ([]byte, error) { return herdr.Run("openspec", proj, args...) }
}
