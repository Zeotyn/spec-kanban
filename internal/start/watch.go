package start

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/zeotyn/openspec-kanban/internal/herdr"
	"github.com/zeotyn/openspec-kanban/internal/source/openspec"
	"github.com/zeotyn/openspec-kanban/internal/work"
	"github.com/zeotyn/openspec-kanban/internal/worktree"
)

// Watched is a Change started from the Board whose agent Watch follows.
type Watched struct {
	Project, Change string
	Main            string // pane the Board was opened from, "" when unknown
	Base            string // configured base, "" to detect it
	On              bool   // Start works in the Change's checkout (config worktree)
	Load            func() ([]work.Item, error)
	Record          func(herdr.Session)
}

// agent is what Watch reads from herdr agent get.
type agent struct {
	Agent   string         `json:"agent"`
	Status  string         `json:"agent_status"`
	Pane    string         `json:"pane_id"`
	Seq     int            `json:"state_change_seq"`
	Session *herdr.Session `json:"agent_session"`
}

func getAgent(run herdr.Runner, target string) (agent, error) {
	out, err := run("agent", "get", target)
	if err != nil {
		return agent{}, err
	}
	var got struct {
		Result struct {
			Agent agent `json:"agent"`
		} `json:"result"`
	}
	if err := json.Unmarshal(out, &got); err != nil {
		return agent{}, err
	}
	return got.Result.Agent, nil
}

// Watch follows the agent of w.Change until it goes away or hands off: at each settle it records
// the agent's session, then either hands the finished Change off to the Main agent or shows a
// notification and waits for the agent's next turn. git runs git inside w.Project.
func Watch(w Watched, run, git herdr.Runner) {
	if _, err := w.get(run); err != nil {
		return
	}
	run("agent", "wait", w.Change, "--until", "working", "--timeout", "60000") // it may have finished its turn already
	for {
		if _, err := run("agent", "wait", w.Change); err != nil {
			return // the agent exited, its pane was closed or herdr went away
		}
		a, err := w.get(run)
		if err != nil {
			return
		}
		if a.Status != "idle" && a.Status != "done" && a.Status != "blocked" {
			continue // its next turn began before it was read
		}
		text, finished, stopped := w.check(git, a)
		if finished {
			w.handoff(run, a.Pane, text)
			return
		}
		run("notification", "show", "openspec-kanban: "+w.Change+" stopped", "--body", stopped)
		if b, err := w.get(run); err != nil {
			return
		} else if b.Seq != a.Seq {
			continue // it moved on while being checked, maybe through a whole turn
		}
		if _, err := run("agent", "wait", w.Change, "--until", "working"); err != nil {
			return
		}
	}
}

// get reads the watched agent and records its session.
func (w Watched) get(run herdr.Runner) (agent, error) {
	a, err := getAgent(run, w.Change)
	if err == nil && a.Session != nil {
		w.Record(*a.Session)
	}
	return a, err
}

// check tells whether the Change of the settled agent a is finished, and returns the Handoff text
// for it and the notification body for when the agent stopped without finishing it.
func (w Watched) check(git herdr.Runner, a agent) (text string, finished bool, stopped string) {
	changes, err := w.Load()
	switch {
	case err != nil && a.Status == "blocked":
		return "", false, w.Change + " waits for input"
	case err != nil:
		return "", false, err.Error()
	}
	checkout := ""
	if w.On {
		for _, c := range worktree.List(git) {
			if c.Branch == w.Change && filepath.Clean(c.Path) != filepath.Clean(w.Project) {
				checkout = c.Path
			}
		}
	}
	var open *work.Item
	inProject := false // an archived Change may share an open one's name
	for i, c := range changes {
		if c.Name == w.Change && c.Phase == work.Archived {
			inProject = true
		} else if c.Name == w.Change {
			open = &changes[i]
		}
	}
	// A Change archived in its checkout is still open in the Project until merged.
	archived := checkout != "" && openspec.ArchivedIn(checkout, w.Change) || open == nil && inProject
	progress := ""
	if open != nil {
		progress = fmt.Sprintf("%d/%d tasks", open.Done, open.Total)
	}
	switch {
	case a.Status == "blocked" && open != nil:
		return "", false, w.Change + " waits for input at " + progress
	case a.Status == "blocked":
		return "", false, w.Change + " waits for input"
	case !archived && open == nil:
		return "", false, w.Change + " is neither open nor archived"
	case !archived && (open.Total == 0 || open.Done < open.Total):
		return "", false, w.Change + " stopped at " + progress
	}
	return w.text(git, checkout, progress, archived, a.Session), true, ""
}

// text is the Handoff text: review, merge and remove the checkout when there is one, archive unless
// archived, and how to reopen the session.
func (w Watched) text(git herdr.Runner, checkout, progress string, archived bool, s *herdr.Session) string {
	c, dir := w.Change, w.Project
	var t string
	switch {
	case checkout == "" && archived:
		t = c + " is finished and archived. Please review it."
	case checkout == "":
		t = fmt.Sprintf("%s is finished (%s). Please review it and archive it with /opsx:archive %s.", c, progress, c)
	default:
		dir = checkout
		base := worktree.Base(w.Base, git)
		if archived {
			t = fmt.Sprintf("%s is finished and archived on branch %s in %s. Please review it, merge branch %s into %s, then", c, c, checkout, c, base)
		} else {
			t = fmt.Sprintf("%s is finished (%s) on branch %s in %s. Please review it, merge branch %s into %s, archive it with /opsx:archive %s, then",
				c, progress, c, checkout, c, base, c)
		}
		t += " remove its worktree with: git worktree remove " + checkout + "." +
			" Never use --force and keep the branch; if the merge or the removal does not go through cleanly, stop and tell the user."
	}
	if s != nil {
		if cmd := ResumeCommand(*s); cmd != "" {
			t += " To look into its session press s on its Card in the Board, or run in " + dir + ": " + cmd
		}
	}
	return t
}

// handoff closes the agent's pane, which closes a tab or worktree workspace with nothing else open,
// then hands text to the Main agent, or shows it as a notification when the Main agent cannot take it.
func (w Watched) handoff(run herdr.Runner, pane, text string) {
	run("pane", "close", pane)
	if w.Main != "" && w.Main != pane {
		if m, err := getAgent(run, w.Main); err == nil && m.Agent != "" && m.Status != "blocked" {
			if _, err := run("agent", "prompt", w.Main, text); err == nil {
				return
			}
		}
	}
	run("notification", "show", "openspec-kanban: "+w.Change+" finished", "--body", text)
}

// resumes holds per agent kind the command that reopens a session, %s standing for its id; only
// kinds whose resume argument was verified are listed.
var resumes = map[string]string{
	"claude":   "claude --resume %s",
	"copilot":  "copilot --resume=%s",
	"codex":    "codex resume %s",
	"opencode": "opencode --session %s",
	"cursor":   "cursor-agent --resume %s",
	"gemini":   "gemini --resume %s",
}

// ResumeArgs returns the args after the agent's executable that reopen s, nil when s cannot be reopened.
func ResumeArgs(s herdr.Session) []string {
	cmd, ok := resumes[s.Agent]
	if !ok || s.Kind != "id" || s.Value == "" {
		return nil
	}
	args := strings.Fields(cmd)[1:]
	for i := range args {
		args[i] = strings.ReplaceAll(args[i], "%s", s.Value)
	}
	return args
}

// ResumeCommand returns the shell command that reopens s, "" when s cannot be reopened.
func ResumeCommand(s herdr.Session) string {
	args := ResumeArgs(s)
	if args == nil {
		return ""
	}
	return strings.Fields(resumes[s.Agent])[0] + " " + strings.Join(args, " ")
}

// Resume focuses tab and starts the agent kind that recorded s in pane as the agent of change,
// reopening s; no prompt is sent.
func Resume(change string, s herdr.Session, pane, tab string, run herdr.Runner) error {
	args := ResumeArgs(s)
	switch {
	case s.Agent == "":
		return fmt.Errorf("no reopenable session for %s", change)
	case args == nil:
		return fmt.Errorf("resume not supported for %s", s.Agent)
	}
	if _, err := run("tab", "focus", tab); err != nil {
		return fmt.Errorf("tab focus: %w", err)
	}
	if _, err := run(append([]string{"agent", "start", change, "--kind", s.Agent, "--pane", pane, "--timeout", "60000", "--"}, args...)...); err != nil {
		return fmt.Errorf("%s did not become ready in tab %q: %w", s.Agent, change, err)
	}
	return nil
}
