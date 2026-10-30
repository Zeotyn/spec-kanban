## MODIFIED Requirements

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
When an agent for the Change is already running, Start SHALL focus that agent instead of creating another tab or agent. Whether it still runs SHALL be decided by the kind herdr reports for that agent, whatever kind Start would launch now. When it has exited, Start SHALL start the chosen kind again in its tab. When another program holds that tab's pane, Start SHALL fail naming that program.

#### Scenario: Start twice
- **WHEN** the user starts `m1-board-view` and then starts it again while the first agent is still running
- **THEN** no second tab is created and the existing agent's tab is focused

#### Scenario: Running agent of another kind
- **WHEN** a Codex agent `m1-board-view` is running and the user starts `m1-board-view` from a Board opened in a Claude pane
- **THEN** no second tab is created and the Codex agent's tab is focused

#### Scenario: Start after the agent exited
- **WHEN** the agent for `m1-board-view` has exited but its tab is still open, and the user starts the Change again
- **THEN** no second tab is created; the tab is focused, then the chosen agent kind is started again in it and receives the configured prompt

#### Scenario: Another program in the tab
- **WHEN** the agent for `m1-board-view` has exited and `vim` runs in its tab
- **THEN** Start creates no tab, starts no agent, and reports that the pane runs `vim`
