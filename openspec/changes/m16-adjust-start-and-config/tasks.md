## 1. Start with adjusted instructions

- [x] 1.1 `boardPrompt` takes an optional override of the configured text and `startOne` passes it through (design D1); table tests: no override, override with and without `prompt`, empty override
- [x] 1.2 Board: `S` opens the input line `start> ` pre-filled through a `startText(change)` callback, `enter` starts with the line, `esc` closes it; running agent focuses as `s`; notice for Cards `s` cannot start and for 2 or more marks (design D2); tests
- [x] 1.3 Card menu entry "start with…" on the Cards where `S` works; key help lists `S`; golden files updated

## 2. Edit the config

- [x] 2.1 Config holder in `cmd/kanban` with `reload`, used by every callback; `editConfig` returns the path and creates `{}` with its directory when missing, or an error without a config directory (design D3); tests
- [x] 2.2 Board: `c` suspends for `$EDITOR` on that path like `e`, then reloads and applies hidden Phases, wave size and the warning, keeping the selection and the filter; notice without a config directory; key help lists `c`; tests

## 3. Docs

- [x] 3.1 README: `S` and `c` in the key table; Settings mention that `c` opens the config
- [x] 3.2 CHANGELOG `## Unreleased`: `S` starts with adjusted instructions, `c` edits the config

## 4. Verification

- [x] 4.1 `go test -race ./...`, `go vet ./...`, `gofmt -l .` pass
- [x] 4.2 `openspec validate m16-adjust-start-and-config --strict` passes

## Smoke test (user, in herdr)

These have no checkboxes: they do not count as tasks, and only the user runs them.

- Press `S` on a Ready Card, change the instructions and start: the agent gets the changed text. Then press `c`, change `instructions`, quit the editor and press `s` on another Card: the agent gets the new instructions
