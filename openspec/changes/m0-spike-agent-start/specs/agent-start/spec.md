## Purpose

Defines how a Change is handed to a new coding-agent session in its own herdr tab, so implementation can begin from the Board with a single action.

## ADDED Requirements

### Requirement: Start opens an agent session for a Change
Starting a Change SHALL create a new herdr tab whose working directory is the Project and whose label is the Change name, start Claude in that tab as an agent that herdr tracks, submit `/opsx:apply <change>` to it, and focus the tab.

#### Scenario: Start a Ready Change
- **WHEN** the user runs `kanban start m1-board-view` inside a Project that contains that open Change
- **THEN** a new tab labelled `m1-board-view` opens in the Project directory, Claude runs in it, its first prompt is `/opsx:apply m1-board-view`, and the tab is focused

#### Scenario: herdr tracks the session
- **WHEN** a Change has been started
- **THEN** `herdr agent list` shows an agent for that Change, so its status appears in the herdr sidebar

### Requirement: Start does not duplicate a running session
When an agent for the Change is already running, Start SHALL focus that agent instead of creating another tab or agent.

#### Scenario: Start twice
- **WHEN** the user starts `m1-board-view` and then starts it again while the first agent is still running
- **THEN** no second tab is created and the existing agent's tab is focused

#### Scenario: Start after the agent exited
- **WHEN** the agent for `m1-board-view` has exited but its tab is still open, and the user starts the Change again
- **THEN** no second tab is created; Claude is started again in that tab, receives `/opsx:apply m1-board-view`, and the tab is focused

### Requirement: Start rejects unknown Changes
Start SHALL check that the Change exists as an open Change of the Project before calling herdr, and SHALL exit with a non-zero status and a message naming the Change otherwise.

#### Scenario: Typo in the Change name
- **WHEN** the user runs `kanban start m1-bord-view` and no such open Change exists
- **THEN** no tab is created and the command prints that `m1-bord-view` is not an open Change of the Project

#### Scenario: Archived Change
- **WHEN** the user runs `kanban start` with the name of a Change that exists only in the archive
- **THEN** no tab is created and the command prints that the Change is not open

### Requirement: Start reports herdr failures
When any herdr step fails, Start SHALL stop, print which step failed with herdr's error, and leave no agent half-started without saying so.

#### Scenario: Claude does not become ready
- **WHEN** the agent does not reach readiness within herdr's start timeout
- **THEN** the command exits non-zero, names the tab it created, and says the prompt was not sent
