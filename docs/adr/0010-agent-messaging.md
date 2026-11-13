---
status: proposed
---

# Agent messaging

kanban reaches agents only by typing a prompt into their pane with `herdr agent prompt`. Claude Code sessions can also find each other with `ListAgents` and message each other with `SendMessage`. The spike m14-spike-agent-messaging tested whether Handoff, Stall, Escalation, `kanban wave escalate` and `kanban wave finish` should move to `SendMessage`. Tested with Claude Code 2.1.323 (model Sonnet) and herdr in a throwaway Project with a Main agent, a coordinator in worktree `wave-1` and a member in worktree `c1-demo`. The raw log is in `spike/m14-agent-messaging/log.md` on branch `m14-spike-agent-messaging`.

**Decision:** all five messages stay on `herdr agent prompt`. f4-no-prompt-into-focus is still needed in full. It handles the common case of the problem, a prompt typed into the user's half-written input while the user works in the Main pane. Only `SendMessage` leaves the input alone in every case, including text left half-typed in a pane that is not focused. We accept that gap for a channel that works for every agent kind and uses only documented interfaces. No follow-up Change moves a message to `SendMessage` now.

## Findings

1. **Busy recipient.** In both runs the message was queued and arrived. The `SendMessage` result warns that a recipient may hold or refuse a message, "usually [in] a different permission mode", which was not tested. An idle recipient starts a new turn by itself. A busy one sees the message at its next tool result, inside the running turn. Its input line is not used.
2. **Disturbance.** A message does not interrupt the running tool call or end the task, but the agent follows requests in it within the same turn. Text half-typed into the recipient's pane (`herdr pane send-text`) stayed intact and unsent, both when the agent was busy and when it was idle. A message to an idle agent makes it run a turn, which herdr reports like any other turn.
3. **Addressing.** A peer name derived by Claude Code is `<basename of cwd>-<2 hex>` (`msgspike-b1`, `wave-1-cb`, `c1-demo-25`) plus a ref like `[351b05]`. A prefix never delivers: a unique prefix returns the full name to resend with, and two sessions in one directory (`c1-demo-25`, `c1-demo-f8`) can only be told apart by suffix and ref. `claude --name <name>` sets the peer name exactly (`msgspike/c1-demo` works), and `claude --resume` keeps it. Duplicate names are allowed, but a send to them is refused as ambiguous, never misdelivered. kanban does not start the Main agent, so it cannot name it. It can only look up the derived name: herdr's `agent_session` id maps to it in `~/.claude/sessions/<pid>.json`, a file Claude Code does not document.
4. **Handoff record.** A record that `kanban handoff` writes only proves that the text was printed, not that it was sent. The prototype makes the watcher look in the agent's transcript (`~/.claude/projects/*/<session id>.jsonl`) for a successful `SendMessage` that carries the record's token. It works: when the agent sent, the watcher typed nothing; when the agent printed but did not send, the watcher fell back to `herdr agent prompt`. The transcript format is not documented either. When the Handoff arrived as a peer message, the Main agent did not act. It said the message "arrived from another session, and you didn't ask for any of it" and waited for the user. The same text typed as a prompt was acted on at once. This was one run, and that Main agent had only run addressing tests; a Main agent that had started the Change may act on the message.
5. **Project boundary.** `ListAgents` lists every Claude Code session on the machine, and any session can message any other. A permission rule like `SendMessage(c1-demo*)` has no effect, and `--disallowedTools SendMessage` removes the tool entirely. A `PreToolUse` hook passed with `--settings` can refuse sends whose `to` lacks the Project prefix. That only covers sessions kanban starts, and only with `--name <project>/<change>`. For the Main agent the boundary can only be stated in the prompt.

## Channel per message

A member of a wave sends its Handoff, Stall and Escalation to the wave's coordinator; every other Change sends them to the Main agent.

| Message | Recipient | Channel | Why |
|---|---|---|---|
| Handoff | Main agent or coordinator | `herdr agent prompt`, with f4's focus wait | The watcher writes it at a settle, and kanban's Go code cannot send. Having the agent send it needs two undocumented files and the agent's help, and the recipient may not act on a peer message. |
| Stall | Main agent or coordinator | `herdr agent prompt`, with f4's focus wait | Only the watcher knows of a Stall. The stalled agent cannot report it. |
| Escalation | Main agent | `herdr agent prompt`, with f4's focus wait | The Change's agent could send it, but only to a Main peer name it would have to look up in an undocumented file. f4 keeps the prompt out of the input while the user works in the Main pane. |
| Escalation | coordinator | `herdr agent prompt` | kanban starts the coordinator, so it could name it exactly (`--name <project>/wave-<N>`). But the user is not expected to type in the coordinator's pane (an assumption, not tested), so `SendMessage` would gain nothing there, and Handoff and Stall to the same coordinator stay on `herdr agent prompt` anyway. One channel per recipient is simpler. |
| `kanban wave escalate` | Main agent | `herdr agent prompt`, with f4's focus wait | The same as Escalation to the Main agent, sent by the coordinator. |
| `kanban wave finish` | Main agent | `herdr agent prompt`, with f4's focus wait | A Handoff of the wave. It asks for a merge and a worktree removal, which a peer message may not get done. |

Only the Handoff to the Main agent was run end to end. The reasons for the other rows follow from findings 1 to 5, not from runs of their own.

**Fallback for non-Claude agent kinds:** none needed, since every message stays on `herdr agent prompt`, which works for every kind. Had a message moved to `SendMessage`, every non-Claude kind and every Claude agent that did not send would have needed the `herdr agent prompt` path anyway, as the prototype's fallback shows.

## Recommendation

- **f4-no-prompt-into-focus:** still needed, in full. Its focus wait is the cheapest fix for the common case of half-typed input, and it works for every agent kind. The spike found nothing that makes it smaller.
- **Follow-up Change:** none for now. Reconsider when Claude Code documents a way for an outside process to send a message or to look up a session's peer name. Then the watcher could send Handoff and Stall itself, without the agent's help. Starting agents with `--name <project>/<change>` and a `PreToolUse` hook would make addressing exact and the Project boundary enforceable.

## Considered Options

- **The agent sends Handoff and Escalation with `SendMessage`, the watcher checks the transcript and falls back to `herdr agent prompt` (the prototype, using both channels with a fallback):** it works and never touches the user's input. It reads two undocumented files, depends on the agent to send, still needs the fallback for other kinds, and lands as a message the Main agent may not act on.
- **kanban talks to Claude Code's messaging socket (`messagingSocketPath` in `~/.claude/sessions/<pid>.json`) directly:** this would let the watcher send. Not tried, because the protocol is undocumented and was not examined, and every session file has a `<pid>.<hash>.key` file next to it.

## Consequences

Nothing changes on `main` but this ADR. The prototype (`kanban handoff`, `kanban spike-watch`, `Watched.Sent`) stays on the spike branch. Claude Code may change these findings in a later version; the test setup in the log can be repeated.
