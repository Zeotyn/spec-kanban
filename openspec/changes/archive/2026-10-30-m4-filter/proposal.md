## Why

With many Changes, finding one by scrolling columns is slow. Vim users expect `/` to search.

## What Changes

- `/` opens a filter prompt in the footer; Cards whose name does not match are hidden in all columns as the user types; column counts show the matching Cards.
- `n`/`N` jump to the next/previous matching Card across columns; `Esc` clears the filter (and only closes the Board when no filter is active).
- A milestone filter: typing `m1` matches exactly the Changes with milestone badge `M1`, not `m10-…`.
- The active filter is shown in the header.

## Capabilities

### New Capabilities

### Modified Capabilities
- `board`: filter prompt and match navigation.

## Impact

- `internal/board/` only.
