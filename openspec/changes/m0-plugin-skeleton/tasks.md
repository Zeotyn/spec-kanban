## 1. Module and build

- [ ] 1.1 Create `go.mod` for `github.com/zeotyn/openspec-kanban` and add bubbletea and lipgloss
- [ ] 1.2 Add `cmd/kanban/main.go` with subcommand dispatch (`open`, `board`, `--version`, `--help`) and an exit code 2 for unknown subcommands
- [ ] 1.3 Add `Makefile` with `build` (`go build -o bin/kanban ./cmd/kanban`) and `test` targets; git-ignore `bin/kanban`
- [ ] 1.4 Delete the `bin/kanban` shell stub

## 2. Project lookup (TDD)

- [ ] 2.1 Write failing tests for `internal/project`: cwd inside a Project, cwd is the Project root, no Project above cwd, symlinked cwd
- [ ] 2.2 Implement the nearest-ancestor lookup until the tests pass

## 3. herdr client and open action (TDD)

- [ ] 3.1 Define the `herdr.Client` interface and a recording fake
- [ ] 3.2 Write failing tests for `kanban open`: Project from focused pane cwd, fallback to workspace cwd, no Project, herdr error is written to stderr with a non-zero exit
- [ ] 3.3 Implement context parsing and the `plugin pane open` call until the tests pass
- [ ] 3.4 Implement the exec-based client using `HERDR_BIN_PATH` with `herdr` on `PATH` as fallback

## 4. Placeholder Board (TDD)

- [ ] 4.1 Set up golden-file helpers with ASCII lipgloss profile and the `-update` flag
- [ ] 4.2 Write failing tests: `q` and `Esc` quit, view with a Project shows its name, view without a Project shows the notice
- [ ] 4.3 Implement the bubbletea model and `kanban board` until the tests pass

## 5. Manifest and docs

- [ ] 5.1 Update `herdr-plugin.toml`: `[[build]]` step, `board` pane entrypoint, action command `kanban open`
- [ ] 5.2 Keep the `prefix+o` snippet in `herdr-config.toml`
- [ ] 5.3 Update `README.md`: requirements (herdr ≥ 0.9.3, Go), install, `make build` for linked checkouts, keybinding

## 6. Verification

- [ ] 6.1 `go test ./...` and `go vet ./...` pass
- [ ] 6.2 Smoke test in herdr: link, `make build`, `prefix+o` inside this repo shows the Board, outside shows the notice, `q` and `Esc` close it
