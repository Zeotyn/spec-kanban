## MODIFIED Requirements

### Requirement: Start opens an agent session for a Change
Starting a Change SHALL create a new herdr tab whose working directory is the Project and whose label is the Change name, focus it, start Claude in that tab as an agent that herdr tracks, and submit the configured prompt with `{change}` replaced by the Change name to it; without a configured prompt this is `/opsx:apply <change>`.

#### Scenario: Start a Ready Change
- **WHEN** the user runs `kanban start m1-board-view` inside a Project that contains that open Change, and no prompt is configured
- **THEN** a new tab labelled `m1-board-view` opens in the Project directory and is focused before Claude is ready, Claude runs in it, and its first prompt is `/opsx:apply m1-board-view`

#### Scenario: Start with a configured prompt
- **WHEN** the configured prompt is `/opsx:apply {change} with ponytail, TDD, subagents and review loops` and the user presses `s` on Card `m1-board-view`
- **THEN** Claude's first prompt in the new tab is `/opsx:apply m1-board-view with ponytail, TDD, subagents and review loops`

#### Scenario: herdr tracks the session
- **WHEN** a Change has been started
- **THEN** `herdr agent list` shows an agent for that Change, so its status appears in the herdr sidebar
