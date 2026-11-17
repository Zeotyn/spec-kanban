## ADDED Requirements

### Requirement: Creation date of a Feature
An open Feature SHALL carry its creation date from the first line of its `spec.md` that starts with `**Created**:` and holds a date `YYYY-MM-DD`. When `spec.md` is missing, has no such line or the line holds no valid date (like the template's `[DATE]`), the Feature SHALL have no creation date, and loading SHALL go on without an error.

#### Scenario: Date from the spec header
- **WHEN** `specs/001-user-login/spec.md` holds the line `**Created**: 2026-10-30`
- **THEN** Feature `001-user-login` has the creation date 2026-10-30

#### Scenario: Placeholder date
- **WHEN** a Feature's `spec.md` holds `**Created**: [DATE]`
- **THEN** the Feature has no creation date and the Board loads without an error
