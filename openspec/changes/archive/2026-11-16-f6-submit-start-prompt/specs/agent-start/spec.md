## ADDED Requirements

### Requirement: A launch makes sure its first prompt was submitted
After a launch submits the first prompt to a new `claude` agent (Start, Verify or Archive in a new tab, a wave's coordinator or member), it SHALL check, within 5 seconds and without typing anything, whether the prompt was taken: the agent is working, or its input box (as found for the waiting of prompts to the Main agent) is empty. When the agent is idle and the box still holds the prompt (its text starts with the prompt's first 20 characters), the launch SHALL send Enter to the agent's pane and check again after 2 seconds, at most 3 times. When the prompt is still not taken then, the launch SHALL fail with `the prompt was typed but not submitted in tab <change>` and the pane's last lines, reported like every other launch failure. When the input box cannot be found, the screen cannot be read, or the box holds other text (the user typed into the new tab), the prompt SHALL count as taken and no Enter SHALL be sent. For other agent kinds the launch SHALL not check.

#### Scenario: Enter got lost
- **WHEN** a Start from the Board launches Claude for `f4-no-prompt-into-input` and, after the prompt was sent, the agent is idle with the whole prompt still in its input box
- **THEN** kanban sends Enter to the pane, the agent starts working on the prompt, and nothing is reported

#### Scenario: Prompt taken at once
- **WHEN** the agent is working within 5 seconds after the prompt was sent
- **THEN** no Enter is sent

#### Scenario: User typed into the new tab
- **WHEN** the agent is idle and its input box holds text that does not start with the prompt
- **THEN** no Enter is sent and the launch succeeds

#### Scenario: Still stuck
- **WHEN** the input box still holds the prompt after the third Enter
- **THEN** herdr shows `spec-kanban: start failed` with `the prompt was typed but not submitted in tab f4-no-prompt-into-input` and the pane's last lines, and the failure log holds it

#### Scenario: Other kind
- **WHEN** a Start launches a `codex` agent
- **THEN** no check runs after the prompt and no Enter is sent
