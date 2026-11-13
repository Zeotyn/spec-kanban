## 1. Setup

- [ ] 1.1 In a throwaway test Project under the scratch directory, open three Claude Code sessions with `herdr` as Main (Project root), coordinator (worktree `wave-1`) and member (worktree named like a Change), and record their peer names from `ListAgents`

## 2. Questions

- [ ] 2.1 Busy recipient: send a message from the member to the coordinator while the coordinator works on a long turn; record whether it is queued or lost, and when and how it arrives (new turn, notification, appended to input)
- [ ] 2.2 Disturbance: send a message to the member mid-task; record whether it interrupts or changes its current work. Type text into the receiving pane with `herdr pane send-text` (without Enter) before the message arrives, and record whether that half-typed input stays intact
- [ ] 2.3 Addressing: find Main, coordinator and member only by peer-name prefix (Project directory, `wave-<N>`, Change name); repeat with two sessions in the same directory and record how they can be told apart
- [ ] 2.4 Handoff record: prototype `kanban handoff <change>` printing the Handoff text and writing a record; check that the watcher can read it and skip its own `herdr agent prompt`, including when the agent never sends
- [ ] 2.5 Project boundary: check whether `ListAgents` or `SendMessage` can be limited to the Project, or whether the boundary can only be stated in the prompt

## 3. Outcome

- [ ] 3.1 Write ADR `docs/adr/0010-agent-messaging.md` with status `proposed`: findings per question, the chosen channel per message (Handoff, Stall, Escalation, wave escalate, wave finish), and the fallback for non-Claude kinds
- [ ] 3.2 Add a recommendation to the ADR: which follow-up Change implements it, and whether f4-no-prompt-into-focus is still needed, needed in a smaller form, or replaced
- [ ] 3.3 Keep throwaway test code on the spike branch; nothing but the ADR goes to `main`, and remove the test Project's worktrees and tabs without `--force`

## Smoke test (user, in herdr)

These have no checkboxes: they do not count as tasks, and only the user runs them.

- While typing in the Main pane, let one test session send a message to it: the typed text stays intact and the message arrives after Enter or as a new turn
