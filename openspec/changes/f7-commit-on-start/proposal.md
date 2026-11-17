## Why

Start creates a Change's worktree from the base branch, so a Change must be committed there first. A freshly planned Change is usually not: Start then refuses with `<change> is not on main: commit it there or set base`, and the user has to commit by hand and press `s` again. The same goes for planning edits made after the last commit, which the worktree silently lacks.

## What Changes

- **New config key `commit`** (boolean, default `false`). When `true`, Start commits the uncommitted files of the Change it starts on the base before it creates the worktree: only the files under that Change's directory, as `docs: plan <change>`, with the repository's hooks. A wave does this for each member.
- It commits only when the Project's checkout is on the base branch itself. On any other branch, or when the Change already has its own branch, nothing is committed and Start behaves as today.
- **Clearer error:** the refusal names the new key: `<change> is not on <ref>: commit it there, set base, or set "commit": true`.
- With the default `false`, nothing changes except the error text.

Non-goals: no push; no commit of files outside the Change's directory; no copying of uncommitted files into the worktree.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `kanban-config`: the new key `commit`.
- `agent-start`: a new requirement that Start commits the plan when configured, and the new error text.

## Impact

- **Code:** `internal/config` (key `commit`), `internal/start` (commit step before the base check in Start and in Wave, error text).
- **Docs:** README (Settings, worktree section), CHANGELOG.
- No new dependencies.
