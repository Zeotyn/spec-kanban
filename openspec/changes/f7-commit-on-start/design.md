## Context

`start` resolves the ref a new worktree starts from: the Change's own branch when it exists, else `worktree.Base(cfg.Base, git)`. It then checks `git cat-file -e <ref>:<src.Rel(change)>` and fails with `<change> is not on <ref>: commit it there or set base`. `Wave` does the same check for every member against the base. `git` is a runner that runs git inside the Project. `config.Config` holds `Worktree` and `Base`; unknown keys and wrong types fall back to the defaults with a warning.

See proposal.md for why.

## Goals / Non-Goals

**Goals:**
- With `commit` on, a freshly planned or edited Change starts without a manual commit.
- Never commit anything but the Change's own files, and never on a branch other than the base.

**Non-Goals:**
- No push, no commit with `commit` off, no change for Changes with their own branch.

## Decisions

### D1: `commitPlan(git, base, rel, change)` before the base check
It runs only when `cfg.Commit` is true and the ref is the base (no Change branch):
1. `git branch --show-current` must equal `base`. A remote base like `origin/main` never matches, so nothing is committed.
2. `git status --porcelain -- <rel>` must list something; otherwise there is nothing to commit.
3. `git add -- <rel>`, then `git commit -m "docs: plan <change>" -- <rel>`. With paths, `git commit` commits only those paths and leaves other staged files staged.

A failing step returns git's error, and Start fails before anything is created. Hooks run, since there is no `--no-verify`. The existing `cat-file` check runs afterwards unchanged, so a commit that did not happen still gives the refusal.
Alternative: copy the uncommitted folder into the worktree. That lost because the merge back into the base then fails on the same untracked files.

### D2: The message is `docs: plan <change>` for every Source
It is the same for OpenSpec and Spec Kit, short and in Conventional Commits form. The Source's `Rel` already gives the directory, so no Source change is needed.

### D3: Wave commits each member, then creates the branch
`Wave` calls `commitPlan` for each member in order before its check loop. One commit per member keeps the history readable. When one commit fails, the wave fails before `wave-<N>` is created; earlier members stay committed, which is harmless, since they are the user's own plans.

## Risks / Trade-offs

- [The user did not expect a commit on `main`] → Off by default, and only the Change's own files are committed. The README states it.
- [A pre-commit hook is slow or interactive] → Start waits for it like any git call. An interactive hook fails, and Start reports git's error.
- [The Board shows the Change as unchanged afterwards] → The commit does not change files, only git state, so Progress and the Phase stay the same.
