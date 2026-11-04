# change-phases Specification

## Purpose

Defines how each Change is placed in exactly one Phase (one Board column) and how Cards are ordered within a Phase, so the Board and the sidebar agree on where every Change stands.
## Requirements
### Requirement: Every Change has exactly one Phase
Each Change SHALL be placed in exactly one Phase by the first matching rule:

1. archived → Archived
2. no tasks → Draft
3. all tasks completed → Review
4. some, but not all, tasks completed → In Progress
5. no task completed and planning complete → Ready
6. no task completed and planning incomplete or unknown → Draft

Progress therefore outranks planning: once a task is completed, missing planning artifacts no longer move a Change back to Draft.

#### Scenario: Archived Change
- **WHEN** a Change is archived, whatever its tasks
- **THEN** its Phase is Archived

#### Scenario: No tasks
- **WHEN** an open Change has a complete plan but 0 tasks
- **THEN** its Phase is Draft

#### Scenario: All tasks done
- **WHEN** an open Change has 5 of 5 tasks completed
- **THEN** its Phase is Review

#### Scenario: Work in progress with incomplete planning
- **WHEN** an open Change has 2 of 5 tasks completed and no `design.md`
- **THEN** its Phase is In Progress

#### Scenario: Fully planned, not started
- **WHEN** an open Change has complete planning and 0 of 4 tasks completed
- **THEN** its Phase is Ready

#### Scenario: Planning not finished
- **WHEN** an open Change has 0 of 4 tasks completed and its planning is incomplete
- **THEN** its Phase is Draft

#### Scenario: Planning state unknown
- **WHEN** an open Change has 0 of 4 tasks completed and its planning state could not be read
- **THEN** its Phase is Draft

### Requirement: A running agent does not change the Phase
The Phase SHALL depend only on the Change's archive state, tasks and planning, never on whether an agent is working on it.

#### Scenario: Agent still working after the last task
- **WHEN** all tasks of a Change are completed and its agent is still running
- **THEN** its Phase is Review

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

