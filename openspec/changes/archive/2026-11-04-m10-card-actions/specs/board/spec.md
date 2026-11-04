## ADDED Requirements

### Requirement: Verify and archive from the Board
`v` on the selected Card SHALL Verify its Change and `a` SHALL Archive it (see "Verify and Archive run the Source's command"), when the Card is in Review, its Source names that command, and no agent of the Change is busy. Otherwise the key SHALL start nothing and show a notice:
- on a Card outside Review: `<name> is not in Review`;
- when the Source names no such command: `<Source> has no verify` or `<Source> has no archive`;
- when the Change's agent is busy: `<name> is <status>: wait until it is idle`, where `<status>` is herdr's agent status, for example `working`.

While the command is handed over, the footer SHALL say which Change it is verifying or archiving and the Board SHALL ignore keys and mouse events as it does while starting; once handed over, the Board SHALL close. A failure before the Board closes SHALL show as a notice; a failure after it SHALL show as a herdr notification, as for Start. In a column without Cards, `v` and `a` SHALL do nothing. The detail view SHALL not react to `v` and `a` outside an open Card menu. The key help SHALL list `v` and `a` under Board; the footer SHALL stay unchanged.

#### Scenario: Verify a Review Card
- **WHEN** Review Card `m1-board-view` has no agent and the user presses `v`
- **THEN** the footer says `m1-board-view` is verifying, the Board closes, and a new agent in the Change's checkout receives `/opsx:verify m1-board-view`

#### Scenario: Archive through an idle agent
- **WHEN** Review Card `m1-board-view` has an idle agent and the user presses `a`
- **THEN** no tab is created, that agent receives `/opsx:archive m1-board-view`, its tab is focused and the Board closes

#### Scenario: Busy agent
- **WHEN** Review Card `m1-board-view` has an agent whose status is `working` and the user presses `v`
- **THEN** nothing is sent and the footer says `m1-board-view is working: wait until it is idle`

#### Scenario: Outside Review
- **WHEN** an In Progress Card `m1-board-view` is selected and the user presses `a`
- **THEN** nothing is sent and the footer says `m1-board-view is not in Review`

#### Scenario: Spec Kit Feature
- **WHEN** Review Card `003-search` of a Spec Kit Project is selected and the user presses `v`
- **THEN** nothing is sent and the footer says `Spec Kit has no verify`

## MODIFIED Requirements

### Requirement: Card menu
A right click on a Card SHALL select it and open the Card menu at the pointer, shifted so it stays fully inside the popup. The Card menu SHALL hold, each entry with its key:
- `▶ Start  s` on a Ready or In Progress Card, `↻ Resume  s` on a Review Card, and `▶ Start  s` in the faint style on a Draft or Archived Card;
- below it `✓ Verify  v` and `⇥ Archive  a` when the Card's Source names that command, in the default style on a Review Card and faint otherwise; a busy agent shows as the key's notice on click, not as a faint entry. A Source without the command shows no entry for it.

A left click on an entry, or its key while the menu is open, SHALL close the menu and do what that key does on the Card, including its notices. A left click outside the entries, `Esc` or any other key SHALL close the menu without any other effect; a right click on another Card SHALL open the menu for that Card instead; a right click outside a Card SHALL close the menu. In the detail view a right click anywhere SHALL open the Card menu for the shown Card. The Card menu SHALL use only the terminal palette, like the rest of the Board.

#### Scenario: Start from the menu
- **WHEN** the user right-clicks Ready Card `m1-board-view` and clicks `▶ Start`
- **THEN** the footer says `m1-board-view` is starting and the Start runs as with `s`

#### Scenario: Resume from the menu
- **WHEN** the user right-clicks Review Card `m1-board-view` whose `claude` session was recorded
- **THEN** the menu shows `↻ Resume  s`, and a click on it resumes the session as with `s`

#### Scenario: Faint entry on a Draft Card
- **WHEN** the user right-clicks a Draft Card and clicks the faint `▶ Start`
- **THEN** the menu closes, nothing is started and the footer says the Change must be planned first

#### Scenario: Verify and Archive entries
- **WHEN** the user right-clicks Review Card `m1-board-view` of an OpenSpec Project, without an agent
- **THEN** the menu shows `↻ Resume  s`, `✓ Verify  v` and `⇥ Archive  a`, and a click on `✓ Verify` verifies the Change as with `v`

#### Scenario: Faint Verify outside Review
- **WHEN** the user right-clicks In Progress Card `m1-board-view` of an OpenSpec Project
- **THEN** `✓ Verify  v` and `⇥ Archive  a` are faint, and a click on one shows the notice of its key

#### Scenario: Spec Kit menu
- **WHEN** the user right-clicks a Review Card of a Spec Kit Project
- **THEN** the menu shows only `↻ Resume  s`

#### Scenario: Close the menu
- **WHEN** the Card menu is open and the user presses `j`
- **THEN** the menu closes and the selection is unchanged

#### Scenario: Menu at the edge
- **WHEN** the user right-clicks a Card in Archived next to the right edge of the popup
- **THEN** the whole menu is drawn inside the popup, to the left of the pointer

#### Scenario: Menu in the details
- **WHEN** the detail view of In Progress Card `m1-board-view` is open and the user right-clicks the document and presses `s`
- **THEN** `m1-board-view` starts as with `s` on the Board
