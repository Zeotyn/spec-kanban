## Why

This is the Board itself: Changes in Phase columns, navigable with vim keys, readable in a popup on wide and narrow terminals. After this change the Board is usable day to day.

## What Changes

- Five columns (Draft, Ready, In Progress, Review, Archived) with Card counts and an overall Progress bar in the header.
- Cards show Change name, Progress bar with done/total, artifact badges (P S D T), age, and a milestone badge (`M1`) parsed from the `m<N>-` name prefix.
- Keys: `h`/`l` column, `j`/`k` Card, `gg`/`G`, `ctrl-d`/`ctrl-u`, `1`–`5` jump to column, `r` reload, `q`/`Esc` close.
- Narrow terminals show one column at a time with a column indicator.
- Cards load from `list` first; planning badges fill in as `status` results arrive.

## Capabilities

### New Capabilities
- `board`: layout, Card content, keyboard navigation, and loading behaviour of the Board.

### Modified Capabilities
- `plugin-host`: the placeholder Board is replaced by the real Board.

## Impact

- `internal/board/` grows into the main UI; golden views at 60, 120, and 200 columns.
