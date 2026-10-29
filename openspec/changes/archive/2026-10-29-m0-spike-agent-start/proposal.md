## Why

Start is the Board's only action (ADR 0003), and ADR 0005 proposes running it through herdr's agent API, but that flow is only known from `--help` output. Proving it now with a minimal command de-risks M4 and settles ADR 0005 before the Board depends on it. The spike also measures how long `openspec` CLI calls take, which ADR 0002 relies on.

## What Changes

- Add a `kanban start <change>` subcommand that opens a new herdr tab in the Project, starts Claude there as a herdr agent, and submits `/opsx:apply <change>`.
- If an agent for that Change is already running, focus it instead of starting a second one.
- Reject unknown Change names before touching herdr.
- Record the outcome: set ADR 0005 to `accepted` or revise it, and add measured `openspec` CLI timings to ADR 0002.

No keybinding on the Board yet; that is `m4-start-action`.

## Capabilities

### New Capabilities
- `agent-start`: starting an agent session for a Change in its own herdr tab, including duplicate and error handling.

### Modified Capabilities

## Impact

- New: `internal/start/`, `start` subcommand in `cmd/kanban`, extensions to `internal/herdr` (tab create, agent start/prompt/get/focus).
- Docs: ADR 0005 status, ADR 0002 consequences.
- Depends on `m0-plugin-skeleton` (module, herdr client, Project lookup).
