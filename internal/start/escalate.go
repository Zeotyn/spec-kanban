package start

import (
	"regexp"
	"strings"
	"time"

	"github.com/zeotyn/spec-kanban/internal/herdr"
)

// Waiting for the Main agent's input line: how often Tell checks and for how long.
const (
	tellEvery = 2 * time.Second
	tellFor   = 10 * time.Minute
)

// sleep and now are variables so tests do not wait.
var (
	sleep = time.Sleep
	now   = time.Now
)

// Tell submits text to the agent in pane main and reports whether it took it. It submits nothing
// when main is empty or own (the Change's own pane), runs no agent, or the agent is blocked; it
// reports false then and when herdr rejects the prompt. Before it submits, it waits until clear,
// checking every tellEvery, and reports false without submitting after tellFor.
func Tell(run herdr.Runner, main, own, text string) bool {
	kind, ok := tellable(run, main, own)
	if !ok {
		return false
	}
	for end := now().Add(tellFor); !clear(run, main, kind); sleep(tellEvery) {
		if !now().Before(end) {
			return false
		}
	}
	return submit(run, main, text)
}

// TellNow is Tell without waiting: when the agent could take text but is not clear, it submits
// nothing and reports waiting.
func TellNow(run herdr.Runner, main, own, text string) (taken, waiting bool) {
	kind, ok := tellable(run, main, own)
	if !ok {
		return false, false
	}
	if !clear(run, main, kind) {
		return false, true
	}
	return submit(run, main, text), false
}

// tellable returns the kind of the agent in pane main when Tell may submit to it.
func tellable(run herdr.Runner, main, own string) (kind string, ok bool) {
	if main == "" || main == own {
		return "", false
	}
	m, err := getAgent(run, main)
	return m.Agent, err == nil && m.Agent != "" && m.Status != "blocked"
}

func submit(run herdr.Runner, main, text string) bool {
	_, err := run("agent", "prompt", main, text)
	return err == nil
}

// clear reports whether a prompt to the agent of kind in pane may go out now, without sending keys
// to it: for a claude agent whose input box is on the screen, when the box is empty; otherwise when
// the pane is not focused. An agent herdr cannot read counts as clear, so its prompt fails at once.
func clear(run herdr.Runner, pane, kind string) bool {
	if kind == "claude" {
		if out, err := run("pane", "read", pane, "--source", "visible", "--format", "ansi"); err == nil {
			if text, found := inputBox(string(out)); found {
				return text == ""
			}
		}
	}
	a, err := getAgent(run, pane)
	return err != nil || !a.Focused
}

var plain = regexp.MustCompile(`^[A-Za-z0-9/._-]+$`)

// Sentence is what a Start from the Board appends to its prompt: how the agent of change escalates
// with the plugin binary self, single-quoted for the shell unless it is plain.
func Sentence(self, change string) string {
	return "If you cannot continue without a human, run " + quote(self) + " escalate " + change + ` "<what you need>" and end your turn.`
}

// quote single-quotes s for the shell unless it is plain.
func quote(s string) string {
	if plain.MatchString(s) {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
