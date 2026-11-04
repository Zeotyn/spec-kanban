## 1. Source commands

- [x] 1.1 `internal/work`: `Action` (`Verify`, `Archive`) and `Command(action, name) string` on the Source interface ("" = none)
- [x] 1.2 `internal/source/openspec`: `/opsx:verify <name>`, `/opsx:archive <name>`; `internal/source/speckit`: "" for both (tests per scenario of "Verify and archive commands")

## 2. Run a command for a Change

- [x] 2.1 `internal/start`: `Run` reads the Change's agent like `Prepare` (foreign agent refused); a foreground agent with status `idle` gets `agent prompt` and `tab focus`; any other status returns an error carrying it; otherwise `Prepare` places a tab and `Run` returns its pane and tab (tests with fake herdr output for idle, working, blocked, exited, none, foreign)
- [x] 2.2 `cmd/kanban`: subcommand `kanban run <change> <pane> <tab> <kind> <command>` runs `start.Launch` without `Watch`, reports a failure as a herdr notification; usage text updated
- [x] 2.3 `cmd/kanban`: the Board's `act` callback resolves the kind like Start, calls `start.Run`, spawns `kanban run` when a tab was placed, and passes the Source's command without `prompt` or `instructions`

## 3. Board

- [x] 3.1 `v` and `a`: notices outside Review, without the Source's command, and for a busy agent (`<name> is <status>: wait until it is idle`); otherwise the footer says `verifying <name>…` / `archiving <name>…` and the Board ignores input until `act` returns, then closes or shows the error (tests per scenario of "Verify and archive from the Board")
- [x] 3.2 Card menu: entries `✓ Verify  v` and `⇥ Archive  a` below Start/Resume when the Source has the command, faint outside Review; a click or the entry's key runs it; Spec Kit Cards keep one entry; menu edge and click tests extended to three lines
- [x] 3.3 Key help: `v` and `a` under Board; the footer and its goldens stay unchanged

## 4. Docs

- [ ] 4.1 README: Verify and Archive (when available, where they run, no Handoff, merge after Archive in a worktree stays manual); CONTEXT.md: Verify and Archive as Card actions; CHANGELOG `## Unreleased`

## 5. Verification

- [ ] 5.1 `go test -race ./...`, `go vet ./...`, `gofmt -l .` pass; only the help goldens change
- [ ] 5.2 `openspec validate m10-card-actions --strict` passes

## Smoke test (user, in herdr)

These have no checkboxes: they do not count as tasks, and only the user runs them.

- `v` on a Review Card without agent: a new tab in its worktree runs `/opsx:verify <change>`; nothing follows when it finishes
- `a` on a Review Card whose agent is idle: the prompt goes into the existing tab
- `v` while that agent works: notice `… is working: wait until it is idle`
- `v` on an In Progress Card: notice; right-click menu shows Verify and Archive faint there
- Spec Kit spike: right-click on a Review Feature shows only Resume; `v` says `Spec Kit has no verify`
