## 1. Editor command (TDD)

- [ ] 1.1 Write failing tests for building the editor command: `$EDITOR` unset → `vi`, `$EDITOR` with arguments, one file, several spec files
- [ ] 1.2 Implement the command builder until the tests pass

## 2. Detail view (TDD)

- [ ] 2.1 Write failing tests: `e` on each tab returns an exec command for the right files; missing document and archived Change give a notice and no command; after the editor message the documents are re-read and the tab is kept; an editor error becomes a notice; the next key clears it
- [ ] 2.2 Implement `e` in the detail view until the tests pass

## 3. Board (TDD)

- [ ] 3.1 Write failing tests: `e` on a Card with tasks returns an exec command for its `tasks.md`; without tasks or archived gives a notice; after the editor message the Board reloads and keeps the selection
- [ ] 3.2 Implement `e` on the Board until the tests pass

## 4. Docs and verification

- [ ] 4.1 README: `e` and the `$EDITOR` note for GUI editors
- [ ] 4.2 `go test -race ./...` and `go vet ./...` pass
- [ ] 4.3 Smoke test in herdr: `e` on the Board ticks a task and the Card updates; `e` on Proposal and Specs in the detail view; `e` on an archived Change shows the notice
