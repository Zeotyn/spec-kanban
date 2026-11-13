## Context

kanban reaches agents only through `herdr agent prompt` (Handoff, Stall, Escalation, `kanban wave escalate`, `kanban wave finish`). Claude Code agents can also use `ListAgents` and `SendMessage`. kanban's Go code cannot send such messages itself. See proposal.md for the questions.

## Goals / Non-Goals

**Goals:**
- Answer the five questions with observed behaviour, not documentation alone.
- Recommend a channel per message, with a fallback for non-Claude agent kinds.

**Non-Goals:**
- No production code, no spec deltas, no change to the watcher on `main`.
- No messaging across Projects.

## Decisions

### D1: Test in a throwaway Project driven by herdr
All sessions run in a test Project under the scratch directory, opened and driven with `herdr` commands (`tab create`, `pane run`, `pane send-text`, `agent read`). This keeps real Projects untouched and makes every step repeatable without a human.
Alternative: test in this repository. That lost because stray messages could reach real agents.

### D2: Record raw observations first, then decide
Each question gets a short log of what was sent, when, and what the receiver showed (from `herdr agent read`). The ADR cites these logs. The recommendation follows from them, not from expectations.

### D3: The prototype stays on the spike branch
`kanban handoff <change>` (task 2.4) is a throwaway prototype. Only the ADR is merged into `main`. A follow-up Change implements the chosen design with specs.

## Risks / Trade-offs

- [SendMessage behaviour depends on the Claude Code version] → The ADR names the version tested with.
- [Peer names change with directory names] → Task 2.3 records the naming rule, and the ADR states how stable it is.
- [The spike ends without a clear winner] → The ADR then recommends keeping `herdr agent prompt` together with f4-no-prompt-into-focus.
