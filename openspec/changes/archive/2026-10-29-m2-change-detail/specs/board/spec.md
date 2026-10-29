## ADDED Requirements

### Requirement: Open a Card's details
`Enter` on the selected Card SHALL open the detail view of its Change. In a column without Cards, `Enter` SHALL do nothing.

#### Scenario: Enter on a Card
- **WHEN** Card `m1-board-view` is selected and the user presses `Enter`
- **THEN** the detail view of `m1-board-view` is shown

#### Scenario: Enter without a Card
- **WHEN** the Board has no Changes and the user presses `Enter`
- **THEN** the Board stays as it is
