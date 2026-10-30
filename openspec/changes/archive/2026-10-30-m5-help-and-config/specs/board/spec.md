## MODIFIED Requirements

### Requirement: Phase columns with a summary header
The Board SHALL show five columns in the order Draft, Ready, In Progress, Review, Archived, each titled with its Phase name and Card count, except the Phases the config hides. A hidden Phase's column SHALL not be drawn, and navigation SHALL treat it as a column without Cards; `1`–`5` keep their Phase numbers. A header SHALL show the Project name and the overall Progress of all open Changes as completed/total tasks with a percentage, including Changes in hidden Phases. An empty column SHALL say so.

#### Scenario: Board of a Project
- **WHEN** the Board opens for Project `magentic` with 9 open Changes holding 120 of 287 tasks completed
- **THEN** the header shows `magentic` and `120/287 42%`, and the five columns show their Phase names with Card counts that add up to all Changes

#### Scenario: Empty Phase
- **WHEN** no Change is in Review
- **THEN** the Review column shows its title with count 0 and a short "nothing here" note

#### Scenario: Archived hidden
- **WHEN** the config hides Archived and the Project has archived Changes
- **THEN** the Board shows the four columns Draft, Ready, In Progress and Review using the full width, and `5` does nothing

#### Scenario: Hidden Phase with open Changes
- **WHEN** the config hides Draft, and Draft Changes have tasks
- **THEN** no Draft column is shown, `h` from Ready does not move focus, and the header's Progress still counts the Draft Changes' tasks

## ADDED Requirements

### Requirement: Key help
`?` on the Board or in the detail view SHALL show the key help in place of the view: every key with what it does, grouped under Board, Filter prompt and Details. The next key SHALL close the help and return to the view it was opened from, without any other effect; `ctrl+c` SHALL close the Board. In the filter prompt `?` SHALL be typed into the filter, and while Start runs it SHALL be ignored like every other key. The footers of the Board and the detail view SHALL mention `?`.

#### Scenario: Help on the Board
- **WHEN** the Board is open and the user presses `?`
- **THEN** the key help is shown with the groups Board, Filter prompt and Details

#### Scenario: Closing the help does nothing else
- **WHEN** the key help is shown over the Board and the user presses `q`
- **THEN** the help closes, the Board stays open, and the selection is unchanged

#### Scenario: Help from the detail view
- **WHEN** the detail view of `m1-board-view` is open and the user presses `?` and then `j`
- **THEN** the key help is shown, and after `j` the detail view of `m1-board-view` is shown again at the same scroll position

#### Scenario: Question mark in the filter
- **WHEN** the filter prompt is open and the user types `?`
- **THEN** no help is shown and the filter contains `?`
