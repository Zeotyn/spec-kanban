package start

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/zeotyn/spec-kanban/internal/herdr"
	"github.com/zeotyn/spec-kanban/internal/work"
	"github.com/zeotyn/spec-kanban/internal/worktree"
)

// Watched is a Change started from the Board whose agent Watch follows.
type Watched struct {
	Project, Change string
	Main            string // pane the Board was opened from, "" when unknown
	Base            string // configured base, "" to detect it
	On              bool   // Start works in the Change's checkout (config worktree)
	Load            func() ([]work.Item, error)
	Source          work.Source // for ArchivedIn and Finish
	Record          func(herdr.Session)
	Escalated       func() bool      // whether the Change escalated since its agent last worked
	Mark            func(why string) // sets the Change's attention marker: "stall" or "handoff"
	Clear           func()           // clears the Change's attention marker
	Resumed         bool             // a reopened session: evaluated only after it worked once
}

// stallAfter is how long an unfinished agent may stay settled before the Main agent hears of it.
const stallAfter = 10 * time.Minute

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
// the agent's session, then either hands the finished Change off to the Main agent or waits for the
// agent's next turn, telling the Main agent of a Stall when an unfinished agent that did not
// escalate stays settled for stallAfter. A Stall and a Handoff the Main agent did not take set the
// attention marker; working again after a marked settle and a taken Handoff clear it. git runs git
// inside w.Project.
func Watch(w Watched, run, git herdr.Runner) {
	if _, err := w.get(run); err != nil {
		return
	}
	name := herdr.AgentName(w.Change)
	if w.Resumed { // a reopened session has nothing to evaluate before its next turn
		if _, err := run("agent", "wait", name, "--until", "working"); err != nil {
			return
		}
	} else {
		run("agent", "wait", name, "--until", "working", "--timeout", "60000") // it may have finished its turn already
	}
	for {
		if _, err := run("agent", "wait", name); err != nil {
			return // the agent exited, its pane was closed or herdr went away
		}
		a, err := w.get(run)
		if err != nil {
			return
		}
		if a.Status != "idle" && a.Status != "done" && a.Status != "blocked" {
			continue // its next turn began before it was read
		}
		text, finished, progress := w.check(git, a)
		if finished {
			w.handoff(run, a.Pane, text)
			return
		}
		if b, err := w.get(run); err != nil {
			return
		} else if b.Seq != a.Seq {
			continue // it moved on while being checked, maybe through a whole turn
		}
		escalated := w.Escalated() // only this marker is cleared: one written later belongs to the next settle
		stalled := false
		if a.Status != "blocked" && !escalated {
			if _, err := run("agent", "wait", name, "--until", "working", "--timeout", strconv.FormatInt(stallAfter.Milliseconds(), 10)); err == nil {
				continue
			}
			if b, err := w.get(run); err != nil {
				return // gone, not timed out
			} else if b.Seq != a.Seq {
				continue // the wait failed for another reason while it moved on
			}
			Tell(run, w.Main, a.Pane, w.stall(progress))
			w.Mark("stall")
			stalled = true
		}
		if _, err := run("agent", "wait", name, "--until", "working"); err != nil {
			return
		}
		if escalated || stalled {
			w.Clear()
		}
	}
}

// stall is the Stall prompt for the Change at progress, "" when unknown.
func (w Watched) stall(progress string) string {
	at, name := "", herdr.AgentName(w.Change)
	if progress != "" {
		at = " at " + progress
	}
	return w.Change + " stopped" + at + " 10 minutes ago without asking for help. Read its output with: herdr agent read " + name + " --lines 40." +
		" If it waits for something that finishes by itself, do nothing. If it needs a human, tell the user what it needs." +
		" To let it continue, run: herdr agent prompt " + name + ` "<what to do>". Do not close its tab.`
}

// get reads the watched agent and records its session.
func (w Watched) get(run herdr.Runner) (agent, error) {
	a, err := getAgent(run, herdr.AgentName(w.Change))
	if err == nil && a.Session != nil {
		w.Record(*a.Session)
	}
	return a, err
}

// check tells whether the Change of the settled agent a is finished, and returns the Handoff text
// for it and its Progress ("3/7 tasks", "" when unknown). A blocked agent is never finished.
func (w Watched) check(git herdr.Runner, a agent) (text string, finished bool, progress string) {
	if a.Status == "blocked" {
		return "", false, ""
	}
	changes, err := w.Load()
	if err != nil {
		return "", false, ""
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
	archived := checkout != "" && w.Source.ArchivedIn(checkout, w.Change) || open == nil && inProject
	if open != nil {
		progress = fmt.Sprintf("%d/%d tasks", open.Done, open.Total)
	}
	if !archived && (open == nil || open.Total == 0 || open.Done < open.Total) {
		return "", false, progress
	}
	return w.text(git, checkout, progress, archived, a.Session), true, progress
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
		t = fmt.Sprintf("%s is finished (%s). Please review it", c, progress)
		if f := w.Source.Finish(c, ""); f != "" {
			t += " and " + f
		}
		t += "."
	default:
		dir = checkout
		base := worktree.Base(w.Base, git)
		if archived {
			t = fmt.Sprintf("%s is finished and archived on branch %s in %s. Please review it, merge branch %s into %s, then", c, c, checkout, c, base)
		} else {
			t = fmt.Sprintf("%s is finished (%s) on branch %s in %s. Please review it, merge branch %s into %s,", c, progress, c, checkout, c, base)
			if f := w.Source.Finish(c, checkout); f != "" {
				t += " " + f + ","
			}
			t += " then"
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
// then hands text to the Main agent and clears the attention marker, or sets it for a Handoff when
// the Main agent cannot take it.
func (w Watched) handoff(run herdr.Runner, pane, text string) {
	run("pane", "close", pane)
	if Tell(run, w.Main, pane, text) {
		w.Clear()
	} else {
		w.Mark("handoff")
	}
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

// Resume focuses tab, waits for the shell prompt of pane and starts the agent kind that recorded s
// in it as the agent of change, reopening s; no prompt is sent.
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
	if err := ready(change, pane, run); err != nil {
		return err
	}
	if _, err := run(append([]string{"agent", "start", herdr.AgentName(change), "--kind", s.Agent, "--pane", pane, "--timeout", "60000", "--"}, args...)...); err != nil {
		return fmt.Errorf("%s did not become ready in tab %q: %w%s", s.Agent, change, err, tail(pane, run))
	}
	return nil
}
