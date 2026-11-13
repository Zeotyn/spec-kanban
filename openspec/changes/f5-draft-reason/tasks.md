## 1. Reason

- [x] 1.1 Add `draftReason(work.Item) string` in `internal/board` (design D1); table tests: not Draft, design missing, several missing in order, absent key counts as missing, all done without tasks, Artifacts nil, unknown ids ignored

## 2. Footer

- [x] 2.1 In `footer()`, without marks and with a reason for the selected Card, write `<reason> · ` first, choose full or short keys against the remaining width, and shift every hit by the prefix width (D2); tests for the spec scenarios (design missing, Ready Card, marks win, notice wins) and a golden view of a Board with a Draft Card selected
- [x] 2.2 Test that a click on `q close` after the reason closes the Board, and that a click on the reason does nothing

## 3. Docs

- [x] 3.1 README: the Draft reason in the Board footer
- [x] 3.2 CHANGELOG `## Unreleased`: the Draft reason

## 4. Verification

- [x] 4.1 `go test -race ./...`, `go vet ./...`, `gofmt -l .` pass
- [x] 4.2 `openspec validate f5-draft-reason --strict` passes

## Smoke test (user, in herdr)

These have no checkboxes: they do not count as tasks, and only the user runs them.

- Open the Board and select a Draft Card: the footer starts with `draft: … missing` or `draft: no tasks`
- Move to a Ready Card: the reason is gone
