## ADDED Requirements

### Requirement: Mouse in the detail view
A left click on a tab in the tab bar SHALL show that document, like its number key, in the full and in the short tab bar. Turning the mouse wheel down or up SHALL scroll the document three lines per step, stopping at both ends like `j`/`k`. In the detail footer a left click on `e edit`, `? help` or `q back` SHALL act like its key, in the full and in the short footer; while the footer shows a notice, clicks on it SHALL do nothing. A right click SHALL open the Card menu for the shown Card (see the Board's "Card menu"). Other clicks SHALL change nothing.

#### Scenario: Switch tabs with a click
- **WHEN** the detail view of `m1-board-view` shows Proposal and the user clicks `Design` in the tab bar
- **THEN** the Design document is shown

#### Scenario: Scroll with the wheel
- **WHEN** the Proposal is 80 lines long, shown from the top, and the user turns the wheel down 2 steps
- **THEN** the Proposal is shown from its seventh line

#### Scenario: Back with a click
- **WHEN** the user opened the details of the third Card in Ready and clicks `q back`
- **THEN** the Board is shown with the third Card in Ready selected
