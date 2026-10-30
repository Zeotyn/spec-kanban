## Why

The Board is plain text on the terminal's default colour: Cards are separated only by a blank line, the selected Card by a one-cell `▌`, and nothing tells the Phases apart except the column title. In a popup full of similar names the eye has no anchor. The Board should look like a kanban board while staying inside the terminal's own palette, so every theme (light or dark) still decides the actual colours.

## What Changes

- Cards are drawn as boxes with a dim rounded border; the selected Card gets a thick border in the accent colour and a bold name instead of the `▌` marker. The blank line between Cards goes away; a Card takes 5 lines instead of 4 and two cells less content width.
- When the popup is too low for one bordered Card (fewer than 13 rows), the Board falls back to today's borderless Cards with the `▌` marker; narrow single-column mode keeps the borders.
- Column titles, their underline and each Card's Progress bar fill are coloured by Phase: Draft and Archived grey, Ready blue, In Progress yellow, Review green; the empty part of the bar is dim.
- Typographic hierarchy: Change names bold, the meta line (planning marks, age, archive date) faint, the milestone badge magenta; in the footer the keys stay normal and their descriptions are faint. The Running agent badge stays green.
- Key help (`?`): its first line names the plugin and its version (`openspec-kanban <version>`, the value `kanban --version` prints) and says that any key closes it; group titles are bold in the accent colour, keys normal, descriptions faint; the help is framed by a dim rounded border when the popup is large enough for the framed help, and drawn unframed as today otherwise.
- Detail view: the active tab is bold in the accent colour, inactive tabs are faint. Nothing else in the detail view changes.
- Only the 16 ANSI colours (`0`–`15`), bold and faint are used; no hex, 256 or true colours.
- Behaviour is unchanged: navigation, filter, the help keys, Start, notices, the header line (title and Progress) and narrow mode work as before.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `board`: new requirements for how Cards look (borders, selected Card, low-popup fallback), Phase colours, and the typographic emphasis of Cards and footer, all within the terminal's 16 colours; the Key help requirement gains the version line, styling and frame.
- `change-detail`: new requirement for how the tab bar marks the active and inactive tabs.

## Impact

- Code: `internal/board` (card and column rendering, height arithmetic, footer, help), `cmd/kanban` (passes its version string into `board.New`), `internal/detail` (tab bar), a small new palette package shared by both.
- Tests: board goldens change for the bordered layout and the framed help with its version line (they render without colour); detail goldens stay as they are; new focused style tests render with a colour profile.
- Sequencing: implemented after `m7-source-seam`, against its `work.Item`/`work.Phase` names (now merged).
- No new dependencies; lipgloss already provides borders and styles.
