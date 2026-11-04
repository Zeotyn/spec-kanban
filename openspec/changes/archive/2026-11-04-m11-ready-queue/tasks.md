## 1. Queue order in `internal/work`

- [x] 1.1 Tests first: the name rank for `f1-…`, `f12-…` (fix), `m9-…`, `m10-…` (milestone), and `fix-login`, `m-foo`, `c2-logo`, `001-user-login`, `add-dark-mode` (other); natural compare for `c2` < `c10`, `001` < `002` < `010`, equal values with leading zeros, and a digit run longer than an int
- [x] 1.2 Implement the unexported rank helper and the natural compare with the standard library only, as in design.md decisions 1 to 3
- [x] 1.3 Tests first, one per scenario of "Cards are ordered within a Phase": Two Draft Changes, Two Ready Changes (Ready ignores modification time), Fixes before milestones, Milestones by number, Other Changes last in natural order, Spec Kit Features in Ready, Archived on the same day. Move the newest-first cases of `TestGroup` from Ready to Draft, and add one for In Progress and Review
- [x] 1.4 `Group`: Ready uses the queue order, ties by name; the other Phases keep their order; update the doc comment

## 2. Callers

- [x] 2.1 Run the board, mouse, sidebar, cache and Source tests; where a fixture with several Ready Cards relied on modification order, change the fixture (names or Phase), not the expected Board. The goldens stay unchanged; if one changes, say why in the commit

## 3. Docs

- [x] 3.1 CHANGELOG `## Unreleased`: Ready is a queue (fixes, then milestones by number, then the rest in natural name order); the other columns keep newest first. The README does not describe the Card order and stays

## 4. Verification

- [x] 4.1 `go test -race ./...`, `go vet ./...`, `gofmt -l .` pass
- [x] 4.2 `openspec validate m11-ready-queue --strict` passes

## Smoke test (user, in herdr)

These have no checkboxes: they do not count as tasks, and only the user runs them.

- Open the Board in this repo: Ready lists fixes first, then milestones by number (`m9` before `m10`), then the rest
- Fix a typo in a Ready Change's proposal: the Card keeps its place, and only its age line changes
- Edit a Draft Change: its Card moves to the top of Draft as before
- `gg` in Ready, then `s`: the first Card in the queue starts
- A Spec Kit Project: Ready Features read `001-…`, `002-…`, `010-…`
