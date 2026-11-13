## Context

m12-escalation added these pieces:
- a Main pane record per Project and Change in the user cache directory (`cache.WriteMain`/`ReadMain`);
- `kanban escalate`, which reads that record;
- a watcher (`start.Watch` with `Watched.Main` and `Watched.Base`) that hands a finished Change to the Main pane or prompts it on a Stall.

Start places a Change in a worktree through `start.place`. It runs `herdr worktree create --branch <change> --base <ref>`, where `ref` comes from `worktree.Base(wt.Base, git)`, and refuses a ref that does not hold the Change. The Handoff text already asks to merge the Change's branch into `Watched.Base` and to archive it, then to remove the worktree. A Start from the Board runs detached (`spawn("launch", …)`), and `launch` calls `watch`, which writes the Main record and runs `start.Watch`.

See proposal.md for why waves exist and specs/waves/spec.md for the behaviour.

## Goals / Non-Goals

**Goals:**
- Reuse m12's Main record, escalation and watcher for members unchanged: a member differs from a Board Start only in its base, its Main pane and one prompt sentence.
- Keep every judgment (order, dependencies, conflicts, dropping) with the coordinator agent. kanban only does what is deterministic.

**Non-Goals:**
- No watcher for the coordinator itself (no Stall, no automatic finish). A silent coordinator is visible as its open tab and the members' badges. A follow-up can add a Stall once waves are in use.
- No dependency field in the Change metadata, no automatic ordering.
- No waves for Spec Kit Projects and none with `worktree` off.
- No resume of a coordinator from the Board, and no wave Card.
- No concurrency limit beyond `wave.size`: the coordinator decides how many members run at once.

## Decisions

### D1: The coordinator starts the members, not kanban
The coordinator runs `kanban wave start <N> <change>` for each member when it decides to.
Alternatives: kanban starts all members at once. That lost because a member that builds on another one would branch off `wave-<N>` before its dependency is merged, the same problem waves are meant to solve. Starting them in marking order lost because marking order does not express dependencies.

### D2: The coordinator becomes the member's Main pane through the existing record
`wave start` writes the coordinator's pane as the member's Main record and passes it as `Watched.Main`. Handoff, Stall and `kanban escalate` then reach the coordinator without any change to them.
Alternatives: a separate "handoff target" field next to the Main record. That lost because it duplicates the record and needs every reader to choose between the two.

### D3: The wave branch is the member's base
`wave start` uses `Worktree{On: true, Base: "wave-<N>"}` for placement and `Watched.Base = "wave-<N>"` for the watcher. The existing Handoff text then already says "merge `<change>` into `wave-<N>`, archive, remove the worktree", which is exactly the coordinator's job.
Alternatives: a wave-specific Handoff text. That lost because the generic text already fits, and a second text would drift.

### D4: Members do not archive; the coordinator archives after each merge
`/opsx:archive` rewrites `openspec/specs/<capability>/spec.md`. Two members that touch the same capability and archive on their own branches would conflict on every merge. The member prompt therefore forbids archiving. The coordinator archives each Change on `wave-<N>` right after its merge, one after the other, so each archive applies onto the specs of the previous one.
Alternatives: let members archive and resolve spec conflicts while merging. That lost because spec deltas conflict almost always, and resolving them by hand is error-prone.

### D5: Wave records live in the user cache directory, one JSON file per wave
The path is `<user cache>/spec-kanban/waves/<project hash>/<N>.json`, written atomically like the other records. The `kanban wave` subcommands run inside agent panes, without `HERDR_PLUGIN_STATE_DIR`, so they cannot use the plugin's state directory (the same reason m12 moved the Main records). A subcommand finds the Project by scanning the wave records for number N whose worktree path contains the current directory, else whose Project contains it.
Alternatives: one file per Project holding all waves. That lost because concurrent subcommands (two members dropped at once) would race on one file.

### D6: The wave is placed by a shared worktree helper
`start.place`'s "create a worktree for a branch from a base" part moves into a helper. Start calls it with the Change name and src's Rel check. The wave launch calls it with `wave-<N>`, the base and no Rel check. The coordinator is then launched through the same `start.Launch` (shell-ready probe, agent start, prompt), named `wave-<N>`.
Alternatives: run `git worktree add` directly. That lost because herdr would not know the workspace, and the label, root pane and later `worktree open` would differ from Start.

### D7: Members start without taking focus
`start.Launch` focuses the tab first. For members, a variant skips that step, so the coordinator starting three members does not move the user's focus three times. The shell-ready probe works without focus, because it only uses `pane run` and `wait-output`.
Alternatives: focus as usual. That lost because focus would jump while the user is reading the coordinator.

### D8: "Archived on the wave branch" is read from the wave worktree
`wave finish` asks the Source for the open Changes of the wave's worktree (`src.Open(path)`). A member that is no longer open there is archived on `wave-<N>`.
Alternatives: trust the coordinator's word, or parse `git log`. That lost because the first is unchecked and the second depends on commit messages.

### D9: The Board learns about waves only from the records
The Board reads the Project's wave records when it loads, to draw `≋<N>` for members that are neither dropped nor in a finished wave. Marks are plain Board model state (a set of Change names), dropped on close. The wave start runs detached (`spawn("wave", …)`) with the marked names and `KANBAN_MAIN_PANE`, like `launch`.

## Risks / Trade-offs

- [The coordinator's context grows with each Handoff and conflict] → Handoff prompts stay short (branch, Progress, worktree). `wave.size` caps a wave at 5 by default.
- [A member with an existing checkout keeps its old base (e.g. `main`)] → The spec states it, and the coordinator merges it like any branch. A conflict is its job.
- [The coordinator stops silently] → Non-goal for now. The coordinator's tab stays visible, and m13-attention's marker can later cover it.
- [A Change named `wave-<N>` would collide with the coordinator's agent name] → The wave number skips existing `wave-<number>` branches. A Change named like that would be refused by Start's existing check that the agent name is taken.
- [Several Boards start waves at the same time] → N counts records and branches, and `worktree create` fails on an existing branch, so the second wave fails cleanly and records nothing.
- [The coordinator merges something broken into `wave-<N>`] → The Main agent still reviews `wave-<N>` before merging into the base, and the plugin never merges itself.

## Migration Plan

- Merge m12-escalation first: m14 builds on its Main records, `kanban escalate` and watcher, and its deltas modify requirements as they read after m12.
- No data migration: projects without wave records behave as before. The `wave` config key is optional.
- Rollback: remove the `wave` key from `config.json` (it is then simply unused) or install the previous binary. Leftover `wave-<N>` branches and worktrees are ordinary git state.
