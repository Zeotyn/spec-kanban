## Context

The Board (`internal/board`) and the detail view (`internal/detail`) are bubbletea models that react only to `tea.KeyMsg`. `cmd/kanban` starts the program without mouse options. Drawing is plain string work:
- The Board's `View` joins the header, a blank line, the columns and the footer.
- Column widths come from `m.width` and the number of shown columns. Card heights come from `cardLines()`: 5 bordered, 4 below `borderMinHeight`. Each column has an optional `↑ n more` line.
- The detail view stacks the header, the tab bar, a rule, the body, a rule and the footer.

bubbletea v1.3 reports clicks and the wheel as `tea.MouseMsg`, with cell coordinates relative to the program's terminal, which here is the whole popup. lipgloss v1 has no layers.

## Goals / Non-Goals

**Goals:**
- Hit-testing that cannot drift from what is drawn.
- One Start path: the Card menu calls the same `startCard` as `s`.
- No new module downloads.

**Non-Goals:**
- Hover effects, drag and drop of Cards (the Board stays view-only), double-click timing.
- A config switch to turn the mouse off. We add one only if text selection turns out to bother people.
- More Card menu entries than the one Start/Resume entry.

## Decisions

1. **Mouse reporting: `tea.WithMouseCellMotion()`** on the Board program. It reports presses, releases and the wheel. All-motion reporting would only add hover events, which nothing uses. Every action fires on the press. Releases and motion are ignored.

2. **One layout for drawing and hit-testing.**
   - `View` has a value receiver and cannot remember where it drew things. A small pure function therefore computes the geometry from the model: x ranges per shown column, y ranges per visible Card (top offset, the `↑` line, `cardLines()`), the `◀`/`▶` cells in single-column mode, and the x ranges of the footer actions.
   - `View` and the mouse handler both use it. The mouse handler maps a cell to a Card or an action, then reuses the key path (`nav`, `enter`, `startCard`, …).
   - Rejected: bubblezone, which marks zones with escape sequences in the output. It is another dependency, and `lipgloss.Width`/`MaxWidth` would have to ignore its markers in every cut.

3. **Tests click where things are drawn.** Each test renders `View()`, strips the ANSI codes, finds the target text (a Card name, `▶`, `q close`, a tab title) and sends a `tea.MouseMsg` at that cell. A layout that drifts from the drawing fails a test, and no golden has to change.

4. **A second click on the selected Card opens it, instead of a double-click.** A double-click needs a clock and a time threshold. The second click is deterministic, and on a Card that is not selected the first click is the selecting click anyway.

5. **The wheel moves the selection, not only the view.** One wheel step equals `j`/`k` in the column under the pointer. This keeps the "selected Card stays visible" rule, and the existing `nav` scrolling does the rest. In the detail view the wheel scrolls three lines per step, a common terminal default.

6. **The Board owns the Card menu, also over the detail view.**
   - The state lives in the Board model: the Card, its Phase and the anchor cell.
   - In the detail view the Board catches the right click before it forwards mouse events, because the detail model knows nothing about Start.
   - The menu is a small box drawn with the theme's dim border. It is shifted left or up when it would cross the popup edge. It is spliced into the finished `View` lines with `ansi.Cut`/`ansi.Truncate` from `github.com/charmbracelet/x/ansi`, which lipgloss already depends on and which only becomes a direct requirement in `go.mod`. Cutting by display cells keeps colour codes and wide characters intact.
   - Rejected: lipgloss v2 layers, which mean a major-version migration for one box.

7. **Footers from segments.**
   - Today the Board's short footer is a hand-written string. Both footer forms will be built from the same list of segments: label, description and an optional action key. The full form shows every segment. The short form shows the short labels.
   - The layout records the x ranges of the segments that have an action. A click on one feeds its key to the existing key handling, so `q close` and `q` cannot diverge.
   - The detail footer gets the same treatment for `e`, `?` and `q`.

8. **Mouse events are ignored while prompting and starting**, the same states in which keys are already restricted. The key help closes on any press.

## Risks / Trade-offs

- [herdr keeps the right click for its own pane menu, or does not forward mouse events to plugin popups] → Task 1 is a spike in a real popup before any code. If the right click does not arrive, stop and replan the menu trigger (for example a `⋯` cell on the Card) with the user.
- [Mouse reporting stops plain text selection in the popup] → Most terminals select with Shift held while dragging; the README says so. A config switch waits until someone needs it.
- [Layout code duplicates the drawing arithmetic] → Decision 2 makes the layout the single source that `View` also uses, and decision 3 tests against the drawing.
- [Splicing the menu over styled lines] → Cutting by display cells with `x/ansi` and focused tests with a colour profile, including a wide character in a Card name.

## Migration Plan

None. Keys behave as before, and mouse reporting only starts with the new binary.
