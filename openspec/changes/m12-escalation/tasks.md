## 1. Records in `internal/cache`

- [x] 1.1 Tests first: `WriteMain`/`ReadMain` round trip, an empty pane removes the record, a missing or unreadable record reads as "", equally named Changes in two Projects stay apart ("Main pane recorded per Change")
- [x] 1.2 Implement `WriteMain`/`ReadMain` under `mains/<project hash>/<change>` with `writeAtomic`
- [x] 1.3 Tests first, then implement `WriteEscalation`, `Escalated` and `ClearEscalation` under `escalations/<project hash>/<change>`

## 2. Delivery and sentence in `internal/start`

- [x] 2.1 Tests first for `Tell`: Main agent idle takes the prompt; no Main pane, Main pane equals the Change's own pane, no agent in it, Main agent blocked, and a rejected `agent prompt` each show `notification show <title> --body <text>` ("Main agent falls back to a notification")
- [x] 2.2 Implement `Tell` and make `handoff` use it; the existing Handoff tests stay green
- [x] 2.3 Tests first for `Sentence`: a plain path ("Escalation sentence from the Board"), a path with a space ("Binary path with a space"), a path with a single quote
- [x] 2.4 Implement `Sentence` (design.md decision 5)

## 3. Watch: no notifications, Stall, Resume

- [x] 3.1 Tests first with the fake herdr runner of `watch_test.go`: an unfinished idle settle and a blocked settle show no notification and prompt nobody ("Agent stops with tasks left", "Agent asks for approval"); "Agent works through several turns"
- [x] 3.2 Remove both notifications; `check` returns the Progress text instead of a notification body
- [x] 3.3 Tests first, one per scenario of "Stall": "Worker stops silently" (timed wait fails, agent still idle → one Stall prompt with the exact text), "Worker goes on by itself" and "User answers in time" (timed wait succeeds → nothing sent), "Second silent stop"; plus: the agent gone after the timed wait ends watching silently; a blocked settle runs no timed wait; Progress unknown leaves out `at <done>/<total> tasks`; "Stall without a Main agent" falls back to `spec-kanban: <change> stalled`
- [x] 3.4 Tests first: an escalation record at the settle skips the timed wait, and the record is cleared once the agent works again ("Escalation is followed by no Stall")
- [x] 3.5 Implement Stall in `Watch` with the constant `stallAfter` and `Watched.Escalated`/`Watched.Clear` callbacks (design.md decisions 2 and 6)
- [x] 3.6 Tests first: with `Resumed` the watch waits `--until working` without timeout before the first evaluation, and a resumed agent that never works gets no Handoff and no Stall ("Resumed session of a finished Change", "Resumed session finishes the Change")
- [x] 3.7 Implement `Watched.Resumed`

## 4. `cmd/kanban`

- [x] 4.1 Tests first: `startOne`'s prompt is `cfg.PromptFor(...)` plus one space and `start.Sentence(self, change)`, also with a configured `prompt`; `kanban start` sends no sentence ("Start with a configured prompt", "Start with configured instructions", "Start a Spec Kit Feature", "Start a Ready Change")
- [x] 4.2 Implement the sentence in `startOne` with `os.Executable`
- [x] 4.3 One helper builds the `Watched` for `launch` and `resumeChange`; both write the Main pane record from `$KANBAN_MAIN_PANE` and clear the escalation record before watching; `resumeChange` watches after a successful `Resume` with `Resumed: true`, only when `$KANBAN_PROJECT` is set (tests for the records written and for no watch without `$KANBAN_PROJECT`)
- [x] 4.4 Tests first, one per scenario of "Escalation": "Worker cannot publish" (exact prompt start, tasks path in the worktree, prints that the Main agent received it), "Without a worktree", "No Main agent" (`spec-kanban: <change> needs you`, prints that a notification received it), "Unknown Change" (exit non-zero, nothing sent), "No reason" (usage, exit 2); plus: the reason joins all remaining arguments; the Project is found from a linked worktree through its Main record; the escalation record is written in every delivered case
- [x] 4.5 Implement `kanban escalate` (design.md decision 3) and add it to the usage text: `kanban escalate <change> <reason>  Ask the Main agent for a human's help with <change> (run by the Change's agent)`

## 5. Docs

- [x] 5.1 CONTEXT.md: add **Escalation** (a worker's request for a human, delivered to the Main agent at once) and **Stall** (a watched agent that stopped unfinished for 10 minutes without escalating) under Board, next to Running agent
- [x] 5.2 README Handoff section: Escalation, the sentence in the Start prompt, Stall, a watched Resume, and that the "stopped" and "waits for input" notifications are gone
- [x] 5.3 CHANGELOG `## Unreleased`: Escalation (`kanban escalate`, sentence in the Board's Start prompt), Stall after 10 minutes, Resume from the Board is watched, and the removed "stopped" and "waits for input" notifications
- [ ] 5.4 At archive, update the Purpose of `openspec/specs/agent-handoff/spec.md`: it names Escalation and Stall instead of "a notification when it is not finished"

## 6. Verification

- [ ] 6.1 `go test -race ./...`, `go vet ./...`, `gofmt -l .` pass
- [ ] 6.2 `openspec validate m12-escalation --strict` passes

## Smoke test (user, in herdr)

These have no checkboxes: they do not count as tasks, and only the user runs them.

- Open the Board from a Claude pane and press `s` on a Ready Card: the worker's first prompt ends with the escalation sentence naming the plugin binary's absolute path
- In that worker's tab, ask it to run the escalate command with a reason: the Main agent gets the Escalation prompt at once, with Progress, worktree and tasks file; the worker's tab stays open
- Tell the worker to stop and wait without escalating, and leave it idle for 10 minutes: the Main agent gets one Stall prompt; no "stopped" notification appears at the settle
- Answer a stopped worker within 10 minutes: no Stall prompt
- Let a worker hit a permission dialog: no notification from spec-kanban, herdr shows it as blocked
- Open the Board from a plain shell pane, start a Change, and escalate from its worker: herdr shows `spec-kanban: <change> needs you`
- Resume a Review Card with `s` and leave it idle: the tab stays open; ask it something and let it settle: the Handoff follows
