// Command kanban is the openspec-kanban herdr plugin.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/zeotyn/openspec-kanban/internal/board"
	"github.com/zeotyn/openspec-kanban/internal/herdr"
	"github.com/zeotyn/openspec-kanban/internal/openspec"
	"github.com/zeotyn/openspec-kanban/internal/project"
	"github.com/zeotyn/openspec-kanban/internal/start"
)

const version = "0.1.0"

const usage = `kanban — OpenSpec kanban board for herdr

Usage:
  kanban open       Open the Board popup for the invoking pane's project (plugin action)
  kanban board      Show the Board for $KANBAN_PROJECT (set by kanban open)
  kanban start <change>
                    Run /opsx:apply <change> in a Claude tab of the current directory's project
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
	case "board":
		err = showBoard(os.Getenv("KANBAN_PROJECT"))
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
	openspec := func(args ...string) ([]byte, error) { return herdr.Run("openspec", proj, args...) }
	return start.Start(proj, change, herdr.Exec, openspec)
}

// showBoard runs the Board UI. No cwd fallback: herdr starts the pane in the plugin root,
// which is itself an OpenSpec project.
func showBoard(proj string) error {
	load := func() ([]openspec.Change, error) {
		return openspec.Load(proj, func(args ...string) ([]byte, error) { return herdr.Run("openspec", proj, args...) })
	}
	_, err := tea.NewProgram(board.New(proj, load, time.Now), tea.WithAltScreen()).Run()
	return err
}
