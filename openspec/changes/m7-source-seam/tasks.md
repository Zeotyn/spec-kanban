Every group ends with `go test ./...` green and no golden file changed (`git status internal/*/testdata` clean). A task that needs a golden update is wrong.

## 1. Move OpenSpec code into the Source package

- [ ] 1.1 `git mv internal/openspec internal/source/openspec` (package name stays `openspec`); fix the imports in `board`, `cache`, `detail`, `docs`, `phase`, `sidebar`, `start`, `cmd/kanban`; tests green, contract tests included
- [ ] 1.2 Move `internal/docs` (`Dir`, `Read`, `Docs` and their tests) into `internal/source/openspec/docs.go`; fix `board` and `detail`; delete `internal/docs`

## 2. Neutral domain types (TDD)

- [ ] 2.1 Write failing tests in `internal/work`: `Phase.String` for all five Phases; `Group` places Items by their `Phase` field and orders them as the change-phases spec says (port the ordering cases of `phase_test.go` onto `Item` with `Phase` set: newest `Modified` first, Archived by `ArchivedOn` newest first, undated last, ties by name)
- [ ] 2.2 Implement `work.Phase` (constants, `String`, comment that the int is cached), `work.Item` (design.md), `work.Group`, `work.Doc`/`work.Docs`; `internal/phase` stays for now
- [ ] 2.3 Write failing tests in `internal/source/openspec` for `phaseOf` (port the `phase.Of` table: archived, no tasks, all done, some done with incomplete planning, complete, incomplete, unknown) and for a temporary `Items([]Change) []work.Item` (fields copied, `Phase` from `phaseOf`, `Files` → `Stamp`, `Archived` → Phase Archived)
- [ ] 2.4 Implement `phaseOf` and `Items`

## 3. Switch consumers one at a time

- [ ] 3.1 `sidebar`: `Summary` and `Sync` take `[]work.Item`, count `c.Phase`, skip `Phase == work.Archived`; `syncLoad` in `cmd/kanban` converts with `openspec.Items`; summary tests build Items with `Phase` set
- [ ] 3.2 `board` and `config`: `board.New`, `load`, snapshot, columns and Cards use `work.Item`/`work.Group`/`work.Phase`; `config.Hide` is indexed by `work.Phase`; `cmd/kanban` converts snapshot and loads; board tests set `Phase` instead of `Planning`/`Archived`
- [ ] 3.3 Write a failing `detail` test with a three-tab fake `work.Docs` (tab bar, `1-3` footer, first tab choice from `Tasks`/`First`, `next ▸` on the tasks tab, `e` opens the tab's `Files`, empty tab notice); add `openspec.ReadDocs(project string, it work.Item) work.Docs` (Tasks, Proposal, Design, Specs; Tasks 0, First 1) with tests ported from `docs_test.go`
- [ ] 3.4 `detail.New(read func() work.Docs, it work.Item, header, w, h)` renders `Tabs`, re-calls `read` after the editor; `Dir` and `Read` become unexported; `board.edit` opens `Tabs[Tasks].Files[0]` with the existing stat check; detail and board goldens unchanged
- [ ] 3.5 `start`: `Watched.Load` returns `[]work.Item`, `check` reads `Phase == work.Archived`; `launch` converts with `openspec.Items`; watch tests build Items

## 4. Load returns Items, cache version 2 (TDD)

- [ ] 4.1 Change the `Load` tests to expect `[]work.Item` with `Phase` decided and `Stamp` set; port the reuse tests to the new rule (same `Dir`, same non-zero `Modified`, same non-empty `Stamp`, non-nil `Artifacts` → `Phase` and `Artifacts` copied, no `status` call); keep the change-cache scenarios: nothing changed, one edited, artifact deleted, setup or CLI changed, unknown planning re-read
- [ ] 4.2 Write failing cache tests: round trip of `work.Item`; a version 1 file is ignored; an entry with empty name or a Phase outside Draft…Archived invalidates the file
- [ ] 4.3 Implement: `Load` returns `[]work.Item` (`status` failure → nil `Artifacts`, Phase Draft); `cache` stores `[]work.Item` with `Version = 2`; `loader` keys `known` by name for `Phase != work.Archived`; delete `Change`, `Planning`, `Items` and `internal/phase`

## 5. Put the remaining specifics behind `Source` (TDD)

- [ ] 5.1 Add `work.Source` (design.md) and `openspec.New(cli) *Source` with the methods `Detect`, `Load`, `Open` (was `OpenNames`), `ArchivedIn`, `Docs` (was `ReadDocs`), `Rel`, `Prompt`, `Finish`; tests: compile-time `var _ work.Source = (*Source)(nil)`, `Detect` true only with `openspec/changes`, `Rel("x") == "openspec/changes/x"`, `Prompt() == "/opsx:apply {change}"`, `Finish("x") == "archive it with /opsx:archive x"`
- [ ] 5.2 `project.Find(start string, detect func(dir string) bool)`; port the project tests with `src.Detect`; `open` and `startChange` pass it
- [ ] 5.3 `config`: `Default().Prompt == ""`, `PromptFor(change, def string)` uses `def` when no prompt is configured; config tests: no file and every invalid file still send `/opsx:apply <change>` through `src.Prompt()`
- [ ] 5.4 `start`: `Start` and `Prepare` take a `work.Source`; `checkOpen` uses `src.Open`, the ref check `git cat-file -e <ref>:` + `src.Rel(change)`; `Watched` gets `Source` for `ArchivedIn` and `Finish`; start and watch tests assert the same herdr/git calls and Handoff texts byte for byte, with a fake Source in the cases that do not need the real one
- [ ] 5.5 `cmd/kanban`: build one `openspec.New(openspecCLI)` and pass it to `project.Find`, `loader`, `syncLoad`, `start`, `launch`, `showBoard` (detail `read`); `loader` calls `src.Load`
- [ ] 5.6 Check the boundary: `grep -rn 'opsx\|openspec/changes\|internal/source/openspec' --include='*.go' internal cmd` hits only `internal/source/openspec` and the Source construction in `cmd/kanban`

## 6. Verification

- [ ] 6.1 `go test -race ./...` (with the openspec contract tests), `go vet ./...` and `gofmt -l .` pass; `git diff --stat main -- '*/testdata/*'` is empty
- [ ] 6.2 Smoke test in herdr against this repo, compared with 0.2.0: Board opens from cache and reloads (first open after upgrade without snapshot), columns and Cards with planning marks and milestone badges, filter `m6`, detail tabs 1-4, `e` on tasks and specs, sidebar rows; `s` on a Ready Card creates the worktree and sends `/opsx:apply <change>`; a finished agent hands off with `/opsx:archive`; `s` on Review resumes; `kanban start <change>` from a shell; a configured prompt and `hide: ["Archived"]` still apply
