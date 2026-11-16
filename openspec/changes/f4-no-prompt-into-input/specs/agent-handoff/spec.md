## ADDED Requirements

### Requirement: A prompt to the Main agent waits for an empty input line
Before kanban submits a prompt to a Main agent (Handoff, Stall, Escalation, and the wave prompts of the waves spec), it SHALL check whether the user has typed into that agent's input line:
- For a `claude` agent it SHALL read the pane's visible screen and find the input box: the lines between the last two lines made only of `─`, the first of them starting with `❯`. The box SHALL count as empty when, after `❯`, it holds only spaces or the dimmed placeholder Claude Code shows in an empty box.
- When the box is empty, kanban SHALL submit the prompt at once, whether the pane is focused or not.
- When the box holds text, kanban SHALL check again every 2 seconds and submit the prompt once the box is empty.
- When the box cannot be found, the screen cannot be read, or the agent is of another kind, kanban SHALL instead wait while herdr reports the pane as focused, with the same 2-second checks, and submit once it is not focused.
- When it has waited for 10 minutes, kanban SHALL not type into the pane. The prompt SHALL count as not taken, so the fallback for a missing Main agent applies.

The checks SHALL never send keys to the pane. A prompt into the Change's own pane (Start, Verify, Archive) SHALL not wait.

#### Scenario: User sits in the Main pane without typing
- **WHEN** a Change finishes while the Main pane is focused, its Claude agent is working, and its input box is empty
- **THEN** the Handoff prompt is submitted at once

#### Scenario: User is typing
- **WHEN** a Change finishes while the Main pane's input box holds `/opsx:propose wenn`
- **THEN** nothing is typed into the pane; once the user submits or clears the input and the box is empty, the Handoff prompt is submitted within about 2 seconds

#### Scenario: Input left for 10 minutes
- **WHEN** the Main pane's input box holds text for 10 minutes after a Change finished
- **THEN** nothing is typed into the pane, the Change's attention marker is set as for a Handoff no Main agent took, and the input stays as the user left it

#### Scenario: Other agent kind
- **WHEN** the Main agent is a `codex` agent and its pane is focused when a Stall is due
- **THEN** the Stall prompt is submitted once the pane is no longer focused, or the attention marker is set after 10 minutes

### Requirement: Subcommands queue a waiting prompt
`kanban escalate`, `kanban wave escalate` and `kanban wave finish` SHALL not wait for an empty input line themselves. When their prompt to the Main agent cannot be submitted at once, they SHALL hand its delivery to a detached process, which waits and falls back as above, and SHALL print `Queued for the Main agent. End your turn.` in place of what they print when the Main agent received it. Everything else they record (the attention marker, a finished wave) SHALL be recorded before they return, as without waiting.

#### Scenario: Escalation while the user types
- **WHEN** the agent of `m1-board-view` runs `kanban escalate m1-board-view needs a tag` while the Main pane's input box holds text
- **THEN** the command returns at once, prints `Queued for the Main agent. End your turn.`, the attention marker of `m1-board-view` is set, and the Main agent receives the Escalation prompt once its input box is empty
