## 1. Planning reuse in the adapter (TDD)

- [x] 1.1 Add JSON tags to `openspec.Change`
- [x] 1.2 Write failing tests for `Load` with known Changes: unchanged → no `status` call; modified → `status` for that Change only; known Unknown → `status` again; `nil` known → today's behaviour
- [x] 1.3 Implement the `known` parameter and update callers until the tests pass

## 2. Cache package (TDD)

- [x] 2.1 Write failing tests for `cache.Read`: missing file, unreadable, invalid JSON, other version, other Project, entry without name → nothing; valid file → its Changes
- [x] 2.2 Write failing tests for `cache.Write`: round trip through `Read`, atomic replace (no temp file left behind), unwritable directory returns an error without panicking
- [x] 2.3 Implement `Dir`, `Read` and `Write` until the tests pass

## 3. Board with a snapshot (TDD)

- [x] 3.1 Write failing tests: snapshot shows columns at once with "refreshing…", fresh load replaces it and clears the marker, selection survives the swap, no snapshot shows the loading notice
- [x] 3.2 Add the optional snapshot to `board.New` until the tests pass; add a golden for the refreshing header

## 5. Review follow-ups

- [x] 5.1 Fingerprint of the Change's file paths in the reuse check; zero Modified never matches; tests for a deleted artifact
- [x] 5.2 `cache.Read` rejects unknown planning values; archived entries do not shadow open ones in `known`
- [x] 5.4 Include config.yaml, project schemas and the openspec CLI version in the fingerprint; test that a config edit and an upgrade force `status`
- [x] 5.3 Extract the loader from `main` and test it: write on success only, re-read per load, unwritable directory ignored

## 4. Wiring and verification

- [x] 4.1 Wire `cache.Dir/Read/Write` and `openspec.Load` with known Changes into `kanban board`
- [x] 4.2 `go test -race ./...` and `go vet ./...` pass
- [ ] 4.3 Measure in herdr: second `prefix+o` in this repo and in magentic shows the Board without the loading notice; time a refresh with nothing changed (expect about one `openspec list`)
