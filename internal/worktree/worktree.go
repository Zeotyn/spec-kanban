// Package worktree lists the git checkouts of a Project and picks the base of a new one.
package worktree

import (
	"strings"

	"github.com/zeotyn/spec-kanban/internal/herdr"
)

// Checkout is a git checkout of the repository: its top directory and branch.
type Checkout struct{ Path, Branch string }

// List returns the checkouts on a branch, the Project's own included, read by git (run inside the
// Project). It is empty when the Project is not the top of a checkout or git fails.
func List(git herdr.Runner) []Checkout {
	prefix, err := git("rev-parse", "--show-prefix")
	if err != nil || strings.TrimSpace(string(prefix)) != "" {
		return nil
	}
	out, err := git("worktree", "list", "--porcelain")
	if err != nil {
		return nil
	}
	var all []Checkout
	for _, block := range strings.Split(string(out), "\n\n") {
		var c Checkout
		skip := false
		for _, line := range strings.Split(block, "\n") {
			key, val, _ := strings.Cut(line, " ")
			switch key {
			case "worktree":
				c.Path = val
			case "branch":
				c.Branch = strings.TrimPrefix(val, "refs/heads/")
			case "bare", "detached", "prunable":
				skip = true
			}
		}
		if !skip && c.Path != "" && c.Branch != "" {
			all = append(all, c)
		}
	}
	return all
}

// Base names what a new worktree branches from: configured, else the default branch of origin
// (its local branch when there is one), else main, master or develop, else the current branch.
func Base(configured string, git herdr.Runner) string {
	if configured != "" {
		return configured
	}
	local := func(name string) bool {
		_, err := git("rev-parse", "--verify", "--quiet", "refs/heads/"+name)
		return err == nil
	}
	if out, err := git("symbolic-ref", "--short", "refs/remotes/origin/HEAD"); err == nil {
		remote := strings.TrimSpace(string(out))
		if name := strings.TrimPrefix(remote, "origin/"); remote != "" {
			if local(name) {
				return name
			}
			return remote
		}
	}
	for _, name := range []string{"main", "master", "develop"} {
		if local(name) {
			return name
		}
	}
	if out, err := git("branch", "--show-current"); err == nil {
		if name := strings.TrimSpace(string(out)); name != "" {
			return name
		}
	}
	return "HEAD"
}
