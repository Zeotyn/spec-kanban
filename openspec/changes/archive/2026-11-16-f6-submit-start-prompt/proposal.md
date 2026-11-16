## Why

The Start of f4-no-prompt-into-input sent its prompt, but the prompt was never submitted. It stood complete in the agent's input box, the agent stayed idle, and only a Stall 10 minutes later showed it. Sending one Enter to the pane fixed it. kanban reported success, so neither a notification nor `kanban failures` held anything. The prompt was most likely typed while Claude Code was still settling after it reported ready, and the Enter got lost.

A Start that silently does nothing costs 10 minutes and a manual look into the pane. kanban can see the case itself. f4 added a parser for Claude Code's input box, so kanban can tell whether the prompt is still sitting there.

## What Changes

- **Check that the prompt went in.** After a launch submits its first prompt to a new `claude` agent, kanban checks within a few seconds whether the prompt was taken:
  - the agent is working, or its input box is empty: done;
  - the box still holds the prompt while the agent is idle: kanban sends Enter to the pane and checks again, up to 3 times;
  - still not taken after that: the launch fails with `the prompt was typed but not submitted in tab <change>`, with the pane's last lines. After the Board closed, that shows as `spec-kanban: <step> failed` and lands in the failure log.
- **Applies to every launch of a new agent with a prompt:** Start from the Board and with `kanban start`, Verify and Archive in a new tab, a wave's coordinator and its members.
- **Other agent kinds:** unchanged, because kanban cannot read their input box.
- No change when the box cannot be found or holds other text, for example what the user typed into the new tab: the check counts the prompt as taken, as today, and sends no Enter.

Non-goals: no check for prompts to an agent that already runs (Verify or Archive on an idle agent, Handoff, Escalation); no retry of the whole launch.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `agent-start`: a new requirement that a launch makes sure its first prompt was submitted.

## Impact

- **Code:** `internal/start` (`LaunchUnfocused` checks after `herdr agent prompt`, reusing `inputBox` and `getAgent`; Enter through `herdr pane send-keys <pane> enter`). The callers report the new error as they report every launch error.
- **Docs:** README (troubleshooting), CHANGELOG.
- No new dependencies.
