## MODIFIED Requirements

### Requirement: Start a Change from the Board
`s` on the selected Card SHALL Start its Change when the Card is in Ready or In Progress. On a Review Card, `s` SHALL focus the Change's agent when it is running, otherwise resume the Change's recorded session when it is reopenable; with a recorded session of a kind that cannot be reopened it SHALL not Start and SHALL show the notice `resume not supported for <kind>`; without a recorded session it SHALL not Start and SHALL show a notice saying all tasks are done. While Start or resume runs, the footer SHALL say which Change is starting and the Board SHALL ignore every key except `ctrl+c`. Once a running agent's tab is focused, or the agent tab exists and Claude's launch is handed to the background, the Board SHALL close without waiting for Claude to become ready; the background launch SHALL focus the agent tab only after the Board handed it over, so the user ends up in the agent tab. When Start fails before the Board closes, the Board SHALL stay open and show the error as a notice in the footer; when it fails after the Board closed, herdr SHALL show the error as a notification. On a Draft or Archived Card, `s` SHALL not Start and SHALL show a notice saying why. In a column without Cards, `s` SHALL do nothing. The detail view SHALL not react to `s`.

#### Scenario: Start a Ready Card
- **WHEN** Card `m1-board-view` is selected in Ready and the user presses `s`
- **THEN** the footer says `m1-board-view` is starting, a new tab labelled `m1-board-view` is created and the Board closes before Claude is ready; the tab is then focused, Claude starts in it and receives `/opsx:apply m1-board-view`

#### Scenario: Start a Change whose agent is running
- **WHEN** Card `m1-board-view` in In Progress has a Running agent and the user presses `s`
- **THEN** no new tab is created, the existing agent's tab is focused, and the Board closes

#### Scenario: Start fails before the Board closes
- **WHEN** the user presses `s` on a Ready Card and herdr cannot create the tab
- **THEN** the Board stays open and the footer shows the error

#### Scenario: Start fails after the Board closed
- **WHEN** the user presses `s` on a Ready Card and Claude does not become ready
- **THEN** herdr shows a notification `openspec-kanban: start failed` naming the tab and saying the prompt was not sent

#### Scenario: Draft Card
- **WHEN** a Draft Card is selected and the user presses `s`
- **THEN** nothing is started and the footer says the Change must be planned first

#### Scenario: Review Card
- **WHEN** Review Card `m1-board-view` without a Running agent is selected, its `claude` session was recorded, and the user presses `s`
- **THEN** the footer says `m1-board-view` is starting, the Board closes, and a new focused tab labelled `m1-board-view` resumes that session without a prompt

#### Scenario: Review Card without a recorded session
- **WHEN** a Review Card without a Running agent and without a recorded session is selected and the user presses `s`
- **THEN** nothing is started and the footer says all its tasks are done

#### Scenario: Review Card of an agent kind without resume
- **WHEN** Review Card `m1-board-view` without a Running agent has a recorded `codex` session and the user presses `s`
- **THEN** nothing is started and the footer says `resume not supported for codex`

#### Scenario: Review Card with a Running agent
- **WHEN** Review Card `m1-board-view` has a Running agent and the user presses `s`
- **THEN** no new tab is created, the existing agent's tab is focused, and the Board closes

#### Scenario: Archived Card
- **WHEN** an Archived Card is selected and the user presses `s`
- **THEN** nothing is started and the footer says archived Changes cannot be started

#### Scenario: Keys while starting
- **WHEN** a Start is running and the user presses `q`, `j` or `s`
- **THEN** the Board does not change and the Start continues
