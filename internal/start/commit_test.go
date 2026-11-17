package start

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/zeotyn/spec-kanban/internal/herdr"
)

func TestCommitPlan(t *testing.T) {
	const rel = "openspec/changes/" + change
	var (
		branch = "branch --show-current"
		status = "status --porcelain -- " + rel
		add    = "add -- " + rel
		commit = "commit -m docs: plan " + change + " -- " + rel
	)
	tests := []struct {
		name    string
		base    string
		out     map[string]string
		want    []string
		wantErr string
	}{
		{"new Change", "main", map[string]string{branch: "main\n", status: "?? " + rel + "/\n", add: "", commit: ""},
			[]string{branch, status, add, commit}, ""},
		{"edited plan", "main", map[string]string{branch: "main\n", status: " M " + rel + "/tasks.md\n", add: "", commit: ""},
			[]string{branch, status, add, commit}, ""},
		{"other branch", "main", map[string]string{branch: "feature-x\n"}, []string{branch}, ""},
		{"remote base", "origin/main", map[string]string{branch: "main\n"}, []string{branch}, ""},
		{"detached HEAD", "main", map[string]string{branch: "\n"}, []string{branch}, ""},
		{"nothing to commit", "main", map[string]string{branch: "main\n", status: ""}, []string{branch, status}, ""},
		{"commit fails", "main", map[string]string{branch: "main\n", status: "?? " + rel + "/\n", add: ""},
			[]string{branch, status, add, commit, "reset -q -- " + rel}, "exit status 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &fakeGit{out: tt.out}
			err := commitPlan(g.run, tt.base, rel, change)
			if got := fmt.Sprint(err); err != nil && got != tt.wantErr || err == nil && tt.wantErr != "" {
				t.Errorf("err %q, want %q", got, tt.wantErr)
			}
			if !slices.Equal(g.calls, tt.want) {
				t.Errorf("calls %q, want %q", g.calls, tt.want)
			}
		})
	}
}

// TestCommitPlanGit commits a new Change in a real repository while another file is staged.
func TestCommitPlanGit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("no git")
	}
	dir := t.TempDir()
	git := func(args ...string) ([]byte, error) { return herdr.Run("git", dir, args...) }
	must := func(args ...string) string {
		t.Helper()
		out, err := git(args...)
		if err != nil {
			t.Fatal(err)
		}
		return strings.TrimSpace(string(out))
	}
	write := func(name, text string) {
		t.Helper()
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	must("init", "-q", "-b", "main")
	must("config", "user.name", "t")
	must("config", "user.email", "t@example.com")
	must("config", "commit.gpgsign", "false")
	write("README.md", "one\n")
	must("add", "README.md")
	must("commit", "-q", "-m", "init")
	write("README.md", "two\n")
	must("add", "README.md")
	const rel = "openspec/changes/" + change
	write(rel+"/proposal.md", "# plan\n")

	if err := commitPlan(git, "main", rel, change); err != nil {
		t.Fatal(err)
	}
	if got := must("log", "-1", "--format=%s"); got != "docs: plan "+change {
		t.Errorf("last commit %q", got)
	}
	if got := must("show", "--name-only", "--format=", "HEAD"); got != rel+"/proposal.md" {
		t.Errorf("committed %q", got)
	}
	if got := must("diff", "--cached", "--name-only"); got != "README.md" {
		t.Errorf("staged %q, want README.md", got)
	}

	write(rel+"/tasks.md", "- [ ] 1.1\n") // a hook refuses: git's error, nothing of the Change left staged
	write(".git/hooks/pre-commit", "#!/bin/sh\nexit 1\n")
	if err := os.Chmod(filepath.Join(dir, ".git/hooks/pre-commit"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := commitPlan(git, "main", rel, change); err == nil || !strings.Contains(err.Error(), "git commit") {
		t.Errorf("got %v, want git's error", err)
	}
	if got := must("diff", "--cached", "--name-only"); got != "README.md" {
		t.Errorf("staged %q after a refused commit, want README.md", got)
	}
}
