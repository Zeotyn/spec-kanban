## ADDED Requirements

### Requirement: An agent is started only at the shell prompt
Before the plugin starts an agent in a pane (Start, Resume, Verify and Archive alike, in a new tab or in the Change's reused tab), it SHALL wait until the pane's shell is at its interactive prompt. herdr reports no such state, so the plugin SHALL probe: it types into the pane, as one line, a space followed by a command that prints a marker which does not appear in the typed line itself (`echo kanban-rea""dy-<nonce>` prints `kanban-ready-<nonce>`), and waits until the pane shows that marker. A `<nonce>` SHALL be unique per probe, so that neither the typed echo nor an earlier probe's output in the same pane can satisfy the wait. A probe that is not answered within 2 seconds SHALL be followed by a new probe with a fresh nonce; the plugin SHALL give up after 60 seconds in all, and SHALL then start no agent and report the failure as in "Start reports herdr failures". A pane that cannot take the probe (herdr refuses it, for example because the pane is gone) SHALL end the wait at once with that error. Only after the marker appeared SHALL the agent command be typed.

#### Scenario: Shell rc asks a question
- **WHEN** the user's `.zshrc` ends in a prompt that reads one key, like oh-my-zsh's `[Y/n]` update question, and the user presses `s` on Card `m1-board-view`
- **THEN** the first probe's space answers that question, the shell runs the rest of the probe at its prompt and prints `kanban-ready-<nonce>`, then Claude is started and receives `/opsx:apply m1-board-view`; `laude` is never typed

#### Scenario: Shell rc reads a whole line
- **WHEN** the user's rc reads a whole line before continuing, and the user presses `s` on Card `m1-board-view`
- **THEN** the first probe is swallowed as that line, a second probe with another nonce is typed after 2 seconds and answered at the prompt, and Claude is started once

#### Scenario: Slow rc without a dialog
- **WHEN** the user's rc takes 3 seconds (nvm, conda) and the user presses `s`
- **THEN** the probes typed meanwhile run as soon as the prompt appears, the wait ends with the first marker the pane shows, and Claude is started

#### Scenario: Shell already at its prompt
- **WHEN** the agent of `m1-board-view` has exited, its shell is at its prompt, and the user starts the Change again
- **THEN** the first probe is answered at once and Claude starts in that tab without a noticeable delay

#### Scenario: Earlier marker in the same tab
- **WHEN** the tab of `m1-board-view` still shows `kanban-ready-a1b2c3` from an earlier Start, and the Change is started again
- **THEN** the new probe uses another nonce and only its own marker ends the wait

#### Scenario: Resume waits too
- **WHEN** the Change `m1-board-view` with the recorded `claude` session `abc` is resumed in a new tab
- **THEN** the probe runs before `claude --resume abc` is typed

#### Scenario: Shell never reaches its prompt
- **WHEN** the shell of the new tab shows a wizard that takes the probes as input and never reaches a prompt, and the user presses `s`
- **THEN** after 60 seconds no agent is started, and the failure says the shell did not reach its prompt in the Change's tab and shows the pane's last lines

#### Scenario: Pane gone during the wait
- **WHEN** the user closes the new tab while the plugin is still probing
- **THEN** the wait ends at once with herdr's error and no further probe is typed

## MODIFIED Requirements

### Requirement: Start reports herdr failures
When any herdr step fails, Start SHALL stop, report which step failed with herdr's error, and leave no agent half-started without saying so. When the shell did not reach its prompt or the agent did not become ready, the report SHALL also show the last lines of the pane (up to 8 non-empty lines of recent output, without escape sequences), so the user sees what the shell was doing; when the pane cannot be read, the report says so instead. A failure in the part that runs after the Board closed SHALL be reported as a herdr notification.

#### Scenario: Claude does not become ready
- **WHEN** the agent does not reach readiness within herdr's start timeout
- **THEN** the command exits non-zero, names the tab it created, says the prompt was not sent, and shows the pane's last lines

#### Scenario: Claude does not become ready after the Board closed
- **WHEN** a Start from the Board has closed the Board and the agent does not reach readiness
- **THEN** herdr shows the notification `spec-kanban: start failed` with that error and the pane's last lines

#### Scenario: Shell did not reach its prompt after the Board closed
- **WHEN** a Start from the Board has closed the Board and the shell of the new tab answers no probe within 60 seconds
- **THEN** herdr shows the notification `spec-kanban: start failed` saying the shell did not reach its prompt in tab `m1-board-view`, with the pane's last lines

#### Scenario: Pane cannot be read for the report
- **WHEN** the agent does not become ready and `pane read` fails as well
- **THEN** the report carries the start error and says the pane could not be read, with that error
