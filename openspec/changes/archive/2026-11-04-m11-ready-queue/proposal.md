## Why

Ready is a queue: it answers "which Change do I start next with `s`?". Today it is sorted like the other open Phases, by last modification, newest first. That order does not fit a queue:
- A small edit, such as a typo fix in a proposal, moves a Card to the top.
- After a clone or checkout all files have about the same time, and the name tie-break puts `m10-…` before `m2-…`.

## What Changes

- Ready gets its own order, as a queue:
  1. Fixes first: names `f<digits>-…`, by number ascending.
  2. Then milestones: names `m<digits>-…`, by number ascending (`m9` before `m10` before `m11`).
  3. Then everything else (`c<N>-…` special changes, Spec Kit Features such as `001-…`, unprefixed names such as `add-dark-mode`) in natural name order: runs of digits compare as numbers, so `2` comes before `10`.
  - Ties are ordered by name. A name like `fix-login` is not a fix, since no digit follows the `f`.
- Draft, In Progress and Review stay newest first. Archived stays as it is.
- The Card's age line stays as it is.
- No cache version bump: the Phase values do not change.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `change-phases`: the requirement "Cards are ordered within a Phase" gets the queue order for Ready.

## Impact

- Code: `internal/work` (`Group` and a small name-key helper with tests). The Board and the sidebar need no change: the Board takes the order from `Group`, and the sidebar picks its Change by modification time on its own.
- Docs: CHANGELOG `## Unreleased`. The README does not describe the Card order, so it stays.
- Dependencies: none.
