## Context

ADR 0002 makes the `openspec` CLI the data source. Measured in M0 (ADR 0002): every call costs about 0.65 s; `status` for 9 Changes takes 5.95 s sequentially and 1.45 s with 8 in parallel. `kanban start` currently calls `openspec list --json` directly (`internal/start`). JSON shapes seen with openspec 1.8.0:

- `list --json`: `{"changes":[{"name","completedTasks","totalTasks","lastModified","status"}]}`, `status` one of `no-tasks`, `in-progress`, `complete`.
- `status --change X --json`: `isPlanningComplete` and `artifacts[{id,status}]` with `done`, `ready`, `blocked`, `skipped`; exit 1 with an error JSON for an unknown Change.

## Goals / Non-Goals

**Goals:**
- One package that turns a Project into a list of Changes, ready for Phase derivation and the Board.
- Keep the number of `status` calls minimal.

**Non-Goals:**
- Caching, file watching, task lists and documents (M2), Phase rules (`m1-phase-derivation`).

## Decisions

### One entry point: `Load`
`internal/openspec` exposes a `Change` struct and `Load(project string, run herdr.Runner) ([]Change, error)`, where `run` executes `openspec` inside the Project (main wires `herdr.Run("openspec", project, …)`, as `kanban start` does today). `Load` calls `list`, then `status` for Changes with `completedTasks == 0`, then reads the archive directory. Separate public `List`/`Status` functions are not needed by any caller yet.

```go
type Change struct {
    Name       string
    Done, Total int
    Modified   time.Time
    Planning   Planning           // Unknown, Incomplete, Complete; Unknown when not read or failed
    Artifacts  map[string]string  // artifact id → done|ready|blocked|skipped, only when read
    Archived   bool
    ArchivedOn string             // "2026-10-29"; "" when open or the folder has no date
}
```

### Status only where it decides something
Planning state matters only for Changes without completed tasks (Draft vs Ready, and the P S D T badges on those Cards). Changes with progress skip `status`, which on magentic cuts 9 calls to 3.

### Bounded parallelism
`status` calls run with at most 8 in flight (`errgroup` is not needed: a buffered channel as semaphore plus a `sync.WaitGroup`). A failing call sets that Change's `Planning` to `Unknown` and is otherwise ignored, per the spec.

### `kanban start` shares the list parsing
`internal/start.checkOpen` currently parses `list --json` itself. It switches to a small `openspec.OpenNames(run)` helper sharing the same list parsing, so the JSON shape lives in one place. `Load` uses it too.

### CLI not found
`exec.ErrNotFound` from the runner is turned into "openspec CLI not found: install it with npm i -g @fission-ai/openspec".

### Tests
- Unit (TDD): a fake runner returning canned JSON covers list parsing, status only for 0-done Changes, Unknown on a failing status, CLI not found, archive folder parsing (dated, undated, missing dir) with `t.TempDir()`.
- Contract: `testdata/project/` is a real OpenSpec project with one Change per case (proposal only, skip_specs, tasks none done, some done, all done, no tasks.md, nested specs) and one dated archive folder. A test runs the real CLI against it and checks the parsed result; it calls `t.Skip` when `openspec` is not on `PATH`.

## Risks / Trade-offs

- [openspec changes its JSON] → the contract test fails in CI (`m5-release-pipeline` installs openspec), not in front of the user.
- [Load time grows with many fresh Changes] → bounded by 8 parallel calls, about 1.5 s for 9; a cache keyed by `lastModified` can come later if it hurts.
- [`lastModified` in `list` is whole-change, not per artifact] → fine for ordering and age.
