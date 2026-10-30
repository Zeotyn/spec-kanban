## 1. Running agent lookup (TDD, fake `herdr.Runner`)

- [x] 1.1 Capture `herdr agent list` with one named agent started by `kanban start` into `testdata/herdr/agent-list.json` (keep the unnamed agents in it); hand-written as `internal/start/testdata/herdr/agent-list.json` in the herdr 0.9.5 shape with fake paths, so no tabs or agents were started
- [x] 1.2 Write failing table tests for `start.Running`: named agent in the Project counted; agent in a subdirectory counted; same name in another Project ignored; unnamed agents ignored; every `agent_status` counts; empty list; herdr error returned; unparsable output is an error
- [x] 1.3 Implement `Running` until the tests pass

## 2. Start changes (TDD)

- [x] 2.1 Write failing tests: `agent get` returns an agent whose `cwd` is another Project → no focus, no restart, no tab, error names the Change and the directory; new tab path calls `tab create … --no-focus` and ends with `agent focus <change>`
- [x] 2.2 Adjust the existing call-order tests to the new focus order, implement until all `internal/start` tests pass

## 3. Board: Running agent badge (TDD)

- [x] 3.1 Write failing tests: `New` with an agent lookup returns it from `Init` together with the load; the agents message marks matching open Cards with `●`; Archived Card with an agent name gets no badge; a lookup error shows no badge and no error; badges survive a Change reload and a cached-to-fresh swap; `r` looks them up again; the cached snapshot renders before the lookup answers
- [x] 3.2 Implement the badge (ANSI green, name fitted to the remaining width) until the tests pass

## 4. Board: `s` key (TDD, fake start func)

- [x] 4.1 Write failing tests: `s` on Ready and In Progress calls start with the Change name and shows `starting <change>…`; success quits; error shows `start failed: …` and keeps the Board; Draft, Review and Archived show their notices and do not call start; empty column does nothing; while starting, `q`, `j` and `s` are ignored and `ctrl+c` quits; `s` in the detail view does nothing
- [x] 4.2 Implement until the tests pass; add `s start` to both footer variants

## 5. Wiring

- [x] 5.1 Wire `start.Start` and `start.Running` into `board.New` in `cmd/kanban`; `kanban start <change>` keeps its usage and behaviour
- [x] 5.2 README: `s` key, Phase gating, the Running agent badge

## 6. Verification

- [ ] 6.1 `go test -race ./...` and `go vet ./...` pass
- [ ] 6.2 Smoke test in herdr: rebuild; open the Board, `s` on a Ready Card → popup closes, new tab labelled with the Change runs `/opsx:apply` in the invoking workspace; reopen → Card shows `●`; `s` again → existing tab focused, no second tab; `s` on Draft/Review/Archived → notices; quit Claude in that tab, reopen the Board and note whether the badge lingers (answers the Open Question); `kanban start <change>` from a shell still works
