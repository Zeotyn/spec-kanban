## 1. Input box

- [ ] 1.1 Add `inputBox(screen) (text, found)` in `internal/start` (design D1); capture real Claude Code screens with `herdr pane read <pane> --source visible` into `testdata` (empty box, one line, several lines, no box) and table-test them, including non-breaking spaces after `❯`

## 2. Waiting

- [ ] 2.1 Add `clear(run, pane, kind)`: for `claude` with a found box the box is empty, otherwise the pane is not focused; a read error counts as "not found" (D2); tests with a fake runner
- [ ] 2.2 Make `Tell` poll `clear` every 2 seconds for at most 10 minutes before `herdr agent prompt`, returning false after the limit, with sleep and clock as package variables; add `TellNow` (one check: taken, waiting); tests for at once, after a few polls, the limit, and the existing agent checks
- [ ] 2.3 Check that the watcher's Handoff and Stall use the waiting `Tell` (D4); adjust the watch tests for a box with text that empties after two polls

## 3. Subcommands

- [ ] 3.1 Add the detached `kanban tell <pane> <title> <text>` entry: run `Tell`, and on false show `title` as a notification unless it is empty (D3); tests
- [ ] 3.2 `kanban escalate`: set the marker, try `TellNow`, on `waiting` spawn `tell` with an empty title and print `Queued for the Main agent. End your turn.`; tests for sent, queued and no Main agent
- [ ] 3.3 `kanban wave escalate` and `kanban wave finish`: same with their notification titles; `wave finish` records the wave finished before it spawns; tests for sent, queued and no Main agent

## 4. Docs

- [ ] 4.1 README: prompts to the Main agent wait while you type, the 10-minute limit, and `Queued for the Main agent`
- [ ] 4.2 CHANGELOG `## Unreleased`: no prompt into typed input

## 5. Verification

- [ ] 5.1 `go test -race ./...`, `go vet ./...`, `gofmt -l .` pass
- [ ] 5.2 `openspec validate f4-no-prompt-into-input --strict` passes

## Smoke test (user, in herdr)

These have no checkboxes: they do not count as tasks, and only the user runs them.

- Type half a line in the Main pane without Enter and let a started Change finish: nothing is typed into your line; clear it and the Handoff arrives within a few seconds
- Sit in the Main pane without typing while a Change finishes: the Handoff arrives at once
