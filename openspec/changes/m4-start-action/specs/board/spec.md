## ADDED Requirements

### Requirement: Start a Change from the Board
`s` on the selected Card SHALL Start its Change when the Card is in Ready or In Progress. While Start runs, the footer SHALL say which Change is starting and the Board SHALL ignore every key except `ctrl+c`. Once a running agent's tab is focused, or the agent tab exists and Claude's launch is handed to the background, the Board SHALL close without waiting for Claude to become ready; the background launch SHALL focus the agent tab only after the Board handed it over, so the user ends up in the agent tab. When Start fails before the Board closes, the Board SHALL stay open and show the error as a notice in the footer; when it fails after the Board closed, herdr SHALL show the error as a notification. On a Draft, Review or Archived Card, `s` SHALL not Start and SHALL show a notice saying why. In a column without Cards, `s` SHALL do nothing. The detail view SHALL not react to `s`.

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
- **WHEN** a Review Card is selected and the user presses `s`
- **THEN** nothing is started and the footer says all its tasks are done

#### Scenario: Archived Card
- **WHEN** an Archived Card is selected and the user presses `s`
- **THEN** nothing is started and the footer says archived Changes cannot be started

#### Scenario: Keys while starting
- **WHEN** a Start is running and the user presses `q`, `j` or `s`
- **THEN** the Board does not change and the Start continues

### Requirement: Running agent badge
An open Card SHALL show a Running agent badge when herdr reports an agent named after the Change whose working directory lies in the Project. The badge SHALL use a terminal colour, not change the Card's Phase or column, and not appear on Archived Cards. Looking up Running agents SHALL not delay showing the Board: the Cards SHALL appear without badges and gain them when herdr answers. `r` SHALL look them up again. When herdr cannot be asked, the Board SHALL show no badges and no error.

#### Scenario: Agent working on a Change
- **WHEN** herdr reports an agent named `m1-board-view` working in the Project and the Board opens
- **THEN** Card `m1-board-view` shows the Running agent badge and stays in its Phase column

#### Scenario: Same name in another Project
- **WHEN** herdr reports an agent named `m1-board-view` whose working directory is another Project
- **THEN** Card `m1-board-view` shows no badge

#### Scenario: Cached Board opens at once
- **WHEN** the Board opens with a cached snapshot and herdr answers later
- **THEN** the cached Cards show immediately and the badges appear once herdr has answered

#### Scenario: herdr unavailable
- **WHEN** the agent lookup fails
- **THEN** the Board shows its Cards without badges and no error
