## Context

`start.Start(project, change, run, openspecRun)` (M0) names the herdr agent after the Change, looks it up with `agent get <change>`, focuses a running one, restarts an exited one in its tab (checked with `pane process-info`), or creates a tab with `tab create --cwd <project> --label <change> --focus` and then runs `agent start` (about 4 s to ready) and `agent prompt`. It is reachable only as `kanban start <change>` from a shell.

The Board runs as a herdr popup pane (`[[panes]] placement = "popup"`); quitting the Bubble Tea program closes the popup. It opens instantly from the cached snapshot (f1) while the fresh load runs in a `tea.Cmd`. Footer notices already exist (`m.notice`, cleared by the next key).

herdr 0.9.5 facts relevant here (from `herdr agent list`, the M0 fixtures and `herdr --skill`):
- `agent list` returns all agents of the session in one call with `name` (only for agents started with a name), `cwd`, `pane_id`, `agent_status`.
- Names match `[a-z][a-z0-9_-]{0,31}`, are unique among live agents across all workspaces, and follow the pane occupant.
- `agent_status` (`idle`, `working`, `blocked`, `done`, `unknown`) is a lifecycle state; `done` means a finished turn not yet seen, not that the agent exited.
- `tab create` accepts `--no-focus`.

## Goals / Non-Goals

**Goals:**
- `s` starts or focuses a Change's agent from the Board, reusing `start.Start` unchanged in behaviour.
- One `herdr agent list` call per Board open and per `r`, never on the path to the first frame.
- Fakes for herdr and Start in Board tests; table tests for the agent list parsing.

**Non-Goals:**
- `s` in the detail view (Esc then `s` is enough; the detail has its own footer and notice).
- Showing the agent's status (`working`, `blocked`, …) on the Card; the herdr sidebar already shows it.
- Phase gating for `kanban start` on the command line; the shell command stays the M0 escape hatch.
- Refreshing badges on herdr events while the Board is open (no watcher, ADR 0003 spirit).
- Detecting agents not started by Start (they carry no Change name).

## Decisions

### Running agent = named agent in the Project
`start.Running(project string, run herdr.Runner) (map[string]bool, error)` calls `herdr agent list` once and returns the names of agents whose `name` is set and whose `cwd` equals the Project or lies below it. The naming rule (agent name = Change name) stays the single convention from M0, now used by both Start and the badge. `agent_status` is ignored: every status of a live agent counts as running, since `done` only means "finished a turn". Alternative: `agent get` per Card (N calls, slower, and misses the cwd check); `pane process-info` per match to rule out an exited Claude (see Open Questions).

### Badge loads asynchronously
`board.New` takes the lookup as an injected `func() (map[string]bool, error)` (nil in tests that do not care). `Init` batches it with the Changes load, and `r` starts both. The result arrives as its own message and is stored by name, independent of the columns, so it survives Change reloads and cache swaps. An error clears nothing visible and shows nothing: outside herdr there are simply no badges. The lookup is a local socket call (tens of ms), but even so it never blocks the cached first frame.

### Badge look
A `●` before the Change name on the Card's first line, in ANSI green (`lipgloss.Color("2")`), so the terminal theme decides the shade. The name loses two cells of width (`FitName` already handles it). Not on Archived Cards. Alternatives: a word like `agent` on the third line (collides with the plan marks and age in narrow columns); colouring the whole Card (conflicts with the bold selection).

### `s` runs Start inside the Board, then quits
`board.New` also takes `start func(change string) error`, wired in `cmd/kanban` to `start.Start(proj, change, herdr.Exec, openspecCLI(proj))`. `s` checks the Phase, sets `starting = <change>` and returns a `tea.Cmd` that calls Start. On success the Board returns `tea.Quit`, which closes the popup; on error it clears `starting` and shows `start failed: <err>` as the notice. Alternatives: quit first and run `kanban start` detached (errors vanish with the popup); open the tab from the Board directly (duplicates Start).

While `starting` is set, every key except `ctrl+c` is ignored and the footer reads `starting <change>…`. Quitting mid-Start would kill the process between `tab create` and `agent prompt`, which is exactly the half-started state the agent-start spec forbids leaving unexplained; `ctrl+c` stays as the user's explicit escape.

### Focus moves last
Start creates the tab with `--no-focus` instead of `--focus` and always ends with `agent focus <change>` (today only the reuse path does). The visible result is the same (the agent's tab is focused when Start returns), but focus no longer moves away from the popup while Start is still running in it. This removes the question of what herdr does to a popup whose tab loses focus mid-Start, and makes the three paths end alike.

### Phase gating lives in the Board
Ready and In Progress start; the notices are `<change> is a draft: plan it first`, `all tasks of <change> are done`, `archived changes cannot be started`. The Board already knows the Phase of each column, so no extra lookup is needed. Start's own `checkOpen` still runs (one `openspec list`, ~0.65 s) and guards against a Board that is out of date.

### Project check for an existing agent
`Start` reads `cwd` from the `agent get` result it already parses. If it is outside the Project, it returns `agent <change> already runs in <dir>` before any focus or restart. This closes the collision the M0 design deferred to M4 and keeps the badge and `s` consistent: a Card without a badge never focuses a foreign agent.

## Risks / Trade-offs

- [herdr keeps an exited Claude registered (M0 spike: `agent_status: "done"` after exit), so the badge may linger until the tab closes] → accepted; `s` still does the right thing because Start checks `pane process-info` and restarts.
- [The popup pane's herdr context decides the workspace of `tab create`; it may differ from the Project's workspace] → smoke test checks the tab lands in the invoking workspace; if not, pass the workspace id from `kanban open`'s context through an env var to `--workspace`.
- [A Start takes ~5 s in the popup with only a footer line as feedback] → acceptable; a running agent is focused in well under a second.
- [Board and agent lookup race on `r`] → results are independent messages; the last one wins, both are fresh.

## Migration Plan

Rebuild the plugin; no config changes. Rollback is reverting the change; `kanban start` behaves as in M0 apart from the foreign-agent check and focusing last.

## Open Questions

- Should the badge confirm with `pane process-info` that Claude is still in the foreground (one extra call per badge) to hide exited agents? Deferred until the smoke test shows whether herdr 0.9.5 clears the name on exit, as its docs say, or keeps it, as the M0 spike saw.

## Notes from review

- The foreign-agent check reads the agent's pane `cwd`. If Claude exited and the user then `cd`'d out of the Project in that shell, Start refuses instead of restarting there. Accepted: restarting would run Claude outside the Project anyway. An agent without a `cwd` is not checked.
- A failed Start (for example "prompt not sent") may leave a named agent behind, so the Board looks up Running agents again after a failure.
