## MODIFIED Requirements

### Requirement: Loading and reload
When a valid cached snapshot of the Project's Changes exists, the Board SHALL show it as soon as it opens, mark the header as refreshing, and replace it with the freshly read Changes when they arrive. Without a valid snapshot, the Board SHALL show a loading notice until the Project's Changes are read, then the columns. `r` SHALL read them again and keep the selected Card selected by Change name, as long as it still exists in the same column; this also applies when the fresh Changes replace a cached snapshot.

#### Scenario: Reload after a task was completed
- **WHEN** Card `m1-board-view` is selected in In Progress, one of its tasks gets completed, and the user presses `r`
- **THEN** the Board shows the new Progress and `m1-board-view` is still selected

#### Scenario: Open with a cached snapshot
- **WHEN** the Board was opened for Project `magentic` before and is opened again
- **THEN** the columns from the last load show immediately with "refreshing…" in the header, and the header loses that marker once the fresh Changes are shown

#### Scenario: Selection survives the swap
- **WHEN** the cached Board is showing, the user selects `m1-board-view`, and then the fresh Changes arrive with `m1-board-view` in the same column
- **THEN** `m1-board-view` is still selected

#### Scenario: First open of a Project
- **WHEN** the Board opens for a Project without a valid cache
- **THEN** it shows the loading notice until the Changes are read
