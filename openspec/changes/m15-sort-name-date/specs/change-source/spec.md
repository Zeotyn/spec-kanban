## ADDED Requirements

### Requirement: Creation date of a Change
An open Change SHALL carry its creation date from the `created` key of its `.openspec.yaml`, as a calendar date `YYYY-MM-DD`. When the file is missing, cannot be read, has no `created` key or holds no valid date, the Change SHALL have no creation date, and loading SHALL go on without an error. A Change read from another checkout SHALL take the date from that checkout.

#### Scenario: Date from the metadata
- **WHEN** `openspec/changes/m1-board-view/.openspec.yaml` holds `created: 2026-11-17`
- **THEN** Change `m1-board-view` has the creation date 2026-11-17

#### Scenario: No date
- **WHEN** a Change's `.openspec.yaml` has no `created` key
- **THEN** the Change has no creation date and the Board loads without an error
