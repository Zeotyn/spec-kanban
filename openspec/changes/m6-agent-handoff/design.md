## Context

A Start from the Board runs `start.Prepare` in the Board, then spawns `kanban launch <change> <pane> <tab> <prompt>` (prompt argument from m5-help-and-config) detached with `Setsid`; it inherits the Board's environment (`KANBAN_PROJECT`, without `HERDR_PANE_ID`/`HERDR_TAB_ID`) and its cwd, the Project. `launch` runs `start.Launch` (`tab focus`, `agent start … --timeout 60000`, `agent prompt`) and exits. m5-help-and-config adds the config key `agent` (any herdr `--kind`, default `claude`) and detects a running agent by herdr's `agent` field instead of `argv0 == claude`; m5 also makes Start's kind default to the invoking pane's agent (config `agent` overrides, `claude` is the fallback) and has `kanban open` pass that pane as `KANBAN_MAIN_PANE`; `agent get <focused_pane_id>` returns its kind in `agent` and `agent_session.agent`. m6 builds on that and hardcodes no kind. `kanban open` reads only `focused_pane_cwd` and `workspace_cwd` from `HERDR_PLUGIN_CONTEXT_JSON`. The cache directory is `$HERDR_PLUGIN_STATE_DIR`, else `os.UserCacheDir()/openspec-kanban`.

herdr 0.9.5 facts (`--help`, `herdr --skill`, `herdr api schema`, herdr docs):
- The action context (`PluginInvocationContext`) carries `focused_pane_id`, `focused_pane_agent` and `focused_pane_status` besides the cwd fields.
- `agent list`/`agent get` report `agent` (the kind), `agent_session: {agent, kind: "id"|"path", source, value}` (nullable), `pane_id`, `tab_id`, `agent_status`. Agent commands accept an agent name or a pane id. herdr's API also has a `resume_argv` that an agent can report about itself, but `agent get` does not return it, so there is no generic resume command to read.
- Only Claude sessions could be observed locally (`kind: id`, a session UUID); the session shape of other kinds is unverified.
- `agent wait <target> [--until <state>]… [--timeout <ms>]`: states `idle`, `working`, `blocked`, `done`, `unknown`; without `--until` it matches the settled states `idle`, `done`, `blocked`; without `--timeout` it waits forever. `done` is `idle` not yet seen.
- `agent prompt <target> <text> [--wait]` rejects a blocked target with `agent_blocked` before typing. With `--wait` it requires observed `working`/`blocked` within 5 s (`agent_prompt_stalled` otherwise), then waits for the first settled state. It does not track turns.
- `agent start <name> --kind <kind> --pane <id> [-- <args>]` passes extra args to the agent's executable.
- A name is cleared when the agent exits or its pane goes away; `pane close <id>` closes one pane, `tab close <id>` the whole tab. herdr itself notifies when an agent finishes or needs input.

## Goals / Non-Goals

**Goals:**
- The watcher lives in the existing detached `kanban launch`; no daemon, no event hook, no new dependency.
- Everything is testable with the fake `herdr.Runner` and a temp state dir.
- Watch, clean check, Handoff and Main agent prompt are kind-agnostic; only reopening a session is kind-specific.

**Non-Goals:**
- Handoff for `kanban start` from a shell (it runs in the foreground; the shell's user is right there).
- A configurable Handoff prompt (m5 config could carry it later).
- Detecting a question asked in plain text: an agent that stops with all tasks completed and asks "shall I also …?" counts as finished; the session can be resumed.
- Treating a Change that the agent archived itself as finished (see Open Questions).
- Showing the recorded session on the Card.
- Resume arguments for kinds other than `claude` (added one line each once a kind's session shape is confirmed).

## Decisions

### The Main agent travels as an env var (from m5)
m6 reuses m5's `KANBAN_MAIN_PANE` (the invoking pane's `focused_pane_id`, set by `kanban open` on the Board pane). The Board does not read it for m6; the spawned `kanban launch` inherits it with the rest of the Board's environment, so no argument changes. The Main agent is whatever agent runs in that pane, of any kind. Whether that pane runs an agent is checked only at Handoff time (`agent get <pane>`), because the user may start or quit Claude there meanwhile. Alternatives: an extra `launch` argument (more plumbing for the same value); `focused_pane_agent` at open time (stale by Handoff time). If m5 ships without the env var, m6 adds it in `kanban open` (one line).

### Watch loop in `kanban launch`
After `agent start` succeeds, `launch` records the session (below). It sends the prompt with `agent prompt <change> <prompt> --wait`, which returns at the first settle of the new turn, so an agent still idle from startup is not mistaken for a finished one; `agent_prompt_stalled` falls through to waiting for `working`. Then the loop:

1. Settled: `agent get <change>` → status, pane, session; record the session; `openspec list --json` in the Project → the Change's Progress.
2. Clean (status `idle` or `done`, Change open with tasks, all completed) → Handoff, exit.
3. Otherwise → notification `openspec-kanban: <change> stopped` with `3/7 tasks` or `waits for input`, then `agent wait <change> --until working`, then `agent wait <change>` (settled), back to 1.
4. Any `agent_not_found` (or other error) from `wait`/`get` → exit 0 silently: Claude exited, the pane/tab was closed, or herdr went away. A Start of the exited agent spawns a fresh `launch` with its own loop.

No overall time limit: the process sleeps in a blocking `herdr agent wait`, one per Change started from the Board, and ends with its agent. The loop lives in `internal/start` as `Watch(project, change, run, openspecRun)` with `launch` as a thin caller. Alternatives: a `pane.agent_status_changed` event hook (stateless, but would have to re-derive which agent belongs to which Main agent and debounce bursts; the hook already runs `kanban sync`); polling `agent list` (wakes up for nothing).

### Handoff order: record, close, prompt
Record the session, `pane close <agent pane>` (not `tab close`: a pane the user split off survives; a single-pane tab closes with it), then prompt the Main agent. Closing before prompting means a failed prompt never leaves the finished tab open. The Main agent is prompted only when `KANBAN_MAIN_PANE` is set, differs from the agent's pane, `agent get <main>` finds an agent (any kind) and its status is not `blocked`; then `agent prompt <main> <text>` without `--wait` (Claude Code queues input that arrives while it works; other kinds are covered by the same Open Question). Any failure or rejected check → `notification show "openspec-kanban: <change> finished" --body <text>`.

Text: `<change> is finished (7/7 tasks). Please review it and archive it with /opsx:archive <change>. To look into its session press s on its Card in the Board, or run: claude --resume <session>` (the last sentence only when the session is reopenable, with that kind's command).

### Session record: one small file per Project and Change
`cache.WriteSession(dir, project, change, s)` writes `s` = herdr's `agent_session` as reported (`{agent, kind, value}`, small JSON) to `<dir>/sessions/<project hash>/<change>` atomically (temp file + rename, like the Changes cache); `cache.ReadSession` returns the zero value when missing or unreadable. Storing herdr's shape as is keeps the record kind-agnostic; an agent without `agent_session` records nothing. One file per Change means concurrent watchers never race on a shared file. It lives next to the Changes cache because both are disposable plugin state. Alternatives: a field in the Changes cache (rewritten by every load, a read-modify-write race with watchers); one JSON map per Project (same race).

### Reopening is kind-specific: a small resume table
`start.ResumeArgs(s Session) []string` maps a recorded session to the agent args, decided by `agent_session.agent`: `claude` with `kind == "id"` → `["--resume", value]`; everything else → nil, meaning not reopenable, and `s` on its Review Card shows `resume not supported for <kind>`. A new kind is one more case once its session shape and CLI flag are confirmed (for example codex `resume <id>`, unverified). Alternatives: a config key for the resume argument (YAGNI until someone runs another kind); using herdr's `resume_argv` (not exposed by `agent get` in 0.9.5). The kind to start is the recorded `agent`, not the current config, so a session always reopens in the agent that wrote it.

### `s` on Review resumes
The Board gets a `resumable func(change string) bool` (record read + `ResumeArgs != nil`) and, as for Start, an injected func. On Review: a Running agent → `start.Prepare` focuses it as today; otherwise with a recorded session → `Prepare` (new tab, or the exited agent's tab) and spawn `kanban resume <change> <pane> <tab>`, which reads the record and runs `tab focus`, `agent start <change> --kind <recorded agent> --pane <pane> --timeout 60000 -- <ResumeArgs>` and nothing else (no prompt, no watch, so a resumed session in Review is not closed again at once). Failure → notification `openspec-kanban: resume failed`. Without a record → today's notice; a record of another kind → `resume not supported for <kind>`. A separate subcommand keeps `launch`'s argument list as m5 defines it.

### Glossary
New terms for CONTEXT.md (proposed, not edited here): **Main agent** (the agent, of any kind, in the pane the Board was opened from; _Avoid_: parent, orchestrator), **Handoff** (closing a finished Change's agent tab and telling the Main agent; _Avoid_: completion, callback), **Recorded session** (the agent kind and session herdr last reported for a Change's agent). **Start** changes: no longer only "hands a Change to a new agent session" — on a Review Card it resumes the Recorded session.

## Risks / Trade-offs

- [A pending `agent wait` might not return when Claude exits or the tab closes, leaving a sleeping process] → smoke test checks it; if it hangs, wait in slices (`--timeout 600000`) and re-check `agent get` between them.
- [The prompt is typed into the Main agent while it is in a dialog herdr does not recognise as blocked] → accepted; herdr checks `blocked`, and the text is plain prose.
- [The agent settles with all tasks ticked but uncommitted or broken work] → the Main agent is asked to review, and the session is resumable; the user explicitly wants auto-close.
- [`s` on a Review Card no longer only explains] → it now opens a tab; the footer still says what starts.
- [One notification per unfinished settle may duplicate herdr's own done/blocked notification] → accepted; ours adds the Progress.
- [Plugin state lost (cache dir cleared)] → only resume from the Board is lost; the resume command from the Handoff text still works.
- [Kinds other than `claude` cannot be reopened from the Board] → they get the notice; the Handoff itself works for them, and the session usually can be reopened in the agent's own UI.

## Migration Plan

Rebuild the plugin. Changes started before this change have no record and no watcher; their Review Cards show the old notice. Rollback: revert; leftover session files are ignored.

## Open Questions

- Should a Change that is no longer open because the agent archived it (possible with an m5 prompt like `/opsx:apply {change} then archive`) count as finished? Planned: no, it gets the "stopped" notification; easy to widen after the smoke test.
- Should watching have an upper bound (for example 24 h) for agents left idle with open tasks forever? Planned: no bound; revisit if leftover processes show up.
- Which resume argument do other kinds need, and what does herdr report as their `agent_session`? Answer per kind when someone uses it; adds one case to `ResumeArgs`.
- Does Claude Code (or another Main agent kind) reliably queue a herdr prompt submitted while the Main agent is `working`? The smoke test checks it; if not, wait for the Main agent to settle (bounded) before prompting.
