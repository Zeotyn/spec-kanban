## Why

The Start command exists since the M0 spike but is only reachable from a shell. The Board should start a Change with one key and show which Changes already have a Running agent.

## What Changes

- `s` on a Card starts the Change (agent-start capability), closes the popup, and leaves the user in the agent's tab.
- Cards show a Running agent badge when herdr reports an agent for that Change in this Project; the Phase is unchanged.
- Start is offered for Ready and In Progress Cards; on other Phases `s` shows why not.
- Start no longer focuses another Project's agent that happens to carry the same name.
- `kanban start <change>` keeps working as before.

## Capabilities

### New Capabilities

### Modified Capabilities
- `board`: `s` key with Phase gating, Start progress and errors in the footer, Running agent badge.
- `agent-start`: Start refuses an agent of the same name that works in another Project.

## Impact

- `internal/board/`, `internal/start/` (Running agent lookup via `herdr agent list`, project check), `cmd/kanban/` (wiring).
