## Why

Keys are only discoverable from the README, and Start always sends `/opsx:apply <change>`, so users cannot hand their agent their own way of working (TDD, review loops, subagents). Before a first release the Board needs in-app help and the few settings users actually ask for.

## What Changes

- `?` shows all keys grouped by view (Board, filter prompt, Details), on the Board and in the detail view; any key closes it.
- An optional config file `config.json` in herdr's plugin config dir with two settings:
  - `prompt`: the text Start sends, a template where `{change}` is replaced by the Change name; default `/opsx:apply {change}`.
  - `hide`: Phases whose columns the Board does not show (e.g. `["Archived"]`); default none.
- An invalid config falls back to all defaults; the Board shows a warning in its footer when it opens, and `kanban start` prints it.
- Cut from the original scope: placement (`popup`/`tab`), popup size and agent kind (see design.md, Non-Goals).

## Capabilities

### New Capabilities
- `kanban-config`: where the config file lives, its settings and defaults, and how an invalid config is handled.

### Modified Capabilities
- `board`: help overlay; hidden Phases are not shown as columns.
- `agent-start`: the prompt Start sends comes from the config.

## Impact

- New `internal/config/` (stdlib `encoding/json`, no new dependency).
- `internal/start/`: `Launch` sends a given prompt; `kanban launch` receives it from the Board.
- `internal/board/`: help view, hidden columns, initial notice; footers of Board and detail view mention `?`.
- `cmd/kanban/main.go`: loads the config for `board` and `start`.
- README: `?` in the keys table, a Settings section.
