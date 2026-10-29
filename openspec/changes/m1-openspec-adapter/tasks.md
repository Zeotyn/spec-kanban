## 1. Fixture project

- [ ] 1.1 Create `testdata/project/` with `openspec/config.yaml` and one Change per case: proposal only, skip_specs fully planned, tasks none done, some done, all done, no `tasks.md`, nested specs; plus `openspec/changes/archive/2026-10-01-old-change/`
- [ ] 1.2 Check each fixture Change with `openspec list --json` and `openspec status` in the fixture and note the expected values in the contract test

## 2. Open Changes and planning state (TDD)

- [ ] 2.1 Write failing tests with a fake runner: list parsing, `status` called only for Changes with 0 done, planning Complete/Incomplete and artifact states, a failing `status` leaves Planning Unknown, empty list, CLI not found message
- [ ] 2.2 Implement `Change`, `OpenNames` and the open-Change part of `Load` with at most 8 parallel `status` calls until the tests pass

## 3. Archived Changes (TDD)

- [ ] 3.1 Write failing tests: dated folder, undated folder, files next to folders ignored, missing archive directory
- [ ] 3.2 Implement archive reading in `Load` until the tests pass

## 4. Contract and reuse

- [ ] 4.1 Add the contract test that runs the real CLI against `testdata/project/` and skips when `openspec` is not on `PATH`
- [ ] 4.2 Switch `internal/start` to `openspec.OpenNames`; its tests stay green

## 5. Verification

- [ ] 5.1 `go test ./...` and `go vet ./...` pass, the contract test runs (not skipped) locally
