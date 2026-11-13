// Command kanban is the spec-kanban herdr plugin.
package main

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/zeotyn/spec-kanban/internal/board"
	"github.com/zeotyn/spec-kanban/internal/cache"
	"github.com/zeotyn/spec-kanban/internal/config"
	"github.com/zeotyn/spec-kanban/internal/herdr"
	"github.com/zeotyn/spec-kanban/internal/project"
	"github.com/zeotyn/spec-kanban/internal/sidebar"
	"github.com/zeotyn/spec-kanban/internal/source/openspec"
	"github.com/zeotyn/spec-kanban/internal/source/speckit"
	"github.com/zeotyn/spec-kanban/internal/start"
	"github.com/zeotyn/spec-kanban/internal/work"
	"github.com/zeotyn/spec-kanban/internal/worktree"
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
  kanban run <change> <pane> <tab> <kind> <command>
                    Focus tab, start an agent of kind in pane and send command, without watching it
                    (internal, spawned by the Board for Verify and Archive)
  kanban resume <change> <pane> <tab>
                    Focus tab, reopen the session recorded for <change> in pane and watch it like launch
                    (internal, spawned by the Board)
  kanban escalate <change> <reason>
                    Ask the Main agent for a human's help with <change> (run by the Change's agent)
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
		detect, _ := sources(config.Default()) // the config only matters once the Project is found
		err = open(os.Getenv("HERDR_PLUGIN_CONTEXT_JSON"), herdr.Exec, detect)
	case "start":
		if len(os.Args) != 3 {
			fmt.Fprintf(os.Stderr, "kanban: start needs exactly one change name\n\n%s", usage)
			os.Exit(2)
		}
		err = startChange(os.Args[2], os.Stderr, herdr.Exec)
	case "launch":
		if len(os.Args) != 7 {
			fmt.Fprintf(os.Stderr, "kanban: launch needs a change name, a pane, a tab, an agent kind and a prompt\n\n%s", usage)
			os.Exit(2)
		}
		err = launch(os.Args[2], os.Args[3], os.Args[4], os.Args[5], os.Args[6], herdr.Exec)
	case "run":
		if len(os.Args) != 7 {
			fmt.Fprintf(os.Stderr, "kanban: run needs a change name, a pane, a tab, an agent kind and a command\n\n%s", usage)
			os.Exit(2)
		}
		err = runCommand(os.Args[2], os.Args[3], os.Args[4], os.Args[5], os.Args[6], herdr.Exec)
	case "resume":
		if len(os.Args) != 5 {
			fmt.Fprintf(os.Stderr, "kanban: resume needs a change name, a pane and a tab\n\n%s", usage)
			os.Exit(2)
		}
		err = resumeChange(os.Args[2], os.Args[3], os.Args[4], herdr.Exec)
	case "escalate":
		if len(os.Args) < 4 || strings.TrimSpace(strings.Join(os.Args[3:], "")) == "" {
			fmt.Fprintf(os.Stderr, "kanban: escalate needs a change name and a reason\n\n%s", usage)
			os.Exit(2)
		}
		cfg, _ := config.Load(config.Dir(herdr.Exec)) // only the Source preference matters here
		detect, sourceFor := sources(cfg)
		var cwd string
		if cwd, err = os.Getwd(); err == nil {
			err = escalate(os.Args[2], os.Args[3:], cwd, os.Stdout, herdr.Exec, detect, sourceFor, gitCLI)
		}
	case "board":
		err = showBoard(os.Getenv("KANBAN_PROJECT"))
	case "sync":
		dir := cache.Dir()
		cfg, _ := config.Load(config.Dir(herdr.Exec)) // the Board and kanban start show a broken config
		detect, sourceFor := sources(cfg)
		err = refresh(dir, herdr.Exec, detect, sourceFor, gitCLI)
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

// open opens the board pane for the project of the pane the action was invoked from; detect
// recognises a Project.
func open(contextJSON string, run herdr.Runner, detect func(dir string) bool) error {
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
	args := []string{"plugin", "pane", "open", "--plugin", "spec-kanban", "--entrypoint", "board", "--focus"}
	if proj := project.Find(cwd, detect); proj != "" {
		args = append(args, "--cwd", proj, "--env", "KANBAN_PROJECT="+proj)
	}
	if ctx.FocusedPaneID != "" { // the Board starts the agent kind running there
		args = append(args, "--env", "KANBAN_MAIN_PANE="+ctx.FocusedPaneID)
	}
	out, err := run(args...)
	os.Stdout.Write(out) // shows up in herdr plugin log list
	return err
}

// startChange starts change from a shell through run; a config warning goes to stderr and Start goes on.
func startChange(change string, stderr io.Writer, run herdr.Runner) error {
	cfg, warn := config.Load(config.Dir(run))
	if warn != nil {
		fmt.Fprintln(stderr, "kanban:", warn)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	detect, sourceFor := sources(cfg)
	proj := project.Find(cwd, detect)
	if proj == "" {
		return fmt.Errorf("no OpenSpec or Spec Kit project at or above %s", cwd)
	}
	src := sourceFor(proj)
	kind := start.Kind(cfg.Agent, os.Getenv("HERDR_PANE_ID"), run)
	wt := start.Worktree{On: cfg.Worktree, Base: cfg.Base}
	return start.Start(proj, change, kind, cfg.PromptFor(change, src.Prompt(proj, kind)), wt, run, src, gitCLI(proj))
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
	_, sourceFor := sources(cfg)
	src := sourceFor(proj)
	git := gitCLI(proj)
	wt := start.Worktree{On: cfg.Worktree, Base: cfg.Base}
	startOne := func(change string) error { // no checkOpen: the Board only offers open Changes, and it costs ~0.65 s
		pane, tab, err := start.Prepare(proj, change, wt, herdr.Exec, src, git)
		if err != nil || pane == "" {
			return err
		}
		kind := start.Kind(cfg.Agent, os.Getenv("KANBAN_MAIN_PANE"), herdr.Exec) // resolved on s, not on open
		self, err := os.Executable()
		if err != nil {
			return err
		}
		return spawn("launch", change, pane, tab, kind, boardPrompt(cfg, src, proj, change, kind, self))
	}
	act := func(change string, a work.Action) error { // the Source's command alone: prompt and instructions are for Start
		command := src.Command(a, change)
		pane, tab, err := start.Run(proj, change, command, wt, herdr.Exec, src, git)
		if err != nil || pane == "" {
			return err
		}
		kind := start.Kind(cfg.Agent, os.Getenv("KANBAN_MAIN_PANE"), herdr.Exec)
		return spawn("run", change, pane, tab, kind, command)
	}
	reopenable := func(change string) bool { return start.ResumeArgs(cache.ReadSession(dir, proj, change)) != nil }
	resumable := func(change string) (string, bool) {
		return cache.ReadSession(dir, proj, change).Agent, reopenable(change)
	}
	resume := func(change string) error {
		pane, tab, err := start.Prepare(proj, change, wt, herdr.Exec, src, git)
		if err != nil || pane == "" {
			return err
		}
		if !reopenable(change) { // a Running badge without a record, and the agent was idle
			return fmt.Errorf("no reopenable session for %s", change)
		}
		return spawn("resume", change, pane, tab)
	}
	_, err := tea.NewProgram(board.New(proj, func(it work.Item) work.Docs { return src.Docs(proj, it) }, loader(dir, proj, src, git), time.Now, cache.Read(dir, proj), lookup(proj, dir, cache.UserDir(), herdr.Exec, git), startOne,
		resumable, resume, act, lacking(src), cfg.Hide, notice, versionString()), tea.WithAltScreen(), tea.WithMouseCellMotion()).Run()
	return err
}

// lookup returns the Board's agent lookup: the names of proj's Running agents, and the Changes
// whose attention marker in user still holds for those agents and the Items cached in dir.
func lookup(proj, dir, user string, run, git herdr.Runner) func() (running, waiting map[string]bool, err error) {
	return func() (map[string]bool, map[string]bool, error) {
		agents, err := start.Running(proj, run, git) // nil on error: only Handoff markers hold
		names := map[string]bool{}
		for name := range agents {
			names[name] = true
		}
		return names, start.Waiting(cache.ReadMarks(user, proj), agents, cache.Read(dir, proj)), err
	}
}

// boardPrompt is the prompt of a Start from the Board: what kanban start sends, then the sentence
// that tells the agent how to escalate with the plugin binary self.
func boardPrompt(cfg config.Config, src work.Source, proj, change, kind, self string) string {
	return cfg.PromptFor(change, src.Prompt(proj, kind)) + " " + start.Sentence(self, change)
}

// lacking returns why src cannot do an action, "" when it names the action's command.
func lacking(src work.Source) func(work.Action) string {
	return func(a work.Action) string {
		if src.Command(a, "") != "" {
			return ""
		}
		name := "OpenSpec"
		if _, ok := src.(*speckit.Source); ok {
			name = "Spec Kit"
		}
		return name + " has no " + string(a)
	}
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
// A Start from the Board is then watched.
func launch(change, pane, tab, kind, prompt string, run herdr.Runner) error {
	if err := start.Launch(change, kind, prompt, pane, tab, run); err != nil {
		run("notification", "show", "spec-kanban: start failed", "--body", err.Error())
		return err
	}
	watch(change, false, run)
	return nil
}

// watch records the Main pane of change for Escalation, clears its attention marker, and watches
// the agent of change until it hands off or goes away; only for the Board ($KANBAN_PROJECT set).
func watch(change string, resumed bool, run herdr.Runner) {
	proj := os.Getenv("KANBAN_PROJECT")
	if proj == "" {
		return
	}
	cfg, _ := config.Load(config.Dir(run)) // a broken config was already shown by the Board
	detect, sourceFor := sources(cfg)
	src := sourceFor(proj)
	dir, git, loadGit := cache.Dir(), gitCLI(proj), gitCLI(proj)
	if !cfg.Worktree { // the agent works in the Project: a leftover checkout of the Change must not supply its Progress
		loadGit = func(...string) ([]byte, error) { return nil, errors.New("worktree off") }
	}
	mainPane, user := os.Getenv("KANBAN_MAIN_PANE"), cache.UserDir() // where kanban escalate finds them
	_ = cache.WriteMain(user, proj, change, mainPane)                // a lost record only leaves Escalation to the marker
	cache.ClearMark(user, proj, change)
	start.Watch(start.Watched{Project: proj, Change: change, Main: mainPane, Base: cfg.Base, On: cfg.Worktree, Resumed: resumed,
		Load:      loader(dir, proj, src, loadGit),
		Source:    src,
		Record:    func(s herdr.Session) { _ = cache.WriteSession(dir, proj, change, s) }, // a lost record only stops resume from the Board
		Escalated: func() bool { return cache.ReadMarks(user, proj)[change] == "escalation" },
		Mark:      marker(proj, change, run, detect, sourceFor, gitCLI),
		Clear:     func() { cache.ClearMark(user, proj, change) },
	}, run, git)
}

// runCommand runs start.Launch with command and reports a failure as a herdr notification; nothing
// watches the agent.
func runCommand(change, pane, tab, kind, command string, run herdr.Runner) error {
	err := start.Launch(change, kind, command, pane, tab, run)
	if err != nil {
		run("notification", "show", "spec-kanban: run failed", "--body", err.Error())
	}
	return err
}

// resumeChange reopens the session recorded for change of $KANBAN_PROJECT, reports a failure as a
// herdr notification, and watches a resume from the Board.
func resumeChange(change, pane, tab string, run herdr.Runner) error {
	s := cache.ReadSession(cache.Dir(), os.Getenv("KANBAN_PROJECT"), change)
	if err := start.Resume(change, s, pane, tab, run); err != nil {
		run("notification", "show", "spec-kanban: resume failed", "--body", err.Error())
		return err
	}
	watch(change, true, run)
	return nil
}

// escalate hands change to its Main agent because it cannot go on without a human, for reason (its
// words joined by spaces), and tells stdout who received it. cwd is a checkout the Change's agent
// works in; the Project is the first of it and its git repository's checkouts that holds a Main
// record for change, else cwd's checkout.
func escalate(change string, reason []string, cwd string, stdout io.Writer, run herdr.Runner,
	detect func(string) bool, sourceFor func(string) work.Source, git func(string) herdr.Runner) error {
	checkout := project.Find(cwd, detect)
	if checkout == "" {
		return fmt.Errorf("no OpenSpec or Spec Kit project at or above %s", cwd)
	}
	dir, user, proj, mainPane := cache.Dir(), cache.UserDir(), checkout, ""
	for _, c := range append([]worktree.Checkout{{Path: checkout}}, worktree.List(git(checkout))...) {
		p, err := filepath.EvalSymlinks(c.Path) // the Board's Project has its symlinks resolved
		if err != nil {
			continue
		}
		if pane, ok := cache.ReadMain(user, p, change); ok {
			proj, mainPane = p, pane
			break
		}
	}
	src := sourceFor(proj)
	open, err := src.Open(proj)
	if err != nil {
		return err
	}
	if !slices.Contains(open, change) {
		return fmt.Errorf("%q is not an open change of %s", change, filepath.Base(proj))
	}
	progress := ""
	changes, _ := loader(dir, proj, src, git(proj))() // without Progress the prompt only leaves it out
	for _, c := range changes {
		if c.Name == change && c.Phase != work.Archived {
			progress = fmt.Sprintf(" (%d/%d tasks)", c.Done, c.Total)
		}
	}
	where := ""
	if filepath.Clean(checkout) != filepath.Clean(proj) {
		where = ", worktree " + checkout + " on branch " + change
	}
	agent := herdr.AgentName(change)
	text := fmt.Sprintf("%s needs a human%s: %s. Its agent %s waits in tab %s%s."+
		" 1. Read its open tasks in %s. 2. If you may do the step yourself, do it and tick its task; otherwise tell the user exactly what to do and wait for them."+
		` 3. Then tell its agent to continue: herdr agent prompt %s "<what changed>". Do not close its tab; when all tasks are done, the normal Handoff follows.`,
		change, progress, strings.Join(reason, " "), agent, change, where, filepath.Join(checkout, src.Rel(change), "tasks.md"), agent)
	_ = cache.WriteMark(user, proj, change, "escalation") // a lost marker only lets a Stall follow
	if start.Tell(run, mainPane, os.Getenv("HERDR_PANE_ID"), text) {
		fmt.Fprintln(stdout, "Sent to the Main agent. End your turn.")
	} else {
		fmt.Fprintln(stdout, "No Main agent took it; the Board and the sidebar show ⚑. End your turn.")
	}
	_ = refresh(dir, run, detect, sourceFor, git) // the sidebar shows the marker without waiting for a herdr event
	return nil
}

// marker returns what sets the attention marker of change of proj for a reason and then runs one
// sidebar pass: a Stall happens without a herdr event that would refresh the sidebar.
func marker(proj, change string, run herdr.Runner, detect func(string) bool, sourceFor func(string) work.Source, git func(string) herdr.Runner) func(why string) {
	return func(why string) {
		_ = cache.WriteMark(cache.UserDir(), proj, change, why) // a lost marker only hides the ⚑
		_ = refresh(cache.Dir(), run, detect, sourceFor, git)
	}
}

// refresh runs one sidebar pass, as kanban sync does, caching loads in dir. Its lock lives in
// cache.UserDir, which kanban sync, the watcher and kanban escalate share: a worker's pane has no
// $HERDR_PLUGIN_STATE_DIR, and an older pass must not report after a newer one. Without a user
// cache dir it locks in dir.
func refresh(dir string, run herdr.Runner, detect func(string) bool, sourceFor func(string) work.Source, git func(string) herdr.Runner) error {
	return sidebar.Coalesce(cmp.Or(cache.UserDir(), dir), func() error {
		return sidebar.Sync(run, detect, syncLoad(dir, sourceFor, git), syncWaiting(cache.UserDir(), git))
	})
}

// syncWaiting returns the Changes of proj whose attention marker in user still holds for herdr's
// agents (nil when agent list failed) and proj's Items.
func syncWaiting(user string, git func(proj string) herdr.Runner) func(string, []start.Agent, []work.Item) map[string]bool {
	return func(proj string, agents []start.Agent, changes []work.Item) map[string]bool {
		marks := cache.ReadMarks(user, proj)
		var in map[string]string // nil: only Handoff markers hold
		for _, why := range marks {
			if why != "handoff" && agents != nil { // InProject runs git: only when a marker needs it
				in = start.InProject(agents, proj, git(proj))
				break
			}
		}
		return start.Waiting(marks, in, changes)
	}
}

// loader reads the Project's Changes and those worked on in its other checkouts, reusing cached
// planning state, and caches every successful load.
func loader(dir, proj string, src work.Source, git herdr.Runner) func() ([]work.Item, error) {
	return func() ([]work.Item, error) {
		known := map[string]work.Item{}
		for _, c := range cache.Read(dir, proj) { // re-read each time, so every refresh reuses the newest load
			if c.Phase != work.Archived { // an archived Change may share an open one's name
				known[c.Name] = c
			}
		}
		changes, err := src.Load(proj, worktree.List(git), known)
		if err == nil {
			_ = cache.Write(dir, proj, changes) // cache problems never break the Board
		}
		return changes, err
	}
}

// syncLoad loads a Project for the sidebar through the Board's loader and its Source, so every sync
// warms the Board's cache.
func syncLoad(dir string, sourceFor func(proj string) work.Source, git func(proj string) herdr.Runner) func(string) ([]work.Item, error) {
	return func(proj string) ([]work.Item, error) { return loader(dir, proj, sourceFor(proj), git(proj))() }
}

// sources returns detect, which accepts a dir any Source recognises, and sourceFor, which picks
// proj's Source: Spec Kit when it recognises proj and cfg prefers it or OpenSpec does not, else OpenSpec.
func sources(cfg config.Config) (detect func(dir string) bool, sourceFor func(proj string) work.Source) {
	op, sk := openspec.New(openspecCLI), speckit.New(gitCLI, cfg.Base)
	detect = func(dir string) bool { return op.Detect(dir) || sk.Detect(dir) }
	sourceFor = func(proj string) work.Source {
		if sk.Detect(proj) && (cfg.Source == "speckit" || !op.Detect(proj)) {
			return sk
		}
		return op
	}
	return detect, sourceFor
}

// openspecCLI runs the openspec CLI inside proj.
func openspecCLI(proj string) herdr.Runner {
	return func(args ...string) ([]byte, error) { return herdr.Run("openspec", proj, args...) }
}

// gitCLI runs git inside proj.
func gitCLI(proj string) herdr.Runner {
	return func(args ...string) ([]byte, error) { return herdr.Run("git", proj, args...) }
}
