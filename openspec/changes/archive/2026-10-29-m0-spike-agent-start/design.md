## Context

ADR 0005 proposes Start via `herdr tab create` → `herdr agent start --kind claude` → `herdr agent prompt`. Only the CLI help text has been read so far; the JSON shapes of the results (for example, where `tab create` returns the new pane id) are unverified. This change builds on the module, herdr client, and Project lookup from `m0-plugin-skeleton`.

## Goals / Non-Goals

**Goals:**
- Prove or refute ADR 0005 with working, tested code that `m4-start-action` can call unchanged.
- Measure `openspec list`, `status`, and `instructions apply` latency on a real Project (magentic, 9 Changes).

**Non-Goals:**
- Board keybinding, Running agent badge (both M4), agents other than Claude.

## Decisions

### Agent name = Change name
`herdr agent start <name>` takes a name; using the Change name lets Start find an existing session with `herdr agent get <change>` and lets M4 match Running agents to Cards without extra state. Alternative (prefixing, e.g. `kanban:<change>`) is kept in reserve if herdr rejects names or collides with user-named agents.

### Sequence

```
kanban start <change>
  ├─ project.Find(cwd)                      no Project → error
  ├─ openspec list --json                   Change not open → error
  ├─ herdr agent get <change>
  │    ├─ found, Claude in the foreground   → herdr agent focus <change>; done
  │    ├─ found, other program in front     → error naming it
  │    └─ found, pane back at the shell     → reuse that pane (skip tab create)
  ├─ herdr tab create --cwd <project> --label <change> --focus   → result.root_pane.pane_id
  ├─ herdr agent start <change> --kind claude --pane <pane-id> --timeout 60000
  ├─ herdr agent prompt <change> "/opsx:apply <change>"
  └─ herdr agent focus <change>             only when a pane was reused
```

The openspec check calls the CLI directly; the proper adapter arrives with `m1-openspec-adapter` and replaces it.

### Spike findings (herdr 0.9.5, 2026-10-29)

Captured outputs live in `testdata/herdr/`.

- `tab create` returns `result.root_pane.pane_id` and `result.tab.tab_id`; no extra lookup needed.
- `agent start <name>` registers the name (`result.agent.name`); `agent get/prompt/focus <name>` accept it. Start to ready took 3.8 s.
- Errors are JSON on stderr with exit 1, e.g. `{"error":{"code":"agent_not_found",…}}`.
- After Claude exits, herdr keeps the agent registered (`agent_status: "done"`) until the tab closes, so "agent exists" does not mean "running". `pane process-info --pane <id>` tells them apart: Claude runs when a foreground process has `argv0: "claude"`; it exited when the foreground group is the shell (`foreground_process_group_id == shell_pid`). Any other foreground program is reported as an error rather than guessed at.
- `agent start` with the same name in the same pane works after an exit, so a stale tab is reused instead of opening a second one.
- The fallback below was not needed.

### Fallback if the agent API does not fit
If `agent start` cannot be used (for example, it refuses the fresh tab's shell), fall back to `herdr pane run <pane> claude "/opsx:apply <change>"` and record in ADR 0005 why. The spec's observable behaviour stays the same.

### Tests
TDD against the recording fake herdr client: happy path call order, duplicate → focus only, unknown and archived Change → no herdr calls, each herdr step failing. One manual end-to-end run in herdr confirms the real JSON shapes; captured outputs go to `testdata/herdr/` as fixtures for the client's parsing tests.

## Risks / Trade-offs

- [Claude startup exceeds herdr's default 30 s readiness timeout] → pass `--timeout 60000`.
- [Agent names are global per session, so two Projects with equally named Changes collide: Start focuses the other Project's agent] → accepted for now; revisit in M4 by comparing the agent's `cwd` with the Project.
- [Closing only the agent's pane (not its tab) was not tested; if herdr keeps the agent, Start reports `pane_not_found`] → recover by treating it like `agent_not_found` if it shows up.
