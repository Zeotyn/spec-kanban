## Why

Keys are only discoverable from the README, and the popup placement is fixed. Before a first release the Board needs in-app help and the few settings users will want.

## What Changes

- `?` shows all keys grouped by view.
- A config file in herdr's plugin config dir with: placement (`popup` or `tab`), popup size, Phases to hide (e.g. Archived), agent kind for Start, and the prompt Start sends (a template with `{change}`, default `/opsx:apply {change}`), so users can add their own way of working such as TDD or review loops.
- Invalid config falls back to defaults with a visible warning.

## Capabilities

### New Capabilities
- `kanban-config`: the user settings, their defaults, and invalid-config handling.

### Modified Capabilities
- `board`: help overlay.
- `plugin-host`: placement and size come from config.
- `agent-start`: the prompt comes from config.

## Impact

- New `internal/config/`; `internal/start/` takes the prompt; README settings section.
