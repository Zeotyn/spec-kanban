## 1. Domain and Sources

- [ ] 1.1 Add `Created string` to `work.Item` and a `Sort` type (`Default`, `ByName`, `ByDate`); `work.Group(items, sort)` orders the open columns by it and leaves Archived as is (design D1, D2); table tests for each sort, undated last and ties by name
- [ ] 1.2 OpenSpec Source: read `created` from each open Change's `.openspec.yaml` in the checkout it comes from; tests for a date, no key, an invalid value and a missing file
- [ ] 1.3 Spec Kit Source: read the first `**Created**:` date from each open Feature's `spec.md`; tests for a date and the `[DATE]` placeholder
- [ ] 1.4 Bump `cache.Version` to 3 and adjust the cache tests

## 2. Board

- [ ] 2.1 Key `o` cycles the sort, regroups, keeps the selection on its Card and works with the filter; `o` in the filter prompt is typed; the sort starts as `default` (design D2)
- [ ] 2.2 Footer shows `sort: name` or `sort: date` before the keys while not `default`; key help lists `o`
- [ ] 2.3 Card line shows `MM-DD · <age>`, or `YYYY-MM-DD · <age>` for another year, and drops the date first when the line does not fit (design D3); golden or render tests for both and for no date

## 3. Docs

- [ ] 3.1 README: `o` in the key table, the creation date on the Card
- [ ] 3.2 CHANGELOG `## Unreleased`: sort by name or date with `o`, creation date on open Cards

## 4. Verification

- [ ] 4.1 `go test -race ./...`, `go vet ./...`, `gofmt -l .` pass
- [ ] 4.2 `openspec validate m15-sort-name-date --strict` passes

## Smoke test (user, in herdr)

These have no checkboxes: they do not count as tasks, and only the user runs them.

- Open the Board, press `o` twice: the open columns are sorted by name, then by date, the footer shows the sort, and every open Card shows its creation date
