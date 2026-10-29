## Why

The user wants to see each workspace's OpenSpec state at a glance in herdr's Space rows without opening the Board: the Most recently modified Change with its Progress, and a short Phase overview. herdr-openspec used to fill these rows; it is disabled now, so its rows in the sidebar are empty.

## What Changes

- New `kanban sync` command: for every herdr workspace it finds the Project its panes work in and reports two Space tokens, the Most recently modified Change with a Progress mini bar (e.g. `▰▰▰▱▱▱ 3/6 m1-board-view`) and the number of open Changes per Phase (e.g. `3 draft · 2 ready · 1 review`).
- Runs on herdr startup and on agent and pane events; bursts of events are coalesced so that at most one extra pass runs.
- Clears the tokens of workspaces outside any Project; keeps them unchanged when a Project cannot be read.
- Reuses the change cache, so a sync usually costs one `openspec list` per Project and keeps the Board's cache warm.
- Config snippet with the two sidebar rows and a bold rule (when the Most recently modified Change is in Review); the README documents copying it.
- This plugin takes over the sidebar from herdr-openspec (ADR 0006).

## Capabilities

### New Capabilities
- `sidebar-summary`: what the plugin writes into Space rows, for which Project, when it refreshes, and how it behaves on errors.

### Modified Capabilities

## Impact

- `herdr-plugin.toml`: one `[[startup]]` hook and `[[events]]` hooks running `kanban sync`.
- New `internal/sidebar/` package; new `sync` subcommand in `cmd/kanban`, sharing the existing loader and cache.
- `herdr-config.toml` and README: sidebar rows and colour rule; users replace the `$openspec`/`$openspec_change` rows of herdr-openspec.
- New ADR `docs/adr/0006-kanban-owns-the-sidebar.md`.
