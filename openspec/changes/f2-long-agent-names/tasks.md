## 1. Agent name

- [x] 1.1 `internal/herdr`: `AgentName(item)` moved from `internal/start`; a name over 32 characters becomes its first 25 characters, a trailing `-` trimmed, `-` and the first 6 hex characters of `sha256(item)` (tests: `update-a-change-from-its-base-branch` → `update-a-change-from-its-302ef9`, `001-a-feature-name-far-too-long-for-herdr` → `f-001-a-feature-name-far-8b5dbb`, `add-a-change-of-thirty-two-chars` unchanged, `001-user-login-email` → `f-001-user-login-email`, `m1-board-view` unchanged, every result at most 32 characters, same input gives the same result)

## 2. Start

- [x] 2.1 `internal/start`: every herdr call uses `herdr.AgentName` (Prepare, Run, Launch, Resume, Watch); `agentName`, `itemName` and the 32-character refusal in `prepare` go away; `TestPrepareNameTooLong` becomes a test that starting `update-a-change-from-its-base-branch` runs `agent get`, `agent start` and `agent prompt` with `update-a-change-from-its-302ef9` and labels the tab with the full name (scenario "Name too long for herdr")
- [x] 2.2 Test that a second Start of that Change, while its shortened agent holds the foreground, focuses it and creates nothing (scenario "Same long name, same agent")
- [x] 2.3 `Running` returns the agent names herdr reports for the Project, without `itemName`; `TestRunningFeatureName` expects `{"f-001-a", "f-x"}`

## 3. Board

- [x] 3.1 `internal/board`: the Running badge and Resume on `s` look up `m.running[herdr.AgentName(c.Name)]` (tests: a Card `update-a-change-from-its-base-branch` with running agent `update-a-change-from-its-302ef9` shows the badge and resumes; a Card `001-a` with running agent `f-001-a` shows the badge; scenario "A shortened agent is found again")

## 4. Docs

- [x] 4.1 README: replace "A name whose agent name would exceed herdr's 32 characters cannot be started" with the shortening rule and its example
- [x] 4.2 CHANGELOG `## Unreleased`: under Fixed, a Change or Feature whose agent name would exceed 32 characters starts with a shortened agent name instead of being refused

## 5. Verification

- [x] 5.1 `go test -race ./...`, `go vet ./...`, `gofmt -l .` pass
- [x] 5.2 `openspec validate f2-long-agent-names --strict` passes

## Smoke test (user, in herdr)

These have no checkboxes: they do not count as tasks, and only the user runs them.

- Create the Change `update-a-change-from-its-base-branch` with tasks, commit it, press `s` on its Card: a workspace labelled with the full name opens on branch `update-a-change-from-its-base-branch`, and `herdr agent list` shows `update-a-change-from-its-302ef9`
- The long label reads fine in the tab strip and the sidebar
- Reopen the Board while the agent works: the Card shows the Running badge
- Press `s` on the Card again: the agent's tab is focused, nothing new opens
- A Change with a short name still starts and shows its badge as before
