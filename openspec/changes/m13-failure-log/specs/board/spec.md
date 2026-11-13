## ADDED Requirements

### Requirement: Unseen failure on opening
When the Board opens and the failure log holds a failure of the Board's Project that is newer than the Board's previous opening for that Project, the footer SHALL show the newest such failure once as a notice, in place of the key footer until the next key: `<step> failed <change> <age> ago: <first line of the error>`, cut to the footer's width like every notice. The age SHALL be rounded down to seconds, minutes, hours or days (`40s`, `4m`, `2h`, `3d`). Opening the Board SHALL record the opening time for the Project in the user's cache directory. A failure shown once, or older than the previous opening, SHALL not be shown again. When the Board opens with a config warning, the config warning SHALL be shown instead and the opening SHALL not be recorded, so the failure shows at the next opening. The first opening of a Project with no recorded opening SHALL show the newest failure of the last 24 hours. When the log or the opening record cannot be read or written, the Board SHALL open without the notice and without an error.

#### Scenario: Verify failed while the Board was closed
- **WHEN** Verify on `s001-plugin-bundles-hooks` failed 4 minutes ago with `shell did not reach its prompt in tab s001-plugin-bundles-hooks`, and the user opens the Board of that Project
- **THEN** the footer shows `run failed s001-plugin-bundles-hooks 4m ago: shell did not reach its prompt in tab s001-plugin-bundles-hooks` until the next key

#### Scenario: Shown only once
- **WHEN** the user closes the Board after that notice and opens it again without a new failure
- **THEN** the footer shows the key footer

#### Scenario: Config warning first
- **WHEN** the Board opens with an invalid `config.json` and an unseen failure
- **THEN** the footer shows the config warning, and the next opening with a valid config shows the failure

#### Scenario: Another Project's failure
- **WHEN** the only unseen failure belongs to another Project
- **THEN** the footer shows the key footer
