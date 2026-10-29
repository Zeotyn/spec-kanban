## 1. Module and build

- [x] 1.1 Create `go.mod` for `github.com/zeotyn/openspec-kanban` and add bubbletea (lipgloss deferred until M1 needs layout)
- [x] 1.2 Add `cmd/kanban/main.go` with subcommand dispatch (`open`, `board`, `--version`, `--help`) and an exit code 2 for unknown subcommands
- [x] 1.3 Document the build command (`go build -o bin/kanban ./cmd/kanban`) in README and git-ignore `bin/`
- [x] 1.4 Delete the `bin/kanban` shell stub

## 2. Project lookup (TDD)

- [x] 2.1 Write failing tests for `internal/project`: cwd inside a Project, cwd is the Project root, no Project above cwd, symlinked cwd
- [x] 2.2 Implement the nearest-ancestor lookup until the tests pass

## 3. herdr client and open action (TDD)

- [x] 3.1 Define the `herdr.Runner` func type so tests can record calls
- [x] 3.2 Write failing tests for `kanban open`: Project from focused pane cwd, fallback to workspace cwd, no Project, herdr error is returned
- [x] 3.3 Implement context parsing and the `plugin pane open` call until the tests pass
- [x] 3.4 Implement `herdr.Exec` using `HERDR_BIN_PATH` with `herdr` on `PATH` as fallback

## 4. Placeholder Board (TDD)

- [x] 4.1 Set up the golden-file helper with the `-update` flag
- [x] 4.2 Write failing tests: `q` and `Esc` quit, view with a Project shows its name, view without a Project shows the notice
- [x] 4.3 Implement the bubbletea model and `kanban board` until the tests pass

## 5. Manifest and docs

- [x] 5.1 Update `herdr-plugin.toml`: `[[build]]` step, `board` pane entrypoint, action command `kanban open`
- [x] 5.2 Keep the `prefix+o` snippet in `herdr-config.toml`
- [x] 5.3 Update `README.md`: requirements (herdr ≥ 0.9.3, Go), install, build for linked checkouts, keybinding

## 6. Verification

- [x] 6.1 `go test ./...` and `go vet ./...` pass
- [x] 6.2 Smoke test in herdr: link, build, `prefix+o` inside this repo shows the Board, outside shows the notice, `q` and `Esc` close it
