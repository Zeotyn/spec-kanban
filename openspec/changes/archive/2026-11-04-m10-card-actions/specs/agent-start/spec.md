## ADDED Requirements

### Requirement: Verify and Archive run the Source's command
Verify and Archive SHALL send the Source's verify or archive command for the Change (see "Verify and archive commands"), never the configured `prompt` or `instructions`. When a herdr agent of the Change exists in the Project or one of its checkouts and its agent runs in the foreground of its pane:
- with agent status `idle`, or `done` (its turn ended and nobody has looked yet), the command SHALL be sent to it as a prompt and its tab focused, without opening a tab;
- with any other status nothing SHALL be sent, and the caller reports the status.

Otherwise the command SHALL go to a new agent of the configured or detected kind in the Change's checkout, opened as for Start, reusing the Change's tab when its agent has exited. No watcher SHALL follow Verify or Archive: when the agent finishes, no Handoff prompt and no notification is sent. An agent of the Change's name outside the Project and its checkouts SHALL be refused as for Start.

#### Scenario: New agent for Verify
- **WHEN** no agent runs for Review Change `m1-board-view`, whose worktree exists, and Verify runs
- **THEN** a tab in that worktree is created, Claude starts in it and receives `/opsx:verify m1-board-view`, and nothing watches it

#### Scenario: Idle agent receives Archive
- **WHEN** the agent `m1-board-view` is idle in its tab and Archive runs
- **THEN** no tab is created, the agent receives `/opsx:archive m1-board-view` and its tab is focused

#### Scenario: Blocked agent
- **WHEN** the agent `m1-board-view` waits for a permission answer (status `blocked`) and Verify runs
- **THEN** nothing is sent and the status `blocked` is reported

#### Scenario: Instructions are not appended
- **WHEN** the config sets `instructions` to `with ponytail` and Verify runs for `m1-board-view`
- **THEN** the agent receives exactly `/opsx:verify m1-board-view`
