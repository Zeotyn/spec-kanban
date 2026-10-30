Prerequisite: `m7-source-seam` is merged on `main`. Names below are the post-refactor ones (`work.Item`, `work.Phase`, `internal/source/openspec`, `detail.New(read, it, header, w, h)`). Every group ends with `go test ./...` green; only groups 2 and 6 change golden files.

## 1. Palette (TDD)

- [x] 1.1 Write failing tests in `internal/theme`: `Accent`, `Dim`, `Milestone`, `Running` are `lipgloss.Color` values `"4"`, `"8"`, `"5"`, `"2"`; `Faint` is faint, `AccentBold` is bold with foreground `Accent`
- [x] 1.2 Implement `internal/theme` (design.md); replace `runningStyle` in `board` by `theme.Running`; tests green, goldens unchanged

## 2. Bordered Cards and the low-popup fallback (TDD)

- [x] 2.1 Write failing board tests: at 120×30 each Card is 5 lines in a rounded box (`╭`, `╰`) with no blank line between Cards and 4 Cards visible per column; the selected Card has a thick box (`┏`, `┃<name>`) and no `▌`; at 12 rows Cards have no border, a blank line between them and `▌` on the selected Card; at 13 rows borders are shown; at 60×30 (single column) borders are shown; every line of a 120- and 60-wide Board is exactly `width` cells
- [x] 2.2 Implement `borderMinHeight = 13`, `bordered()`, `cardLines()` (5 or 4) and `visible()` on it; `card` wraps the content lines (width `w-2`) in `RoundedBorder` with `theme.Dim`, the selected one in `ThickBorder` with `theme.Accent`; below 13 rows keep today's mark-column rendering
- [x] 2.3 Extend `TestViewFitsHeight` down to 10 rows; switch selection assertions that look for `▌<name>` (e.g. `TestResizeKeepsSelectionVisible`) to `┃<name>`, keeping `▌` where a test runs below 13 rows
- [x] 2.4 Rewrite the board goldens with `go test ./internal/board -update` and review the diff: closed boxes, column widths unchanged, `board-scrolled` (14 rows) still one Card per column with scroll markers, `board-60` bordered, `help-120`, `error`, `loading`, `no-project` unchanged; detail goldens unchanged

## 3. Phase colours (TDD)

- [x] 3.1 Write failing tests: `phaseColor` returns `8`, `4`, `3`, `2`, `8` for Draft, Ready, In Progress, Review, Archived; with the profile switched to `termenv.ANSI256` (restored to `Ascii` in `t.Cleanup`), a rendered 120×30 Board contains the Ready title and rule in blue, the In Progress Card's `█` run in yellow and its `░` run faint, and the narrow indicator in the shown Phase's colour; `format_test` cases still pass under ASCII
- [x] 3.2 Implement `phaseColor`; colour the title and rule in `column` and the narrow indicator in `columns`; `progressBar` takes the fill colour, styles `█` with it and `░` faint, label unchanged; goldens unchanged

## 4. Typography and footer (TDD)

- [ ] 4.1 Write failing ANSI256 tests: Card name bold, `M<N>` in `theme.Milestone`, the planning-marks/age line and the archive-date line faint, `●` still in `theme.Running` without ending the bold name; the long footer shows `h/l` default and `column` faint; the short footer shows keys default and `close` faint; notice, prompt and `starting …` footers carry no style; no `38;5;`, `48;5;` or `38;2;` in the coloured Board
- [ ] 4.2 Implement the per-segment Card styling (name, badge, running badge, meta line) for bordered and fallback Cards; turn the footer into key/description pairs rendered long or keys-only, with the width check on `lipgloss.Width`; goldens unchanged

## 5. Detail tab bar (TDD)

- [ ] 5.1 Write failing ANSI256 tests in `detail`: the active tab's part (`▌1 Tasks`) is rendered with `theme.AccentBold`, the inactive ones faint, also in the short form; no `38;5;`, `48;5;` or `38;2;` in the view
- [ ] 5.2 Implement the styling in the tab bar loop over `Docs.Tabs`; marks, separators and the short-form switch unchanged; detail goldens unchanged

## 6. Key help: version, styling, frame (TDD)

- [ ] 6.1 Write failing tests: `board.New` takes a `version` (tests pass `"test"`); the help's first line is `openspec-kanban test · any key closes`, followed by the groups Board, Filter prompt and Details with every key of today's help; at 120×30 the help is framed by a rounded border (`╭`, `╰`) and fits the height; at 120×20 it has no border and no more than 20 lines; with ANSI256 the group titles use `theme.AccentBold`, keys no style, descriptions `theme.Faint`, the border `theme.Dim`; help from the detail view shows the same text; existing help tests (close without effect, `?` in the filter, ignored while starting) stay green
- [ ] 6.2 Implement: help groups as data, `helpView()` with the frame only when it fits (design.md); `cmd/kanban` passes `versionString()` to `board.New`; add a `cmd/kanban` test or assertion that the Board receives the same string `--version` prints, if the existing tests reach `showBoard`, otherwise cover it in the smoke test
- [ ] 6.3 Rewrite `help-120` and add golden `help-low` (120×20) with `go test ./internal/board -update`; review: framed help with `openspec-kanban test`, no line wider than the view, unframed help cut at 20 lines
- [ ] 6.4 Update README's help sentence if it names the help's layout (`?` shows all keys and the version; any key closes the help)

## 7. Goldens and docs

- [ ] 7.1 Run `go test ./... -update` once more and confirm it rewrites nothing (`git status '*/testdata/*'` shows only the changes of groups 2 and 6), proving colour and emphasis do not leak into the ASCII goldens
- [ ] 7.2 Optional: README note or screenshot of the bordered, coloured Board

## 8. Verification

- [ ] 8.1 `go test -race ./...`, `go vet ./...` and `gofmt -l .` pass; `grep -rn 'lipgloss.Color("' --include='*.go' internal` shows only values `0`–`15`
- [ ] 8.2 Smoke test in herdr against this repo, in a dark and (if available) a light terminal theme: bordered Cards, thick accent border moving with `j`/`k`/`h`/`l`, Phase colours on titles, rules and bars, bold names, magenta badges, faint meta line and footer descriptions, green `●`; shrink the popup below 13 rows (borders drop, `▌` appears) and narrow it below five columns (single column keeps borders); filter, `s` notices, the help (`?`: version matches `kanban --version`, accent group titles, faint descriptions, frame at full size, no frame in a low popup) and the detail tab bar (accent active, faint inactive); everything readable in both themes
