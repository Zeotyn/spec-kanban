# board Specification

## Purpose

Defines the Board: how a Project's Changes are laid out in Phase columns, what each Card shows, how the user moves around with vim-style keys, and how loading, small terminals and errors look.
## Requirements
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
Exactly one Card SHALL be selected when the focused column has Cards. Only columns with Cards SHALL receive focus. The keys SHALL act as follows: `h`/`l` focus the previous/next column that has Cards, `j`/`k` select the next/previous Card, `gg`/`G` select the first/last Card, `ctrl-d`/`ctrl-u` move the selection half a column height down/up, `1`–`5` focus that column if it has Cards and do nothing otherwise. When the focused column has no Cards after a load, focus SHALL move to the first column with Cards to its right, otherwise to the first one to its left, otherwise to Draft. Each column SHALL remember its own selected Card when focus leaves and returns. Moves past the first or last column or Card SHALL stop there.

#### Scenario: Move between columns
- **WHEN** Ready and In Progress both have Cards, the third Card of Ready is selected, and the user presses `l` and then `h`
- **THEN** In Progress is focused with its previously selected Card, and then Ready is focused again with its third Card selected

#### Scenario: Jump to the end
- **WHEN** a column has 12 Cards and the user presses `G` and then `gg`
- **THEN** the twelfth and then the first Card is selected

#### Scenario: Jump to a column
- **WHEN** Archived has Cards and the user presses `5`
- **THEN** the Archived column is focused

#### Scenario: Empty columns are skipped
- **WHEN** Draft is focused, Ready is empty and In Progress has Cards, and the user presses `l`
- **THEN** In Progress is focused

#### Scenario: Jump to an empty column
- **WHEN** Review is empty and the user presses `4`
- **THEN** focus stays where it was

#### Scenario: Focused column empties on reload
- **WHEN** the only Card in Review is archived and the user presses `r` while Review is focused
- **THEN** focus moves to Archived

#### Scenario: Nothing to show
- **WHEN** the Project has no Changes at all
- **THEN** Draft is focused, every column says it is empty, and no Card is selected

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

### Requirement: Errors replace the columns
When the Changes cannot be read, the Board SHALL show the error message in place of the columns with a hint that `r` retries; closing SHALL keep working.

#### Scenario: openspec missing
- **WHEN** the openspec CLI is not installed and the Board opens
- **THEN** the Board shows that the openspec CLI was not found and that `r` retries, and `q` closes it

### Requirement: Open a Card's details
`Enter` on the selected Card SHALL open the detail view of its Change. In a column without Cards, `Enter` SHALL do nothing.

#### Scenario: Enter on a Card
- **WHEN** Card `m1-board-view` is selected and the user presses `Enter`
- **THEN** the detail view of `m1-board-view` is shown

#### Scenario: Enter without a Card
- **WHEN** the Board has no Changes and the user presses `Enter`
- **THEN** the Board stays as it is

### Requirement: Edit tasks from the Board
`e` on the selected Card SHALL open the Change's `tasks.md` in the user's editor (`$EDITOR`, `vi` when unset), suspend the Board while the editor runs, and reload the Changes afterwards, keeping the selection by name. Without `tasks.md`, or for an archived Change, `e` SHALL show a short notice instead.

#### Scenario: Tick a task from the Board
- **WHEN** Card `m1-board-view` (5/11) is selected, the user presses `e`, ticks one task and quits the editor
- **THEN** the Board shows `m1-board-view` with `6/11` and it is still selected

#### Scenario: No tasks yet
- **WHEN** a Draft Card without `tasks.md` is selected and the user presses `e`
- **THEN** no editor opens and the Board says the Change has no tasks to edit

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

### Requirement: Filter Cards by name
`/` SHALL open a filter prompt in place of the footer. Every typed character SHALL update the filter at once; `Backspace` SHALL delete the last character and close the prompt when it is empty; `Enter` SHALL close the prompt and keep the filter; `Esc` SHALL close the prompt and clear the filter; `ctrl+c` SHALL close the Board. While the prompt is open, all other keys SHALL be typed into the filter. A Change SHALL match when its name contains the filter, ignoring case; a filter of the form `m<N>` (in either case) SHALL instead match exactly the Changes with milestone badge `M<N>`. While a filter is active, every column SHALL show only matching Cards, its count SHALL be the number of matching Cards, the header SHALL show the filter, and the header's overall Progress SHALL stay that of all open Changes. Navigation SHALL treat columns without matching Cards as empty, and the selection SHALL be kept by Change name when the filter changes, as after a reload. The filter SHALL stay active across reloads and while the detail view is open. With no filter active and no prompt open, `Esc` SHALL close the Board as before; with a filter active, `Esc` SHALL clear it instead.

#### Scenario: Filter by part of a name
- **WHEN** the Project has Changes `m1-board-view`, `m2-change-detail` and `f1-instant-board-open`, and the user types `/board`
- **THEN** only `m1-board-view` and `f1-instant-board-open` are shown, each column's count counts only them, and the header shows the filter `board`

#### Scenario: Case is ignored
- **WHEN** the user types `/BOARD`
- **THEN** the same Cards as for `board` are shown

#### Scenario: Milestone filter
- **WHEN** the Project has Changes `m1-board-view`, `m10-cleanup` and `fix-m1-typo`, and the user types `/m1`
- **THEN** only `m1-board-view` is shown

#### Scenario: Keep the filter
- **WHEN** the user types `/m4`, presses `Enter` and then `j`
- **THEN** the prompt is closed, the filter `m4` stays, and `j` selects the next matching Card

#### Scenario: Clear the filter
- **WHEN** a filter is active and the user presses `Esc`
- **THEN** all Cards are shown again, the header no longer shows a filter, and the Board stays open

#### Scenario: Esc without a filter
- **WHEN** no filter is active and the user presses `Esc`
- **THEN** the Board closes

#### Scenario: Focused column has no match
- **WHEN** Draft is focused and the user types a filter that matches only Cards in In Progress
- **THEN** focus moves to In Progress with one of its matching Cards selected

#### Scenario: Nothing matches
- **WHEN** the filter matches no Change
- **THEN** every column says it is empty, no Card is selected, and the header shows the filter

#### Scenario: Filter survives a reload
- **WHEN** the filter `m4` is active and the user presses `r`
- **THEN** the reloaded Board still shows only Changes matching `m4`

### Requirement: Jump between matches
While a filter is active, `n` SHALL select the next matching Card in reading order (down the focused column, then from the top of the next column with matching Cards) and `N` the previous one; past the last match `n` SHALL wrap to the first, and before the first `N` SHALL wrap to the last. Without an active filter, `n` and `N` SHALL do nothing.

#### Scenario: Next match in the next column
- **WHEN** the filter matches the last Card in Ready and two Cards in Review, the Ready Card is selected, and the user presses `n`
- **THEN** Review is focused with its first matching Card selected

#### Scenario: Wrap around
- **WHEN** the last matching Card on the Board is selected and the user presses `n`
- **THEN** the first matching Card of the leftmost column with matches is selected

#### Scenario: Previous match
- **WHEN** the first matching Card of In Progress is selected, Ready has matches, and the user presses `N`
- **THEN** Ready is focused with its last matching Card selected

#### Scenario: No filter
- **WHEN** no filter is active and the user presses `n`
- **THEN** the Board stays as it is

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

