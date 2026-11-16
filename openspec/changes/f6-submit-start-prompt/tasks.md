## 1. Check

- [ ] 1.1 Add `submitted(run, name, pane, prompt)` in `internal/start` (design D1, D2): poll for up to 5 s for `working` or an empty box, treating no box, a read error or other text as taken; then up to 3 times `herdr pane send-keys <pane> enter` and 2 s of polling; return `the prompt was typed but not submitted in tab <change>` with the pane's last lines; table tests with a fake runner and the f4 screens (taken at once, empty box, Enter needed once, still stuck, other text, no box)
- [ ] 1.2 Call it in `LaunchUnfocused` after `herdr agent prompt` for kind `claude` only; adjust the launch tests (Start, Verify/Archive in a new tab, wave coordinator and member) so their fake runners report `working`

## 2. Docs

- [ ] 2.1 README troubleshooting: a stuck first prompt gets Enter, else `start failed` with `the prompt was typed but not submitted`
- [ ] 2.2 CHANGELOG `## Unreleased`: Start submits a stuck first prompt

## 3. Verification

- [ ] 3.1 `go test -race ./...`, `go vet ./...`, `gofmt -l .` pass
- [ ] 3.2 `openspec validate f6-submit-start-prompt --strict` passes

## Smoke test (user, in herdr)

These have no checkboxes: they do not count as tasks, and only the user runs them.

- Start a few Changes from the Board: each agent starts working on its prompt without a manual Enter
