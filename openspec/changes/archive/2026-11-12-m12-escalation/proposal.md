## Why

The user should rarely look at a worker agent's pane. They talk to the Main agent and glance at the Board and the sidebar. Today a worker that cannot go on has no way to say so. Example: a worker on a release-gate Change stopped at 3/7 tasks, because its prompt forbids the two steps left (npm publish, tag). The Handoff only fires when all tasks are done or the Change is archived. So the user only got "stopped" notifications and had to read the pane to learn why.

## What Changes

- **Escalation.** A new subcommand `kanban escalate <change> <reason>`. A worker runs it when it cannot continue without a human.
  - It prompts the Main agent at once, not at the next settle.
  - The prompt names the Change, its Progress, the reason, the agent's tab and worktree, and the tasks file. It tells the Main agent to do the step itself when it may, or to tell the user exactly what to do, and then to let the worker continue with `herdr agent prompt`.
  - The worker's pane stays open. The normal Handoff follows once all tasks are done.
  - The Main pane comes from a record that `kanban launch` and `kanban resume` write per Project and Change in the user cache directory (`spec-kanban` below it), which the worker's pane can reach without `$HERDR_PLUGIN_STATE_DIR`.
  - Without a Main agent, or when herdr rejects the prompt, nothing is typed and herdr shows a notification instead. m13-attention replaces this notification with the attention marker.
- **Start prompt.** A Start from the Board appends a fixed sentence to the prompt, after the Source's command and `instructions`, and also after a configured `prompt`: `If you cannot continue without a human, run <kanban> escalate <change> "<what you need>" and end your turn.` `<kanban>` is the absolute path of the running binary. `kanban start` from a shell, Resume (no prompt) and Verify/Archive get no sentence.
- **Stall.** When a watched agent settles idle or done with the Change unfinished, without having escalated, and does not work again within 10 minutes, the watcher prompts the Main agent once for that settle. The prompt says the worker stopped without asking for help and how to read its output and let it continue. The 10 minutes are fixed.
- **BREAKING (behaviour):** The "stopped at x/y tasks" and "waits for input" notifications at an unfinished or blocked settle are removed. Stall replaces the first; herdr itself shows a blocked agent. Notifications that stay: start, run and resume failures, and the fallbacks when no Main agent takes a prompt.
- **Resume is watched.** A resume from the Board is now watched like a Start: Handoff when finished, Stall, Escalation. The watcher evaluates a resumed agent only after it worked once, so reopening a finished session does not hand it off at once. Verify/Archive (`kanban run`) and `kanban start` from a shell stay unwatched.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `agent-handoff`: watching now covers a resume from the Board; an unfinished or blocked settle no longer notifies; new requirements for Escalation, Stall and the Main pane record; the notification fallback covers Escalation and Stall too.
- `agent-start`: a Start from the Board appends the escalation sentence to its prompt.
- `board`: the "Start a Ready Card" scenario names the escalation sentence; the requirement text stays.

## Impact

- **Code:**
  - `internal/cache`: Main pane record and escalation record per Project and Change.
  - `internal/start`: the escalation sentence, the Escalation prompt, Stall in `Watch`, a resumed watch that waits for the first turn, no notification at unfinished or blocked settles.
  - `cmd/kanban`: `kanban escalate`, the sentence in `startOne`, the Main pane record in `launch` and `resumeChange`, `Watch` after a resume, usage text.
- **Docs:** README (Handoff, Escalation, Stall), CONTEXT.md (Escalation, Stall), CHANGELOG.
- No new dependencies.
