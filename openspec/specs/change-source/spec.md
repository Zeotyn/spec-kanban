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
