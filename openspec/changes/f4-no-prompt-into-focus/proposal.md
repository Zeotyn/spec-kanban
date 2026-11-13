## Why

kanban submits Handoff, Stall and Escalation prompts with `herdr agent prompt` into the Main agent's pane, whatever the user is doing there. When the user is typing in that pane, the prompt is typed into their half-written input and mixes with it. This happened with the m12-escalation Handoff: it was appended to a half-typed `/opsx:propose wenn…`, and the result ran as a propose request. The Main agent's pane is exactly where the user works most, so this will keep happening.

## What Changes

- **No prompt into a focused pane.** Before kanban submits a prompt to the Main agent, it checks `herdr agent get <pane>`.
  - While the pane is `focused`, kanban polls every 2 seconds and submits the prompt once the pane is no longer focused.
  - If the pane is still focused after 10 minutes, kanban does not type into it. It shows the text as a herdr notification instead, the same fallback as when no Main agent takes the prompt.
  - Why the notification and not "prompt anyway": a user focused on the pane for 10 minutes is most likely still typing there, so typing into it would cause exactly the mix this change prevents. m13-attention later replaces this notification with the attention marker.
- **Applies to every prompt to the Main agent:** the Handoff, the Stall prompt and the Escalation prompt, for every agent kind.
- **Escalation stays quick for the worker.** `kanban escalate` runs inside the worker's turn and must not block it for minutes. It returns at once and hands the delivery to a detached process. It prints `Queued for the Main agent. End your turn.` when it waits for focus to move, and as today when the prompt went through at once.
- The fixed values (2 seconds, 10 minutes) are not configurable.

Non-goals: no check of the Main pane's input line content, no change to prompts sent to a worker's own pane (Start, Verify, Archive), no SendMessage between agents (that is the subject of a spike).

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `agent-handoff`:
  - "Handoff of a finished Change": the prompt waits while the Main pane is focused.
  - "Main agent falls back to a notification": also falls back when the Main pane stays focused for 10 minutes.
  - "Escalation": waits for focus to move, delivered detached, new `Queued…` output.
  - "Stall": waits for focus to move.

## Impact

- **Code:**
  - `internal/start`: `Tell` waits while the Main pane is focused (poll 2 s, limit 10 min, then notification). The watcher's Handoff and Stall use it unchanged.
  - `cmd/kanban`: `escalate` delivers through a detached process; new output line.
  - Tests with a fake herdr: focused then unfocused, focused for the whole limit, not focused at all.
- **Docs:** README (Handoff and Escalation timing), CHANGELOG.
- **Builds on m12-escalation** (`Tell`, Stall, Escalation exist only there): implement it after m12 is merged into `main`, and archive it after m12.
- No new dependencies.
