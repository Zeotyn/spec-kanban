package start

import (
	"strings"
	"testing"
)

func TestCoordinatorPrompt(t *testing.T) {
	got := CoordinatorPrompt("/opt/kanban/bin/kanban", 1, "/wt/wave-1", "main", []string{"f5-bar", "c4-foo"}, src)
	for _, want := range []string{
		"wave 1", "branch wave-1", "worktree /wt/wave-1", "base main",
		"f5-bar (/wt/wave-1/openspec/changes/f5-bar/proposal.md)", "c4-foo (/wt/wave-1/openspec/changes/c4-foo/proposal.md)",
		"/opt/kanban/bin/kanban wave start 1 <change>", "/opt/kanban/bin/kanban wave drop 1 <change>",
		`/opt/kanban/bin/kanban wave escalate 1 "<what you need>"`, "/opt/kanban/bin/kanban wave finish 1",
		"merged into wave-1", "right away", "resolve every conflict yourself", "run the Project's tests", "/opsx:archive <change>",
		"never with --force", "keep its branch", "Drop a member", "final report", "only when you cannot decide",
		"archived on wave-1 or dropped",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("prompt lacks %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "\n") {
		t.Errorf("prompt spans lines:\n%s", got)
	}
	if q := CoordinatorPrompt("/my apps/kanban", 2, "/wt", "main", []string{"a", "b"}, src); !strings.Contains(q, `'/my apps/kanban' wave finish 2`) {
		t.Errorf("binary with a space is not quoted:\n%s", q)
	}
}
