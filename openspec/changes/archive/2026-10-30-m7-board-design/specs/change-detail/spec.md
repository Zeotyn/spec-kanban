## ADDED Requirements

### Requirement: Tab bar emphasis
The tab bar SHALL show the active tab bold in the accent colour (ANSI blue, `4`) and every other tab faint, besides the existing marks (`▌` for the active tab, `·` for a tab whose document does not exist). It SHALL use only the 16 ANSI terminal colours, bold and faint. The rest of the detail view SHALL keep its look.

#### Scenario: Active tab
- **WHEN** the details of `m1-board-view` open on Tasks in a colour terminal
- **THEN** `1 Tasks` is bold and blue, and `2 Proposal`, `3 Design` and `4 Specs` are faint

#### Scenario: No colour support
- **WHEN** the detail view is shown in a terminal without colours
- **THEN** the tab bar looks as before, with `▌` before the active tab
