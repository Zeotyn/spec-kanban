package start

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/zeotyn/spec-kanban/internal/work"
)

// CoordinatorPrompt is what the coordinator of wave n receives: the wave, its worktree path and base,
// each member with its proposal in the worktree (src's Rel), the kanban wave commands with the plugin
// binary self, and the coordinator's rules.
func CoordinatorPrompt(self string, n int, path, base string, members []string, src work.Source) string {
	self, branch := quote(self), fmt.Sprintf("wave-%d", n)
	list := make([]string, len(members))
	for i, m := range members {
		list[i] = fmt.Sprintf("%s (%s)", m, filepath.Join(path, src.Rel(m), "proposal.md"))
	}
	cmd := fmt.Sprintf("%s wave %%s %d", self, n)
	return fmt.Sprintf("You coordinate wave %d: branch %s in worktree %s, from base %s. Its members are %s."+
		" Commands: %s <change> starts a member, which hands off to you when done; %s <change> drops one; %s \"<what you need>\" asks the Main agent; %s finishes the wave."+
		" Rules: 1. Read the proposals and decide the order. Start a member only after every member it builds on is merged into %s; start independent members right away."+
		" 2. When a member hands off, review it, merge its branch into %s, resolve every conflict yourself, run the Project's tests, archive the Change on %s with /opsx:archive <change>, and remove the member's worktree, never with --force, and keep its branch."+
		" 3. Drop a member you cannot make work, and say why in your final report."+
		" 4. Answer a member's escalation yourself when you can; escalate to the Main agent only when you cannot decide."+
		" 5. Finish the wave when every member is archived on %s or dropped.",
		n, branch, path, base, strings.Join(list, ", "),
		fmt.Sprintf(cmd, "start"), fmt.Sprintf(cmd, "drop"), fmt.Sprintf(cmd, "escalate"), fmt.Sprintf(cmd, "finish"),
		branch, branch, branch, branch)
}
