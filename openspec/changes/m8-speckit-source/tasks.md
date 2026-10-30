Every group ends with `go test ./...` green. OpenSpec goldens and Handoff texts stay byte-identical; the only golden allowed to change is `internal/board/testdata/golden/no-project.txt` (group 7).

## 1. Widen the Source interface (TDD, no behaviour change)

- [x] 1.1 Write failing tests: `openspec.Source.Prepare(dir, name)` returns nil and writes nothing; `Prompt(proj, "codex") == "/opsx:apply {change}"`; `Finish("x", "") == Finish("x", "/wt") == "archive it with /opsx:archive x"`; in `start`, `Watched.text` with a fake Source whose `Finish` returns `""` gives `X is finished (9/9 tasks). Please review it.` without a checkout and `…, merge branch X into main, then remove its worktree with: git worktree remove <path>. …` with one (existing OpenSpec text cases unchanged)
- [x] 1.2 Change `work.Source` (design.md: `Prepare`, `Prompt(project, kind)`, `Finish(name, checkout)`), adapt `openspec.Source`, the fake Sources in `start` tests, `Watched.text`, and the `Prompt` callers in `cmd/kanban` (pass `proj` and the resolved kind)
- [x] 1.3 Write failing `start.Prepare` tests with a recording fake Source: `Prepare(dir, change)` is called with the Project (no checkouts / Project on the branch), the existing worktree path (`worktree open`), `result.worktree.path` of `worktree create`, and the checkout containing an idle agent's `cwd`; not called when a busy agent is only focused; a `Prepare` error returns no pane and an error naming the change and dir
- [x] 1.4 Implement the calls in `start.Prepare` (read `worktree.path` in `created` or next to it); all herdr/git call sequences in existing tests unchanged

## 2. Spec Kit detection and Source choice (TDD)

- [x] 2.1 New package `internal/source/speckit` with `New(git func(dir string) herdr.Runner, base string) *Source` and `var _ work.Source = (*Source)(nil)`; failing tests: `Detect` true only with both `.specify/` and `specs/` directories (not `.specify` alone, not files of those names); `Rel("001-a") == "specs/001-a"`; `ArchivedIn` always false
- [x] 2.2 Implement `Detect`, `Rel`, `ArchivedIn`
- [x] 2.3 Config: failing tests for `source` (`openspec`, `speckit` accepted; `spec-kit`, `""`, a number rejected with defaults and a warning naming `source`); implement `Config.Source`
- [x] 2.4 Config `instructions`: failing tests for `Parse`/`Load`/`PromptFor`: `{"instructions": "with ponytail"}` → `PromptFor("m1", "/opsx:apply {change}") == "/opsx:apply m1 with ponytail"` and `PromptFor("m1", "/speckit.implement") == "/speckit.implement with ponytail"`; `""` → `/opsx:apply m1`, no trailing space, no warning; text passed verbatim (a `{change}` in it is not replaced, surrounding spaces kept); a list or number → defaults and a wrong-type warning naming `instructions`; `prompt` + `instructions` → `Load` keeps all settings (e.g. `hide`), `Instructions` empty, `PromptFor` uses `prompt`, and the warning names the file and says `instructions` is ignored because `prompt` is set; existing `prompt` cases unchanged
- [x] 2.5 Implement `Config.Instructions`, the append in `PromptFor`, and the kept-settings warning in `Parse`/`Load`; a `cmd/kanban` test that `kanban start` with that config prints the warning to stderr and sends the `prompt`
- [x] 2.6 Write failing tests in `cmd/kanban` for `sources(cfg)`: `detect` accepts either marker; `sourceFor` picks OpenSpec for an OpenSpec-only dir, Spec Kit for a Spec Kit-only dir, OpenSpec for both without preference, Spec Kit for both with `source: speckit`, and OpenSpec for an OpenSpec-only dir with `source: speckit`; implement it

## 3. Features, Progress and Phase (TDD)

- [x] 3.1 Add fixtures under `internal/source/speckit/testdata/` modelled on `~/code/speckit-spike` (real `spec.md` header, `tasks.md` with `## Phase N:` headings, `T001 [P] [US1]` lines, `[x]` and `[X]`): spec only; spec+plan+research; tasks none checked; some checked; all checked; folder without `spec.md`; `spec.md` with `**Status**: Draft` and all tasks checked; a timestamp-named folder
- [x] 3.2 Write failing `Load` tests: one Item per folder with `spec.md`, named like the folder; `Done`/`Total` from checkboxes; `Modified` = newest file at any depth; Phase per the speckit-source spec (Draft, Draft, Ready, In Progress, Review, none, Review); `Artifacts` → `planMarks` renders `- S · ·`, `- S D ·`, `- S D T`; `Stamp` empty; unreadable `specs/` → error; `Open` returns the Feature names
- [x] 3.3 Write failing tests for other checkouts: a checkout on branch `b` with `specs/b/spec.md` replaces the Project's `b` (Progress from there, `Dir` set) or adds it; the Project's own entry in `others`, a branch without such a folder, and a non-Feature branch are skipped
- [x] 3.4 Implement `Load`, `Open`, `phaseOf` and the checkbox reader (copy of the OpenSpec regex)

## 4. Merged branch archives a Feature (TDD)

- [x] 4.1 Write failing tests with a fake git runner: no Review Feature → no git call at all; a Review Feature whose name `git branch --merged <base> --format=%(refname:short) %(worktreepath)` lists unchecked-out → Archived with empty `ArchivedOn`; not listed → Review; a Ready Feature whose branch is listed (fresh branch) → Ready; a merged branch still checked out → Review; git failing → Review; configured base passed through, else `worktree.Base` detection
- [x] 4.2 Write one real-git test in a temp repo: branch with all tasks checked merged into `main` → Archived; unmerged branch with commits → Review; branch just created from `main` with tasks left → Ready; a worktree on a fresh branch with all tasks checked but not committed → Review
- [x] 4.3 Implement the merged rule in `Load`

## 5. Detail tabs (TDD)

- [x] 5.1 Write failing `Docs` tests: all four files → Tasks, Spec, Plan, Research with `Files` per tab, `Tasks: 0`, `First: 1`, `Next` = `next ▸ T004 …` for the first open line or `all tasks done`; spec only → Tasks (empty body, `Files` = `tasks.md` path) and Spec; plan without research → no Research tab; an Item with `Dir` reads from that checkout
- [x] 5.2 Implement `Docs`; a `detail` test with the Spec Kit `Docs` of a spec-only Feature: two tabs, opens on Spec, `3` does nothing, the Tasks tab says there are no tasks

## 6. Prepare, prompt and Finish (TDD)

- [x] 6.1 Write failing `Prepare` tests: writes `.specify/feature.json` = `{"feature_directory":"specs/<name>"}` in `dir`, creating `.specify/`, replacing an existing file; an unwritable `dir` returns an error naming the file
- [x] 6.2 Write failing `Prompt(project, kind)` tests against `integration.json` fixtures: the spike's file + `claude` → `/speckit-implement`; kind `codex` installed → `$speckit-implement`; kind not installed → the default integration's command; `gemini` with separator `.` → `/speckit.implement`; no file, broken JSON, no separator → `/speckit-implement`
- [x] 6.3 Write failing `Finish` tests: `Finish(n, "") == ""`; `Finish(n, "/wt") == "delete /wt/.specify/feature.json unless git tracks it"`; a `start` watch test with the Spec Kit Source asserts both Handoff texts of the speckit-source spec
- [x] 6.4 Implement `Prepare`, `Prompt`, `Finish`

## 7. Wiring and texts

- [x] 7.1 `cmd/kanban`: build `openspec.New(openspecCLI)` and `speckit.New(gitCLI, cfg.Base)` through `sources(cfg)`; `open` uses `detect`; `start`, `board`, `launch` use `sourceFor(proj)`; `sync` loads the config (warnings ignored) and passes `detect` and a `syncLoad` that picks the Source per Project; existing `main_test.go` cases unchanged
- [x] 7.2 No-project notice `No project here.` / `… (a directory containing openspec/changes, or .specify and specs).`, update the `no-project` golden; `kanban start` outside a Project: `no OpenSpec or Spec Kit project at or above <cwd>`
- [x] 7.3 Check the boundary: `grep -rn 'speckit\|\.specify\|feature\.json' --include='*.go' internal cmd` hits only `internal/source/speckit`, `internal/config` (the `source` value) and `cmd/kanban`

## 8. Docs

- [x] 8.1 README: section "Spec Kit" (detection and `source`, Phases and the merged-branch rule with its Review fallback, planning marks `- S D T`, tabs, `feature.json` written on `s` and deleted in the Handoff, default prompt per integration plus `instructions`, and that a configured `prompt` applies to every Project, commit `specs/<feature>` before `s`); `source` and `instructions` rows in the Settings table (`prompt` row: complete override, wins over `instructions`); the Settings example switches from `prompt` to `{"instructions": "with ponytail, TDD and review loops"}`; the `s` paragraph says `/opsx:apply <change>` plus `instructions` unless `prompt` is configured
- [x] 8.3 CHANGELOG `Unreleased`: Spec Kit Source; new config keys `source` and `instructions` (appended to the Source's command; `prompt` still wins)
- [x] 8.2 CONTEXT.md: **Feature** (Spec Kit kind of Work item), Source "OpenSpec or Spec Kit", Project with both markers, Archived also for a Feature whose branch is merged; ADR 0009 `docs/adr/0009-speckit-source.md` (merged branch counts as Archived; one Source per Project, OpenSpec wins a tie, `source` prefers; considered: Review as final column, a "Done" rename, mixing Sources)

## 9. Verification

- [x] 9.1 `go test -race ./...`, `go vet ./...`, `gofmt -l .` pass; `git diff --stat main -- '*/testdata/golden/*'` lists only `no-project.txt`
- [x] 9.2 `openspec validate m8-speckit-source --strict` passes

## 10. Smoke test (user, in herdr, with `~/code/speckit-spike`)

- [ ] 10.1 Prepare the spike on `main` (from its root; `001-user-login-email` already has `spec.md` only → Draft):
  - `002-export`: `mkdir -p specs/002-export && cp specs/001-user-login-email/spec.md specs/002-export/ && printf '# Plan\n' > specs/002-export/plan.md && printf '# Research\n' > specs/002-export/research.md` → Draft `- S D ·`
  - `003-search`: spec + plan + `tasks.md` with `## Phase 1: Setup` and `- [ ] T001 [P] [US1] Create search index`, `- [ ] T002 [US1] Add query endpoint`, `- [ ] T003 [US1] Wire UI` → Ready
  - `004-reports`: same, `T001` as `[X]` → In Progress
  - `005-audit`: all three `[x]` on `main`, no branch → Review
  - `006-billing`: spec + plan + three open tasks; `git checkout -b 006-billing`, check all three, commit, `git checkout main && git merge --no-ff 006-billing` → Archived
  - commit `specs/` (not `.specify/feature.json`) on `main`
- [ ] 10.2 Board: open it from a pane in `speckit-spike/specs`; columns as listed in 10.1, no milestone badges, Archived Card `006-billing` without a date; `/00` filters all, `/search` one; Enter on `002-export` shows Tasks, Spec, Plan, Research and opens on Spec; `e` on Plan edits `plan.md`; `e` on the Board for `003-search` opens `tasks.md`; `r` after checking a task in an editor moves `003-search` to In Progress (then revert)
- [ ] 10.3 Sidebar: a workspace in the spike shows the most recently modified Feature with its Progress and `2 draft · 1 ready · 1 wip · 1 review`-style counts
- [ ] 10.4 Start: `s` on `003-search` creates worktree `~/.herdr/worktrees/speckit-spike/003-search`, whose `.specify/feature.json` names `specs/003-search` before Claude starts; Claude receives `/speckit-implement` and its prerequisite check finds the Feature; while it works the Card moves to In Progress (read from the worktree); after it checks all tasks it hands off with a prompt asking to review, merge branch `003-search` into `main`, delete `<worktree>/.specify/feature.json` unless git tracks it, then remove the worktree, with no archive command; after the Main agent merges and removes, `r` shows `003-search` in Archived
- [ ] 10.5 Shell and config: `kanban start 004-reports` from a spike pane with `{"worktree": false}` writes the spike's own `.specify/feature.json` for `004-reports` and starts a tab in the spike; `kanban start 999-x` fails as not open; with `{"agent": "codex"}` (codex installed as integration or not) the prompt follows 6.2; with `{"instructions": "with ponytail"}` Claude receives `/speckit-implement with ponytail`, and with `prompt` added too the Board footer warns and the `prompt` is sent
- [ ] 10.6 Both markers: `mkdir -p openspec/changes` in the spike; the Board shows the (empty) OpenSpec Board; with `{"source": "speckit"}` it shows the Features again; `{"source": "spec-kit"}` shows the config warning; remove `openspec/` afterwards
- [ ] 10.7 Regression: this repo's Board, detail, `s` and Handoff behave as before this change (`/opsx:apply`, `/opsx:archive` texts unchanged); with `{"instructions": "with ponytail"}` `s` sends `/opsx:apply <change> with ponytail`
