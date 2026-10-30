---
status: accepted
---

# Start in a git worktree

Start runs the agent in the Change's checkout instead of a tab in the Project, so the agent and the user no longer edit the same working tree. herdr creates the checkout (`herdr worktree create --cwd <project> --branch <change> --base <base> --label <change> --no-focus`), by default at `<worktrees.directory>/<repo>/<change>` (herdr's `[worktrees] directory`, default `~/.herdr/worktrees`), and opens it as a workspace; its root pane is the agent's tab. Launching the agent stays as in ADR 0005. This supersedes the tab location of ADR 0005.

A checkout that already has the Change's branch is reused: a closed linked worktree is opened again (`herdr worktree open --path <path>`), and when it is already open, Start adds a new tab to its workspace, because the root pane may be in use. The Project's own checkout on that branch gets a tab as before.

The base of a new worktree is the configured `base`, else the default branch `origin/HEAD` points at (its local branch when one exists, otherwise the remote ref), else the first existing of `main`, `master`, `develop`, else the current branch. The local branch wins because the user plans and merges locally before pushing. Before creating anything Start checks that the Change exists on the start ref (the Change's own branch if it exists, else the base) and fails with a clear message when it does not, instead of leaving an agent that cannot find its Change.

Progress is read from the Change's checkout, so the Board and the sidebar show the agent's work before the branch is merged. Start never removes a worktree or branch; that belongs to m6-agent-handoff. Without a git repository, with the Project below the top of its checkout, or with `worktree: false`, Start opens a tab in the Project.

## Considered Options

- **`--path <project>-<change>` next to the Project**: predictable, but it clutters the parent directory and duplicates a herdr setting.
- **Plain `git worktree add` plus `tab create`**: no herdr worktree support needed, but herdr would not group the workspace and we would manage paths ourselves.
- **`origin/main` as the base without a local check**: simpler, but fails whenever local `main` is ahead of the remote.
