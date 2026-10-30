## 1. Recorded session (TDD)

- [ ] 1.1 Write failing tests for `cache.WriteSession`/`cache.ReadSession` in a temp dir: round trip of `{agent, kind, value}`; missing or broken file → zero value; same Change name in two Projects kept apart; overwrite keeps the newest; empty dir → error on write, `""` on read
- [ ] 1.2 Implement them (atomic temp file + rename under `sessions/<project hash>/`) until the tests pass

## 2. Main agent from m5

- [ ] 2.1 Check that m5's `KANBAN_MAIN_PANE` reaches the spawned `kanban launch` (childEnv keeps it; add a `childEnv` test case); if m5 shipped without it, add it to `kanban open` with its table test

## 3. Watch and Handoff (TDD, fake `herdr.Runner` and openspec runner)

- [ ] 3.1 Capture `herdr agent get` output with `agent_session` into `internal/start/testdata/herdr/` (hand-written in the 0.9.5 shape, fake paths)
- [ ] 3.2 Write failing tests for `start.Watch`: clean settle → session recorded, `pane close <agent pane>`, `agent get <main>`, `agent prompt <main> <text>` in that order with the expected text; tasks left → notification with `3/7`, then `wait --until working`, then `wait`; blocked with all tasks done → "waits for input" notification, no close; `agent_not_found` from `wait` or `get` → returns nil, no notification; Change archived by the agent → Handoff with the "finished and archived" text, no archive request; Change neither open nor archived → not clean
- [ ] 3.3 Write failing tests for the Main agent checks: a Main agent of another kind (e.g. `codex`) is prompted; a watched `codex` agent hands off with `codex resume <id>` in its text; a watched `pi` agent hands off the same way and its text has no resume hint; no `KANBAN_MAIN_PANE`, main pane equals the agent pane, `agent get <main>` fails or finds no agent, status `blocked`, `agent prompt` rejected → each shows `openspec-kanban: <change> finished` with the Handoff text, and the pane is still closed
- [ ] 3.4 Implement `Watch` until the tests pass
- [ ] 3.5 `start.Launch` records the session after `agent start` and sends the prompt with `--wait` (stalled falls through); `kanban launch` calls `Watch` after a successful Launch, with the Project from `KANBAN_PROJECT`; adapt the Launch call-order tests

## 4. Resume (TDD)

- [ ] 4.1 Write failing table tests over the resume map (design.md) for `start.ResumeArgs` and `start.ResumeCommand`: one row per kind with `kind: id` → `claude --resume <id>`, `copilot --resume=<id>`, `codex resume <id>`, `opencode --session <id>`, `cursor-agent --resume <id>`, `gemini --resume <id>`; `claude`/`path`, `pi`, unknown kind, empty session → nil and `""`
- [ ] 4.2 Write failing tests for `start.Resume(change, session, pane, tab, run)`: `tab focus`, then `agent start <change> --kind <recorded agent> --pane <pane> --timeout 60000 -- <ResumeArgs>`, no prompt; not reopenable → error without herdr calls; `kanban resume` shows `openspec-kanban: resume failed` on error
- [ ] 4.3 Implement `ResumeArgs`, `Resume` and the internal `kanban resume <change> <pane> <tab>` (reads the record); update the usage text

## 5. Board: `s` on Review (TDD, fake funcs)

- [ ] 5.1 Write failing tests: Review Card with a reopenable record calls the resume func and quits on success; with a Running agent focuses it through the same func; without a record shows `all tasks of <change> are done`, with a `pi` record shows `resume not supported for pi`, a `codex` record calls the resume func, both calling nothing; resume error shows the notice and keeps the Board; Draft/Archived unchanged
- [ ] 5.2 Implement, wire `resumable` (`cache.ReadSession` + `ResumeArgs`) and the resume path (`Prepare`, then spawn `kanban resume`) in `showBoard`

## 6. Docs

- [ ] 6.1 README: Handoff (when the tab closes, what the Main agent receives, the notification fallback), `s` on Review, where sessions are recorded

## 7. Verification

- [ ] 7.1 `go test -race ./...`, `go vet ./...` and `gofmt -l .` pass
- [ ] 7.2 Smoke test in herdr with a tiny test Change: open the Board from a Claude pane, `s` → agent finishes → tab closes and that Claude receives the Handoff prompt; if `copilot` (installed here) or another mapped kind is installed, repeat with `agent` set to it → Handoff text names its resume command, `herdr agent get` shows its `agent_session` (fix the map entry if its shape differs), Review `s` resumes it; let an agent archive its Change → Handoff "finished and archived"; repeat from a shell pane → notification; stop an agent with tasks left → notification, tab open; close a watched tab by hand and check with `ps` that its `kanban launch` exits (answers the wait risk in design.md); prompt the Main agent while it is working and check the text is queued; `s` on the Review Card → resumed session in a new tab, not closed again
