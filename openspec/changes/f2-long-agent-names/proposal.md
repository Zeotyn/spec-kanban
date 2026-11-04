## Why

A Change with a long name cannot be started. `update-a-change-from-its-base-branch` has 36 characters, and Start refuses it with `cannot name a herdr agent: … is longer than 32 characters`. herdr allows at most 32 characters in an agent name, but OpenSpec and Spec Kit allow longer names, so the user has to rename the Change to work on it from the Board.

## What Changes

- **Shortened agent names instead of a refusal.** When a Change's agent name would have more than 32 characters, the agent gets a shortened name: the first 25 characters, a trailing `-` dropped, then `-` and the first 6 hex characters of the SHA-256 of the full Change name.
  - Example: `update-a-change-from-its-base-branch` runs as agent `update-a-change-from-its-302ef9`.
  - The same Change always gets the same agent name, so Start, Resume, Verify, Archive, Watch and the Running badge all find it.
- **Names that fit stay as they are.** A Change of up to 32 characters (with the `f-` prefix for a Spec Kit Feature) keeps today's agent name, so agents that already run are still found.
- **No reverse mapping.** The plugin no longer turns an agent name back into a Change name. The Board asks whether the agent name of each Card is running. This also ends the guess for `f-2fa`, which could be the OpenSpec Change `f-2fa` or the Spec Kit Feature `2fa`.
- The tab, the workspace and the worktree branch keep the full Change name. Only the agent name is shortened.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `agent-start`: the requirement "Start opens an agent session for a Change" shortens a long agent name instead of refusing the Change, and finds an agent by the Change's agent name instead of undoing the prefix.

## Impact

- **Code:**
  - `internal/herdr`: the agent name of a Work item, moved here from `internal/start` and shortened when too long.
  - `internal/start`: uses it everywhere (Prepare, Run, Launch, Resume, Watch); the 32-character refusal and `itemName` go away; `Running` returns agent names.
  - `internal/board`: the Running badge and Resume look up a Card by its agent name.
- **Docs:** README (the 32-character note), CHANGELOG.
- No new dependencies: `crypto/sha256` is in the standard library.
