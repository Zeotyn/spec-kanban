## Why

Finding what to work on next means scanning every column by eye. The order is fixed: Ready is a queue by name prefix, the other open columns are sorted by last modification, and a Card shows only the age of its last change. When a Change was created is not visible anywhere.

Neither OpenSpec nor Spec Kit has a priority for a whole Change or Feature, so a priority sort would need a field of our own that no OpenSpec or Spec Kit user knows. Both tools do record a creation date: `created` in a Change's `.openspec.yaml`, and `**Created**:` in a Feature's `spec.md`. Sorting by name or by that date, with the date on the Card, uses only what both tools already write.

## What Changes

- **Sort key `o`:** cycles the order of the open columns through `default` (as today: Ready as a queue, Draft, In Progress and Review newest modification first), `name` (natural order, `m2` before `m10`) and `date` (creation date, newest first, undated last). Archived keeps its archive-date order. The footer shows `sort: name` or `sort: date` while it is not `default`. The order resets to `default` when the Board closes.
- **Creation date on the Card:** open Cards show the creation date next to the age, as `11-17 · 3h` (`2025-11-17 · 4w` for another year). A Card without a creation date shows the age only, as today.
- **Sources:** the OpenSpec Source reads `created` from the Change's `.openspec.yaml`; the Spec Kit Source reads the date from the `**Created**:` line of the Feature's `spec.md`.
- Key help and README list `o`.

Non-goals: no priority field, no dependency detection between Changes, no sort order kept across openings of the Board.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `board`: a new requirement for the sort key; the Card content adds the creation date.
- `change-phases`: the ordering within a Phase becomes the `default` sort.
- `change-source`: a new requirement that a Change carries its creation date from `.openspec.yaml`.
- `speckit-source`: a new requirement that a Feature carries its creation date from `spec.md`.

## Impact

- **Code:** `internal/work` (creation date on the Item, sort by name or date), `internal/source/openspec` and `internal/source/speckit` (read the date), `internal/board` (key `o`, footer, Card line, help), `internal/cache` (format version bump, since the stored Item changes).
- **Docs:** README (keys, Card), CHANGELOG.
- No new dependencies.
