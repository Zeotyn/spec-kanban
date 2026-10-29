## Why

The Start command exists since the M0 spike but is only reachable from a shell. The Board should start a Change with one key and show which Changes already have a Running agent.

## What Changes

- `s` on a Card starts the Change (agent-start capability), closes the popup, and leaves the user in the new tab.
- Cards show a Running agent badge when herdr reports an agent for that Change; the Phase is unchanged.
- Start is offered for Ready and In Progress Cards; on other Phases `s` shows why not.

## Capabilities

### New Capabilities

### Modified Capabilities
- `board`: `s` key and Running agent badge.
- `agent-start`: Phase gating when started from the Board.

## Impact

- `internal/board/`, `internal/start/`, herdr client (`agent list`).
