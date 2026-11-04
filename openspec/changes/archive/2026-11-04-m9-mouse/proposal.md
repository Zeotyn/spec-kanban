## Why

herdr is mouse first: its users click tabs, panes and menus, and expect the same inside a plugin popup. The Board and the detail view react only to keys today, so someone who does not know the vim keys cannot even select a Card or close the popup. Every Board function should be reachable with the mouse alone, while the keys keep working exactly as before.

## What Changes

- The Board and the detail view turn on mouse reporting.
- Board:
  - A left click on a Card selects it and focuses its column; a left click on the already selected Card opens its details (like Enter).
  - The mouse wheel over a column focuses that column and moves the selection one Card down or up (like `j`/`k`).
  - In narrow single-column mode, clicking `◀` or `▶` shows the previous or next column.
- Card menu: a right click on a Card selects it and opens a small menu at the pointer, kept inside the popup. For now it holds one entry:
  - `▶ Start  s` on a Ready or In Progress Card.
  - `↻ Resume  s` on a Review Card.
  - A faint `▶ Start  s` on a Draft or Archived Card.
  - Clicking the entry or pressing `s` does what `s` does on that Card, notices included. A click outside the menu, Esc or any other key closes it without further effect; a right click on another Card opens the menu there.
- Clickable footers: the actions in the Board footer (`⏎ details`, `e edit tasks`, `s start`, `/ filter`, `r reload`, `? help`, `q close`) and in the detail footer (`e edit`, `? help`, `q back`) act like their key when clicked, in the full and the short footer.
- Detail view:
  - A click on a tab in the tab bar shows that document.
  - The mouse wheel scrolls the document three lines per step.
  - A right click opens the Card menu for the shown Card.
- Key help: a click anywhere closes it, like any key.
- While the filter prompt is open or a Start is running, mouse events are ignored.
- Selecting text in the popup needs the terminal's own modifier (usually Shift while dragging); the README says so.
- The new user smoke-test tasks say explicitly that only the user ticks them. The m7 and m8 agents ticked theirs themselves.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `board`: new requirements for mouse selection and scrolling on the Board, the Card menu, the clickable footer and closing the key help with a click.
- `change-detail`: new requirement for clicking tabs, scrolling with the wheel, the clickable footer and the Card menu in the detail view.

## Impact

- Code: `internal/board` (mouse handling, a shared layout used by drawing and by hit-testing, the Card menu and its overlay), `internal/detail` (tab bar and footer hit-testing, wheel), `cmd/kanban` (starts the program with mouse reporting), README (mouse section, text selection hint), CONTEXT.md (Card menu).
- Dependencies: `github.com/charmbracelet/x/ansi`, already in the module graph through lipgloss, becomes a direct dependency for splicing the menu into rendered lines. Nothing new is downloaded.
- Prerequisite: a spike in a herdr popup confirms that left clicks, right clicks and the wheel reach the plugin pane, and that herdr does not keep the right click for its own menu.
