# change-source Specification

## Purpose

Defines which Changes of a Project the plugin knows about and what it knows about each, read through the OpenSpec CLI and the archive directory, including how failures are reported.
## Requirements
### Requirement: Open Changes come from OpenSpec
The change source SHALL report every open Change that `openspec list` reports for the Project, each with its name, number of completed tasks, total number of tasks, and last modification time.

#### Scenario: Project with open Changes
- **WHEN** the Project has the open Changes `a` (3 of 5 tasks done), `b` (no tasks) and `c` (2 of 2 done)
- **THEN** the source reports exactly `a`, `b` and `c` with those counts and their last modification times

#### Scenario: No open Changes
- **WHEN** the Project has no open Changes
- **THEN** the source reports no open Changes and no error

### Requirement: Planning state for Changes without completed tasks
For every open Change with no completed task, the source SHALL report whether its planning is complete and the state of each planning artifact (done, ready, blocked or skipped). For Changes with at least one completed task, planning state SHALL NOT be read.

#### Scenario: Proposal only
- **WHEN** a Change has only `proposal.md`
- **THEN** its planning is reported incomplete, with proposal done and the other artifacts not done

#### Scenario: Specs skipped
- **WHEN** a Change sets `skip_specs: true` and has proposal, design and tasks
- **THEN** its planning is reported complete and its specs artifact as skipped

#### Scenario: Work has started
- **WHEN** a Change has 1 of 4 tasks done
- **THEN** no planning state is read for it

### Requirement: Archived Changes come from the archive directory
The source SHALL report each directory under `openspec/changes/archive/` as an archived Change. A leading `YYYY-MM-DD-` is the archive date and is removed from the name; a directory without it keeps its full name and has no date.

#### Scenario: Dated archive folder
- **WHEN** the archive contains `2026-10-29-m0-plugin-skeleton`
- **THEN** the source reports an archived Change `m0-plugin-skeleton` archived on 2026-10-29

#### Scenario: No archive directory
- **WHEN** `openspec/changes/archive/` does not exist
- **THEN** the source reports no archived Changes and no error

### Requirement: Failures are reported, not hidden
The source SHALL fail with a message that names the cause when the `openspec` CLI is missing or `openspec list` fails. When reading the planning state of a single Change fails, that Change SHALL be reported with unknown planning state and all other Changes SHALL be unaffected.

#### Scenario: CLI not installed
- **WHEN** `openspec` is not on `PATH`
- **THEN** the source fails with a message saying the openspec CLI was not found

#### Scenario: One status call fails
- **WHEN** `openspec status` fails for Change `b` but succeeds for the others
- **THEN** `b` is reported with unknown planning state and every other Change is reported normally

### Requirement: Changes worked on in another checkout
When the Project is the top directory of a git checkout, the source SHALL read each other checkout of that repository (a linked worktree, or the main checkout when the Project is a linked worktree) whose branch is named exactly like a Change and which contains that Change's directory, through `openspec list` run in that checkout. The Change of that name as reported there, with its Progress, last modification time and planning state read there, SHALL replace the Project's open Change of that name, or be added as an open Change when the Project has no Change of that name. The source SHALL report for every open Change the directory its files are read from, which the Board's detail view and `e` use. Other Changes of that checkout, a Change the Project has archived, and a checkout whose read fails or that does not report the Change SHALL leave the Project's Changes as they are, without an error. A Project that is not the top directory of a git checkout SHALL be read as before.

#### Scenario: Agent ticks tasks in its worktree
- **WHEN** Change `m1-board-view` has 0 of 6 tasks done in the Project and 3 of 6 in the linked worktree on branch `m1-board-view`
- **THEN** the source reports `m1-board-view` with 3 of 6 tasks done and the worktree's last modification time, and its tasks file is the worktree's

#### Scenario: Change created in its worktree
- **WHEN** the linked worktree on branch `m2-export` contains open Change `m2-export` and the Project has no Change `m2-export`
- **THEN** the source reports the open Change `m2-export` as read in that worktree

#### Scenario: Other Changes of a worktree
- **WHEN** the worktree on branch `m1-board-view` also contains an open Change `m9-other` that the Project does not have
- **THEN** the source does not report `m9-other`

#### Scenario: Branch without the Change
- **WHEN** a linked worktree is on branch `fix-typo` and no Change `fix-typo` exists in it
- **THEN** the source does not run `openspec list` there and reports the Project's Changes unchanged

#### Scenario: Archived after merge
- **WHEN** the Project has archived `m1-board-view` and a stale worktree on branch `m1-board-view` still reports it open
- **THEN** the source reports `m1-board-view` only as archived

#### Scenario: Worktree read fails
- **WHEN** `openspec list` fails in the worktree on branch `m1-board-view`
- **THEN** the source reports the Project's own `m1-board-view` and no error

