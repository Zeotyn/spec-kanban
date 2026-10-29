## 1. Summary text (TDD)

- [x] 1.1 Write failing table tests for `sidebar.Summary`: Change in progress (`▰▰▰▱▱▱ 3/6 m1-board-view`), 29/30 not full, all done full, no tasks, long name shortened to 28, newest `Modified` wins with ties by name, archived Changes ignored, no open Change (empty change, `no open changes`), Phase overview with zeros omitted, four Phases dropping Draft to fit 28
- [x] 1.2 Implement `Summary` (six-cell floor bar, `board.FitName`, `phase.Of` counts) until the tests pass

## 2. One pass (TDD, fake `herdr.Runner` and fake loader)

- [x] 2.1 Write failing tests for `sidebar.Sync`: workspace mapped by `foreground_cwd` (fallback `cwd`); focused pane wins over another Project; workspace without Project gets both tokens cleared; no open Change reports `kanban` and clears `kanban_change`; shared Project loaded once and reported to both workspaces; failed load leaves its workspaces untouched while others are reported and the error is returned for the log; `workspace list` or `pane list` failure reports nothing and returns herdr's error
- [x] 2.2 Write a failing test that at most 4 Project loads run at once
- [x] 2.3 Implement `Sync` until the tests pass

## 3. Coalescing (TDD)

- [x] 3.1 Write failing tests for `sidebar.Coalesce` in a temp dir: free lock runs one pass; lock held elsewhere touches `sync.pending` and returns without a pass; pending created during a pass causes exactly one more pass; several requests during a pass still cause only one; pass error is returned
- [x] 3.2 Implement `Coalesce` with `syscall.Flock` until the tests pass

## 4. Wiring

- [x] 4.1 Add `kanban sync` to `cmd/kanban` (usage text included): `Coalesce(cache.Dir(), …)` around `Sync(herdr.Exec, load)` with `load` built from the existing `loader`; test that a sync load writes the cache the Board reads
- [x] 4.2 Add `[[startup]]` and `[[events]]` (`pane.agent_status_changed`, `pane.focused`, `pane.created`, `workspace.created`) running `kanban sync` to `herdr-plugin.toml`
- [x] 4.3 Add the sidebar width, rows and Review colour rule to `herdr-config.toml`
- [x] 4.4 README: sidebar section with the two rows, the `wip` label, the bold rule, and replacing herdr-openspec's `$openspec` rows

## 5. Verification

- [x] 5.1 `go test -race ./...` and `go vet ./...` pass
- [x] 5.2 Smoke test in herdr: rebuild, apply the snippet to `~/.config/herdr/config.toml`, reload; workspaces in this repo and in magentic show both lines, a workspace outside a Project shows none; tick a task via an agent turn and see the count update; finish a Change's tasks and see the line turn bold; `herdr plugin log list` shows no errors and a burst of focus changes runs at most two passes
