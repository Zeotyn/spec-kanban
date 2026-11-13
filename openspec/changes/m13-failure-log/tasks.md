## 1. Log in the cache

- [ ] 1.1 Add the failure record (time, project, change, step, error) and `AppendFailure(dir, f)` to `internal/cache`: one `O_APPEND` line of JSON under `dir/failures.log` with mode `0600`, then a rewrite of the newest 100 valid lines through `writeAtomic` when the file holds more (design D1, D2); tests for append, the limit, a broken line and an unwritable directory
- [ ] 1.2 Add `ReadFailures(dir)`, which returns the valid records oldest first and skips invalid lines; tests
- [ ] 1.3 Add `ReadOpened(dir, project)` / `WriteOpened(dir, project, t)` under `dir/opened/<project hash>`, written atomically (D5); tests including a missing and an unreadable record

## 2. Recording failures

- [ ] 2.1 Add `fail(run, step, proj, change, err)` in `cmd/kanban`, which records the failure in `cache.UserDir()`, ignoring a write error, and shows `spec-kanban: <step> failed` with body `<error>\n\nDetails: kanban failures` (D3); tests with a fake runner, including an unwritable cache directory
- [ ] 2.2 Use `fail` in `launch`, `runCommand`, `resumeChange`, `waveMember` (step `start`) and `waveLaunch` (step `wave`, Change = the wave's Changes joined by spaces); adjust the existing notification tests to the new body and check one record per failure

## 3. `kanban failures`

- [ ] 3.1 Add the `failures [--all]` subcommand: find the Project as in D4, print the matching records oldest first as header line plus error indented by two spaces plus a blank line, `no failures` when none match, an error outside any Project without `--all`, and the usage with an error for any other argument; tests for the spec scenarios (current Project, worktree, `--all`, nothing recorded, broken line)
- [ ] 3.2 Add `failures [--all]` to the usage text

## 4. Board notice

- [ ] 4.1 In `showBoard`, without a config warning: read the Project's last opening (24 hours ago when missing), take the newest failure of the Project after it, set the notice `<step> failed <change> <age> ago: <first line>` with the age rounded down to `s`, `m`, `h` or `d`, then write the opening time; with a config warning, read and write neither (D5). Keep the formatting in a function with tests for the spec scenarios (shown once, config warning first, another Project's failure, first opening, unreadable log)

## 5. Docs

- [ ] 5.1 README: troubleshooting section on `kanban failures` and `--all`, the `Details: kanban failures` line and the Board notice
- [ ] 5.2 CONTEXT.md: add **Failure log**
- [ ] 5.3 CHANGELOG `## Unreleased`: failure log, `kanban failures`, the notification hint and the Board notice

## 6. Verification

- [ ] 6.1 `go test -race ./...`, `go vet ./...`, `gofmt -l .` pass
- [ ] 6.2 `openspec validate m13-failure-log --strict` passes

## Smoke test (user, in herdr)

These have no checkboxes: they do not count as tasks, and only the user runs them.

- Make a Start fail after the Board closed (for example an agent kind that does not start): the notification ends with `Details: kanban failures`
- Run `kanban failures` in the Project: it prints that failure with the full error and the pane's last lines
- Open the Board: the footer shows `start failed <change> …` once; close and reopen it: the key footer is back
- Run `kanban failures --all` in a directory outside any Project: it prints the failures of every Project
