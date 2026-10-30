## 1. Checkouts and base (TDD, fake git runner)

- [x] 1.1 Capture `git worktree list --porcelain` output (main + linked, one detached, one prunable) into `internal/worktree/testdata/` with fake paths
- [x] 1.2 Write failing tests for `worktree.List(project, git)`: top-level Project → all non-bare, non-detached, non-prunable checkouts with path and branch; non-empty `rev-parse --show-prefix` → empty list; git failing (outside git) → empty list, no error
- [x] 1.3 Write failing table tests for `worktree.Base(configured, git)`: configured wins without git calls; `origin/HEAD` → `origin/develop` with local `develop` → `develop`, without → `origin/develop`; no `origin/HEAD` with `master` and `develop` → `master`; none of them → current branch; detached → `HEAD`
- [x] 1.4 Implement `internal/worktree` (stdlib only) until the tests pass

## 2. Config (TDD)

- [x] 2.1 Write failing `config.Parse` tests: defaults `Worktree: true`, `Base: ""`; `{"worktree": false}`; `{"base": "develop"}`; `{"worktree": "no"}`, `{"base": ""}`, `{"base": 1}` → defaults and an error naming the key
- [x] 2.2 Implement `worktree` and `base` until the tests pass

## 3. Start in the Change's checkout (TDD, fake herdr and git runners)

- [ ] 3.1 Write hand-written `herdr worktree create` and `worktree open` (`already_open` true and false) results in the 0.9.5 shape into `internal/start/testdata/herdr/`
- [ ] 3.2 Write failing `Prepare` tests for the call sequence: no checkout for the branch → `cat-file` check on the base, then `worktree create --cwd <project> --branch <change> --base <base> --label <change> --no-focus` returning its root pane and tab; existing local branch without checkout → check runs on `<change>`; Change missing on the ref → error naming Change and ref, no herdr call after `agent get`; closed linked checkout → `worktree open --path`; already open → `worktree open` then `tab create --workspace <id> --cwd <path>`; Project's own checkout on the branch, `worktree` off, or no checkouts → today's `tab create --cwd <project>`; no `--trust-repository` and no `--path` anywhere
- [ ] 3.3 Write failing tests for the agent checks with checkouts: agent `cwd` in a linked checkout is focused when running and reused when idle; `cwd` outside the Project and all checkouts is refused as today
- [ ] 3.4 Implement the new `Prepare` (and `Start` passing through) until the tests pass; keep `Launch` and `kanban launch` unchanged
- [ ] 3.5 Write a failing test, then change `start.Running` to take the Project and checkout directories: agent in a linked checkout counts, agent in another Project does not

## 4. Change source reads the Change's checkout (TDD, fake openspec runner per dir)

- [x] 4.1 Write failing `openspec.Load` tests with temp dirs: worktree on branch `a` holding Change `a` with more tasks done → `a` replaced with its counts, modified time and `Dir`; worktree-only Change `n` → appended; other Changes of the worktree ignored; branch without its Change directory → no list call there; Change archived in the Project → only archived; worktree list failing or lacking the Change → Project's entry, no error; planning `status` for a worktree Change runs in its `Dir`; cached planning not reused across different `Dir`
- [x] 4.2 Implement `Change.Dir` and the overlay in `Load` until the tests pass; keep `contract_test.go` passing
- [x] 4.3 Write a failing test, then make `docs.Dir` use `c.Dir` when set, so the detail view and `e` read and edit the worktree's files

## 5. Wiring

- [ ] 5.1 `cmd/kanban/main.go`: add a git runner like `openspecCLI`; the loader (Board and sync) passes `worktree.List` minus the Project's own checkout to `Load`; the Board's agent lookup passes the checkout directories to `Running`; Board Start and `kanban start` pass `Worktree{On, Base}` from the config and the git runner to `Prepare`/`Start`; adapt `main_test.go`
- [ ] 5.2 Update the usage text of `kanban start` (new tab or worktree of the Change)

## 6. Docs

- [ ] 6.1 Write `docs/adr/0007-start-in-a-git-worktree.md` (decision, herdr worktree commands, reuse, base detection, no removal; supersedes the tab location of ADR 0005) and mark ADR 0005's location sentence as superseded
- [ ] 6.2 CONTEXT.md: update **Start**, add **Change checkout**
- [ ] 6.3 README: Start in a worktree (location from herdr's `[worktrees] directory`, reuse, commit planning before `s`, nothing removed automatically), settings `worktree` and `base`

## 7. Verification

- [ ] 7.1 `go test -race ./...`, `go vet ./...` and `gofmt -l .` pass; `openspec validate m6-worktree-start --strict` passes
- [ ] 7.2 Smoke test in herdr with a throwaway committed test Change: `s` from the Board → worktree workspace labelled with the Change under herdr's worktree directory, agent tab focused (answers the `tab focus` open question; add the `workspace focus` fallback if needed), agent receives the prompt; ticked tasks in the worktree show on the Board after `r` and in the sidebar of the main workspace; badge shows; `s` again while running → focused; quit the agent, close the workspace, `s` → worktree reopened, no new branch; with the workspace open and the agent gone → new tab in it; uncommitted Change → error in the footer, nothing created; `{"worktree": false}` → tab in the Project; afterwards remove the test worktree and branch by hand
