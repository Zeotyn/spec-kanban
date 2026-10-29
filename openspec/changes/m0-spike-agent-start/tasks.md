## 1. Explore herdr for real

- [ ] 1.1 Run `tab create`, `agent start --kind claude`, `agent prompt`, `agent get`, `agent focus` by hand in a scratch workspace and save their JSON outputs to `testdata/herdr/`
- [ ] 1.2 Decide name scheme and fallback based on the outputs; note findings in `design.md`

## 2. herdr client extensions (TDD)

- [ ] 2.1 Write failing parsing tests against the saved outputs for tab create, agent get, agent start, agent prompt
- [ ] 2.2 Extend `herdr.Client` and the fake until the tests pass

## 3. Start command (TDD)

- [ ] 3.1 Write failing tests: happy path call order, running agent → focus only, unknown Change, archived Change, each herdr step failing
- [ ] 3.2 Implement `internal/start` and the `kanban start <change>` subcommand until the tests pass

## 4. Measure and record

- [ ] 4.1 Time `openspec list --json`, `status --change X --json` (sequential and 8 in parallel), and `instructions apply --json` in magentic; add the numbers to ADR 0002 Consequences
- [ ] 4.2 Set ADR 0005 to `accepted`, or revise it with what the spike found

## 5. Verification

- [ ] 5.1 `go test ./...` and `go vet ./...` pass
- [ ] 5.2 Manual: `kanban start <change>` in this repo opens a tab with Claude running `/opsx:apply <change>`; running it again focuses the same tab
