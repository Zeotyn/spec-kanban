## Why

Opening the Board shows "Loading changes…" for about 1.4 s on top of herdr's ~0.8 s popup start, which reads like a hang or a bug. Every openspec CLI call starts Node (~0.65 s); `list` plus the parallel `status` calls cannot get much faster, so the Board has to stop waiting for them before it shows anything.

## What Changes

- Keep a per-Project snapshot of the last successfully loaded Changes and show it immediately when the Board opens, marked as refreshing, while a fresh load runs; swap in the fresh result when it arrives.
- Reuse a Change's planning state from the cache when its last modification time is unchanged, so a refresh usually needs only `openspec list`.
- Validate the cache before using any of it: format version, Project path, readable and complete content, and per-Change modification time; anything invalid is ignored and replaced after the next successful load.
- Write the cache atomically after every successful load; cache read or write failures never break the Board.
- First open of a Project (no valid cache) behaves as today.

## Capabilities

### New Capabilities
- `change-cache`: where the last load is kept, when cached data may be used, how it is validated and refreshed.

### Modified Capabilities
- `board`: the loading notice appears only when no valid cached snapshot exists; otherwise the cached Board shows at once with a refreshing marker.

## Impact

- New `internal/cache/` package; `kanban board` wires it between the Board and `openspec.Load`.
- `openspec.Load` gains a way to take known planning states so it can skip `status` calls.
- Files under `$HERDR_PLUGIN_STATE_DIR` (fallback: the user cache directory), one per Project.
