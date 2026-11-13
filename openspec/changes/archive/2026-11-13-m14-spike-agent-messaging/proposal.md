## Why

kanban reaches agents by typing a prompt into their terminal with `herdr agent prompt`. For the Main agent that is the terminal the user is writing in, so a Handoff or Escalation can land in the middle of the user's own input. Claude Code agents can also message each other directly with `SendMessage` and find each other with `ListAgents`. That is convenient and leaves the user's input line alone. m14-waves built its Main → coordinator → member hierarchy on `herdr agent prompt`: Handoff, Stall and Escalation of a member go to the coordinator, `kanban wave escalate` and `kanban wave finish` go to the Main agent. This spike finds out whether agent-to-agent messages are reliable enough to carry these messages instead, or next to the prompt.

## What Changes

Spike only: no production code and no spec deltas. Throwaway test code stays on the spike branch.

- **Scope kept:** the hierarchy Main → coordinators → implementers stays, and agents talk only within their Project.
- **Known limits:**
  - kanban's Go code cannot send such messages, because the `claude` CLI has no send command.
  - kanban does not know peer names. herdr's session id (e.g. `1feaf6c1-…`) does not map to a peer name such as `openspec-kanban-7d [cb1ab0]`. Peer names derive from the working directory, for example `c1-hosting-e6` or `wave-1-…`.
- **Idea to evaluate:**
  - Agents send Handoff and Escalation themselves with `SendMessage`, using text that `kanban handoff <change>` or `kanban escalate` prints.
  - kanban records that the message was sent.
  - The watcher falls back to `herdr agent prompt` when nothing was sent, for example for non-Claude agent kinds or when the agent forgot.
- **Questions:**
  1. Is a message to a busy recipient queued or lost, and how does it arrive?
  2. Does an incoming message disturb an implementer while it works?
  3. Can Main, coordinator and implementer be found reliably by peer-name prefix, and what happens with two sessions in the same directory?
  4. Can the watcher reliably tell "already handed off" from a record written by a kanban subcommand?
  5. Can the Project boundary be enforced, or only stated in the prompt?
- **Outcome:**
  - an ADR in `docs/adr/` with status `proposed`;
  - a recommendation per message (Handoff, Stall, Escalation, wave escalate, wave finish) whether it stays on `herdr agent prompt`, moves to `SendMessage`, or uses both with a fallback;
  - how this relates to f4-no-prompt-into-focus, which keeps `herdr agent prompt` but waits while the Main pane is focused;
  - the follow-up Change that implements the recommendation, if any.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

None (spike, `skip_specs: true`).

## Impact

- **Docs:** one new ADR in `docs/adr/` (next number after 0009).
- **Code:** none on `main`. Test sessions and scripts stay on the spike branch.
- Informs a follow-up to m12-escalation and m14-waves (the Handoff, Escalation and wave paths) and the scope of f4-no-prompt-into-focus.
