## ADDED Requirements

### Requirement: Mouse on the Board
The Board SHALL turn on mouse reporting. A left click on a Card SHALL focus its column and select it; a left click on the Card that is already selected SHALL open its details, like `Enter`. A left click elsewhere on the Board SHALL change nothing, except on a footer action (see "Clickable footer"). Turning the mouse wheel down or up over a column SHALL focus that column, when it has Cards, and select the next or previous Card, like `j`/`k`, scrolling the column so the selected Card stays visible. In single-column mode a click on `◀` or `▶` SHALL show the previous or next column that has Cards, like `h`/`l`, and the wheel SHALL act on the shown column. Clicks SHALL act on the press of the button. While the filter prompt is open or a Start runs, the Board SHALL ignore mouse events. Every key SHALL keep working as before.

#### Scenario: Select a Card with a click
- **WHEN** the second Card in Ready is selected and the user clicks the first Card in In Progress
- **THEN** In Progress is focused with its first Card selected, and Ready still remembers its second Card

#### Scenario: Open details with a second click
- **WHEN** Card `m1-board-view` is selected and the user clicks it
- **THEN** the detail view of `m1-board-view` is shown

#### Scenario: Click on empty space
- **WHEN** the user clicks the header line or below the last Card of a column
- **THEN** the focus and the selection stay as they are

#### Scenario: Wheel over a column
- **WHEN** Ready is focused, In Progress has room for 6 Cards and holds 10, its first Card is selected, and the user turns the wheel down 7 steps over In Progress
- **THEN** In Progress is focused, its eighth Card is selected and visible, and the column indicates Cards above it

#### Scenario: Wheel over an empty column
- **WHEN** Review has no Cards and the user turns the wheel over it
- **THEN** the focus and the selection stay as they are

#### Scenario: Arrows in single-column mode
- **WHEN** the Board shows only Ready as `◀ 2/5 Ready 3 ▶`, In Progress has Cards, and the user clicks `▶`
- **THEN** the Board shows In Progress

#### Scenario: Mouse while starting
- **WHEN** a Start is running and the user clicks a Card or turns the wheel
- **THEN** the Board does not change and the Start continues

### Requirement: Card menu
A right click on a Card SHALL select it and open the Card menu at the pointer, shifted so it stays fully inside the popup. The Card menu SHALL hold one entry with its key: `▶ Start  s` on a Ready or In Progress Card, `↻ Resume  s` on a Review Card, and `▶ Start  s` in the faint style on a Draft or Archived Card. A left click on the entry, or `s` while the menu is open, SHALL close the menu and do what `s` does on that Card, including its notices. A left click outside the menu, `Esc` or any other key SHALL close the menu without any other effect; a right click on another Card SHALL open the menu for that Card instead; a right click outside a Card SHALL close the menu. In the detail view a right click anywhere SHALL open the Card menu for the shown Card. The Card menu SHALL use only the terminal palette, like the rest of the Board.

#### Scenario: Start from the menu
- **WHEN** the user right-clicks Ready Card `m1-board-view` and clicks `▶ Start`
- **THEN** the footer says `m1-board-view` is starting and the Start runs as with `s`

#### Scenario: Resume from the menu
- **WHEN** the user right-clicks Review Card `m1-board-view` whose `claude` session was recorded
- **THEN** the menu shows `↻ Resume  s`, and a click on it resumes the session as with `s`

#### Scenario: Faint entry on a Draft Card
- **WHEN** the user right-clicks a Draft Card and clicks the faint `▶ Start`
- **THEN** the menu closes, nothing is started and the footer says the Change must be planned first

#### Scenario: Close the menu
- **WHEN** the Card menu is open and the user presses `j`
- **THEN** the menu closes and the selection is unchanged

#### Scenario: Menu at the edge
- **WHEN** the user right-clicks a Card in Archived next to the right edge of the popup
- **THEN** the whole menu is drawn inside the popup, to the left of the pointer

#### Scenario: Menu in the details
- **WHEN** the detail view of In Progress Card `m1-board-view` is open and the user right-clicks the document and presses `s`
- **THEN** `m1-board-view` starts as with `s` on the Board

### Requirement: Clickable footer
In the Board footer a left click on the action `⏎ details`, `e edit tasks`, `s start`, `/ filter`, `r reload`, `? help` or `q close` SHALL act like its key, both in the full and in the short footer. Navigation keys in the footer SHALL not react to clicks. While the footer shows a notice, the prompt or the starting Change, clicks on it SHALL do nothing.

#### Scenario: Close with a click
- **WHEN** the user clicks `q close` in the Board footer
- **THEN** the Board closes

#### Scenario: Open the filter with a click
- **WHEN** the user clicks `/ filter` in the Board footer
- **THEN** the filter prompt opens as with `/`

### Requirement: Close the help with a click
While the key help is shown, a click anywhere SHALL close it and return to the view it was opened from, without any other effect, like a key.

#### Scenario: Click on the help
- **WHEN** the key help is shown over the Board and the user clicks where a Card is drawn underneath
- **THEN** the help closes and the selection is unchanged

## MODIFIED Requirements

### Requirement: Start a Change from the Board
`s` on the selected Card SHALL Start its Change when the Card is in Ready or In Progress. On a Review Card, `s` SHALL focus the Change's agent when it is running, otherwise resume the Change's recorded session when it is reopenable (see "Resume a recorded session"); with a recorded session of a kind that cannot be reopened it SHALL not Start and SHALL show the notice `resume not supported for <kind>`; without a recorded session it SHALL not Start and SHALL show a notice saying all tasks are done. While Start or resume runs, the footer SHALL say which Change is starting and the Board SHALL ignore every key except `ctrl+c`. Once a running agent's tab is focused, or the agent tab exists and the agent's launch is handed to the background, the Board SHALL close without waiting for the agent to become ready; the background launch SHALL focus the agent tab only after the Board handed it over, so the user ends up in the agent tab. When Start or resume fails before the Board closes, the Board SHALL stay open and show the error as a notice in the footer; when it fails after the Board closed, herdr SHALL show the error as a notification. On a Draft or Archived Card, `s` SHALL not Start and SHALL show a notice saying why. In a column without Cards, `s` SHALL do nothing. The detail view SHALL not react to `s` outside an open Card menu (see "Card menu").

#### Scenario: Start a Ready Card
- **WHEN** Card `m1-board-view` is selected in Ready and the user presses `s`
- **THEN** the footer says `m1-board-view` is starting, a worktree workspace labelled `m1-board-view` is opened and the Board closes before Claude is ready; its tab is then focused, Claude starts in it and receives `/opsx:apply m1-board-view`

#### Scenario: Start a Change whose agent is running
- **WHEN** Card `m1-board-view` in In Progress has a Running agent and the user presses `s`
- **THEN** no new tab is created, the existing agent's tab is focused, and the Board closes

#### Scenario: Start fails before the Board closes
- **WHEN** the user presses `s` on a Ready Card and herdr cannot create the tab
- **THEN** the Board stays open and the footer shows the error

#### Scenario: Start fails after the Board closed
- **WHEN** the user presses `s` on a Ready Card and Claude does not become ready
- **THEN** herdr shows a notification `spec-kanban: start failed` naming the tab and saying the prompt was not sent

#### Scenario: Draft Card
- **WHEN** a Draft Card is selected and the user presses `s`
- **THEN** nothing is started and the footer says the Change must be planned first

#### Scenario: Review Card
- **WHEN** Review Card `m1-board-view` without a Running agent is selected, its `claude` session was recorded, its worktree is not open in herdr, and the user presses `s`
- **THEN** the footer says `m1-board-view` is starting, the Board closes, and the Change's worktree opens as a workspace whose focused tab resumes that session without a prompt

#### Scenario: Review Card without a recorded session
- **WHEN** a Review Card without a Running agent and without a recorded session is selected and the user presses `s`
- **THEN** nothing is started and the footer says all its tasks are done

#### Scenario: Review Card of an agent kind without resume
- **WHEN** Review Card `m1-board-view` without a Running agent has a recorded `pi` session and the user presses `s`
- **THEN** nothing is started and the footer says `resume not supported for pi`

#### Scenario: Review Card with a Running agent
- **WHEN** Review Card `m1-board-view` has a Running agent and the user presses `s`
- **THEN** no new tab is created, the existing agent's tab is focused, and the Board closes

#### Scenario: Archived Card
- **WHEN** an Archived Card is selected and the user presses `s`
- **THEN** nothing is started and the footer says archived Changes cannot be started

#### Scenario: Keys while starting
- **WHEN** a Start is running and the user presses `q`, `j` or `s`
- **THEN** the Board does not change and the Start continues
