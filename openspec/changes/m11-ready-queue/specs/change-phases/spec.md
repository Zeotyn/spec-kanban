## MODIFIED Requirements

### Requirement: Cards are ordered within a Phase
Within Draft, In Progress and Review, Cards SHALL be ordered by last modification, newest first. Within Ready, Cards SHALL be ordered as a queue, independent of their modification time: first the fixes, named `f<digits>-…`, by number ascending; then the milestones, named `m<digits>-…`, by number ascending; then all other Cards in natural name order, where runs of digits compare as numbers. A name counts as a fix or a milestone only when the `f` or `m` is followed by at least one digit and then `-`. Within Archived, Cards SHALL be ordered by archive date, newest first, with undated Cards last. In every Phase, ties are ordered by name.

#### Scenario: Two Draft Changes
- **WHEN** Draft Change `a` was last modified yesterday and `b` an hour ago
- **THEN** `b` comes before `a`

#### Scenario: Two Ready Changes
- **WHEN** Ready Change `m2-board-view` was last modified yesterday and `m3-filter` an hour ago
- **THEN** `m2-board-view` comes before `m3-filter`

#### Scenario: Fixes before milestones
- **WHEN** Ready holds `m1-board-view`, `f2-cache-race` and `f1-instant-board-open`
- **THEN** the order is `f1-instant-board-open`, `f2-cache-race`, `m1-board-view`

#### Scenario: Milestones by number
- **WHEN** Ready holds `m10-card-actions`, `m9-mouse` and `m11-ready-queue`, all with the same modification time
- **THEN** the order is `m9-mouse`, `m10-card-actions`, `m11-ready-queue`

#### Scenario: Other Changes last in natural order
- **WHEN** Ready holds `m1-board-view`, `c10-cleanup`, `fix-login`, `c2-logo` and `add-dark-mode`
- **THEN** the order is `m1-board-view`, `add-dark-mode`, `c2-logo`, `c10-cleanup`, `fix-login`

#### Scenario: Spec Kit Features in Ready
- **WHEN** Ready holds the Features `010-export`, `002-search` and `001-user-login`
- **THEN** the order is `001-user-login`, `002-search`, `010-export`

#### Scenario: Archived on the same day
- **WHEN** `m0-plugin-skeleton` and `m0-spike-agent-start` were both archived on 2026-10-29
- **THEN** `m0-plugin-skeleton` comes before `m0-spike-agent-start`
