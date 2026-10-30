## Context

A Start from the Board (`showBoard`'s `startOne` in `cmd/kanban/main.go`) runs `start.Prepare(project, change, start.Worktree{On, Base}, run, git)` in the Board, resolves the kind with `start.Kind(cfg.Agent, $KANBAN_MAIN_PANE, run)`, then spawns `kanban launch <change> <pane> <tab> <kind> <prompt>` detached with `Setsid`. The child inherits the Board's environment through `childEnv` (`KANBAN_PROJECT`, `KANBAN_MAIN_PANE`; `HERDR_PANE_ID`/`HERDR_TAB_ID` dropped) and runs `start.Launch(change, kind, prompt, pane, tab, run)`: `tab focus`, `agent start <change> --kind <kind> --pane <pane> --timeout 60000`, `agent prompt <change> <prompt>`, then exits; a failure becomes the notification `openspec-kanban: start failed`. `kanban open` sets `KANBAN_MAIN_PANE` to the invoking pane's `focused_pane_id` (m5).

Since m6-worktree-start, `Prepare` puts the agent in the Change's checkout when config `worktree` is on (default) and the Project is the top of a git checkout: an existing checkout on branch `<change>` is reused (`worktree open`, or a new tab in its open workspace), otherwise `herdr worktree create --cwd <project> --branch <change> [--base <base>] --label <change> --no-focus` creates one under herdr's `[worktrees] directory` as a workspace labelled with the Change. `worktree.List(git)` lists the checkouts (`Checkout{Path, Branch}`, the Project's own included), `worktree.Base(configured, git)` names the base. `openspec.Load(project, cli, others, known)` reads each Change worked on in a linked checkout from there (`Change.Dir` = checkout path); a Change archived inside a worktree is not seen by it (its `openspec/changes/<change>` is gone there and the Project's archive does not have it until merged), so the Project's stale open Change is returned. The loader in `main.go` (`loader(dir, proj, cli, git)`) wraps `Load` with the Changes cache. Nothing removes worktrees yet (README).

herdr 0.9.5 facts (`--help`, `herdr --skill`, `herdr api schema`, herdr docs):
- `agent list`/`agent get` report `agent` (the kind), `agent_session: {agent, kind: "id"|"path", source, value}` (nullable), `pane_id`, `tab_id`, `cwd`, `agent_status`. Agent commands accept an agent name or a pane id. `resume_argv` exists only as an input of the integrations' report requests; neither `agent get`/`agent list` nor `herdr api snapshot` returns it, so there is no generic resume command to read.
- `agent_session` is filled by herdr's official integrations (`herdr integration status`). herdr's session restore docs (https://raw.githubusercontent.com/herdrdev/herdr/v0.9.5/docs/next/website/src/content/docs/session-state.mdx) list the resume command per kind, among them `claude --resume <id>`, `codex resume <id>`, `copilot --resume=<id>`, `opencode --session <id>`, `cursor-agent --resume <id>`. Gemini CLI has no herdr integration in 0.9.5 (the `~/.gemini` one is Antigravity, kind `agy`), so a `gemini` agent reports no `agent_session`. Only Claude sessions were observed locally (`kind: id`, `source: herdr:claude`, a UUID).
- `agent wait <target> [--until <state>]… [--timeout <ms>]`: states `idle`, `working`, `blocked`, `done`, `unknown`; without `--until` it matches the settled states `idle`, `done`, `blocked`; without `--timeout` it waits forever. `done` is `idle` not yet seen.
- `agent prompt <target> <text> [--wait]` rejects a blocked target with `agent_blocked` before typing.
- `agent start <name> --kind <kind> --pane <id> [-- <args>]` passes extra args to the agent's executable.
- A name is cleared when the agent exits or its pane goes away; `pane close <id>` closes one pane, `tab close <id>` the whole tab, `workspace close <id>` a workspace.
- `worktree remove --workspace <ID> [--force]` runs `git worktree remove` on an open linked workspace and never deletes the branch; it needs that workspace open.

## Goals / Non-Goals

**Goals:**
- The watcher lives in the existing detached `kanban launch`; no daemon, no event hook, no new dependency.
- Everything is testable with the fake `herdr.Runner`, a fake loader and a temp state dir.
- Watch, clean check, Handoff and Main agent prompt are kind-agnostic; only reopening a session is kind-specific.
- No work is ever lost by the Handoff: the plugin merges, removes and deletes nothing.

**Non-Goals:**
- Handoff for `kanban start` from a shell (it runs in the foreground; the shell's user is right there).
- Merging, removing a worktree or deleting a branch in the plugin.
- A configurable Handoff prompt.
- Detecting a question asked in plain text: an agent that stops with all tasks completed and asks "shall I also …?" counts as finished; the session can be resumed.
- Showing the recorded session on the Card.
- Resume arguments for kinds without a verified source (one map entry each once verified).
- An upper bound on watching.

## Decisions

### The Main agent travels as an env var (from m5)
`KANBAN_MAIN_PANE` already reaches `kanban launch` (`childEnv` keeps it). The Main agent is whatever agent runs in that pane, of any kind, checked only at Handoff time (`agent get <pane>`), because the user may start or quit an agent there meanwhile. `kanban launch` keeps its five arguments.

### Watch loop in `kanban launch`
`kanban launch` runs `start.Launch` unchanged and, when it succeeds and `KANBAN_PROJECT` is set, `start.Watch(w, run, git)` with

```go
type Watched struct {
	Project, Change, Main, Base string           // Main from KANBAN_MAIN_PANE, Base = config base ("" detects)
	On     bool                                  // config worktree: without it no checkout is looked at
	Load   func() ([]openspec.Change, error)     // main.go's loader for the Project
	Record func(herdr.Session)                   // cache.WriteSession for Project and Change
}
```

`launch` loads the config for `Base` (a warning is ignored) and passes `loader(cache.Dir(), proj, openspecCLI, gitCLI(proj))`, so watching also warms the Board's cache. Steps:

0. `agent get <change>` → record the session (the agent is ready now). `agent wait <change> --until working --timeout 60000`; any error falls through, so an agent that finished its turn before the wait started is still evaluated.
1. `agent wait <change>` (settled).
2. `agent get <change>` → status, pane, session; record the session.
3. A status other than `idle`, `done` or `blocked` (its next turn began before it was read) → back to 1, no notification. Clean = status `idle` or `done`, and the Change is finished (below) → Handoff, return.
4. Otherwise → notification `openspec-kanban: <change> stopped` with `3/7 tasks` or `waits for input` (or the load error), then `agent get <change>` again: its `state_change_seq` changed (it moved on while being checked, maybe through a whole turn) → back to 1; else `agent wait <change> --until working`, back to 1. With `worktree` off the loader gets no other checkouts, so a leftover checkout of the Change does not supply its Progress.
5. `agent_not_found` (or any other error) from `wait`/`get` → return nil silently: the agent exited, its pane/tab was closed, or herdr went away.

Launch keeps `agent prompt` without `--wait`, so `kanban start` and its tests do not change. The loop lives in `internal/start`; `launch` in `main.go` is a thin caller. Alternatives: `agent prompt --wait` inside Launch (would block `kanban start` in the shell or need a flag); a `pane.agent_status_changed` event hook (would have to re-derive which agent belongs to which Main agent and debounce bursts); polling `agent list`.

### Finished: Progress from the checkout, archive in either place
`checkout` is the linked checkout of the Change: the entry of `worktree.List(git)` with `Branch == change` and a path other than the Project ("" when none: worktree off, no git, or the Project itself on that branch). The Change is finished when:
- it is archived: `Load` returns it with `Archived`, or `checkout != ""` and `openspec.ArchivedIn(checkout, change)` (its `openspec/changes/archive` holds `<change>` or `<date>-<change>`); this is checked first, because `Load` then still returns the Project's stale open Change; or
- `Load` returns it open with `Total > 0` and `Done == Total` (read from the worktree through `Dir`).
A Change that is neither (deleted, renamed) is not finished and gets the "stopped" notification. `openspec.ArchivedIn` is the archive scan `Load` already does, factored out and reused there.

### Handoff order: record, close, prompt; the Main agent cleans up
Record the session, `pane close <agent pane>` (not `tab close`: a pane the user split off survives; a single-pane tab closes with it, and so does a worktree workspace that has nothing else open), then prompt the Main agent. Closing before prompting means a failed prompt never leaves the finished tab open. The Main agent is prompted only when `Main` is set, differs from the agent's pane, `agent get <main>` finds an agent (any kind) and its status is not `blocked`; then `agent prompt <main> <text>` without `--wait`. Any failure or rejected check → `notification show "openspec-kanban: <change> finished" --body <text>`.

Cleanup decision: the plugin removes nothing. `herdr worktree remove --workspace <id>` needs the worktree workspace open, but the Handoff just closed it with the agent's pane, and keeping it open would leave the finished tab behind. So the text asks the Main agent, after review and merge, to run `git worktree remove <path>` from the main checkout, which works without herdr and refuses a checkout with uncommitted or untracked changes unless forced. The text forbids `--force` and deleting the branch, and says to stop when the merge or the removal does not go through cleanly, so unmerged commits stay on the branch and uncommitted work stays in the checkout. Alternatives: the plugin removes the worktree after the merge (it would have to watch the merge, and a merge done by hand would be missed); keep the workspace and name `herdr worktree remove --workspace <id>` (the finished tab stays open, and not every Main agent kind knows herdr).

Texts (`<base>` = `worktree.Base(w.Base, git)`, only computed with a checkout):
- no checkout, open: `<change> is finished (7/7 tasks). Please review it and archive it with /opsx:archive <change>.`
- no checkout, archived: `<change> is finished and archived. Please review it.`
- checkout, open: `<change> is finished (7/7 tasks) on branch <change> in <path>. Please review it, merge branch <change> into <base>, archive it with /opsx:archive <change>, then remove its worktree with: git worktree remove <path>. Never use --force and keep the branch; if the merge or the removal does not go through cleanly, stop and tell the user.`
- checkout, archived (by the agent in the worktree): `<change> is finished and archived on branch <change> in <path>. Please review it, merge branch <change> into <base>, then remove its worktree with: git worktree remove <path>. Never use --force …` (same last sentence).
- suffix when the session is reopenable: ` To look into its session press s on its Card in the Board, or run in <dir>: <resume command>` (`<dir>` = checkout, else Project; `start.ResumeCommand`, e.g. `claude --resume <session>`).

The base is recomputed at Handoff time rather than recorded at Start: it is the configured `base` or the default branch, which is what Start used for a new branch; a branch that already existed before Start has no better-known base.

### Session record: one small file per Project and Change
`herdr.Session{Agent, Kind, Value string}` is herdr's `agent_session` as reported. `cache.WriteSession(dir, project, change, s)` writes it as small JSON to `<dir>/sessions/<project hash>/<change>` atomically (temp file + rename, like the Changes cache); `cache.ReadSession` returns the zero value when missing or unreadable. An agent without `agent_session` records nothing. One file per Change means concurrent watchers never race. The record is keyed by the Project, not the checkout, so it survives removing the worktree. Alternatives: a field in the Changes cache (rewritten by every load, a race with watchers); one JSON map per Project (same race).

### Reopening is kind-specific: a plain resume map
`start.ResumeArgs(s herdr.Session) []string` maps a recorded session to the args appended after the executable herdr launches (`agent start … -- <args>`), keyed by `s.Agent` and only for `s.Kind == "id"`; anything else → nil, meaning not reopenable, and `s` on its Review Card shows `resume not supported for <kind>`. `start.ResumeCommand(s)` prefixes the kind's executable for the Handoff text. Only kinds with a verified source are in the map:

| kind | executable | args | source |
|---|---|---|---|
| `claude` | `claude` | `--resume <id>` | `claude --help` 2.1.314 ("`-r, --resume [value]` Resume a conversation by session ID"); herdr session-state docs |
| `copilot` | `copilot` | `--resume=<id>` | `copilot --help` 1.0.105 ("Resume a specific session by ID: `copilot --resume=<session-id>`"); herdr session-state docs |
| `codex` | `codex` | `resume <id>` | https://learn.chatgpt.com/docs/developer-commands?surface=cli (`codex resume` with positional `SESSION_ID`); herdr session-state docs |
| `opencode` | `opencode` | `--session <id>` | https://opencode.ai/docs/cli/ (`--session`/`-s`: session ID to continue); herdr session-state docs |
| `cursor` | `cursor-agent` | `--resume <id>` | https://cursor.com/docs/cli/reference/parameters (`--resume [chatId]`); herdr session-state docs |
| `gemini` | `gemini` | `--resume <id>` | `gemini --help` 0.65.0; https://geminicli.com/docs/cli/session-management/ ("Resume by ID": full session UUID) |

`copilot` takes `--resume=<id>` as one arg because its `--resume` value is optional. `codex` resumes through a subcommand; `-- resume <id>` starts `codex resume <id>`. The `gemini` entry stays inert until herdr reports a `gemini` session. Every other kind (`pi`, `agy`, `kimi`, …) keeps the notice. The kind to start is the recorded `agent`, not the config or the Main agent's kind (`start.Kind` is not used for resume), so a session always reopens in the agent that wrote it.

### `s` on Review resumes through Prepare
`board.New` gets two more injected funcs: `resumable func(change string) (kind string, ok bool)` (`cache.ReadSession` + `ResumeArgs != nil`; `kind` is the recorded agent, "" without a record) and `resume func(change string) error`. On a Review Card: a Running badge or a reopenable record → `resume(change)`, with the same footer, key lock and close-on-success as Start; a record of another kind → `resume not supported for <kind>`; no record → today's `all tasks of <change> are done`.

`showBoard`'s `resume` runs the same `start.Prepare(proj, change, wt, herdr.Exec, git)` as Start: a running agent is focused (pane "" → done); otherwise Prepare returns the exited agent's pane, the root pane of the reopened worktree workspace, a new tab in the already open workspace, or a new worktree from the existing branch when the Main agent already removed the checkout. Before spawning, it re-reads the record and fails with `no reopenable session for <change>` if there is none (a Running badge without a record that Prepare found idle). Then it spawns `kanban resume <change> <pane> <tab>` with the same `childEnv`/`Setsid` as `launch` (`spawnLaunch` generalised to a subcommand and args). `kanban resume` reads the record and runs `start.Resume(change, s, pane, tab, run)`: `tab focus`, `agent start <change> --kind <s.Agent> --pane <pane> --timeout 60000 -- <ResumeArgs(s)>` and nothing else (no prompt, no watch, so a resumed session in Review is not closed again at once). Failure → notification `openspec-kanban: resume failed`. A separate subcommand keeps `launch`'s argument list.

### A Change the agent archived counts as finished
With a prompt like `/opsx:apply {change} then archive` the agent may archive the Change itself, in the Project (worktree off) or inside its worktree. Both are a Handoff; the text drops the archive request, and for a worktree still asks to merge (the archive lives on the branch) and remove the checkout.

### No upper bound on watching
Watching ends with a Handoff or when the agent goes away; closing the agent tab ends it. An agent left idle with open tasks keeps one sleeping `kanban launch`.

### Glossary
New terms for CONTEXT.md (proposed, not edited here): **Main agent** (the agent, of any kind, in the pane the Board was opened from; _Avoid_: parent, orchestrator), **Handoff** (closing a finished Change's agent pane and telling the Main agent to review, merge, archive and remove its checkout; _Avoid_: completion, callback), **Recorded session** (the agent kind and session herdr last reported for a Change's agent). **Start** changes: on a Review Card it resumes the Recorded session in the Change checkout.

## Risks / Trade-offs

- [A pending `agent wait` might not return when the agent exits or the tab closes, leaving a sleeping process] → smoke test checks it; if it hangs, wait in slices (`--timeout 600000`) and re-check `agent get` between them.
- [Closing the last pane of a worktree workspace might leave an empty workspace open] → smoke test; harmless, and `git worktree remove` does not need it closed; if it stays, the Handoff also runs `workspace close` for a workspace with no pane left.
- [The user kept another tab open in the worktree workspace and the Main agent removes the checkout] → git refuses while that pane holds uncommitted work; otherwise the leftover workspace shows a missing directory and the user closes it.
- [The agent left uncommitted work in the worktree] → `git worktree remove` refuses without `--force`, which the text forbids; the Main agent reports it.
- [The base recomputed at Handoff differs from where the branch started (config changed, branch created by hand)] → the Main agent names the base it merges into; the user sees it in the prompt.
- [The prompt is typed into the Main agent while it is in a dialog herdr does not recognise as blocked] → accepted; herdr checks `blocked`, and the text is plain prose.
- [The agent settles with all tasks ticked but broken work] → the Main agent is asked to review before merging, and the session is resumable.
- [`s` on a Review Card no longer only explains] → it now opens a tab; the footer still says what starts.
- [One notification per unfinished settle may duplicate herdr's own done/blocked notification] → accepted; ours adds the Progress.
- [Plugin state lost (cache dir cleared)] → only resume from the Board is lost; the resume command from the Handoff text still works.
- [A resume map entry is verified from docs and `--help` only for `copilot`, `codex`, `opencode`, `cursor`] → smoke test with whichever is installed; a wrong entry only breaks resume for that kind.

## Migration Plan

Rebuild the plugin. Changes started before this change have no record and no watcher; their Review Cards show the old notice (or focus a running agent). Rollback: revert; leftover session files are ignored.

## Open Questions

- Does Claude Code (or another Main agent kind) reliably queue a herdr prompt submitted while the Main agent is `working`? The smoke test checks it; if not, wait for the Main agent to settle (bounded) before prompting.
- Does closing a worktree workspace's last pane close the workspace? (Smoke test; see Risks.)
