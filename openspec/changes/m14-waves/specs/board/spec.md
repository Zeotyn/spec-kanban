## ADDED Requirements

### Requirement: Mark Cards for a wave
`space` on the selected Card SHALL mark it for a wave, or unmark it when it is marked. Only Ready Cards SHALL be markable. On any other Card, `space` SHALL mark nothing and show the notice `only Ready cards join a wave`. Marking a Card while `wave.size` Cards are marked SHALL mark nothing and show the notice `a wave holds at most <size> changes`. A marked Card SHALL show `✓` before its name, in the place where a Running agent badge would be. Marks SHALL be kept by Change name across filtering and reloads. A marked Card that leaves Ready on a reload SHALL lose its mark. Marks SHALL be dropped when the Board closes. While Cards are marked, the footer SHALL show `<n> marked · s wave · esc clear` in place of the key footer. `space` in a column without Cards and in the detail view SHALL do nothing. The key help SHALL list `space` under Board.

#### Scenario: Mark two Cards
- **WHEN** Ready Cards `f5-bar` and `c4-foo` are selected one after the other and the user presses `space` on each
- **THEN** both Cards show `✓` before their names and the footer says `2 marked · s wave · esc clear`

#### Scenario: Unmark a Card
- **WHEN** Card `f5-bar` is marked and the user presses `space` on it
- **THEN** `f5-bar` no longer shows `✓`

#### Scenario: Card outside Ready
- **WHEN** the selected Card is in In Progress and the user presses `space`
- **THEN** nothing is marked and the footer says `only Ready cards join a wave`

#### Scenario: Marks survive the filter
- **WHEN** the user filters `/m12`, marks `m12-escalation`, filters `/c4`, marks `c4-foo` and clears the filter
- **THEN** both `m12-escalation` and `c4-foo` are marked

### Requirement: Start a wave from the Board
With 2 or more Cards marked, `s` SHALL not Start the selected Card. Instead the footer SHALL ask `start wave of <n>? y/n`. `y` SHALL start a wave of the marked Changes (see the waves spec): the footer SHALL say `starting wave…`, the Board SHALL ignore keys and mouse events as it does while starting, and the Board SHALL close once the wave is handed over. Any other key SHALL cancel the question and keep the marks. With 0 or 1 Cards marked, `s` SHALL behave as without marks, on the selected Card. When `worktree` is off, `s` with 2 or more marks SHALL start nothing and show the notice `waves need worktree on`. In a Spec Kit Project it SHALL start nothing and show the notice `waves need OpenSpec`. A failure before the Board closes SHALL show as a notice and keep the marks; a failure after it SHALL show as a herdr notification `spec-kanban: wave failed`.

#### Scenario: Start a wave
- **WHEN** Cards `f5-bar`, `m12-escalation` and `c4-foo` are marked and the user presses `s`, then `y`
- **THEN** the footer asks `start wave of 3? y/n`, then says `starting wave…`, a wave of the three Changes is started and the Board closes

#### Scenario: Cancel the question
- **WHEN** the footer asks `start wave of 3? y/n` and the user presses `n`
- **THEN** no wave is started, the three Cards stay marked and the key footer returns

#### Scenario: One mark
- **WHEN** only Card `f5-bar` is marked and Card `c4-foo` is selected, and the user presses `s`
- **THEN** `c4-foo` is started as without marks and no wave is started

#### Scenario: Worktrees off
- **WHEN** `config.json` is `{"worktree": false}`, two Cards are marked and the user presses `s`
- **THEN** nothing is started and the footer says `waves need worktree on`

### Requirement: Wave badge
An open Card whose Change is a member of an active wave (see the waves spec) SHALL show the badge `≋<N>`, where N is the wave's number, before the milestone badge or, without one, at the end of the name line. The badge SHALL be in the dim colour (ANSI bright black, `8`) and SHALL not change the Card's Phase or column. A Change that was dropped from its wave, or whose wave is finished, SHALL show no wave badge.

#### Scenario: Member of wave 2
- **WHEN** Change `m12-escalation` is a member of active wave 2
- **THEN** its Card shows `≋2` before the badge `M12`

#### Scenario: Dropped member
- **WHEN** `c4-foo` was dropped from wave 2
- **THEN** its Card shows no wave badge

## MODIFIED Requirements

### Requirement: Filter Cards by name
`/` SHALL open a filter prompt in place of the footer. Every typed character SHALL update the filter at once; `Backspace` SHALL delete the last character and close the prompt when it is empty; `Enter` SHALL close the prompt and keep the filter; `Esc` SHALL close the prompt and clear the filter; `ctrl+c` SHALL close the Board. While the prompt is open, all other keys SHALL be typed into the filter. A Change SHALL match when its name contains the filter, ignoring case; a filter of the form `m<N>` (in either case) SHALL instead match exactly the Changes with milestone badge `M<N>`. While a filter is active, every column SHALL show only matching Cards, its count SHALL be the number of matching Cards, the header SHALL show the filter, and the header's overall Progress SHALL stay that of all open Changes. Navigation SHALL treat columns without matching Cards as empty, and the selection SHALL be kept by Change name when the filter changes, as after a reload. The filter SHALL stay active across reloads and while the detail view is open. With no prompt open, `Esc` SHALL clear an active filter; with no filter active it SHALL clear the marks when Cards are marked (see "Mark Cards for a wave"); otherwise it SHALL close the Board.

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

#### Scenario: Esc clears the filter, then the marks, then closes
- **WHEN** the filter `m4` is active, Cards `m4-a` and `m4-b` are marked, and the user presses `Esc` three times
- **THEN** the first `Esc` clears the filter and keeps the marks, the second clears the marks, and the third closes the Board
