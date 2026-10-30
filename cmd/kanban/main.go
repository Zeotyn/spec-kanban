// Command kanban is the openspec-kanban herdr plugin.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime/debug"
	"slices"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/zeotyn/openspec-kanban/internal/board"
	"github.com/zeotyn/openspec-kanban/internal/cache"
	"github.com/zeotyn/openspec-kanban/internal/config"
	"github.com/zeotyn/openspec-kanban/internal/herdr"
	"github.com/zeotyn/openspec-kanban/internal/project"
	"github.com/zeotyn/openspec-kanban/internal/sidebar"
	"github.com/zeotyn/openspec-kanban/internal/source/openspec"
	"github.com/zeotyn/openspec-kanban/internal/start"
	"github.com/zeotyn/openspec-kanban/internal/work"
	"github.com/zeotyn/openspec-kanban/internal/worktree"
)

// version is stamped by release builds with -ldflags "-X main.version=<version>".
var version = ""

// versionString is the stamped version, else the module version Go recorded in the build info.
func versionString() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Version
	}
	return "(devel)"
}

const usage = `kanban — OpenSpec kanban board for herdr

Usage:
  kanban open       Open the Board popup for the invoking pane's project (plugin action)
  kanban board      Show the Board for $KANBAN_PROJECT (set by kanban open)
  kanban start <change>
                    Send the configured prompt for <change> to an agent in a new tab of the current
                    directory's project, or in a worktree of the project on a branch named <change>
  kanban launch <change> <pane> <tab> <kind> <prompt>
                    Focus tab, start an agent of kind in pane, send prompt and watch it until it hands
                    the Change off or goes away (internal, spawned by the Board)
  kanban resume <change> <pane> <tab>
                    Focus tab and reopen the session recorded for <change> in pane (internal, spawned by the Board)
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
		err = startChange(os.Args[2], os.Stderr)
	case "launch":
		if len(os.Args) != 7 {
			fmt.Fprintf(os.Stderr, "kanban: launch needs a change name, a pane, a tab, an agent kind and a prompt\n\n%s", usage)
			os.Exit(2)
		}
		err = launch(os.Args[2], os.Args[3], os.Args[4], os.Args[5], os.Args[6], herdr.Exec)
	case "resume":
		if len(os.Args) != 5 {
			fmt.Fprintf(os.Stderr, "kanban: resume needs a change name, a pane and a tab\n\n%s", usage)
			os.Exit(2)
		}
		err = resumeChange(os.Args[2], os.Args[3], os.Args[4], herdr.Exec)
	case "board":
		err = showBoard(os.Getenv("KANBAN_PROJECT"))
	case "sync":
		dir := cache.Dir()
		err = sidebar.Coalesce(dir, func() error { return sidebar.Sync(herdr.Exec, syncLoad(dir, openspecCLI, gitCLI)) })
	case "--version":
		fmt.Println(versionString())
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
		FocusedPaneID  string `json:"focused_pane_id"`
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
	if ctx.FocusedPaneID != "" { // the Board starts the agent kind running there
		args = append(args, "--env", "KANBAN_MAIN_PANE="+ctx.FocusedPaneID)
	}
	out, err := run(args...)
	os.Stdout.Write(out) // shows up in herdr plugin log list
	return err
}

// startChange starts change from a shell; a config warning goes to stderr and Start goes on with defaults.
func startChange(change string, stderr io.Writer) error {
	cfg, warn := config.Load(config.Dir(herdr.Exec))
	if warn != nil {
		fmt.Fprintln(stderr, "kanban:", warn)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	proj := project.Find(cwd)
	if proj == "" {
		return fmt.Errorf("no OpenSpec project at or above %s", cwd)
	}
	kind := start.Kind(cfg.Agent, os.Getenv("HERDR_PANE_ID"), herdr.Exec)
	wt := start.Worktree{On: cfg.Worktree, Base: cfg.Base}
	return start.Start(proj, change, kind, cfg.PromptFor(change), wt, herdr.Exec, openspecCLI(proj), gitCLI(proj))
}

// showBoard runs the Board UI. No cwd fallback: herdr starts the pane in the plugin root,
// which is itself an OpenSpec project.
func showBoard(proj string) error {
	dir := cache.Dir()
	cfg, warn := config.Load(config.Dir(herdr.Exec))
	notice := ""
	if warn != nil {
		notice = warn.Error()
	}
	git := gitCLI(proj)
	running := func() (map[string]bool, error) { return start.Running(proj, herdr.Exec, git) }
	wt := start.Worktree{On: cfg.Worktree, Base: cfg.Base}
	startOne := func(change string) error { // no checkOpen: the Board only offers open Changes, and it costs ~0.65 s
		pane, tab, err := start.Prepare(proj, change, wt, herdr.Exec, git)
		if err != nil || pane == "" {
			return err
		}
		kind := start.Kind(cfg.Agent, os.Getenv("KANBAN_MAIN_PANE"), herdr.Exec) // resolved on s, not on open
		return spawn("launch", change, pane, tab, kind, cfg.PromptFor(change))
	}
	reopenable := func(change string) bool { return start.ResumeArgs(cache.ReadSession(dir, proj, change)) != nil }
	resumable := func(change string) (string, bool) {
		return cache.ReadSession(dir, proj, change).Agent, reopenable(change)
	}
	resume := func(change string) error {
		pane, tab, err := start.Prepare(proj, change, wt, herdr.Exec, git)
		if err != nil || pane == "" {
			return err
		}
		if !reopenable(change) { // a Running badge without a record, and the agent was idle
			return fmt.Errorf("no reopenable session for %s", change)
		}
		return spawn("resume", change, pane, tab)
	}
	_, err := tea.NewProgram(board.New(proj, asItems(loader(dir, proj, openspecCLI, git)), time.Now, openspec.Items(cache.Read(dir, proj)), running, startOne,
		resumable, resume, cfg.Hide, notice), tea.WithAltScreen()).Run()
	return err
}

// spawn runs kanban with args detached in its own session, so it outlives the Board's popup.
// Only the child moves focus, so a popup that closes on focus loss cannot kill the Board before the spawn.
func spawn(args ...string) error {
	self, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(self, args...)
	cmd.Env = childEnv(os.Environ())
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}

// childEnv drops HERDR_PANE_ID and HERDR_TAB_ID from env: the popup they name is gone.
func childEnv(env []string) []string {
	return slices.DeleteFunc(slices.Clone(env), func(kv string) bool {
		return strings.HasPrefix(kv, "HERDR_PANE_ID=") || strings.HasPrefix(kv, "HERDR_TAB_ID=")
	})
}

// launch runs start.Launch and reports a failure as a herdr notification: nobody reads its stderr.
// A Start from the Board ($KANBAN_PROJECT set) is then watched until it hands off or its agent goes away.
func launch(change, pane, tab, kind, prompt string, run herdr.Runner) error {
	if err := start.Launch(change, kind, prompt, pane, tab, run); err != nil {
		run("notification", "show", "openspec-kanban: start failed", "--body", err.Error())
		return err
	}
	proj := os.Getenv("KANBAN_PROJECT")
	if proj == "" {
		return nil
	}
	cfg, _ := config.Load(config.Dir(run)) // a broken config was already shown by the Board
	dir, git, loadGit := cache.Dir(), gitCLI(proj), gitCLI(proj)
	if !cfg.Worktree { // the agent works in the Project: a leftover checkout of the Change must not supply its Progress
		loadGit = func(...string) ([]byte, error) { return nil, errors.New("worktree off") }
	}
	start.Watch(start.Watched{Project: proj, Change: change, Main: os.Getenv("KANBAN_MAIN_PANE"), Base: cfg.Base, On: cfg.Worktree,
		Load:   loader(dir, proj, openspecCLI, loadGit),
		Record: func(s herdr.Session) { _ = cache.WriteSession(dir, proj, change, s) }, // a lost record only stops resume from the Board
	}, run, git)
	return nil
}

// resumeChange reopens the session recorded for change of $KANBAN_PROJECT and reports a failure as
// a herdr notification.
func resumeChange(change, pane, tab string, run herdr.Runner) error {
	s := cache.ReadSession(cache.Dir(), os.Getenv("KANBAN_PROJECT"), change)
	err := start.Resume(change, s, pane, tab, run)
	if err != nil {
		run("notification", "show", "openspec-kanban: resume failed", "--body", err.Error())
	}
	return err
}

// loader reads the Project's Changes and those worked on in its other checkouts, reusing cached
// planning state, and caches every successful load.
func loader(dir, proj string, cli func(string) herdr.Runner, git herdr.Runner) func() ([]openspec.Change, error) {
	return func() ([]openspec.Change, error) {
		known := map[string]openspec.Change{}
		for _, c := range cache.Read(dir, proj) { // re-read each time, so every refresh reuses the newest load
			if !c.Archived { // an archived Change may share an open one's name
				known[c.Name] = c
			}
		}
		changes, err := openspec.Load(proj, cli, worktree.List(git), known)
		if err == nil {
			_ = cache.Write(dir, proj, changes) // cache problems never break the Board
		}
		return changes, err
	}
}

// syncLoad loads a Project for the sidebar through the Board's loader, so every sync warms the Board's cache.
func syncLoad(dir string, cli, git func(proj string) herdr.Runner) func(string) ([]work.Item, error) {
	return func(proj string) ([]work.Item, error) { return asItems(loader(dir, proj, cli, git(proj)))() }
}

// asItems converts load's Changes to Items.
func asItems(load func() ([]openspec.Change, error)) func() ([]work.Item, error) {
	return func() ([]work.Item, error) {
		changes, err := load()
		return openspec.Items(changes), err
	}
}

// openspecCLI runs the openspec CLI inside proj.
func openspecCLI(proj string) herdr.Runner {
	return func(args ...string) ([]byte, error) { return herdr.Run("openspec", proj, args...) }
}

// gitCLI runs git inside proj.
func gitCLI(proj string) herdr.Runner {
	return func(args ...string) ([]byte, error) { return herdr.Run("git", proj, args...) }
}
