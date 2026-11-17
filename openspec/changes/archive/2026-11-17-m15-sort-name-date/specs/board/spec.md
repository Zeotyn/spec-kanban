## ADDED Requirements

### Requirement: Sort the Cards by name or date
`o` on the Board SHALL switch the order of the Cards in Draft, Ready, In Progress and Review through `default`, `name` and `date`, then back to `default`. `default` SHALL be the order the change-phases spec defines. `name` SHALL order Cards by name in natural order, where runs of digits compare as numbers. `date` SHALL order Cards by creation date, newest first, with Cards without a creation date last. Ties SHALL be ordered by name. Archived SHALL keep its archive-date order in every sort. While the sort is not `default`, the footer SHALL show `sort: name` or `sort: date` before the keys. The selection SHALL stay on the same Card when it is still shown, and the sort SHALL apply together with the filter. The sort SHALL be `default` each time the Board opens. In the filter prompt `o` SHALL be typed into the filter. The key help SHALL list `o`.

#### Scenario: Sort by name
- **WHEN** Ready holds `m9-mouse`, `f2-fix` and `c10-docs` and the user presses `o`
- **THEN** Ready shows `c10-docs`, `f2-fix`, `m9-mouse`, and the footer shows `sort: name`

#### Scenario: Sort by date
- **WHEN** In Progress holds `a-one` created 2026-11-10, `b-two` created 2026-11-17 and `c-three` without a creation date, and the user presses `o` twice
- **THEN** In Progress shows `b-two`, `a-one`, `c-three`, and the footer shows `sort: date`

#### Scenario: Back to default
- **WHEN** the sort is `date` and the user presses `o`
- **THEN** the columns show the default order and the footer shows no sort

#### Scenario: Selection follows its Card
- **WHEN** `m9-mouse` is selected in Ready and the user presses `o`
- **THEN** `m9-mouse` is still selected at its new position

#### Scenario: New opening
- **WHEN** the user closes the Board with the sort `name` and opens it again
- **THEN** the sort is `default`

## MODIFIED Requirements

### Requirement: Card content
Each Card SHALL show the Change name, shortened in the middle with `…` when it does not fit. A Change whose name starts with `m<N>-` SHALL show a milestone badge `M<N>`. Open Cards SHALL show a Progress bar with completed/total tasks, or "no tasks", and the age of the last modification (`5m`, `3h`, `2d`, `4w`). Open Cards with a creation date SHALL show it before the age as `MM-DD · <age>`, or `YYYY-MM-DD · <age>` when it is not in the current year. Draft and Ready Cards SHALL show the planning artifacts as `P S D T`, marking each as done, skipped, or missing. Archived Cards SHALL show their archive date instead of Progress and age.

#### Scenario: Card in progress
- **WHEN** Change `m1-board-view` has 5 of 11 tasks completed, was created on 17 November of the current year and was modified 3 hours ago
- **THEN** its Card shows `m1-board-view`, the badge `M1`, a Progress bar with `5/11`, and `11-17 · 3h`

#### Scenario: Created in another year
- **WHEN** an open Change was created on 2025-11-17 and modified 4 weeks ago, and the current year is 2026
- **THEN** its Card shows `2025-11-17 · 4w`

#### Scenario: No creation date
- **WHEN** an open Change has no creation date and was modified 2 days ago
- **THEN** its Card shows `2d` only

#### Scenario: Draft Card
- **WHEN** Change `add-dark-mode` has only a proposal and no tasks
- **THEN** its Card shows "no tasks" and `P` as done while `S D T` are marked missing, and no milestone badge

#### Scenario: Long name in a narrow column
- **WHEN** a Change name is longer than the column is wide
- **THEN** the name is shortened in the middle with `…`, keeping its start and end
