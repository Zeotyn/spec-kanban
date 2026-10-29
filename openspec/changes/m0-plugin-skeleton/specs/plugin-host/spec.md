## Purpose

Defines how the openspec-kanban plugin is built and hosted by herdr: how it is installed, how the open action finds the Project, and how the Board pane opens and closes.

## ADDED Requirements

### Requirement: Plugin builds on install
The plugin SHALL build its `kanban` binary during `herdr plugin install`, and the installation SHALL fail with herdr's build error when the binary cannot be built. For a linked development checkout, the repository SHALL document a single command that builds the binary at the path the manifest uses.

#### Scenario: Install from GitHub with Go available
- **WHEN** a user runs `herdr plugin install zeotyn/openspec-kanban` on a machine with a Go toolchain
- **THEN** the install succeeds and the plugin's actions can be invoked without further steps

#### Scenario: Install without Go
- **WHEN** a user runs `herdr plugin install zeotyn/openspec-kanban` on a machine without a Go toolchain
- **THEN** the install aborts, the plugin is not registered, and herdr shows the failed build command

#### Scenario: Linked checkout
- **WHEN** a developer links the checkout with `herdr plugin link` and runs the documented build command
- **THEN** invoking the open action launches the freshly built binary

### Requirement: Open action opens the Board for the invoking Project
Invoking the `openspec-kanban-open` action SHALL open the Board as a popup covering 90% of the terminal width and height. The Board SHALL show the Project that contains the working directory of the pane the action was invoked from; when that pane has no working directory, the Board SHALL use the focused workspace's working directory.

#### Scenario: Pane inside a Project
- **WHEN** the action is invoked from a pane whose working directory is `<project>/internal/ui`, and `<project>/openspec/changes` exists
- **THEN** a popup opens showing the Board for `<project>` with the Project's directory name as its title

#### Scenario: Pane outside any Project
- **WHEN** the action is invoked from a pane whose working directory has no ancestor containing `openspec/changes`
- **THEN** a popup opens with a notice that no OpenSpec project was found and how to fix it

#### Scenario: herdr cannot open the popup
- **WHEN** the action is invoked while herdr reports the UI as busy (for example, settings are open)
- **THEN** no popup opens and the action's log entry contains herdr's error

### Requirement: Keybinding opens the Board
The repository SHALL provide a herdr config snippet that binds `prefix+o` to the `openspec-kanban-open` action.

#### Scenario: Snippet applied
- **WHEN** the user copies the snippet into `~/.config/herdr/config.toml`, reloads the config, and presses `prefix+o`
- **THEN** the Board opens exactly as if the action had been invoked

### Requirement: Board closes on q and Esc
The Board SHALL close its popup and return focus to the previously focused pane when the user presses `q` or `Esc`.

#### Scenario: Close with q
- **WHEN** the Board is open and the user presses `q`
- **THEN** the popup closes and the previously focused pane has focus again

#### Scenario: Close with Esc
- **WHEN** the Board is open and the user presses `Esc`
- **THEN** the popup closes and the previously focused pane has focus again
