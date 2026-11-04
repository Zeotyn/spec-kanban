## MODIFIED Requirements

### Requirement: Phase overview
The token `kanban` SHALL list the number of open Changes per Phase in Board column order, using the labels `draft`, `ready`, `wip` and `review`, joined by ` · `, omitting Phases without Changes and leaving out Archived. When the line is longer than 28 characters, Phases SHALL be dropped from the left (Draft first) until it fits. A Project without open Changes SHALL show `no open changes`. When `n` open Changes of the Project wait for the user (see "Attention marker"), the token SHALL start with `⚑<n>` and a space; this prefix SHALL never be dropped, and Phases SHALL be dropped from the left until prefix and Phases together fit in 28 characters.

#### Scenario: Overview of a Project
- **WHEN** a Project has 3 Draft, 2 Ready, no In Progress and 1 Review Change
- **THEN** `kanban` is `3 draft · 2 ready · 1 review`

#### Scenario: Too many Phases for the width
- **WHEN** a Project has 3 Draft, 2 Ready, 1 In Progress and 1 Review Change
- **THEN** `kanban` is `2 ready · 1 wip · 1 review`

#### Scenario: Only archived Changes
- **WHEN** all Changes of the Project are archived
- **THEN** `kanban` is `no open changes`

#### Scenario: One Change waits
- **WHEN** a Project has 1 Ready and 1 Review Change, and the Review Change waits for the user
- **THEN** `kanban` is `⚑1 1 ready · 1 review`

#### Scenario: Prefix keeps its place
- **WHEN** a Project has 2 Ready, 1 In Progress and 1 Review Change, and one of them waits for the user
- **THEN** `kanban` is `⚑1 1 wip · 1 review`

### Requirement: Refresh on herdr activity
The plugin SHALL refresh the tokens of all workspaces when herdr starts, when an agent's status changes, when a pane is focused or created, when a workspace is created, and when the plugin sets an attention marker. When a refresh is requested while one is running, the running refresh SHALL do exactly one more pass after it finishes, however many requests arrived meanwhile, and the requesting command SHALL return at once. Each Project SHALL be read once per pass even when several workspaces show it.

#### Scenario: Agent finishes a task
- **WHEN** an agent completes a task of the Most recently modified Change and its status changes to idle
- **THEN** the workspace's `kanban_change` shows the new count without user action

#### Scenario: Burst of events
- **WHEN** five events arrive while a refresh is running
- **THEN** exactly one further pass runs after it, and the five event commands exit without reading any Project

#### Scenario: Shared Project
- **WHEN** two workspaces work in the same Project
- **THEN** a pass runs `openspec list` for that Project once and reports the same tokens to both

#### Scenario: Stall while nothing else happens
- **WHEN** the agent of `m1-board-view` stalls and the plugin sets its marker while no herdr event occurs
- **THEN** the workspace's `kanban` token gains the `⚑1` prefix without user action
