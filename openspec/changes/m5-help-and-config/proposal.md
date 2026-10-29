## Why

Keys are only discoverable from the README, and the popup placement is fixed. Before a first release the Board needs in-app help and the few settings users will want.

## What Changes

- `?` shows all keys grouped by view.
- A config file in herdr's plugin config dir with: placement (`popup` or `tab`), popup size, Phases to hide (e.g. Archived), and agent kind for Start.
- Invalid config falls back to defaults with a visible warning.

## Capabilities

### New Capabilities
- `kanban-config`: the user settings, their defaults, and invalid-config handling.

### Modified Capabilities
- `board`: help overlay.
- `plugin-host`: placement and size come from config.

## Impact

- New `internal/config/`; README settings section.
