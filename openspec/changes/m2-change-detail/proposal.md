## Why

A Card shows only a summary. To decide what to do next, the user needs the Change's tasks and documents without leaving herdr.

## What Changes

- `Enter` on a Card opens a detail view with tabs Tasks, Proposal, Design, Specs; `1`–`4` and `Tab` switch tabs; `q`/`Esc` returns to the Board with the selection kept.
- Tasks are grouped by their `##` heading with done markers and the next open task highlighted.
- Markdown is rendered for the terminal; nested specs are all shown.
- `o` on a Card toggles an inline task list without leaving the Board.

## Capabilities

### New Capabilities
- `change-detail`: what the detail view shows for a Change and how it is navigated.

### Modified Capabilities
- `board`: `Enter` and `o` open details.

## Impact

- New `internal/detail/`; markdown rendering dependency (glamour or similar, decided in design).
