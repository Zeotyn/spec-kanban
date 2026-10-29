## Purpose

Defines the Board: how a Project's Changes are laid out in Phase columns, what each Card shows, how the user moves around with vim-style keys, and how loading, small terminals and errors look.

## ADDED Requirements

### Requirement: Phase columns with a summary header
The Board SHALL show five columns in the order Draft, Ready, In Progress, Review, Archived, each titled with its Phase name and Card count. A header SHALL show the Project name and the overall Progress of all open Changes as completed/total tasks with a percentage. An empty column SHALL say so.

#### Scenario: Board of a Project
- **WHEN** the Board opens for Project `magentic` with 9 open Changes holding 120 of 287 tasks completed
- **THEN** the header shows `magentic` and `120/287 42%`, and the five columns show their Phase names with Card counts that add up to all Changes

#### Scenario: Empty Phase
- **WHEN** no Change is in Review
- **THEN** the Review column shows its title with count 0 and a short "nothing here" note

### Requirement: Card content
Each Card SHALL show the Change name, shortened in the middle with `…` when it does not fit. A Change whose name starts with `m<N>-` SHALL show a milestone badge `M<N>`. Open Cards SHALL show a Progress bar with completed/total tasks, or "no tasks", and the age of the last modification (`5m`, `3h`, `2d`, `4w`). Draft and Ready Cards SHALL show the planning artifacts as `P S D T`, marking each as done, skipped, or missing. Archived Cards SHALL show their archive date instead of Progress and age.

#### Scenario: Card in progress
- **WHEN** Change `m1-board-view` has 5 of 11 tasks completed and was modified 3 hours ago
- **THEN** its Card shows `m1-board-view`, the badge `M1`, a Progress bar with `5/11`, and `3h`

#### Scenario: Draft Card
- **WHEN** Change `add-dark-mode` has only a proposal and no tasks
- **THEN** its Card shows "no tasks" and `P` as done while `S D T` are marked missing, and no milestone badge

#### Scenario: Long name in a narrow column
- **WHEN** a Change name is longer than the column is wide
- **THEN** the name is shortened in the middle with `…`, keeping its start and end

### Requirement: Vim-style navigation
Exactly one Card SHALL be selected when the focused column has Cards. The keys SHALL act as follows: `h`/`l` focus the previous/next column, `j`/`k` select the next/previous Card, `gg`/`G` select the first/last Card, `ctrl-d`/`ctrl-u` move the selection half a column height down/up, `1`–`5` focus that column. Each column SHALL remember its own selected Card when focus leaves and returns. Moves past the first or last column or Card SHALL stop there.

#### Scenario: Move between columns
- **WHEN** the third Card of Ready is selected, the user presses `l` and then `h`
- **THEN** In Progress is focused with its previously selected Card, and then Ready is focused again with its third Card selected

#### Scenario: Jump to the end
- **WHEN** a column has 12 Cards and the user presses `G` and then `gg`
- **THEN** the twelfth and then the first Card is selected

#### Scenario: Jump to a column
- **WHEN** the user presses `5`
- **THEN** the Archived column is focused

#### Scenario: Empty column
- **WHEN** the user focuses an empty column
- **THEN** no Card is selected and `j`/`k` do nothing

### Requirement: The selected Card stays visible
When a column has more Cards than fit its height, the column SHALL scroll so the selected Card is fully visible, and SHALL show that more Cards exist above or below.

#### Scenario: Scrolling down
- **WHEN** a column has room for 6 Cards, holds 10, and the user presses `j` until the eighth Card is selected
- **THEN** the eighth Card is visible and the column indicates Cards above it

### Requirement: Single column on narrow terminals
When the Board is too narrow for five readable columns, it SHALL show only the focused column, with an indicator of its position and name (for example `◀ 2/5 Ready ▶`); `h`/`l` and `1`–`5` SHALL switch the shown column.

#### Scenario: Narrow popup
- **WHEN** the Board is 60 cells wide
- **THEN** only the focused column is shown with its position indicator, and `l` shows the next column

### Requirement: Loading and reload
The Board SHALL show a loading notice until the Project's Changes are read, then the columns. `r` SHALL read them again and keep the selected Card selected by Change name, as long as it still exists in the same column.

#### Scenario: Reload after a task was completed
- **WHEN** Card `m1-board-view` is selected in In Progress, one of its tasks gets completed, and the user presses `r`
- **THEN** the Board shows the new Progress and `m1-board-view` is still selected

### Requirement: Errors replace the columns
When the Changes cannot be read, the Board SHALL show the error message in place of the columns with a hint that `r` retries; closing SHALL keep working.

#### Scenario: openspec missing
- **WHEN** the openspec CLI is not installed and the Board opens
- **THEN** the Board shows that the openspec CLI was not found and that `r` retries, and `q` closes it
