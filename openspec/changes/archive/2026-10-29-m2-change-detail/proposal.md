## Why

A Card shows only a summary. To decide what to do next, the user needs the Change's tasks and documents without leaving herdr.

## What Changes

- `Enter` on a Card opens a detail view with tabs Tasks, Proposal, Design, Specs; `1`–`4` and `Tab` switch tabs; `q`/`Esc` returns to the Board with the selection kept.
- The Tasks tab shows the next open task on top, then `tasks.md` rendered with its `##` groups and done markers.
- Markdown is rendered for the terminal and scrolls with vim keys; nested specs are all shown, each under its capability path.
- Missing documents say so instead of an empty tab; archived Changes open read-only from the archive folder.

## Capabilities

### New Capabilities
- `change-detail`: what the detail view shows for a Change and how it is navigated.

### Modified Capabilities
- `board`: `Enter` opens a Card's details.

## Impact

- New `internal/detail/`; new dependency glamour for markdown rendering.
- An inline task list on the Board (`o`) was dropped: the detail view covers it.
