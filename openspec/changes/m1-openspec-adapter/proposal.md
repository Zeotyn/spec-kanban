## Why

Every Board feature needs the Changes of a Project. ADR 0002 makes the `openspec` CLI the data source; this change puts all CLI calls behind one adapter, so the unversioned JSON format is handled in one place and guarded by contract tests.

## What Changes

- Read open Changes with `openspec list --json` (name, task counts, last modified, status).
- Read per-Change planning state with `openspec status --change <name> --json`, in parallel and in the background.
- Read archived Changes from `openspec/changes/archive/` on disk.
- Fixture Project under `testdata/project/` covering every Phase, nested specs, and a Change without `tasks.md`.
- Contract tests that run the real CLI against the fixture and are skipped when `openspec` is missing.
- Replace the direct CLI call in `kanban start` with the adapter.

## Capabilities

### New Capabilities
- `change-source`: which Changes of a Project are known, with what data, and how missing or failing CLI calls are reported.

### Modified Capabilities

## Impact

- New `internal/openspec/` package and `testdata/project/`. Runtime dependency on the `openspec` CLI on `PATH`.
