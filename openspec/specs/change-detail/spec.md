# change-detail Specification

## Purpose

Defines the detail view of a Change: which documents it shows in which tabs, how they are rendered and scrolled, and how the user moves between tabs and back to the Board.

## Requirements

### Requirement: Tabs for a Change's documents
The detail view SHALL show one Change with four tabs in this order: Tasks, Proposal, Design, Specs. A header SHALL show the Change name, its milestone badge, its Progress and its Phase. The view SHALL open on Tasks when the Change has tasks and on Proposal otherwise. `1`–`4` SHALL switch to that tab, `Tab` and `shift+Tab` to the next and previous tab, wrapping around.

#### Scenario: Open a Change with tasks
- **WHEN** the user opens the details of `m1-board-view`, which has 5 of 11 tasks completed
- **THEN** the header shows `m1-board-view`, `M1`, `5/11` and `In Progress`, and the Tasks tab is shown

#### Scenario: Open a Draft without tasks
- **WHEN** the user opens the details of a Change that has only a proposal
- **THEN** the Proposal tab is shown

#### Scenario: Switch tabs
- **WHEN** the Tasks tab is shown and the user presses `Tab` four times
- **THEN** Proposal, Design, Specs and then Tasks are shown in turn

### Requirement: Rendered documents
Each tab SHALL show its document rendered as markdown for the terminal, wrapped to the view width. The Tasks tab SHALL start with the next open task, or a note that all tasks are done, followed by `tasks.md`. The Specs tab SHALL show every `spec.md` below the Change's `specs/` directory, at any depth, each under a heading with its capability path. A tab whose document does not exist SHALL say that the Change has no such document.

#### Scenario: Next task on top
- **WHEN** the first unchecked task in `tasks.md` is `2.3 Implement the navigation state`
- **THEN** the Tasks tab starts with `next ▸ 2.3 Implement the navigation state`

#### Scenario: Nested specs
- **WHEN** a Change has `specs/agent-control/session-wait/spec.md` and `specs/board/spec.md`
- **THEN** the Specs tab shows both, under the headings `agent-control/session-wait` and `board`

#### Scenario: Missing design
- **WHEN** a Change has no `design.md` and the user switches to Design
- **THEN** the tab says the Change has no design

### Requirement: Scrolling with vim keys
When a document is longer than the view, `j`/`k` SHALL scroll by one line, `ctrl-d`/`ctrl-u` by half a view, and `gg`/`G` to the top and bottom; scrolling SHALL stop at both ends. Each tab SHALL keep its own scroll position while the detail view is open.

#### Scenario: Scroll and switch back
- **WHEN** the user scrolls the Proposal down 10 lines, switches to Design and back to Proposal
- **THEN** the Proposal is still scrolled 10 lines down

### Requirement: Back to the Board
`q` and `Esc` SHALL close the detail view and show the Board with the same Card selected; `ctrl+c` SHALL close the popup.

#### Scenario: Return to the Board
- **WHEN** the user opened the details of the third Card in Ready and presses `q`
- **THEN** the Board is shown with the third Card in Ready selected

### Requirement: Archived Changes
The details of an archived Change SHALL be read from its folder in `openspec/changes/archive/` and show its archive date in the header instead of the Phase.

#### Scenario: Archived Change
- **WHEN** the user opens the details of `m0-plugin-skeleton`, archived on 2026-10-29
- **THEN** its documents are read from `openspec/changes/archive/2026-10-29-m0-plugin-skeleton/` and the header shows `archived 2026-10-29`

### Requirement: Edit the current document
`e` in the detail view SHALL open the file behind the current tab in the user's editor (`$EDITOR`, `vi` when unset): `tasks.md`, `proposal.md` or `design.md`, and on the Specs tab all of the Change's `spec.md` files. The detail view SHALL be suspended while the editor runs and SHALL show the edited Change afresh afterwards, on the same tab. For a document that does not exist, and for archived Changes, `e` SHALL show a short notice instead of opening an editor.

#### Scenario: Edit the proposal
- **WHEN** the Proposal tab of `m2-editor-open` is shown and the user presses `e`, edits and quits the editor
- **THEN** the detail view returns on the Proposal tab showing the edited text

#### Scenario: Edit all specs
- **WHEN** the Specs tab of a Change with two `spec.md` files is shown and the user presses `e`
- **THEN** the editor opens with both files

#### Scenario: Missing document
- **WHEN** the Design tab of a Change without `design.md` is shown and the user presses `e`
- **THEN** no editor opens and the view says there is no design to edit

#### Scenario: Archived Change
- **WHEN** the details of an archived Change are shown and the user presses `e`
- **THEN** no editor opens and the view says archived Changes are read-only

#### Scenario: Editor fails to start
- **WHEN** `$EDITOR` names a program that does not exist and the user presses `e`
- **THEN** the view shows that the editor could not be started and stays usable

### Requirement: Tab bar emphasis
The tab bar SHALL show the active tab bold in the accent colour (ANSI blue, `4`) and every other tab faint, besides the existing marks (`▌` for the active tab, `·` for a tab whose document does not exist). It SHALL use only the 16 ANSI terminal colours, bold and faint. The rest of the detail view SHALL keep its look.

#### Scenario: Active tab
- **WHEN** the details of `m1-board-view` open on Tasks in a colour terminal
- **THEN** `1 Tasks` is bold and blue, and `2 Proposal`, `3 Design` and `4 Specs` are faint

#### Scenario: No colour support
- **WHEN** the detail view is shown in a terminal without colours
- **THEN** the tab bar looks as before, with `▌` before the active tab
