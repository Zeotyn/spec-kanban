## Why

The plugin today is a shell stub that only prints its root. Before any Board feature can land we need the Go program, its build inside herdr, and the path from keybinding to an open Board pane working end to end, so every later change only adds behaviour on top of a proven host integration.

## What Changes

- Add a Go module `github.com/zeotyn/openspec-kanban` with a single `kanban` binary that replaces the `bin/kanban` shell stub.
- Build the binary through a `[[build]]` step in `herdr-plugin.toml`, so `herdr plugin link` / `install` produces it.
- Declare a `board` pane entrypoint and make the `openspec-kanban-open` action open it as a popup for the Project of the pane the action was invoked from.
- Show a placeholder Board: the Project name, or a "no OpenSpec project" notice when the pane is outside a Project; `q` and `Esc` close it.
- Keep the `prefix+o` keybinding snippet in `herdr-config.toml`.
- Set up the test harness (Go tests, golden files, fake herdr client) that later changes build on.

## Capabilities

### New Capabilities
- `plugin-host`: how the plugin is built by herdr, how the open action finds the Project and opens the Board pane, and how the pane closes.

### Modified Capabilities

## Impact

- New: `go.mod`, `cmd/kanban/`, `internal/…`, `.gitignore` for the built binary.
- Changed: `herdr-plugin.toml` (build step, pane entrypoint, action command), `README.md`.
- Removed: the `bin/kanban` shell stub (the built binary takes its path).
- Dependencies: Go toolchain on the machine that links or installs the plugin; `charmbracelet/bubbletea` and `lipgloss`.
