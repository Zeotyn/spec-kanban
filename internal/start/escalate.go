package start

import (
	"regexp"
	"strings"

	"github.com/zeotyn/spec-kanban/internal/herdr"
)

// Tell submits text to the agent in pane main and reports whether it took it. It shows text as a
// herdr notification titled title instead when main is empty or own (the Change's own pane), runs
// no agent, the agent is blocked, or herdr rejects the prompt.
func Tell(run herdr.Runner, main, own, title, text string) bool {
	if main != "" && main != own {
		if m, err := getAgent(run, main); err == nil && m.Agent != "" && m.Status != "blocked" {
			if _, err := run("agent", "prompt", main, text); err == nil {
				return true
			}
		}
	}
	run("notification", "show", title, "--body", text)
	return false
}

var plain = regexp.MustCompile(`^[A-Za-z0-9/._-]+$`)

// Sentence is what a Start from the Board appends to its prompt: how the agent of change escalates
// with the plugin binary self, single-quoted for the shell unless it is plain.
func Sentence(self, change string) string {
	if !plain.MatchString(self) {
		self = "'" + strings.ReplaceAll(self, "'", `'\''`) + "'"
	}
	return "If you cannot continue without a human, run " + self + " escalate " + change + ` "<what you need>" and end your turn.`
}
