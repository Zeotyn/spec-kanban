## Purpose

Defines what happens after a Change started from the Board settles: the Handoff of a finished Change to the Main agent with its tab closed, a notification when it is not finished, and how its recorded session can be reopened.

## ADDED Requirements

### Requirement: A Start from the Board is watched until its agent goes away
After a Start from the Board has sent its prompt, the plugin SHALL watch the Change's agent in the background and evaluate it each time it settles (idle, done or blocked) after working. Watching SHALL end with a Handoff, or silently when the agent no longer exists (the agent exited, the tab or pane was closed, herdr stopped). A Start through `kanban start`, a Start that only focused a running agent, and a resumed session SHALL not be watched.

#### Scenario: Agent works through several turns
- **WHEN** the agent of `m1-board-view` settles with tasks left, the user answers it, and it later settles with all tasks completed
- **THEN** the first settle shows a notification, and the second one leads to a Handoff

#### Scenario: User closes the tab
- **WHEN** the user closes the agent tab of a watched Change before it is finished
- **THEN** watching ends without a notification and nothing else is closed or prompted

#### Scenario: Start from a shell
- **WHEN** the user runs `kanban start m1-board-view` and the agent later finishes all tasks
- **THEN** the tab stays open and no Main agent is prompted

### Requirement: Handoff of a finished Change
When the watched agent settles in idle or done and the Change is either open with at least one task, all of them completed, or archived, the plugin SHALL record the agent's session for the Change, close the agent's pane (which closes its tab when it is the only pane), and submit a prompt to the Main agent that names the Change, its Progress and asks to review it, and to archive it unless it is already archived. When the recorded session is reopenable, the prompt SHALL also say how to reopen it (`s` on its Card, or the agent's resume command in the Project). The Handoff SHALL work the same for every agent kind.

#### Scenario: Clean finish
- **WHEN** the Claude agent of `m1-board-view` settles idle and all 7 tasks of `m1-board-view` are completed, and the Board was opened from a pane running an agent
- **THEN** the agent's tab closes and the agent in that pane receives a prompt starting with `m1-board-view is finished (7/7 tasks). Please review it and archive it with /opsx:archive m1-board-view.` that names `claude --resume <session>`

#### Scenario: Agent archived the Change itself
- **WHEN** the agent of `m1-board-view` archives the Change and settles idle
- **THEN** the agent's tab closes and the Main agent receives a prompt starting with `m1-board-view is finished and archived. Please review it.`

#### Scenario: Clean finish of another reopenable kind
- **WHEN** the `codex` agent of `m1-board-view` finishes it cleanly with session `abc`
- **THEN** the Main agent's prompt names `codex resume abc`

#### Scenario: Clean finish of a kind without resume
- **WHEN** the configured agent kind is one without a known resume argument (for example `pi`) and its agent finishes `m1-board-view` cleanly
- **THEN** the tab closes and the Main agent receives the same prompt without a way to reopen the session

#### Scenario: Other panes in the agent tab
- **WHEN** the user split the agent tab and a Handoff happens
- **THEN** only the agent's pane closes and the user's other pane stays

### Requirement: An unfinished settle leaves the tab open
When the watched agent settles blocked, or with tasks of the Change not completed, the plugin SHALL leave its tab open, prompt no Main agent, show a herdr notification naming the Change and its Progress (and that it waits for input when blocked), and keep watching.

#### Scenario: Agent stops with tasks left
- **WHEN** the agent of `m1-board-view` settles idle with 3 of 7 tasks completed
- **THEN** the tab stays open and herdr shows a notification that `m1-board-view` stopped at 3/7 tasks

#### Scenario: Agent asks for approval
- **WHEN** the agent of `m1-board-view` settles blocked, even with all tasks completed
- **THEN** the tab stays open and herdr shows a notification that `m1-board-view` waits for input

### Requirement: Main agent falls back to a notification
The Main agent SHALL be the agent, of any kind herdr detects, in the pane the Board was opened from. When there is no such pane, it no longer exists, it is the finished agent's own pane, herdr detects no agent in it, or herdr rejects the prompt (for example because it is blocked), the plugin SHALL show the Handoff text as a herdr notification instead; the Handoff itself (session recorded, pane closed) SHALL still happen.

#### Scenario: Board opened from a shell pane
- **WHEN** the Board was opened from a pane running only a shell (no agent) and a started Change finishes cleanly
- **THEN** the agent's tab closes and herdr shows a notification `openspec-kanban: m1-board-view finished` with the Handoff text

#### Scenario: Main agent blocked
- **WHEN** a Change finishes cleanly while the Main agent waits at an approval dialog
- **THEN** no text is typed into the Main agent and herdr shows the Handoff text as a notification

### Requirement: Session recorded per Change
For every Change started from the Board, the plugin SHALL record the agent kind and the session herdr reports for its agent, once the agent is ready and again at every settle, so the newest session is kept. The record SHALL be kept per Project and Change in the plugin's state directory and survive closing the tab and restarting herdr. Losing the record SHALL only make the session not resumable from the Board.

#### Scenario: Tab closed by the user
- **WHEN** `m1-board-view` was started from the Board and the user closes its tab before it is finished
- **THEN** the session of `m1-board-view` is still recorded

#### Scenario: Equally named Change in two Projects
- **WHEN** Changes named `m1-board-view` in Projects `alpha` and `beta` were both started from the Board
- **THEN** each Project keeps its own recorded session

### Requirement: Resume a recorded session
Whether a recorded session is reopenable SHALL be decided by the agent kind herdr reported with the session: a session id is reopenable for `claude` (`--resume <id>`), `copilot` (`--resume=<id>`), `codex` (`resume <id>`), `opencode` (`--session <id>`), `cursor` (`--resume <id>`) and `gemini` (`--resume <id>`), the arguments following the agent's executable; every other kind, and a session reported as a path, SHALL be not reopenable. Resuming a Change SHALL open a new tab in the Project labelled with the Change name, focus it, and start the recorded agent kind in it as the Change's agent with its resume argument, without submitting a prompt. A failure after the Board closed SHALL be shown as a herdr notification `openspec-kanban: resume failed`.

#### Scenario: Resume a finished Change
- **WHEN** the Change `m1-board-view` has the recorded `claude` session `abc` and is resumed
- **THEN** a tab labelled `m1-board-view` opens in the Project and is focused, Claude runs in it on session `abc`, no prompt is sent, and the tab is not closed when the agent settles

#### Scenario: Resume a Copilot session
- **WHEN** the Change `m1-board-view` has the recorded `copilot` session `abc` and is resumed
- **THEN** Copilot is started in the new tab with `--resume=abc`

#### Scenario: Recorded session no longer exists
- **WHEN** the recorded session of `m1-board-view` was deleted from Claude's history and the Change is resumed
- **THEN** the tab stays open with Claude's own error, or herdr shows `openspec-kanban: resume failed` when Claude does not become ready
