package start

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/zeotyn/spec-kanban/internal/herdr"
	"github.com/zeotyn/spec-kanban/internal/work"
	"github.com/zeotyn/spec-kanban/internal/worktree"
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

// Wave starts wave n of changes, which must be Ready among project's items: it creates branch
// wave-<n> from base (detected when "") in a worktree workspace labelled like it, records the wave
// before the coordinator can act on it, launches the coordinator named like the branch with its
// prompt (self is the plugin binary), and focuses its tab. It starts no member. When a step fails it
// reports the step and forgets a recorded wave. git runs git inside project.
func Wave(project string, n int, changes []string, items []work.Item, base, kind, self string, run herdr.Runner, src work.Source, git herdr.Runner,
	record func(pane, path, base string) error, forget func()) error {
	for _, c := range changes {
		i := slices.IndexFunc(items, func(it work.Item) bool { return it.Name == c && it.Phase != work.Archived })
		switch {
		case i < 0:
			return fmt.Errorf("%q is not an open change of %s", c, filepath.Base(project))
		case items[i].Phase != work.Ready:
			return fmt.Errorf("%s is not in Ready", c)
		}
	}
	branch, ref := fmt.Sprintf("wave-%d", n), worktree.Base(base, git)
	for _, c := range changes { // a member must be on the wave branch to start there
		if _, err := git("cat-file", "-e", ref+":"+src.Rel(c)); err != nil {
			return fmt.Errorf("%s is not on %s: commit it there or set base", c, ref)
		}
	}
	pane, tab, path, err := createWorktree(project, branch, ref, run)
	if err != nil {
		return err
	}
	if err := record(pane, path, ref); err != nil {
		return fmt.Errorf("record wave %d: %w", n, err)
	}
	if err := LaunchUnfocused(branch, kind, CoordinatorPrompt(self, n, path, ref, changes, src), pane, run); err != nil {
		forget()
		return err
	}
	run("tab", "focus", tab) // the wave runs; a tab left unfocused is no failure
	return nil
}
