## Context

`internal/board` is the M0 placeholder (bubbletea model, q/Esc/ctrl+c quit, golden views). This change builds on `openspec.Load` (`m1-openspec-adapter`) and `phase.Group` (`m1-phase-derivation`). The popup is 90% of the terminal, so on a 200-column terminal the Board gets about 180 columns, on a 120-column laptop terminal about 108.

## Goals / Non-Goals

**Goals:**
- A Board that is usable daily: readable on laptop and wide terminals, fast to move around with the keyboard.
- Deterministic golden views for every layout.

**Non-Goals:**
- Card details (`m2-change-detail`), editing (`m2-editor-open`), Start and Running-agent badge (M4), filter (M4), colours beyond what lipgloss gives for free, mouse.

## Decisions

### Injected loader and clock
`board.New(project string, load func() ([]openspec.Change, error), now func() time.Time)`. main passes `openspec.Load` bound to the Project and `time.Now`; tests pass fixed data and a fixed time, which keeps ages and goldens stable. Loading runs as a `tea.Cmd`, its result arrives as a message; `r` issues the same command.

### Layout
- Header line: Project name left, overall Progress right; footer line: key hints.
- Columns share the width equally with a one-cell gap. Below 20 cells per column (104 cells for five), the Board switches to single-column mode with `◀ 2/5 Ready ▶` in the column title.
- A Card is three lines plus a blank line:
  ```
  ▌m1-board-view              M1
  ▌█████░░░░░ 5/11
  ▌P S D T · 3h          (Draft/Ready)   or   3h
  ```
  Archived Cards: name and `archived 2026-10-29`. Done artifacts show their letter, skipped ones `-`, missing ones `·`.
- Selection is shown with the `▌` marker and bold, so it is visible without colour and in goldens.
- `↑ 3 more` / `↓ 4 more` lines mark hidden Cards when a column scrolls.

lipgloss is added for joining columns and width-aware padding; goldens render with the ASCII colour profile so they contain no escape codes.

### Navigation state
The model keeps the focused column, a selected index per column and a scroll offset per column. Every move clamps to the column bounds and then adjusts the offset so the selection is visible. `g` waits for a second `g`; any other key cancels it.

### Reload keeps the selection by name
Before applying new data, the model remembers the selected Change name per column and restores it by name; if it is gone, the index is clamped.

### Age
`now - Modified` rounded down: under 1 h in minutes, under 1 d in hours, under 2 w in days, else weeks.

## Risks / Trade-offs

- [Load takes 1–2 s on Projects with many fresh Changes] → loading notice; later caching is possible (see `m1-openspec-adapter` design).
- [Five columns on a 104-cell Board leave 20 cells per Card] → names shorten in the middle; single-column mode below that.
- [Goldens get brittle with layout tweaks] → `go test ./... -update` and review the diff, as in M0.
