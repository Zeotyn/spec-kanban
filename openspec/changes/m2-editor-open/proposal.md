## Why

Small fixes to a Change's documents (ticking a task, rewording a proposal) should not need a separate terminal.

## What Changes

- `e` opens the file behind the current detail tab (or `tasks.md` on the Board) in `$EDITOR`, suspending the Board and reloading the Change afterwards.
- Archived Changes are read-only: `e` shows a notice instead.

## Capabilities

### New Capabilities

### Modified Capabilities
- `change-detail`: adds editing of the current document.
- `board`: adds `e` on a Card.

## Impact

- `internal/detail/`, `internal/board/`; relies on `$EDITOR` (fallback `vi`).
