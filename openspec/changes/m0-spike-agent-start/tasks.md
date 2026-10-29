## 1. Explore herdr for real

- [x] 1.1 Run `tab create`, `agent start --kind claude`, `agent prompt`, `agent get`, `agent focus` by hand in a scratch workspace and save their JSON outputs to `testdata/herdr/`
- [x] 1.2 Decide name scheme and fallback based on the outputs; note findings in `design.md`

## 2. herdr client extensions (TDD)

- [x] 2.1 Write failing tests that feed the saved outputs (tab create, agent get, process-info) through a recording fake runner
- [x] 2.2 Parse only the needed fields inside `internal/start` (no herdr client methods) and share stderr handling via `herdr.Run`

## 3. Start command (TDD)

- [x] 3.1 Write failing tests: happy path call order, running agent → focus only, exited agent → restart in its tab, Change not open, each herdr step failing
- [x] 3.2 Implement `internal/start` and the `kanban start <change>` subcommand until the tests pass

## 4. Measure and record

- [x] 4.1 Time `openspec list --json`, `status --change X --json` (sequential and 8 in parallel), and `instructions apply --json` in magentic; add the numbers to ADR 0002 Consequences
- [x] 4.2 Set ADR 0005 to `accepted`, or revise it with what the spike found

## 5. Verification

- [x] 5.1 `go test ./...` and `go vet ./...` pass
- [ ] 5.2 Manual: `kanban start <change>` in this repo opens a tab with Claude running `/opsx:apply <change>`; running it again focuses the same tab
