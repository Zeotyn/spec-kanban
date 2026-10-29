## 1. Phase rules (TDD)

- [ ] 1.1 Write a failing table test for `phase.Of` covering every scenario of the change-phases spec, including the precedence cases (archived with open tasks, progress with incomplete planning, unknown planning)
- [ ] 1.2 Implement `Phase` (Draft, Ready, InProgress, Review, Archived with display names) and `Of(openspec.Change)` until the test passes

## 2. Ordering (TDD)

- [ ] 2.1 Write failing tests for `phase.Group`: Changes split into the five Phases in column order, open Phases newest first, Archived by date then name, undated last
- [ ] 2.2 Implement `Group` until the tests pass

## 3. Glossary and verification

- [ ] 3.1 Update the Draft definition in `CONTEXT.md` so progress outranks planning
- [ ] 3.2 `go test ./...` and `go vet ./...` pass
