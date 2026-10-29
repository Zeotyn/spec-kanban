## ADDED Requirements

### Requirement: Edit tasks from the Board
`e` on the selected Card SHALL open the Change's `tasks.md` in the user's editor (`$EDITOR`, `vi` when unset), suspend the Board while the editor runs, and reload the Changes afterwards, keeping the selection by name. Without `tasks.md`, or for an archived Change, `e` SHALL show a short notice instead.

#### Scenario: Tick a task from the Board
- **WHEN** Card `m1-board-view` (5/11) is selected, the user presses `e`, ticks one task and quits the editor
- **THEN** the Board shows `m1-board-view` with `6/11` and it is still selected

#### Scenario: No tasks yet
- **WHEN** a Draft Card without `tasks.md` is selected and the user presses `e`
- **THEN** no editor opens and the Board says the Change has no tasks to edit
