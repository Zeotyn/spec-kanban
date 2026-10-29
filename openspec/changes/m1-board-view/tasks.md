## 1. Model and loading (TDD)

- [x] 1.1 Write failing tests: loading notice before data, columns after a load message, error message with retry hint, `r` triggers a reload, q/Esc/ctrl+c still quit
- [x] 1.2 Change `board.New` to take the loader and clock, run the loader as a `tea.Cmd`, and wire `openspec.Load` and `time.Now` in `kanban board` until the tests pass

## 2. Navigation (TDD)

- [x] 2.1 Write failing tests for every navigation scenario: h/l with remembered selection, j/k bounds, gg/G, ctrl-d/ctrl-u, 1–5, empty column, scroll offset keeps the selection visible, reload keeps the selection by name
- [x] 2.2 Implement the navigation state until the tests pass

- [x] 2.3 Skip empty columns: failing tests for h/l skipping, 1–5 on an empty column, focus moving off a column that empties on reload, all columns empty; then implement

## 3. Rendering (TDD with goldens)

- [x] 3.1 Add lipgloss and render goldens in the ASCII profile; write failing golden tests for a fixture Board at 200, 120 and 60 columns, plus loading and error views
- [x] 3.2 Write failing unit tests for name shortening, age formatting, milestone badge parsing and the `P S D T` marks
- [x] 3.3 Implement header, columns, Cards, scroll markers, footer and single-column mode until the tests pass

## 4. Verification

- [x] 4.1 `go test ./...` and `go vet ./...` pass
- [ ] 4.2 Smoke test in herdr: `prefix+o` in this repo and in magentic shows the right columns; navigate with all keys; `r` after ticking a task updates the Card; a narrow herdr window switches to single-column mode
