## Why

The Board's columns are Phases derived from a Change's artifacts and tasks (ADR 0003, CONTEXT.md). The rule is small but has sharp edges (no tasks, planning incomplete, all done), so it gets its own change, built test-first.

## What Changes

- Derive exactly one Phase per Change: Draft, Ready, In Progress, Review, Archived, following the definitions in CONTEXT.md.
- A Change without tasks stays in Draft; a Running agent never changes the Phase.
- Order Cards within a Phase by last modification, newest first.

## Capabilities

### New Capabilities
- `change-phases`: the rules that place each Change in exactly one Phase and order Cards within it.

### Modified Capabilities

## Impact

- New pure package `internal/phase/`, no I/O. Used by the Board and the sidebar.
