## ADDED Requirements

### Requirement: Wave prompts wait for an empty input line
The prompts of `kanban wave escalate` and `kanban wave finish` to the wave's Main agent, and every Handoff, Stall and Escalation of a member to its coordinator, SHALL wait for an empty input line as the agent-handoff spec requires ("A prompt to the Main agent waits for an empty input line"). `wave escalate` and `wave finish` SHALL queue a waiting prompt as `kanban escalate` does ("Subcommands queue a waiting prompt"). When the wait ends after 10 minutes, `wave escalate` SHALL fall back to the notification `spec-kanban: wave <N> needs you` and `wave finish` to `spec-kanban: wave <N> finished`. `wave finish` SHALL mark the wave finished before it returns, whether its prompt went out or was queued.

#### Scenario: Wave finishes while the user types
- **WHEN** the coordinator of wave 1 runs `kanban wave finish 1` while the Main pane's input box holds text
- **THEN** the command prints `Queued for the Main agent. End your turn.`, wave 1 is finished and its badges are gone, and the Main agent receives the finish prompt once its input box is empty
