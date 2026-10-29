## 1. Documents (TDD)

- [x] 1.1 Write failing tests for reading a Change's documents: open and archived (dated, undated) directories, nested specs joined under their capability paths, missing files, next open task / all done
- [x] 1.2 Implement document reading and the next-task line until the tests pass

## 2. Detail view (TDD)

- [x] 2.1 Add glamour (version compatible with lipgloss v1.1.0); render with the notty style in tests
- [x] 2.2 Write failing tests: header (name, badge, Progress, Phase or archive date), default tab (Tasks vs Proposal), 1–4 and Tab/shift+Tab with wrap, missing document note, scrolling with clamping and per-tab offsets, q/Esc emit close, resize re-renders
- [x] 2.3 Implement `internal/detail.Model` until the tests pass; goldens for a Tasks tab, a Specs tab with nested specs, a missing document, and a narrow width

## 3. Board integration (TDD)

- [x] 3.1 Write failing tests: Enter opens the selected Card's details, Enter in an empty Board does nothing, q in the detail returns to the Board with the same selection, ctrl+c quits from the detail
- [x] 3.2 Wire the detail into `board.Model` until the tests pass

## 4. Verification

- [x] 4.1 `go test -race ./...` and `go vet ./...` pass
- [x] 4.2 Smoke test in herdr: open details of an In Progress, a Draft and an archived Card in this repo and in magentic (nested specs); scroll, switch tabs, return
