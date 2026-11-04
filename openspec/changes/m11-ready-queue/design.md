## Context

`work.Group` in `internal/work/work.go` sorts each column: open Phases by `Modified`, newest first; Archived by `ArchivedOn`, newest first, undated last; ties by name. The Board gets its columns only from `Group`. The sidebar picks its Change by `Modified` on its own and does not use `Group`.

`internal/board/format.go` already parses `m<digits>-` in `milestone()` for the Card badge. `internal/board` imports `internal/work`, so `work` cannot call it.

See proposal.md for why Ready needs a queue, and the `change-phases` delta for the exact order.

## Goals / Non-Goals

**Goals:**
- The queue order lives in `internal/work`, next to the other orders, and is tested there.
- Standard library only, smallest diff.

**Non-Goals:**
- A config key to choose the order per Phase.
- Changing the order of Draft, In Progress, Review or Archived.
- Changing the Card's age line or the milestone badge.
- A shared name parser for `work` and `board`. The two checks are a few lines each, and moving the badge code is not needed for this change.

## Decisions

1. **A rank and a number from the name.** A small unexported helper returns `(rank, n, ok)` for a name: rank 0 for `f<digits>-`, rank 1 for `m<digits>-`, rank 2 for everything else. The Ready comparator orders by rank, then by number (ranks 0 and 1), then by natural name order (rank 2), then by name. The `f`/`m` check mirrors `milestone()`: a letter, at least one digit, then `-`. So `fix-login` and `m-foo` fall into rank 2.

2. **Natural order compares digit runs without parsing to int.** Walk both names; outside digits compare bytes; for two digit runs strip leading zeros, compare by length, then by bytes. This never overflows, also for Spec Kit timestamp names such as `20261030-142501-user-login`. When digit runs are equal in value but differ in leading zeros (`m01-` vs `m1-`), the final tie by name decides.
   - Rejected: `strconv.Atoi` on the digit runs. It fails on very long runs and needs an error path.

3. **The milestone and fix number use the same digit-run compare** as decision 2, so `m010-` and `m10-` sort together and no int parsing is needed at all.

4. **No cache version bump.** The cache stores Items with their Phase, not the order; `Group` sorts after every load.

## Risks / Trade-offs

- [An OpenSpec project that does not use the `f`/`m` prefixes sees only natural name order in Ready] → That is still a stable, predictable queue, better than modification time for a queue.
- [A user who relied on "recently edited first" in Ready loses it] → The age line still shows how old each Card is, and Draft, In Progress and Review keep newest first.
- [Upper case names (`M2-…`) are not milestones] → Change names are lowercase kebab-case by project rule; upper case names fall into natural order.

## Migration Plan

None. The new order shows with the new binary; nothing is stored.
