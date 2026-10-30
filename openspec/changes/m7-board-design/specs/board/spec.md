## ADDED Requirements

### Requirement: Terminal palette only
The Board SHALL style text only with the 16 ANSI terminal colours (`0`–`15`), bold and faint, so the terminal's theme decides the actual colours; it SHALL emit no 256-colour or true-colour sequences. Without colour support the same layout SHALL be shown without styling.

#### Scenario: Colour terminal
- **WHEN** the Board is shown in a terminal with 256 colours
- **THEN** every colour in the output is one of the 16 ANSI colours

#### Scenario: No colour support
- **WHEN** the Board is shown in a terminal without colours
- **THEN** it shows the same columns, borders and Cards without any colour or emphasis

### Requirement: Bordered Cards
Each Card SHALL be drawn as a box with a rounded border in the dim colour (ANSI bright black, `8`) around its three content lines, with no blank line between Cards. The selected Card SHALL instead have a thick border in the accent colour (ANSI blue, `4`); no `▌` marker SHALL be shown. When the Board is fewer than 13 rows high, too low for one bordered Card, Cards SHALL be drawn without borders, separated by a blank line, and the selected Card SHALL be marked with `▌` before each of its lines. The single-column mode on narrow terminals SHALL keep the borders.

#### Scenario: Board with bordered Cards
- **WHEN** the Board is 120 by 30 cells and Ready holds two Cards
- **THEN** each Ready Card is a rounded box of five lines directly below the other, and the column shows four Cards before it scrolls

#### Scenario: Selected Card
- **WHEN** Card `m1-board-view` is selected
- **THEN** its box has a thick border in the accent colour and the other Cards keep their dim rounded border

#### Scenario: Low popup
- **WHEN** the Board is 12 rows high
- **THEN** Cards have no borders, are separated by a blank line, and the selected Card is marked with `▌`

#### Scenario: Narrow popup keeps borders
- **WHEN** the Board is 60 cells wide and 30 rows high
- **THEN** the single shown column draws its Cards with borders

### Requirement: Phase colours
Each column's title and the rule below it SHALL be coloured by the column's Phase: Draft and Archived in ANSI bright black (`8`), Ready in blue (`4`), In Progress in yellow (`3`), Review in green (`2`). The position indicator of the single-column mode SHALL use the colour of the shown Phase. The filled part of a Card's Progress bar SHALL use the colour of the Card's Phase and the empty part SHALL be faint. The header's overall Progress SHALL keep the default colour.

#### Scenario: Coloured column titles
- **WHEN** the Board shows all five columns in a colour terminal
- **THEN** the titles and rules of Draft and Archived are bright black, Ready blue, In Progress yellow and Review green

#### Scenario: Progress bar of an In Progress Card
- **WHEN** Card `m1-board-view` in In Progress has 5 of 11 tasks completed
- **THEN** the filled cells of its bar are yellow and the empty cells faint

### Requirement: Typographic emphasis
On every Card the Change name SHALL be bold, the milestone badge SHALL be in the accent colour for milestones (ANSI magenta, `5`), and the line with planning marks, age or archive date SHALL be faint. The Running agent badge SHALL stay green (`2`). In the Board's key footer each key SHALL be shown in the default style and its description faint. Notices, the filter prompt and the starting message in the footer SHALL keep the default style.

#### Scenario: Card text
- **WHEN** Card `m1-board-view` (Ready, planned, modified 40 minutes ago) is shown in a colour terminal
- **THEN** `m1-board-view` is bold, `M1` magenta, and `P S D T  40m` faint

#### Scenario: Footer
- **WHEN** the Board shows its key footer in a colour terminal
- **THEN** keys such as `h/l` and `q` are in the default style and descriptions such as `column` and `close` are faint

## MODIFIED Requirements

### Requirement: Key help
`?` on the Board or in the detail view SHALL show the key help in place of the view: every key with what it does, grouped under Board, Filter prompt and Details. The help SHALL start with a line naming the plugin and its version as `openspec-kanban <version>`, where the version is the one `kanban --version` prints, and saying that any key closes the help. Group titles SHALL be bold in the accent colour (ANSI blue, `4`), keys in the default style and their descriptions faint. When the popup is large enough for the whole help with a border around it, the help SHALL be framed by a rounded border in the dim colour (ANSI bright black, `8`); otherwise it SHALL be shown without the border, cut to the popup as before. The next key SHALL close the help and return to the view it was opened from, without any other effect; `ctrl+c` SHALL close the Board. In the filter prompt `?` SHALL be typed into the filter, and while Start runs it SHALL be ignored like every other key. The footers of the Board and the detail view SHALL mention `?`.

#### Scenario: Help on the Board
- **WHEN** the Board is open and the user presses `?`
- **THEN** the key help is shown with the groups Board, Filter prompt and Details

#### Scenario: Version in the help
- **WHEN** `kanban --version` prints `0.3.0` and the user presses `?` on the Board
- **THEN** the help starts with `openspec-kanban 0.3.0` and says that any key closes it

#### Scenario: Styled and framed help
- **WHEN** the user presses `?` in a 120 by 40 popup in a colour terminal
- **THEN** the help is framed by a dim rounded border, `Board`, `Filter prompt` and `Details` are bold and blue, keys such as `gg/G` are in the default style and descriptions such as `first/last card` faint

#### Scenario: Help in a low popup
- **WHEN** the popup is too low for the framed help and the user presses `?`
- **THEN** the help is shown without the border, cut to the popup height

#### Scenario: Closing the help does nothing else
- **WHEN** the key help is shown over the Board and the user presses `q`
- **THEN** the help closes, the Board stays open, and the selection is unchanged

#### Scenario: Help from the detail view
- **WHEN** the detail view of `m1-board-view` is open and the user presses `?` and then `j`
- **THEN** the key help is shown, and after `j` the detail view of `m1-board-view` is shown again at the same scroll position

#### Scenario: Question mark in the filter
- **WHEN** the filter prompt is open and the user types `?`
- **THEN** no help is shown and the filter contains `?`
