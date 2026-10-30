## 1. Recorded session (TDD)

- [x] 1.1 Write failing tests for `cache.WriteSession`/`cache.ReadSession` with `herdr.Session{Agent, Kind, Value}` in a temp dir: round trip; missing or broken file → zero value; same Change name in two Projects kept apart; overwrite keeps the newest; empty dir → error on write, zero value on read
- [x] 1.2 Add `herdr.Session` and implement them (atomic temp file + rename under `sessions/<project hash>/`) until the tests pass

## 2. Archive check (TDD)

- [x] 2.1 Write failing tests for `openspec.ArchivedIn(dir, change)`: `archive/<change>` and `archive/<date>-<change>` → true; other names, no archive dir → false
- [x] 2.2 Implement it by factoring out `Load`'s archive scan and reuse it there; existing `Load` tests stay green

## 3. Watch and Handoff (TDD, fake `herdr.Runner`, git runner and loader)

- [x] 3.1 Hand-write `herdr agent get` output with `agent_session` into `internal/start/testdata/herdr/` (0.9.5 shape, fake paths)
- [x] 3.2 Write failing tests for `start.Watch` without a linked checkout: `agent get` + record, `wait --until working --timeout 60000` (an error falls through), `wait`; clean settle → session recorded, `pane close <agent pane>`, `agent get <main>`, `agent prompt <main> <text>` in that order with the no-checkout text; tasks left → notification with `3/7`, then `wait --until working`, then `wait`; blocked with all tasks done → "waits for input", no close; load error → "stopped" notification; `agent_not_found` from `wait` or `get` → returns nil, no notification; Change archived in the Project → "finished and archived. Please review it."; Change neither open nor archived → not clean
- [x] 3.3 Write failing tests with a linked checkout on branch `<change>` (temp dir, fake `git worktree list --porcelain`): loader Progress from the checkout → Handoff text with branch, path, `merge branch <change> into <base>` (configured base, and detected via `worktree.Base`), `/opsx:archive <change>`, `git worktree remove <path>`, `--force` forbidden; Change archived inside the checkout while the loader still returns it open → "finished and archived" text with merge and removal, no archive request; no `worktree remove`, `workspace close`, `git merge` or `git branch -d` call is ever made
- [x] 3.4 Write failing tests for the Main agent checks: a Main agent of another kind (e.g. `codex`) is prompted; a watched `codex` agent hands off with `codex resume <id>` and the checkout as directory in its text; a watched `pi` agent's text has no resume hint; no Main pane, Main pane equals the agent pane, `agent get <main>` fails or finds no agent, status `blocked`, `agent prompt` rejected → each shows `openspec-kanban: <change> finished` with the Handoff text, and the pane is still closed
- [x] 3.5 Implement `Watched` and `Watch` until the tests pass
- [x] 3.6 `kanban launch` calls `Watch` after a successful `start.Launch` when `KANBAN_PROJECT` is set, with `KANBAN_MAIN_PANE`, the config's `base`, `loader(cache.Dir(), proj, openspecCLI, gitCLI(proj))` and `cache.WriteSession`; add a `childEnv` test case that `KANBAN_PROJECT` and `KANBAN_MAIN_PANE` survive

## 4. Resume (TDD)

- [x] 4.1 Write failing table tests over the resume map (design.md) for `start.ResumeArgs` and `start.ResumeCommand`: one row per kind with `kind: id` → `claude --resume <id>`, `copilot --resume=<id>`, `codex resume <id>`, `opencode --session <id>`, `cursor-agent --resume <id>`, `gemini --resume <id>`; `claude`/`path`, `pi`, unknown kind, zero session → nil and `""`
- [x] 4.2 Write failing tests for `start.Resume(change, s, pane, tab, run)`: `tab focus`, then `agent start <change> --kind <s.Agent> --pane <pane> --timeout 60000 -- <ResumeArgs>`, no prompt; not reopenable → error without herdr calls; `kanban resume` shows `openspec-kanban: resume failed` on error
- [x] 4.3 Implement `ResumeArgs`, `ResumeCommand`, `Resume` and the internal `kanban resume <change> <pane> <tab>` (reads the record); generalise `spawnLaunch` to a subcommand with args; update the usage text

## 5. Board: `s` on Review (TDD, fake funcs)

- [x] 5.1 Write failing tests: Review Card with a reopenable record calls the resume func, shows the starting footer and quits on success; with a Running badge calls the resume func; without a record shows `all tasks of <change> are done`, with a `pi` record shows `resume not supported for pi`, both calling nothing; a `codex` record calls the resume func; resume error shows the notice and keeps the Board; Draft/Archived unchanged
- [x] 5.2 Implement `resumable` and `resume` in `board.New`; wire them in `showBoard`: `resumable` = `cache.ReadSession` + `ResumeArgs`, `resume` = `start.Prepare(proj, change, wt, herdr.Exec, git)` (pane "" → done), re-read the record (none → `no reopenable session for <change>`), spawn `kanban resume`

## 6. Docs

- [ ] 6.1 README: Handoff (when the pane closes, what the Main agent receives with and without a worktree, that the Main agent merges and runs `git worktree remove` and the plugin removes nothing, the notification fallback), `s` on Review, where sessions are recorded; replace "Nothing is removed automatically yet"

## 7. Verification

- [ ] 7.1 `go test -race ./...`, `go vet ./...` and `gofmt -l .` pass
- [ ] 7.2 Smoke test in herdr with a tiny test Change committed on the base: open the Board from a Claude pane, `s` → worktree workspace, agent finishes → pane closes (check whether the worktree workspace closes with it), that Claude receives the worktree Handoff prompt, merges, archives and `git worktree remove` succeeds, branch kept; leave an uncommitted file in a worktree → removal refused, nothing lost; let an agent archive its Change in the worktree → "finished and archived" text with merge; `worktree: false` → old text; repeat from a shell pane → notification; stop an agent with tasks left → notification, tab open; close a watched tab by hand and check with `ps` that its `kanban launch` exits; prompt the Main agent while it is working and check the text is queued; `s` on the Review Card → worktree reopened (or recreated after removal), session resumed, not closed again; if `copilot` or another mapped kind is installed, repeat with `agent` set to it and check its `agent_session` shape and resume
