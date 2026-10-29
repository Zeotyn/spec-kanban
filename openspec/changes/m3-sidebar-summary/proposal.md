## Why

The user wants to see each workspace's OpenSpec state at a glance in herdr's Space rows without opening the Board: the Most recently modified Change with its Progress, and a short Phase overview.

## What Changes

- `kanban sync` reports two tokens per workspace: the Most recently modified Change with a Progress mini bar, and counts per Phase (e.g. `3 open · 1 review · 2 draft`).
- Runs on herdr startup and on agent and pane events, coalescing bursts so only one extra pass runs.
- Clears the tokens for workspaces outside any Project.
- Config snippet with sidebar rows and colour rules (green when a Change reaches Review).
- Decide whether this plugin replaces herdr-openspec's sidebar tokens (ADR to follow).

## Capabilities

### New Capabilities
- `sidebar-summary`: what the plugin writes into Space rows and when it refreshes.

### Modified Capabilities

## Impact

- `[[startup]]` and `[[events]]` in the manifest, new `internal/sync/`, `herdr-config.toml` sidebar section; possible ADR on sidebar ownership.
