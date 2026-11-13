## 1. Marker in `internal/cache`

- [x] 1.1 Tests first: `WriteMark`/`ReadMarks` round trip for the three reasons, `ClearMark`, a broken file is skipped, an unknown reason is skipped, equally named Changes in two Projects stay apart
- [x] 1.2 Replace m12's escalation record with `WriteMark`, `ClearMark` and `ReadMarks` under `attention/<project hash>/<change>` (design.md decision 1); update its callers

## 2. Setting and clearing in `internal/start` and `cmd/kanban`

- [x] 2.1 Tests first for `Tell`: no notification in any fallback case; it returns false there and true when the Main agent took the prompt
- [x] 2.2 Remove the notification branch from `Tell`
- [x] 2.3 Tests first with the fake herdr runner: a Stall sets a `stall` marker with and without a Main agent ("Worker stops silently", "Stall without a Main agent"); working again clears it ("Worker continues"); a Handoff without a Main agent sets `handoff` ("Board opened from a shell pane", "Main agent blocked", "Handoff without a Main agent"); a Handoff the Main agent took clears an older marker ("Handoff clears an older marker")
- [x] 2.4 Implement them in `Watch` and `handoff`
- [x] 2.5 Tests first for `kanban escalate`: the `escalation` marker is set with and without a Main agent, no notification is shown, and the printed line says whether the Main agent received it or only the marker shows it ("Worker cannot publish", "No Main agent", "Worker escalates")
- [x] 2.6 Implement them in `escalate`; `launch` and `resumeChange` clear the marker before watching ("New Start")
- [x] 2.7 Tests first, then implement: after setting a marker, `escalate` and the watcher run one sidebar pass through `sidebar.Coalesce` ("Stall while nothing else happens")

## 3. Reader rule

- [ ] 3.1 Tests first for `start.Waiting`, one per scenario of "Attention marker": escalated and idle agent waits; working agent does not ("Watcher gone"); no agent does not ("Agent tab closed"); `handoff` waits while the Change is open and not after it is archived or gone ("Handoff without a Main agent"); nil agents hide `escalation` and `stall` but keep `handoff`
- [ ] 3.2 Implement `Waiting`; split `Running` into fetching `agent list` (with each agent's status) and filtering it for a Project (design.md decisions 3, 4 and 6)

## 4. Board

- [ ] 4.1 `internal/theme`: `Attention` = ANSI `1`; extend `TestPalette`
- [ ] 4.2 Tests first, one per scenario of "Attention marker on the Card": "Escalated Change" (red `⚑`, no `●`, same column), "Agent works again" (`r` → `●`), "Handoff without a Main agent", "Long name" (name width as with `●`), "herdr unavailable" (no `⚑`, no error); plus no `⚑` on an Archived Card and a golden for a Card with `⚑`
- [ ] 4.3 The agents callback returns running and waiting together; `card` draws `⚑ ` in the badge's place; `cmd/kanban` builds waiting from `Running`, `ReadMarks`, `Waiting` and the loaded Items

## 5. Sidebar

- [ ] 5.1 Tests first: "One Change waits" (`⚑1 1 ready · 1 review`), "Prefix keeps its place" (`⚑1 1 wip · 1 review`), no prefix without waiting Changes; the existing Phase overview tests stay green
- [ ] 5.2 `Summary` takes the waiting count and prepends `⚑<n> ` (design.md decision 6)
- [ ] 5.3 Tests first: `Sync` runs `agent list` once per pass for all Projects; a failing `agent list` still reports the tokens, counting only Handoff markers
- [ ] 5.4 Implement it in `Sync` and `syncLoad`'s caller in `cmd/kanban`

## 6. Docs

- [ ] 6.1 CONTEXT.md: add **Attention marker** (`⚑` on a Card and in the sidebar while a Change waits for the user after an Escalation, a Stall or a Handoff no Main agent took) under Board, next to Running agent
- [ ] 6.2 README: the `⚑` on Cards and in the `kanban` token, when it appears and goes away, and that notifications now only report failures
- [ ] 6.3 CHANGELOG `## Unreleased`: the attention marker on Cards and in the sidebar; the `finished`, `needs you` and `stalled` notifications are replaced by it

## 7. Verification

- [ ] 7.1 `go test -race ./...`, `go vet ./...`, `gofmt -l .` pass
- [ ] 7.2 `openspec validate m13-attention --strict` passes

## Smoke test (user, in herdr)

These have no checkboxes: they do not count as tasks, and only the user runs them.

- Start a Change from the Board and let its worker escalate: its Card shows a red `⚑` instead of `●`, and the sidebar's `kanban` line starts with `⚑1`
- Tell the worker to continue: the `⚑` goes away on the next Board lookup (`r`) and in the sidebar
- Leave a worker idle for 10 minutes without escalating: `⚑` appears in the sidebar without touching anything
- Close an escalated worker's tab: the `⚑` goes away
- Open the Board from a plain shell pane and let a started Change finish: no notification, its Review Card shows `⚑` until it is archived
- Make a Start fail (for example an agent kind that does not start): herdr still shows `spec-kanban: start failed`
