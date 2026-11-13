# failure-log Specification

## Purpose
TBD - created by archiving change m13-failure-log. Update Purpose after archive.
## Requirements
### Requirement: Failures are recorded
Every failure that makes herdr show a `spec-kanban: <step> failed` notification SHALL also append one record to the failure log. This covers `start` (a Start from the Board and a wave member's start), `run` (Verify and Archive), `resume` and `wave` (a wave start). A record SHALL hold the time, the Project, the Change, the step and the full error text, including the pane's last lines when the error carries them. For a wave the Change SHALL be the wave's Changes, separated by spaces. The log SHALL be one file `failures.log` in the user's cache directory (`spec-kanban` below the operating system's user cache directory), so that a plain shell or an agent pane can read it. It SHALL keep the newest 100 records over all Projects and drop older ones. Failures shown only as a notice in the Board's footer SHALL not be recorded.

#### Scenario: Verify fails after the Board closed
- **WHEN** Verify on `s001-plugin-bundles-hooks` fails because the shell did not reach its prompt
- **THEN** herdr shows `spec-kanban: run failed`, and the log gains a record with the Project, `s001-plugin-bundles-hooks`, step `run`, the time and the full error with the pane's last lines

#### Scenario: Wave start fails
- **WHEN** a wave of `c1-greet` and `c2-farewell` fails to launch its coordinator
- **THEN** herdr shows `spec-kanban: wave failed`, and the log gains a record with step `wave`, Change `c1-greet c2-farewell` and the error

#### Scenario: Log is full
- **WHEN** the log holds 100 records and another failure is recorded
- **THEN** the log holds the newest 100 records and the oldest one is gone

#### Scenario: Failure before the Board closes
- **WHEN** Start fails while the Board is still open and shows the error in its footer
- **THEN** no record is written

### Requirement: The notification points to the log
The body of every `spec-kanban: <step> failed` notification SHALL keep the error text and end with the line `Details: kanban failures`.

#### Scenario: Run failed
- **WHEN** Verify fails after the Board closed
- **THEN** the notification body shows the error and ends with `Details: kanban failures`

### Requirement: A log problem never breaks a launch
When the record cannot be written, the notification SHALL still be shown, with the same body, and the failing command SHALL report its original error. A log that cannot be read or holds a line that is not a valid record SHALL be read as the valid records it holds; the next write SHALL keep only those.

#### Scenario: Cache directory not writable
- **WHEN** Start fails after the Board closed and the cache directory is not writable
- **THEN** herdr shows `spec-kanban: start failed` with the error, and nothing else changes

#### Scenario: Broken line
- **WHEN** the log holds a line that is not a valid record
- **THEN** `kanban failures` prints the other records and no error

### Requirement: Read the failures
`kanban failures` SHALL print the recorded failures of the current directory's Project, oldest first. The Project SHALL be the OpenSpec or Spec Kit Project at or above the current directory; when the log holds no failure of it, the first other checkout of its git repository that has failures, as `kanban escalate` finds the Project. Outside any Project, `kanban failures` without `--all` SHALL fail saying so. Each failure SHALL be printed as a header line with the local time, the step, the Change and the Project, followed by the full error text indented by two spaces and a blank line. With `--all` it SHALL print the failures of every Project. Without failures to print it SHALL print `no failures` and exit successfully. Any other argument SHALL print the usage and exit with an error.

#### Scenario: Failures of the current Project
- **WHEN** the log holds two failures of `/code/a` and one of `/code/b`, and the user runs `kanban failures` in `/code/a/internal`
- **THEN** it prints the two failures of `/code/a`, oldest first, each with its full error text

#### Scenario: Run from a worktree
- **WHEN** the user runs `kanban failures` in a worktree of `/code/a`'s repository
- **THEN** it prints the failures of `/code/a`

#### Scenario: All Projects
- **WHEN** the user runs `kanban failures --all`
- **THEN** it prints all three failures, oldest first

#### Scenario: Nothing recorded
- **WHEN** the log holds no failure of the current directory's Project
- **THEN** `kanban failures` prints `no failures`

