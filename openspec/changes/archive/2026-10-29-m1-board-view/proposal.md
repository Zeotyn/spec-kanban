## Why

This is the Board itself: Changes in Phase columns, navigable with vim keys, readable in a popup on wide and narrow terminals. After this change the Board is usable day to day.

## What Changes

- Five columns (Draft, Ready, In Progress, Review, Archived) with Card counts and an overall Progress bar in the header.
- Cards show Change name, Progress bar with done/total, age, a milestone badge (`M1`) parsed from the `m<N>-` name prefix, and artifact badges (P S D T) on Draft and Ready Cards.
- Keys: `h`/`l` column, `j`/`k` Card, `gg`/`G`, `ctrl-d`/`ctrl-u`, `1`–`5` jump to column, `r` reload, `q`/`Esc` close.
- Narrow terminals show one column at a time with a column indicator.
- A loading notice shows until the Changes are read; `r` reloads and keeps the selection.
- A missing or failing `openspec` CLI is shown in place of the columns.

## Capabilities

### New Capabilities
- `board`: layout, Card content, keyboard navigation, and loading behaviour of the Board.

### Modified Capabilities

## Impact

- `internal/board/` grows into the main UI; golden views at 60, 120, and 200 columns.
- New dependency: lipgloss for column layout.
