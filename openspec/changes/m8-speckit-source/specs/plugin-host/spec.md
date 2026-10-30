## MODIFIED Requirements

### Requirement: Open action opens the Board for the invoking Project
Invoking the `openspec-kanban-open` action SHALL open the Board as a popup covering 90% of the terminal width and height. The Board SHALL show the Project that contains the working directory of the pane the action was invoked from; when that pane has no working directory, the Board SHALL use the focused workspace's working directory. A Project is the nearest directory at or above it that any Source recognises: one containing `openspec/changes` (OpenSpec), or one containing both `.specify` and `specs` (Spec Kit).

#### Scenario: Pane inside a Project
- **WHEN** the action is invoked from a pane whose working directory is `<project>/internal/ui`, and `<project>/openspec/changes` exists
- **THEN** a popup opens showing the Board for `<project>` with the Project's directory name as its title

#### Scenario: Pane inside a Spec Kit Project
- **WHEN** the action is invoked from a pane whose working directory is `<project>/src`, and `<project>` contains `.specify/` and `specs/`
- **THEN** a popup opens showing the Board for `<project>` with its Spec Kit Features

#### Scenario: Pane outside any Project
- **WHEN** the action is invoked from a pane whose working directory has no ancestor containing `openspec/changes` or both `.specify` and `specs`
- **THEN** a popup opens with a notice that no project was found, naming both markers as the way to fix it

#### Scenario: herdr cannot open the popup
- **WHEN** the action is invoked while herdr reports the UI as busy (for example, settings are open)
- **THEN** no popup opens and the action's log entry contains herdr's error
