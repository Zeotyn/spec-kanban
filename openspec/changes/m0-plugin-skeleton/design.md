## Context

The repository holds a manifest with one action and a `bin/kanban` shell stub. ADR 0001 (standalone binary) and ADR 0004 (Go with bubbletea) fix the platform. herdr facts that shape the design (herdr 0.9.5 docs):

- `[[build]]` commands run only on `herdr plugin install`, not on `plugin link`; they get no herdr context.
- Actions are plain commands with `HERDR_PLUGIN_CONTEXT_JSON` (focused pane cwd, workspace cwd, …). Panes are opened through `plugin.pane.open`; a `popup` is session-modal, receives every key including `Esc`, and closes when its command exits.

## Goals / Non-Goals

**Goals:**
- One binary, one module layout that every later milestone extends without restructuring.
- A test harness in place before the first feature: unit tests, golden view files, a fake herdr client.

**Non-Goals:**
- Reading Changes (M1), the sidebar (M3), Start (M0 spike / M4), configurable placement (M5).

## Decisions

### Module and layout
Module path `github.com/zeotyn/openspec-kanban`.

```
cmd/kanban/main.go        subcommand dispatch only
internal/herdr/           thin client around the herdr CLI (interface + exec implementation)
internal/project/         Project lookup: nearest ancestor with openspec/changes
internal/board/           bubbletea model, update, view
testdata/                 fixture Projects and golden views
```

`internal/` keeps the API private; the binary is the product. Alternative `pkg/` was rejected: nobody imports this module.

### Two-step open, as in herdr-openspec
`kanban open` runs as the action: it reads `HERDR_PLUGIN_CONTEXT_JSON`, resolves the Project, and calls `herdr plugin pane open --plugin openspec-kanban --entrypoint board --placement popup --width 90% --height 90% --cwd <project> --env KANBAN_PROJECT=<project> --focus`. The `board` pane entrypoint runs `kanban board`, which reads `KANBAN_PROJECT`. Only the action sees the invoking pane, so it must pick the Project; the pane just renders it. Opening the pane straight from the manifest action is not possible, since actions are commands, not pane openers.

### Binary location and build
The manifest runs `$HERDR_PLUGIN_ROOT/bin/kanban`. `[[build]]` runs `go build -o bin/kanban ./cmd/kanban`; `bin/kanban` is git-ignored. For linked checkouts, `make build` does the same, and README documents it. The shell stub is deleted.

### herdr behind an interface
`internal/herdr.Client` exposes only what is used (`OpenPane` now). The exec implementation calls `HERDR_BIN_PATH` (fallback: `herdr` on `PATH`). Tests use a fake that records calls, so `kanban open` is tested without herdr.

### Tests and TDD
Red-green-refactor for `project` lookup and `open` argument construction. The bubbletea model is tested by sending key messages to `Update` and asserting `tea.Quit`. `View()` is compared to golden files under `testdata/golden/`, rendered with lipgloss in ASCII profile, rewritten with `go test ./... -update`.

## Risks / Trade-offs

- [Users without Go cannot install] → README states the requirement; prebuilt release binaries come with `m5-release-pipeline`.
- [`plugin link` does not build, so stale binaries confuse development] → `make build` plus a README note; the binary prints its version with `kanban --version`.
- [Popup sizing differs between terminals] → percentages, not cell counts; checked manually in the smoke test.
