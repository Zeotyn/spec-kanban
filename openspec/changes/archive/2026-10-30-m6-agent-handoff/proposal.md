## Why

After `s` the agent tab stays open forever, and nobody tells the agent the user plans with that the work is done: the user has to notice the Review Card, close the tab by hand, ask for a review and, since m6-worktree-start, merge the Change's branch and remove its worktree. Once the tab is closed, the session that did the work is hard to find again.

## What Changes

- The detached `kanban launch` that a Start from the Board spawns stays alive after sending the prompt and watches the agent with `herdr agent wait`.
- When the agent settles and the Change is in Review (all tasks completed) or archived and the agent is not blocked, it is a **Handoff**: the session is recorded, the agent's pane (and with it its tab, and a worktree workspace with nothing else open) is closed, and the **Main agent** (the agent in the pane the Board was opened from, of any kind) receives a prompt. For a Change worked on in its own worktree it reads like `m1-board-view is finished (7/7 tasks). Please review it, merge branch m1-board-view into main, archive it with /opsx:archive m1-board-view, then remove its worktree with git worktree remove <path>. Never use --force and keep the branch. If the merge does not go through cleanly, stop and leave the worktree.` Without a worktree it stays `… Please review it and archive it with /opsx:archive <change>.` Both add how to reopen the session when its kind supports that.
- The plugin itself never merges, removes a worktree or deletes a branch; the Main agent does merge and removal, and `git worktree remove` refuses a checkout with uncommitted work.
- The finished check reads the Progress from the Change's worktree (the m6-worktree-start loader) and counts a Change the agent archived, in the Project or inside its worktree, as finished.
- The Handoff works for every agent kind the config key `agent` allows; nothing in it is Claude-specific.
- When the agent settles in any other state (tasks left, blocked), the tab stays open, herdr shows a notification with the Progress, and watching continues until the next settle.
- When the Main agent cannot take the prompt (pane gone, no agent in it, blocked), the same text is shown as a herdr notification instead.
- The agent session herdr reports (agent kind and session id) for each Change started from the Board is recorded in the plugin's state directory.
- `s` on a Review Card with a reopenable recorded session reopens it in the Change's checkout as Start would (reused tab, reopened worktree workspace, or a new tab), with the same agent kind (per kind from a small verified map: `claude`, `copilot`, `codex`, `opencode`, `cursor`, `gemini`; no prompt); without a recorded session it shows the notice as today; a session of another kind shows `resume not supported for <kind>`.
- Cut: resume arguments for kinds without a verified source, Handoff for `kanban start` from a shell, a configurable Handoff prompt, a time limit for watching, merging or removing worktrees in the plugin (see design.md).

## Capabilities

### New Capabilities
- `agent-handoff`: watching a Started agent, the Handoff to the Main agent (with merge and worktree removal for a worktree Change), the notification fallback, the recorded session per Change and resuming it.

### Modified Capabilities
- `board`: "Start a Change from the Board": `s` on a Review Card focuses the running agent or resumes the Change's recorded session instead of only showing a notice.

## Impact

- Builds on m5-help-and-config (config `agent`, `start.Kind`, `KANBAN_MAIN_PANE` set by `kanban open`) and m6-worktree-start (`start.Prepare` with `Worktree`, `worktree.List`, `worktree.Base`, `openspec.Load` with checkouts, config `worktree`/`base`).
- `cmd/kanban/main.go`: `kanban launch` watches after `start.Launch`; new internal `kanban resume <change> <pane> <tab>`; `showBoard` wires the resume path.
- `internal/start/`: `Watch`, Handoff text, `ResumeArgs`/`ResumeCommand`, `Resume`, all through the fake `herdr.Runner`.
- `internal/openspec/`: a small helper that tells whether a directory's archive holds a Change, reused by `Load`.
- `internal/cache/`: read and write the recorded session per Project and Change.
- `internal/board/`: `s` on Review.
- README: Handoff, merge and worktree removal by the Main agent, resume.
