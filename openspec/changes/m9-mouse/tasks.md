## 1. Prerequisite

- [x] 1.1 Spike result recorded in design.md: left click, right click and wheel events reach a bubbletea program with `tea.WithMouseCellMotion()` in a herdr plugin popup, and herdr shows no menu of its own on the right click. If the spike was not done or any of these failed, stop and tell the user before writing code

## 2. Mouse reporting and layout

- [ ] 2.1 `cmd/kanban`: start the Board program with `tea.WithMouseCellMotion()`
- [ ] 2.2 `internal/board`: a pure layout function computes the x ranges of the shown columns, the y ranges of the visible Cards (including the `↑ n more` line and both Card heights) and the `◀`/`▶` cells of single-column mode. `View` uses it for its geometry
- [ ] 2.3 Test helper: render `View()`, strip ANSI, find a text and return its cell; send a `tea.MouseMsg` press there

## 3. Board mouse

- [ ] 3.1 Left click on a Card focuses its column and selects it; left click on the selected Card opens its details; clicks elsewhere change nothing (tests per scenario of "Mouse on the Board", bordered and borderless)
- [ ] 3.2 Wheel down/up over a column with Cards focuses it and acts like `j`/`k`, scrolling with the existing `nav`; over an empty column nothing changes
- [ ] 3.3 Single-column mode: clicks on `◀`/`▶` act like `h`/`l`; the wheel acts on the shown column
- [ ] 3.4 Mouse events are ignored while the filter prompt is open or a Start runs; a press closes the key help without any other effect

## 4. Card menu

- [ ] 4.1 Board state for the open menu (Card, Phase, anchor); a right click on a Card selects it and opens the menu, a right click elsewhere closes it, a right click on another Card moves it
- [ ] 4.2 Entry by Phase: `▶ Start  s` (Ready, In Progress), `↻ Resume  s` (Review), faint `▶ Start  s` (Draft, Archived); a click on it or `s` closes the menu and calls `startCard`; a click outside, `Esc` or another key only closes it
- [ ] 4.3 Draw the menu as a dim-bordered box spliced into the `View` lines with `github.com/charmbracelet/x/ansi` (direct requirement in `go.mod`), shifted to stay inside the popup; tests at the right and bottom edges and over a Card name with a wide character, with a colour profile
- [ ] 4.4 Detail view: the Board catches a right click and opens the menu for the shown Card; `s` in the open menu starts it; outside the menu the detail view still ignores `s`

## 5. Clickable footers

- [ ] 5.1 Build the Board's full and short footer from one segment list (label, description, optional action key); the short footer's text stays as today
- [ ] 5.2 Clicks on `⏎ details`, `e edit tasks`, `s start`, `/ filter`, `r reload`, `? help`, `q close` feed their key through the existing key handling, in both forms; no effect while a notice, the prompt or the starting Change is shown
- [ ] 5.3 Detail footer the same for `e edit`, `? help`, `q back`

## 6. Detail mouse

- [ ] 6.1 `internal/detail`: clicks on a tab in the full and the short tab bar show that document
- [ ] 6.2 Wheel scrolls three lines per step, stopping at both ends; each tab keeps its scroll position

## 7. Docs

- [ ] 7.1 README: a Mouse section (click, second click, wheel, right click menu, clickable footers, Shift to select text); key help unchanged
- [ ] 7.2 CONTEXT.md: **Card menu**; CHANGELOG `## Unreleased` entry

## 8. Verification

- [ ] 8.1 `go test -race ./...`, `go vet ./...`, `gofmt -l .` pass; the existing board and detail goldens stay unchanged
- [ ] 8.2 `openspec validate m9-mouse --strict` passes

## 9. Smoke test (user only: agents never tick these)

- [ ] 9.1 Board: click to select, click again for details, wheel in a long column, `◀`/`▶` in a narrow popup, `q close` in the footer
- [ ] 9.2 Card menu: right click on Ready (Start runs), on Review (Resume shown), on Draft (faint, notice), at the right edge, Esc closes; right click in the details starts the shown Card
- [ ] 9.3 Details: click tabs, wheel scroll, `q back`; key help closes on a click; Shift-drag still selects text
