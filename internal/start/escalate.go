package start

import (
	"regexp"
	"strings"

	"github.com/zeotyn/spec-kanban/internal/herdr"
)

// Tell submits text to the agent in pane main and reports whether it took it. It submits nothing
// when main is empty or own (the Change's own pane), runs no agent, or the agent is blocked; it
// reports false then and when herdr rejects the prompt.
func Tell(run herdr.Runner, main, own, text string) bool {
	if main != "" && main != own {
		if m, err := getAgent(run, main); err == nil && m.Agent != "" && m.Status != "blocked" {
			if _, err := run("agent", "prompt", main, text); err == nil {
				return true
			}
		}
	}
	return false
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
