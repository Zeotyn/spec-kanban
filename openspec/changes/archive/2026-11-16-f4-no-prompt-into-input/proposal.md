## Why

kanban submits prompts to the Main agent with `herdr agent prompt`, whatever the user is doing in that pane:
- Handoff, Stall and Escalation from m12;
- `kanban wave escalate` and `kanban wave finish` from m14.

When the user has half-typed input in the agent's input line, the prompt is typed into it and the two mix. This happened with the m12-escalation Handoff: it was appended to a half-typed `/opsx:propose wenn…`, and the result ran as a propose request.

A prompt into a pane the user merely looks at is harmless. The f5-draft-reason Handoff arrived while the Main pane was focused and busy, and Claude Code queued it cleanly. Waiting for focus to move would therefore delay every Handoff while the user sits in the Main pane, which is the common case. The risk is the typed input, not the focus.

## What Changes

- **No prompt into typed input.** Before kanban submits a prompt to a Main agent, it reads the pane's visible screen with `herdr pane read`.
  - For a `claude` agent it finds the input box: the lines between the last two rules of `─`, the first one starting with `❯`.
  - When the box holds no text, kanban submits at once.
  - When the box holds text, kanban checks again every 2 seconds and submits once the box is empty.
  - When the box is not found, or the agent is of another kind, kanban falls back to focus: it waits while the pane is `focused` (from `herdr agent get`), with the same 2-second checks.
  - After 10 minutes of waiting, kanban does not type. The prompt counts as not taken, and the existing fallback applies: the attention marker `⚑` for Handoff, Stall and Escalation, and the notifications `spec-kanban: wave <N> needs you` and `spec-kanban: wave <N> finished` for waves.
- **Applies to every prompt to a Main agent:** Handoff, Stall, Escalation, `wave escalate` and `wave finish`. It also covers the coordinator of a wave as its members' Main agent.
- **Subcommands stay quick.** `kanban escalate`, `kanban wave escalate` and `kanban wave finish` run inside an agent's turn and must not block it. When the prompt cannot go out at once, they hand the delivery to a detached process and print `Queued for the Main agent. End your turn.` Otherwise they print what they print today.
- The fixed values (2 seconds, 10 minutes) are not configurable.

Non-goals:
- no change to prompts into the Change's own pane (Start, Verify, Archive);
- no SendMessage between agents (the subject of m14-spike-agent-messaging);
- no input-box detection for other agent kinds beyond the focus fallback.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `agent-handoff`: a new requirement for when a prompt to a Main agent goes out, and the queued delivery for `kanban escalate`.
- `waves`: `wave escalate` and `wave finish` follow the same rule and queue their delivery the same way.

## Impact

- **Code:**
  - `internal/start`: `Tell` waits as above before `herdr agent prompt`, and reports "not taken" after 10 minutes. It gets a pure input-box parser.
  - `cmd/kanban`: a detached `tell` entry for the three subcommands, and the `Queued…` output.
- **Docs:** README (Handoff and Escalation timing), CHANGELOG.
- No new dependencies.
- m14-spike-agent-messaging still refers to this change by its old name `f4-no-prompt-into-focus`.
