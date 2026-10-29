## Why

With many Changes, finding one by scrolling columns is slow. Vim users expect `/` to search.

## What Changes

- `/` opens a filter prompt; Cards whose name does not match are hidden in all columns; counts update.
- `n`/`N` jump to the next/previous match; `Esc` clears the filter.
- A milestone filter: typing `m1` matches the milestone prefix.

## Capabilities

### New Capabilities

### Modified Capabilities
- `board`: filter prompt and match navigation.

## Impact

- `internal/board/` only.
