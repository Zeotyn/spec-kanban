## Why

Several Ready Changes can be worked on at the same time, but today each Start is a single Change whose Handoff goes to the Main agent. Parallel Starts lead to parallel branches. The Main agent then gets several Handoffs at once, and the second or third merge into the base conflicts, because most Changes touch the same files. Dependent Changes also break: a worktree branches off the base, so a Change that builds on another one starts without that one's code. The user wants to start a group of Changes at once and hear back only when the group is merged, or when an agent cannot go on.

## What Changes

- **Wave.** A wave is a group of 2 to `wave.size` Ready Changes of one OpenSpec Project, started together. It is plugin state, not a Change.
  - It gets its own branch `wave-<N>` from the base, in its own worktree, with a **coordinator** agent. N is the next free wave number of the Project.
  - The wave's members branch off `wave-<N>`, not off the base.
- **Coordinator.** The coordinator gets a prompt from kanban that names the members, their proposals, the commands below and these rules:
  - It reads the proposals and decides the order. It starts a member only after the members that member builds on are merged into `wave-<N>`.
  - A finished member hands off to the coordinator, not to the Main agent. The coordinator reviews it, merges it into `wave-<N>` and resolves every conflict itself. It runs the Project's tests, archives the Change on `wave-<N>` and removes the member's worktree.
  - It may drop a member that cannot be made to work. That member leaves the wave and stays an ordinary Change.
  - Members escalate to the coordinator. The coordinator escalates to the Main agent only when it cannot decide itself.
  - When every member is archived on `wave-<N>` or dropped, it finishes the wave. Then **one** Handoff goes to the Main agent: review `wave-<N>`, merge it into the base, remove the wave worktree.
- **New subcommands** for the coordinator:
  - `kanban wave start <N> <change>` starts a member, watched, with the coordinator as its Main agent. The member's prompt adds that it must not archive the Change.
  - `kanban wave drop <N> <change>`
  - `kanban wave escalate <N> "<reason>"`
  - `kanban wave finish <N>` refuses while a member is neither archived on `wave-<N>` nor dropped.
- **Board.**
  - `space` marks or unmarks a Ready Card, shown as `✓` before its name. Marks survive filtering and are dropped when the Board closes.
  - Marking more than `wave.size` Cards is refused with `a wave holds at most <size> changes`.
  - `s` with 2 or more marks asks `start wave of <N>? y/n`. `y` starts the wave and the Board closes. `s` with 0 or 1 mark behaves as today.
  - `Esc` clears the filter, then the marks, then closes the Board.
  - A Card of an active wave's member shows the badge `≋<N>`.
  - A wave needs `worktree` on and an OpenSpec Project. Otherwise `s` with marks starts nothing and says why.
- **Config.** New key `wave`, an object with `size`: the maximum number of members, an integer of at least 2, default 5.

## Capabilities

### New Capabilities

- `waves`: what a wave is and where its state lives, its branch and worktree, the coordinator and its prompt, `kanban wave start|drop|escalate|finish`, and the coordinator as Main agent of the members.

### Modified Capabilities

- `board`: marking Cards, starting a wave, the `≋<N>` badge, and `Esc` clearing marks before it closes the Board.
- `kanban-config`: the `wave` setting and its validation.

## Impact

- **Code:**
  - `internal/config`: `wave.size`.
  - `internal/cache`: wave records per Project in the user cache directory, next to the Main records from m12.
  - `internal/start`: a wave's branch, worktree and coordinator launch; the member start with the wave branch as base and the coordinator as Main pane.
  - `internal/board`: marks, the wave confirmation, the badge, `Esc` order, key help.
  - `cmd/kanban`: the `wave` subcommands and the Board's wave callback.
- **Docs:** README (Waves), CONTEXT.md (Wave, Coordinator), CHANGELOG.
- **Builds on m12-escalation:**
  - It uses the Main pane record, `kanban escalate`, Stall and the Handoff.
  - Its deltas modify requirements as they stand after m12. Archive m14 after m12 is merged into `main`.
- No new dependencies.
