# sidebar-summary Specification

## Purpose

Defines the summary the plugin writes into each herdr workspace's Space rows: which Project a workspace shows, what the two summary lines say, when they are refreshed, and how errors affect them.

## Requirements

### Requirement: Each workspace shows the Project its panes work in
The plugin SHALL assign each herdr workspace the Project that contains the working directory of one of its panes, using the pane's foreground process directory when herdr reports one. When the panes of a workspace work in different Projects, the focused pane's Project SHALL win; otherwise the first pane herdr lists with a Project SHALL decide. A workspace with no pane inside a Project SHALL have no Project.

#### Scenario: Pane in a subdirectory
- **WHEN** a pane of workspace `w1` works in `<project>/internal/board` and `<project>/openspec/changes` exists
- **THEN** `w1` shows the summary of `<project>`

#### Scenario: Two Projects in one workspace
- **WHEN** workspace `w1` has a pane in Project `magentic` and the focused pane in Project `openspec-kanban`
- **THEN** `w1` shows the summary of `openspec-kanban`

#### Scenario: Workspace outside any Project
- **WHEN** no pane of workspace `w2` works inside a Project
- **THEN** both summary tokens of `w2` are cleared, so its Space rows show nothing from this plugin

### Requirement: Most recently modified Change with its Progress
The token `kanban_change` SHALL show the Most recently modified Change of the workspace's Project as a six-cell Progress bar (`▰` completed, `▱` open), the completed/total task count and the Change name, separated by single spaces. The bar SHALL be full only when all tasks are completed. A Change without tasks SHALL show `no tasks` instead of bar and count. The name SHALL be shortened in the middle with `…` so the whole value fits in 28 characters. When the Project has no open Change, the token SHALL be cleared.

#### Scenario: Change in progress
- **WHEN** the Most recently modified Change is `m1-board-view` with 3 of 6 tasks completed
- **THEN** `kanban_change` is `▰▰▰▱▱▱ 3/6 m1-board-view`

#### Scenario: Almost done is not full
- **WHEN** the Most recently modified Change has 29 of 30 tasks completed
- **THEN** its bar shows at least one `▱`

#### Scenario: Change without tasks
- **WHEN** the Most recently modified Change is `add-dark-mode` with no tasks
- **THEN** `kanban_change` is `no tasks add-dark-mode`

#### Scenario: Long name
- **WHEN** the name does not fit next to the bar and count in 28 characters
- **THEN** it is shortened in the middle with `…`, keeping its start and end

#### Scenario: No open Change
- **WHEN** all Changes of the Project are archived
- **THEN** `kanban_change` is cleared

### Requirement: Phase overview
The token `kanban` SHALL list the number of open Changes per Phase in Board column order, using the labels `draft`, `ready`, `wip` and `review`, joined by ` · `, omitting Phases without Changes and leaving out Archived. When the line is longer than 28 characters, Phases SHALL be dropped from the left (Draft first) until it fits. A Project without open Changes SHALL show `no open changes`.

#### Scenario: Overview of a Project
- **WHEN** a Project has 3 Draft, 2 Ready, no In Progress and 1 Review Change
- **THEN** `kanban` is `3 draft · 2 ready · 1 review`

#### Scenario: Too many Phases for the width
- **WHEN** a Project has 3 Draft, 2 Ready, 1 In Progress and 1 Review Change
- **THEN** `kanban` is `2 ready · 1 wip · 1 review`

#### Scenario: Only archived Changes
- **WHEN** all Changes of the Project are archived
- **THEN** `kanban` is `no open changes`

### Requirement: Refresh on herdr activity
The plugin SHALL refresh the tokens of all workspaces when herdr starts, when an agent's status changes, when a pane is focused or created, and when a workspace is created. When a refresh is requested while one is running, the running refresh SHALL do exactly one more pass after it finishes, however many requests arrived meanwhile, and the requesting command SHALL return at once. Each Project SHALL be read once per pass even when several workspaces show it.

#### Scenario: Agent finishes a task
- **WHEN** an agent completes a task of the Most recently modified Change and its status changes to idle
- **THEN** the workspace's `kanban_change` shows the new count without user action

#### Scenario: Burst of events
- **WHEN** five events arrive while a refresh is running
- **THEN** exactly one further pass runs after it, and the five event commands exit without reading any Project

#### Scenario: Shared Project
- **WHEN** two workspaces work in the same Project
- **THEN** a pass runs `openspec list` for that Project once and reports the same tokens to both

### Requirement: Refresh shares the Board's cache
A refresh SHALL read Changes with the same cached planning reuse as the Board, and every successful read SHALL update the Project's cache, so that a following Board open or refresh benefits from it. Cache problems SHALL never stop a refresh.

#### Scenario: Board opens after a refresh
- **WHEN** a refresh has read Project `magentic` and the user then opens the Board for it
- **THEN** the Board shows those Changes at once from the cache

#### Scenario: Unwritable state directory
- **WHEN** the cache directory is not writable
- **THEN** the refresh still reports the tokens

### Requirement: Errors keep the last summary
When a Project's Changes cannot be read (for example, the openspec CLI is missing or fails), the tokens of the workspaces showing that Project SHALL stay as they were, and the error SHALL appear in the plugin's command log. Other workspaces SHALL be refreshed as usual. When herdr cannot list its workspaces or panes, the refresh SHALL report nothing and fail with herdr's error.

#### Scenario: openspec CLI fails for one Project
- **WHEN** `openspec list` fails in Project `magentic` but succeeds in `openspec-kanban`
- **THEN** the workspaces of `magentic` keep their previous tokens and those of `openspec-kanban` are refreshed

#### Scenario: herdr unreachable
- **WHEN** `herdr workspace list` fails
- **THEN** no token is reported or cleared and the command exits with an error

### Requirement: Sidebar config snippet
The repository SHALL provide a herdr config snippet with Space rows that show `$kanban_change` and `$kanban` on their own lines, make `$kanban_change` bold when the Change's Progress bar is full, and widen the sidebar so the 28-character values fit. The README SHALL explain copying the snippet and replacing herdr-openspec's `$openspec` rows.

#### Scenario: Snippet applied
- **WHEN** the user copies the snippet into `~/.config/herdr/config.toml` and reloads the config
- **THEN** each workspace inside a Project shows the two summary lines under its name, and the first one turns bold once all tasks of the Change are completed
