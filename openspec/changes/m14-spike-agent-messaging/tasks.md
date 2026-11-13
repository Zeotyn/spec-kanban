## 1. Setup

- [ ] 1.1 In a throwaway test Project, open three Claude Code sessions as Main (Project root), coordinator (worktree `wave-1`) and implementer (worktree named like a Change), and record their peer names from `ListAgents`

## 2. Questions

- [ ] 2.1 Busy recipient: send a message from the implementer to the coordinator while the coordinator works on a long turn; record whether it is queued or lost, when and how it arrives (new turn, notification, appended to input)
- [ ] 2.2 Disturbance: send a message to the implementer mid-task; record whether it interrupts or changes its current work, and whether the user's half-typed input in the receiving terminal stays intact
- [ ] 2.3 Addressing: find Main, coordinator and implementer only by peer-name prefix (Project directory, `wave-<N>`, Change name); repeat with two sessions in the same directory and record how they can be told apart
- [ ] 2.4 Handoff record: prototype `kanban handoff <change>` printing the Handoff text and writing a record; check that the watcher can read it and skip its own `herdr agent prompt`, including when the agent never sends
- [ ] 2.5 Project boundary: check whether `ListAgents` or `SendMessage` can be limited to the Project, or whether the boundary can only be stated in the prompt

## 3. Outcome

- [ ] 3.1 Write ADR `docs/adr/0010-agent-messaging.md` with status `proposed`: findings per question, the chosen channel per message (Handoff, Escalation, answers, Stall), and the fallback for non-Claude kinds
- [ ] 3.2 Add a recommendation to the ADR: whether m14-waves builds on `SendMessage`, on `herdr agent prompt`, or both with a fallback, and which follow-up Change implements it
- [ ] 3.3 Keep throwaway test code on the spike branch; nothing but the ADR goes to `main`
