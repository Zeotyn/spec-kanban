## Context

The rules live entirely in the change-phases spec. Input is `openspec.Change` from `m1-openspec-adapter` (tasks done/total, planning Unknown/Incomplete/Complete, archived flag and date, last modification).

## Goals / Non-Goals

**Goals:**
- A pure package `internal/phase`, no I/O, shared by the Board (M1) and the sidebar (M3).

**Non-Goals:**
- Running agents (M4 shows them as a badge; they never affect the Phase).

## Decisions

### Rule order as a single switch
`Of` is one `switch` in the spec's precedence order. A table of predicates was considered and rejected: six cases read best top to bottom.

### Grouping returns columns, not a map
`Group([]openspec.Change) [5][]openspec.Change`, indexed by `Phase` in column order (Draft … Archived), so the Board renders columns by index without sorting keys.

### Unknown planning counts as Draft
A failed `status` call must not invite a Start on a Change that may be half planned; Draft is the conservative column.

## Risks / Trade-offs

- [Optional artifacts: a Change without `design.md` reports planning incomplete and stays Draft] → follows OpenSpec's own `isPlanningComplete`; teams that skip design get Draft until tasks start. Revisit if it annoys in practice.
