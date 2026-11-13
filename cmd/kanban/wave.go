package main

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/zeotyn/spec-kanban/internal/board"
	"github.com/zeotyn/spec-kanban/internal/cache"
	"github.com/zeotyn/spec-kanban/internal/config"
	"github.com/zeotyn/spec-kanban/internal/herdr"
	"github.com/zeotyn/spec-kanban/internal/source/speckit"
	"github.com/zeotyn/spec-kanban/internal/start"
	"github.com/zeotyn/spec-kanban/internal/work"
)

// waveLaunch starts a wave of changes of $KANBAN_PROJECT for the Board, whose pane is
// $KANBAN_MAIN_PANE, and reports a failure as a herdr notification: nobody reads its stderr.
func waveLaunch(changes []string, run herdr.Runner, sourceFor func(string) work.Source, git func(string) herdr.Runner) error {
	err := launchWave(os.Getenv("KANBAN_PROJECT"), os.Getenv("KANBAN_MAIN_PANE"), changes, run, sourceFor, git)
	if err != nil {
		run("notification", "show", "spec-kanban: wave failed", "--body", err.Error())
	}
	return err
}

func launchWave(proj, mainPane string, changes []string, run herdr.Runner, sourceFor func(string) work.Source, git func(string) herdr.Runner) error {
	cfg, _ := config.Load(config.Dir(run)) // a broken config was already shown by the Board
	src, g, user := sourceFor(proj), git(proj), cache.UserDir()
	items, err := loader(cache.Dir(), proj, src, g)() // fresh: a Card may have left Ready since it was marked
	if err != nil {
		return err
	}
	self, err := os.Executable()
	if err != nil {
		return err
	}
	branches, _ := g("for-each-ref", "--format=%(refname:lstrip=2)", "refs/heads/wave-*") // none read: the records still count
	n := cache.NextWave(cache.ReadWaves(user, proj), strings.Fields(string(branches)))
	w := cache.Wave{Project: proj, Number: n, Branch: fmt.Sprintf("wave-%d", n), Main: mainPane, State: cache.Dir(), Members: changes}
	return start.Wave(proj, n, changes, items, cfg.Base, start.Kind(cfg.Agent, mainPane, run), self, run, src, g,
		func(pane, path, base string) error {
			w.Coordinator, w.Path, w.Base = pane, path, base
			return cache.WriteWave(user, w)
		},
		func() { cache.RemoveWave(user, w) })
}

// wave runs kanban wave <command> <N> …, which a wave's coordinator runs in cwd: start, drop,
// escalate or finish. It tells stdout what happened.
func wave(args []string, cwd string, stdout io.Writer, run herdr.Runner, sourceFor func(string) work.Source, git func(string) herdr.Runner) error {
	if len(args) < 2 {
		return fmt.Errorf("wave needs a command and a wave number")
	}
	n, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("wave number %q is not a number", args[1])
	}
	cmd, rest, user := args[0], args[2:], cache.UserDir()
	if !slices.Contains([]string{"start", "drop", "escalate", "finish"}, cmd) {
		return fmt.Errorf("unknown wave command %q", cmd)
	}
	w, ok := cache.FindWave(user, n, cwd)
	switch {
	case !ok:
		return fmt.Errorf("unknown wave %d", n)
	case w.Finished:
		return fmt.Errorf("wave %d is finished", n)
	}
	member := func() (string, error) {
		if len(rest) != 1 {
			return "", fmt.Errorf("wave %s needs a wave number and a change name", cmd)
		}
		if !slices.Contains(w.Members, rest[0]) {
			return "", fmt.Errorf("%s is not a member of wave %d", rest[0], n)
		}
		return rest[0], nil
	}
	switch cmd {
	case "start":
		change, err := member()
		if err != nil {
			return err
		}
		if slices.Contains(w.Dropped, change) {
			return fmt.Errorf("%s was dropped from wave %d", change, n)
		}
		return startMember(w, change, stdout, run, sourceFor(w.Project), git(w.Project))
	case "drop":
		change, err := member()
		if err != nil {
			return err
		}
		if !slices.Contains(w.Dropped, change) {
			w.Dropped = append(w.Dropped, change)
			if err := cache.WriteWave(user, w); err != nil {
				return err
			}
		}
		fmt.Fprintln(stdout, "dropped", change)
		return nil
	case "escalate":
		reason := strings.TrimSpace(strings.Join(rest, " "))
		if reason == "" {
			return fmt.Errorf("wave escalate needs a wave number and a reason")
		}
		coord := herdr.AgentName(w.Branch)
		text := fmt.Sprintf("Wave %d needs a human: %s. Its coordinator %s waits in tab %s, worktree %s on branch %s."+
			` Answer it with herdr agent prompt %s "<answer>", or tell the user exactly what to do.`, n, reason, coord, w.Branch, w.Path, w.Branch, coord)
		tell(run, w.Main, text, fmt.Sprintf("spec-kanban: wave %d needs you", n), " End your turn.", stdout)
		return nil
	}
	src := sourceFor(w.Path) // finish
	var pending, archived []string
	for _, m := range w.Members {
		switch {
		case slices.Contains(w.Dropped, m):
		case src.ArchivedIn(w.Path, m):
			archived = append(archived, m)
		default:
			pending = append(pending, m)
		}
	}
	if len(pending) > 0 {
		return fmt.Errorf("wave %d is not done: %s is neither archived on %s nor dropped", n, strings.Join(pending, ", "), w.Branch)
	}
	w.Finished = true
	if err := cache.WriteWave(user, w); err != nil {
		return err
	}
	text := fmt.Sprintf("Wave %d is finished: archived on %s: %s; dropped: %s. 1. Review branch %s in its worktree %s. 2. Merge it into %s."+
		" 3. Remove the worktree with git worktree remove %s, never with --force, and keep the branch."+
		" 4. If the merge or the removal does not go through cleanly, stop and tell the user.",
		n, w.Branch, list(archived), list(w.Dropped), w.Branch, w.Path, w.Base, w.Path)
	tell(run, w.Main, text, fmt.Sprintf("spec-kanban: wave %d finished", n), "", stdout)
	return nil
}

// list joins names with commas, "none" when there are none.
func list(names []string) string {
	if len(names) == 0 {
		return "none"
	}
	return strings.Join(names, ", ")
}

// tell submits text to the Main agent in pane main, else shows it as a herdr notification titled
// title, and tells stdout which, followed by after.
func tell(run herdr.Runner, main, text, title, after string, stdout io.Writer) {
	if start.Tell(run, main, os.Getenv("HERDR_PANE_ID"), text) {
		fmt.Fprintln(stdout, "Sent to the Main agent."+after)
		return
	}
	run("notification", "show", title, "--body", text)
	fmt.Fprintln(stdout, "No Main agent took it; herdr shows a notification."+after)
}

// startMember starts change as a member of wave w: as a Start from the Board, but in a worktree
// branched off the wave, without focus, with the coordinator as its Main pane and a prompt that
// forbids archiving. The agent is launched and watched detached, so the coordinator goes on.
func startMember(w cache.Wave, change string, stdout io.Writer, run herdr.Runner, src work.Source, git herdr.Runner) error {
	cfg, warn := config.Load(config.Dir(run))
	if warn != nil {
		fmt.Fprintln(os.Stderr, "kanban:", warn)
	}
	pane, _, err := start.PrepareUnfocused(w.Project, change, start.Worktree{On: true, Base: w.Branch}, run, src, git)
	if err != nil {
		return err
	}
	self, err := os.Executable()
	if err != nil {
		return err
	}
	kind := start.Kind(cfg.Agent, w.Coordinator, run)
	prompt := boardPrompt(cfg, src, w.Project, change, kind, self) +
		fmt.Sprintf(" Do not archive %s: the coordinator of wave %d merges and archives it.", change, w.Number)
	if err := spawn("wave-member", w.Project, strconv.Itoa(w.Number), change, pane, kind, prompt); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "started", change)
	return nil
}

// waveMember launches the agent of member change of wave n of proj in pane without focus, reports a
// failure as a herdr notification, and watches it as the Board's watcher would, with the coordinator
// as its Main pane and the wave branch as its base.
func waveMember(proj string, n int, change, pane, kind, prompt string, run herdr.Runner) error {
	w, ok := cache.FindWave(cache.UserDir(), n, proj)
	err := fmt.Errorf("unknown wave %d", n)
	if ok {
		err = start.LaunchUnfocused(change, kind, prompt, pane, run)
	}
	if err != nil {
		run("notification", "show", "spec-kanban: start failed", "--body", err.Error())
		return err
	}
	if w.State != "" { // the coordinator's pane has no $HERDR_PLUGIN_STATE_DIR; the Board reads sessions there
		os.Setenv("HERDR_PLUGIN_STATE_DIR", w.State)
	}
	watch(w.Project, change, w.Coordinator, w.Branch, false, run)
	return nil
}

// waveMain runs the wave commands of main with args.
func waveMain(cmd string, args []string) error {
	cfg, _ := config.Load(config.Dir(herdr.Exec)) // only the Source preference matters here
	_, sourceFor := sources(cfg)
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	switch cmd {
	case "wave-launch":
		if len(args) < 2 {
			return usageError("wave-launch needs at least two change names")
		}
		return waveLaunch(args, herdr.Exec, sourceFor, gitCLI)
	case "wave-member":
		if len(args) != 6 {
			return usageError("wave-member needs a project, a wave number, a change name, a pane, an agent kind and a prompt")
		}
		n, err := strconv.Atoi(args[1])
		if err != nil {
			return usageError("wave-member needs a wave number")
		}
		return waveMember(args[0], n, args[2], args[3], args[4], args[5], herdr.Exec)
	}
	if len(args) < 2 || args[0] == "escalate" && strings.TrimSpace(strings.Join(args[2:], "")) == "" {
		return usageError("wave needs start|drop|escalate|finish, a wave number and a change name or reason")
	}
	return wave(args, cwd, os.Stdout, herdr.Exec, sourceFor, gitCLI)
}

// usageError prints msg and the usage to stderr and exits 2, as main does for other commands.
func usageError(msg string) error {
	fmt.Fprintf(os.Stderr, "kanban: %s\n\n%s", msg, usage)
	os.Exit(2)
	return nil
}

// boardWave is what the Board of proj needs for waves: the size, why none can start, the badges
// from the wave records, and the detached wave start.
func boardWave(cfg config.Config, src work.Source, proj string) board.Wave {
	blocked := ""
	if _, ok := src.(*speckit.Source); ok {
		blocked = "waves need OpenSpec"
	} else if !cfg.Worktree {
		blocked = "waves need worktree on"
	}
	return board.Wave{Size: cfg.Wave(), Blocked: blocked,
		Of:    func() map[string]int { return cache.Badges(cache.ReadWaves(cache.UserDir(), proj)) },
		Start: func(changes []string) error { return spawn(append([]string{"wave-launch"}, changes...)...) },
	}
}
