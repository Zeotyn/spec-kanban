## 1. Planning reuse in the adapter (TDD)

- [ ] 1.1 Add JSON tags to `openspec.Change`
- [ ] 1.2 Write failing tests for `Load` with known Changes: unchanged → no `status` call; modified → `status` for that Change only; known Unknown → `status` again; `nil` known → today's behaviour
- [ ] 1.3 Implement the `known` parameter and update callers until the tests pass

## 2. Cache package (TDD)

- [ ] 2.1 Write failing tests for `cache.Read`: missing file, unreadable, invalid JSON, other version, other Project, entry without name → nothing; valid file → its Changes
- [ ] 2.2 Write failing tests for `cache.Write`: round trip through `Read`, atomic replace (no temp file left behind), unwritable directory returns an error without panicking
- [ ] 2.3 Implement `Dir`, `Read` and `Write` until the tests pass

## 3. Board with a snapshot (TDD)

- [ ] 3.1 Write failing tests: snapshot shows columns at once with "refreshing…", fresh load replaces it and clears the marker, selection survives the swap, no snapshot shows the loading notice
- [ ] 3.2 Add the optional snapshot to `board.New` until the tests pass; add a golden for the refreshing header

## 4. Wiring and verification

- [ ] 4.1 Wire `cache.Dir/Read/Write` and `openspec.Load` with known Changes into `kanban board`
- [ ] 4.2 `go test -race ./...` and `go vet ./...` pass
- [ ] 4.3 Measure in herdr: second `prefix+o` in this repo and in magentic shows the Board without the loading notice; time a refresh with nothing changed (expect about one `openspec list`)
