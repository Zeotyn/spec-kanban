# change-cache Specification

## Purpose

Defines how the plugin remembers a Project's last loaded Changes so the Board can show them immediately and skip redundant openspec CLI calls, and how that memory is validated so stale or broken data is never trusted.

## Requirements

### Requirement: The last successful load is kept per Project
After every successful load of a Project's Changes, the plugin SHALL store them, together with each Change's planning state and the Project path, in a cache file of its own for that Project. The file SHALL be replaced atomically, so a concurrent reader sees either the previous or the new content, never a partial one. A failed load SHALL leave the cache unchanged.

#### Scenario: Cache written after loading
- **WHEN** the Board finishes loading the Changes of Project `magentic`
- **THEN** a cache file for `magentic` holds those Changes and their planning states

#### Scenario: Failed load
- **WHEN** a load fails because the openspec CLI reports an error
- **THEN** the cache file for that Project is unchanged

### Requirement: Cached data is validated before use
Before any cached data is used, the cache file SHALL be checked; the whole file SHALL be ignored when it cannot be read or parsed, when its format version differs from the plugin's, when it belongs to a different Project path, or when an entry lacks a Change name or has an unknown planning value. An ignored file SHALL be replaced by the next successful load.

#### Scenario: Corrupted file
- **WHEN** the cache file contains invalid JSON
- **THEN** the Board opens as if no cache existed and the next successful load overwrites the file

#### Scenario: Older plugin version
- **WHEN** the cache file was written with a different format version
- **THEN** none of its data is used

#### Scenario: Different Project
- **WHEN** the cache file found for a Project names another Project path
- **THEN** none of its data is used

### Requirement: Planning state is reused only for unchanged Changes
During a load, a Change's cached planning state SHALL be used instead of calling `openspec status` only when the cached entry has the same name, exactly the same non-empty last modification time as `openspec list` reports now, the same set of file paths in the Change's directory, the same OpenSpec setup (the Project's `openspec/config.yaml`, its project schemas and the installed openspec CLI), and a known planning state. Otherwise `openspec status` SHALL be called as without a cache. The file set is checked because `openspec list` reports the newest file time, which does not change when an older file is deleted or renamed.

#### Scenario: Nothing changed since the last load
- **WHEN** a Project has 7 Changes without completed tasks, none modified since the last load
- **THEN** the load calls `openspec list` and no `openspec status`

#### Scenario: One Change edited
- **WHEN** one of those 7 Changes was modified since the last load
- **THEN** the load calls `openspec status` for that Change only

#### Scenario: Artifact deleted
- **WHEN** `design.md` of a Change is deleted while a newer file of that Change remains
- **THEN** the load calls `openspec status` for that Change

#### Scenario: Schema or CLI changed
- **WHEN** `openspec/config.yaml` was edited, or the openspec CLI was upgraded, since the last load
- **THEN** the load calls `openspec status` for every Change without completed tasks

#### Scenario: Unknown planning is not reused
- **WHEN** a Change's cached planning state is unknown because its last `status` call failed
- **THEN** the load calls `openspec status` for it again

### Requirement: Cache problems never break the Board
When the cache cannot be read, written or located, the Board SHALL behave exactly as without a cache.

#### Scenario: Read-only state directory
- **WHEN** the cache directory is not writable
- **THEN** the Board loads and shows the Changes normally and shows no cache error
