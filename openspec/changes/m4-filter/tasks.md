## 1. Matching (TDD)

- [x] 1.1 Write failing table tests for `matches`: substring anywhere, case ignored, empty filter matches all, `m1` matches `m1-board-view` but not `m10-cleanup` or `fix-m1-typo`, `M1` same as `m1`, `m` alone is a substring filter, `m1-` is a substring filter
- [x] 1.2 Implement `matches` until the tests pass

## 2. Match navigation (TDD)

- [x] 2.1 Write failing tests for `nav.match`: next within a column; next jumps to the first Card of the next column with Cards, skipping empty ones; wraps from the last Card to the first; `N` goes to the last Card of the previous column and wraps backwards; single Card stays selected; all columns empty does nothing; scroll offset follows
- [x] 2.2 Implement `nav.match` until the tests pass

## 3. Filtered Board (TDD)

- [x] 3.1 Write failing Board tests: `/` opens the prompt and the footer shows it; typing filters columns and counts at once; `j`, `q`, `n` typed in the prompt become filter text; backspace deletes and closes on empty; `enter` keeps the filter and closes the prompt (no detail opens); `esc` in the prompt clears; `ctrl+c` quits from the prompt; header shows ` · /<filter>` and keeps the unfiltered Progress; focused column without matches moves focus as on reload; selection kept by name when the filter changes; filter survives `r` and the cached-to-fresh swap; open detail header still updates while filtered; `esc` on the Board clears an active filter and quits without one; `n`/`N` without a filter do nothing
- [x] 3.2 Implement `all`/`filter`/`prompting` in the Model, the prompt key handling, header and footer text, `n`/`N` until the tests pass

## 4. Docs

- [x] 4.1 README: `/`, `n`/`N`, `Esc` clearing the filter, the milestone filter rule

## 5. Verification

- [x] 5.1 `go test -race ./...` and `go vet ./...` pass
- [ ] 5.2 Smoke test in herdr: rebuild, open the Board on a Project with several milestones; `/m1` shows only M1 Cards with fitting counts and the header filter; `enter`, `n`/`N` walk the matches across columns and wrap; `/BOARD` matches case-insensitively; `esc` restores all Cards, a second `esc` closes the popup; narrow popup still works with a filter
