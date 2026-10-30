## Context

See proposal.md for motivation and specs/ for the visible behaviour.

**Ordering.** This change is implemented after `m7-source-seam` and written against its names: `work.Item` instead of `openspec.Change`, `work.Phase` and `work.Group` instead of `internal/phase`, `internal/source/openspec` instead of `internal/openspec` and `internal/docs`, `Phase == work.Archived` instead of `Archived`, and `detail.New(read func() work.Docs, it work.Item, header, w, h)` with a tab bar built from `Docs.Tabs`. That refactor kept every golden unchanged and is now merged on `main`, so this change starts from the current goldens and the current `board.New(project, docs, load, now, snapshot, agents, start, resumable, resume, hide, notice)`.

Today's rendering (`internal/board/board.go`):

- `card` returns 3 content lines plus a blank one (`cardLines = 4`); each line is prefixed by a one-cell mark, `▌` for the selected Card (whose lines are also bold), so content width is `w-1`.
- `visible()` is `max(1, (height-chromeLines-2)/cardLines)` with `chromeLines = 6`; the 2 are the scroll markers.
- The only colour is `runningStyle` (`lipgloss.Color("2")`) for `●`, rendered apart so its reset does not end the bold name.
- Column titles, rules, the Progress bar (`progressBar` in `format.go`) and the footer are plain strings.
- The key help is the constant `helpText` (29 lines: `Keys`, the three groups, `any key closes`), rendered with `MaxWidth`/`MaxHeight` and no styling; golden `help-120`.
- `cmd/kanban` has `var version` (stamped with `-ldflags "-X main.version=<version>"`) and `versionString()`, which falls back to the module version in the build info and then `(devel)`; `kanban --version` prints it. The Board does not know it.
- `detail.View` builds the tab bar from marks `▌`/`·`, number and name, joined by ` │ `, with a short form when it is too wide.
- Goldens in `internal/board/testdata/golden` and `internal/detail/testdata/golden` render with `lipgloss.SetColorProfile(termenv.Ascii)` set in `TestMain`: colours and bold vanish, border characters remain. `detail_test` already has one test that switches to `termenv.ANSI256` and restores `Ascii` (`TestTerminalStyleUsesOnlyANSI16`).

## Goals / Non-Goals

**Goals:**
- One place for the palette, so board and detail cannot drift apart.
- Goldens change only through the bordered layout; colours and emphasis are checked by a few focused tests and the smoke test.
- Width and height arithmetic stays exact: every line `<= width`, the view never taller than the popup (existing `TestViewFitsHeight`, extended to the fallback heights).

**Non-Goals:**
- Styling the header line, scroll markers (`↑ n more`), "nothing here", error and loading texts, notices, the filter prompt or the sidebar.
- Colouring the focused column differently from the others.
- Configurable colours or themes; a light/dark switch.
- Anything in the detail view besides the tab bar (glamour's `terminal.json` is unchanged).

## Decisions

### A small `internal/theme` package holds the palette

```go
package theme

var (
	Accent    = lipgloss.Color("4") // selected Card border, active tab
	Dim       = lipgloss.Color("8") // Card borders, Draft and Archived
	Milestone = lipgloss.Color("5") // M<N> badge
	Running   = lipgloss.Color("2") // ● Running agent badge
)
```

Plus the derived styles both views need (`Faint`, `AccentBold`). `board` and `detail` both import it; `detail` cannot import `board` (board imports detail). The Phase colour map lives in `board` as `phaseColor(p work.Phase) lipgloss.Color` (Draft/Archived `8`, Ready `4`, In Progress `3`, Review `2`), because only the Board knows Phases visually.

Only `lipgloss.Color("0".."15")`, `Bold` and `Faint` are allowed; a test asserts no `38;5;`/`38;2;` sequences (see Tests).

Alternatives: constants duplicated in board and detail (two accents that can drift); a `Theme` struct passed through `New` (config surface nobody asked for).

### Cards: rounded box, thick box when selected

`card` builds the 3 content lines at content width `w-2` and wraps them in `lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(theme.Dim)`; the selected Card uses `lipgloss.ThickBorder()` and `theme.Accent`. `Width` is set so the box is exactly `w` cells wide (lipgloss adds the border outside `Width`; the implementation checks this with a width assertion, not by assumption). The `▌` mark and the per-line bold of the selected Card go away in bordered mode; the name is bold on every Card, so selection is shown by the border alone.

Card height becomes 5 and the blank separator line goes away (the borders separate Cards). At 120×30 that is still 4 visible Cards per column (`(30-8)/5`), at the golden's 14 rows still one.

The Card's line content keeps its order and texts: first line name (+ `● ` when running) and badge, second line Progress bar or archive date, third line planning marks + age (faint). Each segment is rendered separately and concatenated, the same trick `runningStyle` uses today, so one segment's reset never ends another's style; padding between name and badge stays unstyled.

### Fallback below 13 rows

```go
const borderMinHeight = 13 // chromeLines 6 + 2 scroll markers + one 5-line bordered Card
```

`m.bordered()` is `m.height >= borderMinHeight`; `cardLines()` returns 5 or 4 and `visible()` uses it. Below 13 rows the Card is rendered as today (mark column, `▌` on the selected Card, blank line after it) but with the new typography, so a 12-row popup still shows one Card per column. Width plays no part: single-column narrow mode keeps borders.

Alternative: dropping borders in narrow mode too (not needed: at ≥ 20 cells per column there is room, and the single column is wide).

`nav.fit` is driven by `visible()`, so a resize across 13 rows re-fits every column as any other resize does.

### Phase colours on titles, rules and bar fill

`column` renders the title and the `─` rule with `phaseColor(i)`; the narrow-mode indicator `◀ 3/5 In Progress 1 ▶` likewise. `progressBar(done, total, width int, fill lipgloss.Color)` styles the `█` run with `fill` and the `░` run faint; the ` done/total` label and "no tasks" stay default. Under the ASCII profile its output is byte-identical to today, so `format_test` keeps its expectations and only gains the parameter.

### Footer as key/description pairs

The long footer becomes a list of pairs (`{"h/l", "column"}, {"j/k", "card"}, …`); it is joined as today (`key desc` separated by ` · `) with the description faint. The width check measures `lipgloss.Width`, which ignores escape codes, so the switch to the short footer happens at the same width. The short footer (`h/l j/k ⏎ e s / gg/G ^d/^u 1-5 · r · ? · q close`) is the same list rendered keys-only, except `close` after `q`, which stays as its description. Notice, prompt and `starting …` footers are unchanged.

### Key help: version line, styling, frame

`board.New` gets a `version string` parameter (after `notice`); `cmd/kanban` passes `versionString()`, the same value `--version` prints. `board` does no build-info lookup of its own, so tests pass a fixed value (`"test"`) and the goldens stay stable.

`helpText` becomes data instead of one string: a list of groups, each a title and key/description rows, using the same pair type as the footer. `helpView()` renders:

```
openspec-kanban <version> · any key closes

Board
  h/l j/k      column, card
  …
```

The first line replaces both `Keys` and the trailing `any key closes`, so the help is 28 lines instead of 29. Group titles use `theme.AccentBold`, keys default (padded to the key column as today), descriptions `theme.Faint`, the first line default.

Frame: `lipgloss.RoundedBorder()` with `theme.Dim`, used only when the framed help fits, i.e. `height >= helpLines+2` and `width >= helpWidth+2`, computed from the rendered content rather than a fixed constant (28+2 = 30 rows, so the default 30-row popup shows the frame). Otherwise the help is drawn unframed and cut with `MaxWidth`/`MaxHeight` as today. This is the same idea as the Card fallback: a border is drawn only when it costs no content. The help stays top-left; no centring.

Alternatives: keep `Keys` as a second title line (31 framed rows, so no frame at the default popup height); the version in the Board header (the header's width is already shared with the filter and Progress); reading the build info inside `board` (a second source of truth for the version and unstable goldens).

### Detail tab bar

In the tab bar loop the part `mark + number + name` of the active tab is rendered with `theme.AccentBold`, every other part with `theme.Faint`; separators ` │ ` and marks are unchanged, and so is the short form. The detail goldens do not change.

### Tests

- Board goldens are rewritten with `go test ./internal/board -update` once, after the card and fallback groups, and the diff is reviewed by eye: borders closed, column widths unchanged (23 at 120, 60 in narrow mode), nothing wider than the view, `board-scrolled` (14 rows) still one Card per column with scroll markers, `error`, `loading`, `no-project` and `help-120` unchanged. The help group later rewrites `help-120` (30 rows) once more: it changes to the framed help with `openspec-kanban test`, and a new golden `help-low` (120×20) shows the unframed, cut help. Detail goldens must not change.
- Tests that find the selection through `▌<name>` (e.g. `TestResizeKeepsSelectionVisible`) look for the thick border `┃<name>` instead; a fallback test keeps `▌`.
- `TestViewFitsHeight` starts at 10 rows instead of 14 so it covers the fallback and the 12/13 boundary; a new test asserts each line of a 120- and 60-wide Board is exactly `width` cells.
- Style tests switch the global profile to `termenv.ANSI256` and restore `termenv.Ascii` with `t.Cleanup` (as `detail_test` already does; no `t.Parallel` in these packages). Expected fragments are built with the same lipgloss styles (e.g. `lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Render("Ready 2")`) rather than hand-written escape codes. They check: `phaseColor` for all five Phases; Ready's title and a yellow bar fill in a rendered Board; the selected Card's thick border in the accent; bold name, magenta badge, faint meta line; a faint footer description; help group titles accent+bold, keys default, descriptions faint, the frame in `theme.Dim`; the active tab accent+bold and an inactive tab faint in detail; no `38;5;`, `48;5;`, `38;2;` anywhere in a coloured Board or detail view.
- Everything else about colour (how it looks in real themes) is verified by the smoke test.

## Risks / Trade-offs

- [ANSI `8` is unreadable in some themes, e.g. close to the background in Solarized dark] → it is used only for borders and the Draft/Archived titles, never for Card text; the smoke test checks a light and a dark theme.
- [Yellow In Progress can be pale on light backgrounds] → it colours only the title, rule and bar fill, not text the user must read; smoke test on a light theme.
- [Faint is ignored by some terminals] → it degrades to normal text; no information is carried by faint alone.
- [Bordered Cards show one Card fewer per column at mid heights, e.g. 3 instead of 4 at 24 rows] → accepted for the clearer layout; scrolling and scroll markers are unchanged.
- [Content width shrinks by one cell] → names shorten one character earlier; `FitName` already handles it, the goldens show the effect.
- [Global colour profile in tests] → switched only inside tests that restore it with `t.Cleanup`; packages do not run tests in parallel.

## Migration Plan

Nothing to migrate: no config, cache or CLI changes. Rollback is a revert.

## Open Questions

- Whether the README should get an updated screenshot; optional and not needed for the specs.
