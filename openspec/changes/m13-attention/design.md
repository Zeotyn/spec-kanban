## Context

This change builds on m12-escalation:
- `escalate` writes an escalation record under `escalations/<project hash>/<change>`; the watcher skips Stall while it exists and removes it when the agent works again; `launch` and `resumeChange` remove it.
- `start.Tell` prompts the Main agent or shows a notification (`finished`, `needs you`, `stalled`).

Today:
- The Board gets `running map[string]bool` (agent names) from `start.Running`, asynchronously, and draws a green `●` in a 2-cell prefix before the name (`card` in `internal/board/board.go`).
- The sidebar's `Summary` builds the `kanban` token and drops Phases from the left until it fits 28 runes. `Sync` lists workspaces and panes, but no agents.
- `theme` holds the ANSI-16 colours.

## Goals / Non-Goals

**Goals:**
- One marker file per Change, one reader rule, used by the Board and the sidebar.
- A marker never outlives its reason by more than one Board lookup or sidebar pass.
- No new herdr calls on the Board; one `agent list` per sidebar pass.

**Non-Goals:**
- Showing the escalation reason or the Handoff text on the Board.
- A marker for a blocked agent. herdr's own sidebar shows it.
- A key or action to dismiss a marker.

## Decisions

1. **Marker file.** m12's escalation record becomes `attention/<project hash>/<change>` with JSON `{"why":"escalation"|"stall"|"handoff"}`.
   - `cache.WriteMark(dir, project, change, why)`, `cache.ClearMark(...)`, and `cache.ReadMarks(dir, project) map[string]string` (Change → why) for readers.
   - The watcher's "escalated" check becomes `why == "escalation"`. A Stall is only written when no Escalation marker exists, since m12 skips the Stall then. An Escalation overwrites a Stall marker: it is the newer reason.
   - Rejected: a field in the session record. Sessions are written at every settle by the watcher; two writers on one file would race with `escalate`.

2. **Who writes and clears.**
   - Set: `escalate` (always, before or after delivery), the watcher at a Stall (always), the watcher at a Handoff when `Tell` returns false.
   - Clear: the watcher when the agent works again after a settle; the watcher after a Handoff `Tell` returned true; `launch` and `resumeChange` before watching.
   - `Tell` drops its notification branch and only returns whether the Main agent took the prompt.

3. **Self-healing reader.** A pure function `start.Waiting(marks map[string]string, agents map[string]string, items []work.Item) map[string]bool`, keyed by Change name:
   - `escalation`/`stall`: shown when `agents[herdr.AgentName(change)]` is present and not `working`.
   - `handoff`: shown when the Change is in `items` and not Archived.
   - `agents` nil (herdr failed) hides `escalation`/`stall`.
   - Readers never delete files. Writers own them, and stale files are harmless because they are not shown.
   - Rejected: showing every marker file. A crashed or killed watcher would leave a permanent `⚑`.

4. **`start.Running` returns statuses.** It returns `map[string]string` (agent name → `agent_status`) from the same `agent list` call, instead of `map[string]bool`. The Board's agents callback becomes `func() (running map[string]bool, waiting map[string]bool, err error)`, built in `cmd/kanban` from `Running`, `ReadMarks`, `Waiting` and the cached Items. One message carries both, so `⚑` and `●` change together.

5. **Card: `⚑` replaces `●`.** One 2-cell prefix: `⚑ ` in `theme.Attention` (ANSI red, `1`) when waiting, else `● ` when running, else none.
   - The name width and the milestone badge do not change, so no golden moves except where a marker is shown.
   - `⚑` (U+2691) has no emoji presentation by default and is one cell wide in common terminals, like `●`.
   - Rejected: both glyphs (`●⚑ `). One more cell from every waiting Card's name, and an escalated agent always also has `●`, which adds nothing.
   - Rejected: red border or name. The border already means selection; the name colour is bold default.

6. **Sidebar prefix.** `Summary(changes, waiting int)` prepends `⚑<n> ` to the `kanban` token when `waiting > 0`, and drops Phases from the left until prefix plus Phases fit 28 runes. It keeps at least one Phase, as today. `kanban_change` stays. A count, not names, because 28 characters leave no room for names.
   - `Sync` runs `agent list` once per pass. A failed `agent list` does not fail the pass: it gives nil statuses, and only Handoff markers count.
   - Agents are matched to a Project by working directory, as `start.Running` does. `Running` splits into fetching the list and filtering it for a Project, so `Sync` fetches once per pass and filters per Project.

7. **Refresh after a marker.** `escalate` and the watcher run the same code as `kanban sync` (through `sidebar.Coalesce`) after setting a marker. A Stall happens with no herdr event, so without this the sidebar would show it only at the next event. Clearing needs no refresh: the agent's status change already triggers one.

8. **No fallback notifications.** With 2 and 3, nothing is shown as a notification any more except start, run and resume failures. `escalate` prints `No Main agent took it; the Board and the sidebar show ⚑. End your turn.` in that case.

## Risks / Trade-offs

- [A Handoff without a Main agent loses its text: merge, archive and worktree removal instructions] → The `⚑` on the Review Card points the user there; `v`, `a` and `s` (resume) work from the Card. Accepted as decided; see Open Questions.
- [The Board's `⚑` is only as fresh as the last lookup] → Same as the Running badge; `r` refreshes.
- [A terminal draws `⚑` two cells wide] → The Card would shift one cell. The goldens fix the width as 1, like `●`; if it shows up in practice, swap the glyph in one constant.
- [Sidebar pass gets one more herdr call] → `agent list` is cheap compared to `openspec list`, and runs once per pass for all Projects.

## Migration Plan

m12's `escalations/` directory is not read any more. A record left there only means a watcher from before the upgrade still runs; it ends with its agent. No migration code.

## Open Questions

- Should the Handoff text be kept for the user when no Main agent took it, for example in the Card's detail view? Not needed for the marker itself; a later change can add it without touching this one.
