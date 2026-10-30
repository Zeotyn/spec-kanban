## ADDED Requirements

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
