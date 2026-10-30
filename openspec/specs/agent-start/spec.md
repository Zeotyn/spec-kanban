# agent-start Specification

## Purpose

Defines how a Change is handed to a new coding-agent session in its own herdr tab, so implementation can begin from the Board with a single action.
## Requirements
### Requirement: Start opens an agent session for a Change
Starting a Change SHALL create a new herdr tab whose working directory is the Project and whose label is the Change name, focus it, start an agent of the chosen kind in that tab as an agent that herdr tracks, and submit the configured prompt with `{change}` replaced by the Change name to it; without a configured prompt this is `/opsx:apply <change>`. The kind SHALL be the configured `agent`; without one, the kind herdr reports for the agent running in the invoking pane (the pane the Board was opened from, or the pane `kanban start` runs in); `claude` when that pane runs no agent, herdr cannot tell, or the reported kind is not one Start can launch.

#### Scenario: Start a Ready Change
- **WHEN** the user runs `kanban start m1-board-view` inside a Project that contains that open Change, from a plain shell, and nothing is configured
- **THEN** a new tab labelled `m1-board-view` opens in the Project directory and is focused before the agent is ready, Claude runs in it, and its first prompt is `/opsx:apply m1-board-view`

#### Scenario: Start with a configured prompt
- **WHEN** the configured prompt is `/opsx:apply {change} with ponytail, TDD, subagents and review loops` and the user presses `s` on Card `m1-board-view`
- **THEN** the agent's first prompt in the new tab is `/opsx:apply m1-board-view with ponytail, TDD, subagents and review loops`

#### Scenario: Board opened from a Codex pane
- **WHEN** the Board was opened from a pane where herdr reports a `codex` agent, no `agent` is configured, and the user presses `s` on Card `m1-board-view`
- **THEN** Codex is started in the new tab and receives the configured prompt

#### Scenario: Board opened from a plain shell
- **WHEN** the Board was opened from a pane that runs no agent, no `agent` is configured, and the user presses `s`
- **THEN** Claude is started in the new tab

#### Scenario: herdr tracks the session
- **WHEN** a Change has been started
- **THEN** `herdr agent list` shows an agent for that Change, so its status appears in the herdr sidebar

### Requirement: Start does not duplicate a running session
When an agent for the Change is already running, Start SHALL focus that agent instead of creating another tab or agent. The agent SHALL count as running while anything but the shell holds the foreground of its pane, whatever its kind and whatever kind Start would launch now. Only when the shell is idle again SHALL Start start the chosen kind again in that tab. A program other than the agent in the foreground SHALL be treated like a running agent: its tab is focused and nothing is started.

#### Scenario: Start twice
- **WHEN** the user starts `m1-board-view` and then starts it again while the first agent is still running
- **THEN** no second tab is created and the existing agent's tab is focused

#### Scenario: Running agent of another kind
- **WHEN** a Codex agent `m1-board-view` is running and the user starts `m1-board-view` from a Board opened in a Claude pane
- **THEN** no second tab is created and the Codex agent's tab is focused

#### Scenario: Start after the agent exited
- **WHEN** the agent for `m1-board-view` has exited but its tab is still open, and the user starts the Change again
- **THEN** no second tab is created; the tab is focused, then the chosen agent kind is started again in it and receives the configured prompt

#### Scenario: Agent shown as node
- **WHEN** a Gemini agent `m1-board-view` is running and its process shows as `node`
- **THEN** no second tab is created and the Gemini agent's tab is focused

#### Scenario: Another program in the tab
- **WHEN** the agent for `m1-board-view` has exited and `vim` runs in its tab
- **THEN** Start creates no tab, starts no agent, and focuses that tab

### Requirement: Start rejects unknown Changes
Start SHALL check that the Change exists as an open Change of the Project before calling herdr, and SHALL exit with a non-zero status and a message naming the Change otherwise.

#### Scenario: Typo in the Change name
- **WHEN** the user runs `kanban start m1-bord-view` and no such open Change exists
- **THEN** no tab is created and the command prints that `m1-bord-view` is not an open Change of the Project

#### Scenario: Archived Change
- **WHEN** the user runs `kanban start` with the name of a Change that exists only in the archive
- **THEN** no tab is created and the command prints that the Change is not open

### Requirement: Start reports herdr failures
When any herdr step fails, Start SHALL stop, report which step failed with herdr's error, and leave no agent half-started without saying so. A failure in the part that runs after the Board closed SHALL be reported as a herdr notification.

#### Scenario: Claude does not become ready
- **WHEN** the agent does not reach readiness within herdr's start timeout
- **THEN** the command exits non-zero, names the tab it created, and says the prompt was not sent

#### Scenario: Claude does not become ready after the Board closed
- **WHEN** a Start from the Board has closed the Board and the agent does not reach readiness
- **THEN** herdr shows the notification `openspec-kanban: start failed` with that error

### Requirement: Start does not take another Project's agent
herdr agent names are unique across all Projects. When an agent named after the Change exists but works in a directory outside the Project, Start SHALL neither focus nor restart it, create no tab, and fail with a message naming the Change and the other agent's directory.

#### Scenario: Equally named Change in two Projects
- **WHEN** an agent `m1-board-view` works in Project `alpha` and the user starts `m1-board-view` in Project `beta`
- **THEN** no tab is created, the `alpha` agent is not focused, and Start reports that `m1-board-view` already runs in `alpha`

