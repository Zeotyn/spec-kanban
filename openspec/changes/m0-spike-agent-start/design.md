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
  ├─ herdr agent get <change>               exists → herdr agent focus <change>; done
  ├─ herdr tab create --cwd <project> --label <change> --focus   → pane id
  ├─ herdr agent start <change> --kind claude --pane <pane-id>   → waits for readiness
  └─ herdr agent prompt <change> "/opsx:apply <change>"
```

The openspec check calls the CLI directly; the proper adapter arrives with `m1-openspec-adapter` and replaces it.

### Fallback if the agent API does not fit
If `agent start` cannot be used (for example, it refuses the fresh tab's shell), fall back to `herdr pane run <pane> claude "/opsx:apply <change>"` and record in ADR 0005 why. The spec's observable behaviour stays the same.

### Tests
TDD against the recording fake herdr client: happy path call order, duplicate → focus only, unknown and archived Change → no herdr calls, each herdr step failing. One manual end-to-end run in herdr confirms the real JSON shapes; captured outputs go to `testdata/herdr/` as fixtures for the client's parsing tests.

## Risks / Trade-offs

- [`tab create` does not return a pane id] → look it up via `herdr pane list` filtered by the new tab id; note in ADR 0005.
- [Claude startup exceeds herdr's default 30 s readiness timeout] → pass `--timeout 60000`.
- [Agent names are global per session, so two Projects with equally named Changes collide] → accepted for now; revisit in M4 if it happens.
