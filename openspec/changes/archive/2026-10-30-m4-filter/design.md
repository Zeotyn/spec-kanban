## Context

`board.Model` keeps the Changes grouped by Phase in `m.cols` (`phase.Group`) and a `nav` that works purely on per-column counts: `next` skips empty columns, `keep(old, fresh, visible)` restores each column's selection by name and moves focus off a column that became empty (right first, then left, then Draft). Keys in use on the Board: `h j k l`, `g`/`gg`, `G`, `ctrl+d`/`ctrl+u`, `1`–`5`, `enter`, `e`, `r`, `q`, `esc`, `ctrl+c`, plus `s` from m4-start-action. `/`, `n` and `N` are free. `esc` currently closes the Board. Milestone badges come from `milestone(name)` (`m<digits>-` → `M<N>`).

## Goals / Non-Goals

**Goals:**
- Filtering reuses `nav` unchanged: hidden Cards are simply absent from the columns nav sees.
- Pure, table-tested matching and match navigation; Board behaviour tested through `Update` like the existing keys.

**Non-Goals:**
- Fuzzy matching, regular expressions, or matching anything but the Change name.
- Filter history, or remembering the filter between Board opens.
- Highlighting the matched part of a name.
- Filtering in the detail view or the sidebar.
- Milestone filters for `f<N>`/`c<N>` prefixes; they are not milestones and substring matching already finds them.

## Decisions

### Keep all Changes, derive the visible columns
The Model gets `all [columns][]openspec.Change` (what `phase.Group` returned) and `filter string`; `cols` becomes `filtered(all, filter)`, recomputed on every load and every filter change, followed by `nav.keep(names(oldCols), names(cols), visible)`. Everything that renders or navigates keeps reading `cols`, so counts, empty-column notes, scrolling, focus fallback and selection-by-name all apply to the filtered Board without new code paths. The header's Progress and `updateDetailHeader` read `all`, so a filter never changes project-level numbers or loses the open detail's Change. Alternative: a `hidden` flag per Card that nav would have to skip (touches every nav function).

### Matching
`matches(name, filter)`: if `filter` matches `^[mM][0-9]+$`, compare `milestone(name)` with the upper-cased filter; otherwise `strings.Contains(strings.ToLower(name), strings.ToLower(filter))`. Case-insensitive substring is what users expect from a quick filter and costs nothing; the milestone rule makes `m1` mean milestone 1 rather than also `m10-…` and `fix-m1-…`. An empty filter matches everything and counts as no filter.

### Prompt
`prompting bool` in the Model. While prompting, keys go to the prompt before anything else: printable runes (`tea.KeyRunes`, including space) append, `backspace` deletes a rune or closes the prompt when empty, `enter` closes it keeping the filter, `esc` closes it and clears the filter, `ctrl+c` quits. The footer shows `/` followed by the filter and a `█` cursor in place of the key help. Filtering is incremental (every key re-filters); with tens of Changes this is instant. The prompt is plain stdlib string handling, not `bubbles/textinput`: no cursor movement or clipboard is needed, and it avoids a new dependency.

The m4-start-action "ignore keys while starting" check stays first in `Update`, so a Start cannot be interrupted by typing.

### Esc and the active filter
On the Board (prompt closed), `esc` clears an active filter and otherwise quits as today. `q` always quits. This mirrors how Esc backs out one level in the detail view, and keeps a stray Esc from closing the Board while the user is looking at a filtered set.

### Display
The header title gets ` · /<filter>` after the Project name (before `· refreshing…`), so the filter stays visible while notices replace the footer. Column titles show the filtered count. An empty column keeps its existing `nothing here` note; with an active filter, the header already explains why. The footer key help gains `/ filter`; the narrow variant gains `/`.

### n / N
`nav.match(dir int, counts)` walks Cards in reading order: within the focused column from the current selection, then from the top (or, for `N`, the bottom) of the next column with Cards in that direction, wrapping from the last column to the first. Because only matching Cards are present, "next match" equals "next Card on the filtered Board", so `n`/`N` need no knowledge of the filter beyond "is one active". Without a filter they do nothing, which keeps `n` free of surprising Board-wide jumps. With a single match, `n` leaves it selected.

## Risks / Trade-offs

- [Esc no longer always closes the Board] → only while a filter is active, and the header shows it; `q` and `ctrl+c` always close.
- [Typed text cannot contain `ctrl` keys or paste with newlines] → Change names are kebab-case; nothing is lost.
- [Both M4 changes edit `board.Update` and the footer] → whichever lands second rebases; the order of checks (starting, prompt, detail, Board keys) is stated above.
