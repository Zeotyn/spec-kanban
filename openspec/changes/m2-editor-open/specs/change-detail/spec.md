## ADDED Requirements

### Requirement: Edit the current document
`e` in the detail view SHALL open the file behind the current tab in the user's editor (`$EDITOR`, `vi` when unset): `tasks.md`, `proposal.md` or `design.md`, and on the Specs tab all of the Change's `spec.md` files. The detail view SHALL be suspended while the editor runs and SHALL show the edited Change afresh afterwards, on the same tab. For a document that does not exist, and for archived Changes, `e` SHALL show a short notice instead of opening an editor.

#### Scenario: Edit the proposal
- **WHEN** the Proposal tab of `m2-editor-open` is shown and the user presses `e`, edits and quits the editor
- **THEN** the detail view returns on the Proposal tab showing the edited text

#### Scenario: Edit all specs
- **WHEN** the Specs tab of a Change with two `spec.md` files is shown and the user presses `e`
- **THEN** the editor opens with both files

#### Scenario: Missing document
- **WHEN** the Design tab of a Change without `design.md` is shown and the user presses `e`
- **THEN** no editor opens and the view says there is no design to edit

#### Scenario: Archived Change
- **WHEN** the details of an archived Change are shown and the user presses `e`
- **THEN** no editor opens and the view says archived Changes are read-only

#### Scenario: Editor fails to start
- **WHEN** `$EDITOR` names a program that does not exist and the user presses `e`
- **THEN** the view shows that the editor could not be started and stays usable
