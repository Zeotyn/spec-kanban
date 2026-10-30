## Why

After `s` the agent tab stays open forever, and nobody tells the agent the user plans with that the work is done: the user has to notice the Review Card, close the tab by hand and ask for a review. Once the tab is closed, the session that did the work is hard to find again.

## What Changes

- The detached `kanban launch` that a Start from the Board spawns stays alive after sending the prompt and watches the agent with `herdr agent wait`.
- When the agent settles and the Change is in Review (all tasks completed) or archived and the agent is not blocked, it is a **Handoff**: the session is recorded, the agent's pane (and with it the tab) is closed, and the **Main agent** (the agent in the pane the Board was opened from, of any kind) receives a prompt like `<change> is finished (7/7 tasks). Please review it and archive it with /opsx:archive <change>.` plus how to reopen the session when its kind supports that.
- The Handoff works for every agent kind the m5 config key `agent` allows; nothing in it is Claude-specific.
- When the agent settles in any other state (tasks left, blocked), the tab stays open, herdr shows a notification with the Progress, and watching continues until the next settle.
- When the Main agent cannot take the prompt (pane gone, no agent in it, blocked), the same text is shown as a herdr notification instead.
- The agent session herdr reports (agent kind and session id) for each Change started from the Board is recorded in the plugin's state directory.
- `s` on a Review Card with a reopenable recorded session reopens it in a new tab with the same agent kind (per kind from a small verified map: `claude`, `copilot`, `codex`, `opencode`, `cursor`, `gemini`; no prompt); without a recorded session it shows the notice as today; a session of another kind shows `resume not supported for <kind>`.
- A Change the agent archived itself counts as finished.
- Cut: resume arguments for kinds without a verified source, Handoff for `kanban start` from a shell, a configurable Handoff prompt, a time limit for watching (see design.md).

## Capabilities

### New Capabilities
- `agent-handoff`: watching a Started agent, the Handoff to the Main agent, the notification fallback, the recorded session per Change and resuming it.

### Modified Capabilities
- `board`: "Start a Change from the Board": `s` on a Review Card resumes the Change's recorded session instead of only showing a notice.

## Impact

- Builds on m5-help-and-config: its `agent` setting and Start kind defaulting to the invoking pane's agent, and its `KANBAN_MAIN_PANE` env.
- `cmd/kanban/main.go`: `kanban launch` watches after the prompt, using `KANBAN_MAIN_PANE` (introduced by m5-help-and-config); new internal `kanban resume`.
- `internal/start/`: watch loop, Handoff, resume launch, all through the fake `herdr.Runner`.
- `internal/cache/`: read and write the recorded session per Project and Change.
- `internal/board/`: `s` on Review.
- README: Handoff and resume.
