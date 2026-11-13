## Why

m12-escalation lets a worker ask the Main agent for help and catches a silent stop. But the user still has no glance-able sign that a Change waits for them. When no Main agent takes a prompt, m12 falls back to a herdr notification, which is gone after a few seconds. The user wants notifications for errors only, and to see from the Board and the sidebar which Changes wait for a human.

## What Changes

- **Attention marker `⚑`.** A Change waits for the user after an Escalation, after a Stall prompt, and after a Handoff that no Main agent took. The plugin keeps this as a marker file per Project and Change in its state directory.
  - Escalation and Stall set it; the watcher clears it when the agent works again. A Handoff to the Main agent clears it; a Handoff without one sets it. A new Start or resume clears it.
  - Self-healing: the Board and the sidebar ignore an Escalation or Stall marker while the Change's agent works or no longer exists, and a Handoff marker once the Change is no longer open. A dead watcher leaves no permanent `⚑`.
- **Card.** A Card whose Change waits shows a red `⚑` before its name, in the place of the green `●`. The Card's width and layout stay.
- **Sidebar.** The `kanban` token starts with `⚑<n> ` when `n` Changes of the Project wait, for example `⚑1 1 ready · 1 review`. The prefix is never dropped; Phases are dropped from the left as before so the token fits 28 characters. The plugin refreshes the sidebar when it sets a marker.
- **BREAKING (behaviour): no fallback notifications.** The notifications `spec-kanban: <change> finished`, `… needs you` and `… stalled` are removed; the marker replaces them. Only failure notifications stay (start, run and resume failed). The blocked notification already went in m12.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `agent-handoff`: new requirement for the attention marker (set, clear, self-healing); Escalation and Stall set it; the fallback becomes the marker instead of a notification.
- `board`: the Card shows `⚑` instead of `●` while its Change waits.
- `sidebar-summary`: the `kanban` token carries the `⚑<n>` prefix; a refresh also follows a new marker.

## Impact

- **Code:**
  - `internal/cache`: the m12 escalation record becomes the marker with a reason (`escalation`, `stall`, `handoff`), and all markers of a Project can be read at once.
  - `internal/start`: `Tell` writes no notification; Stall and the Handoff fallback set the marker; `Running` also returns each agent's status; a pure function decides which markers show.
  - `internal/board`: the `⚑` badge; the agents callback also returns the waiting Changes.
  - `internal/sidebar`: the prefix in `Summary`; `Sync` reads agent statuses and markers.
  - `internal/theme`: an `Attention` colour (ANSI red, `1`).
  - `cmd/kanban`: wiring, and a sidebar refresh after setting a marker.
- **Docs:** README, CONTEXT.md (attention marker), CHANGELOG, `herdr-config.toml` unchanged.
- Builds on m12-escalation; archive m13 after m12.
- No new dependencies.
