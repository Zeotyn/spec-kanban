## Why

`s` starts the agent in a tab of the Project's own checkout, so the agent and the user edit the same working tree: the user cannot keep working, committing or switching branches in parallel without getting in the agent's way. herdr 0.9.5 creates and opens git worktrees as workspaces natively (`herdr worktree create/open`), so each Started Change can get its own checkout on a branch named after it.

## What Changes

- Start opens the agent in a git worktree of the Project's repository on a branch named after the Change: herdr creates the checkout at its configured location (`<worktrees.directory>/<repo>/<change>`) from the base and opens it as a new workspace labelled with the Change; the agent tab is its root pane.
- A checkout that already has the Change's branch (a rerun, resumed work, or the user's own checkout on that branch) is reused instead of failing: a closed linked worktree is opened as a workspace again, an open one gets a new tab.
- The base of a new worktree is detected: the default branch `origin/HEAD` points at (its local branch when one exists), else the first existing of `main`, `master`, `develop`, else the current branch. Start fails before creating anything when the Change does not exist on that ref.
- New config keys: `worktree` (bool, default `true`; `false` keeps today's tab in the Project) and `base` (overrides the detected base).
- The change source reads a Change that is worked on in another checkout of the Project's repository (branch named like the Change, Change directory present there) from that checkout, so the Board, the detail view, `e` and the sidebar show the agent's Progress before the branch is merged; a Change that exists only in that checkout is shown too.
- The Running agent badge and Start's "other Project" check also accept an agent whose working directory lies in another checkout of the Project's repository.
- Never deletes a worktree or branch; cleanup belongs to m6-agent-handoff.
- Without a git repository Start falls back to a tab in the Project.

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `agent-start`: Start opens its tab in the Change's worktree (create, reuse, base detection, Change-on-base check) and accepts agents working in a worktree of the Project as its own.
- `kanban-config`: new settings `worktree` and `base`, and their validation.
- `change-source`: Changes worked on in another checkout of the Project's repository are read from that checkout.
- `board`: the Running agent badge counts agents in worktrees; the Start scenario describes the worktree workspace.

## Impact

- New `internal/worktree`: list the repository's checkouts (`git worktree list --porcelain`, `git rev-parse --show-prefix`) and detect the base; all git calls through a `herdr.Runner`-shaped fake in tests.
- `internal/start`: `Prepare` chooses the checkout and calls `herdr worktree create`/`worktree open`/`tab create`; `Running` and the other-Project check take the checkout directories. `Launch` and the `kanban launch` arguments stay as they are.
- `internal/openspec`: `Change` gains `Dir` (the checkout's Project directory, empty for the Project itself); `Load` overlays Changes from other checkouts; planning state and file fingerprint read there.
- `internal/docs`, `internal/board`: detail view and `e` use the Change's `Dir`.
- `internal/config`: `worktree` and `base`.
- `cmd/kanban/main.go`: wire git runner, config and checkouts into Board, `kanban start`, loader and sync.
- Docs: README (worktree Start, settings), CONTEXT.md (Start, new term Change checkout), ADR 0007 superseding the tab location of ADR 0005.
- m6-agent-handoff, planned after this change, must adapt its Handoff to worktrees (see design.md).
