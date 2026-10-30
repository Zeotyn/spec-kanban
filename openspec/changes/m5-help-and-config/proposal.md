## Why

Keys are only discoverable from the README, and Start always starts Claude with `/opsx:apply <change>`, so users cannot hand their agent their own way of working (TDD, review loops, subagents) or use another agent herdr supports. Before a first release the Board needs in-app help and the few settings users actually ask for.

## What Changes

- `?` shows all keys grouped by view (Board, filter prompt, Details), on the Board and in the detail view; any key closes it.
- Start launches the agent kind that runs in the pane the Board was opened from (herdr reports it), `claude` when that pane runs no agent or herdr cannot tell.
- An optional config file `config.json` in herdr's plugin config dir with three settings:
  - `agent`: overrides the detected kind, one of herdr's `agent start --kind` values; default unset (detect).
  - `prompt`: the text Start sends, a template where `{change}` is replaced by the Change name; default `/opsx:apply {change}`.
  - `hide`: Phases whose columns the Board does not show (e.g. `["Archived"]`); default none.
- An invalid config falls back to all defaults; the Board shows a warning in its footer when it opens, and `kanban start` prints it.
- Start checks whether the agent still runs by comparing the pane's foreground process with the agent's kind instead of the fixed `claude`.
- Cut from the original scope: placement (`popup`/`tab`) and popup size (see design.md, Non-Goals).

## Capabilities

### New Capabilities
- `kanban-config`: where the config file lives, its settings and defaults, and how an invalid config is handled.

### Modified Capabilities
- `board`: help overlay; hidden Phases are not shown as columns.
- `agent-start`: Start uses the invoking pane's agent kind or the configured one, and the configured prompt; the still-running check works for every kind.

## Impact

- New `internal/config/` (stdlib `encoding/json`, no new dependency).
- `internal/start/`: `Launch` starts a given kind and sends a given prompt; `kanban launch` receives both from the Board; `start.Kind` resolves the kind; `agentRunning` matches the kind herdr reports instead of `claude`.
- `internal/board/`: help view, hidden columns, initial notice; footers of Board and detail view mention `?`.
- `cmd/kanban/main.go`: loads the config for `board` and `start`; `kanban open` passes the invoking pane to the Board as `KANBAN_MAIN_PANE`.
- README: `?` in the keys table, a Settings section; texts that say Claude say "the agent" where any kind is meant.
