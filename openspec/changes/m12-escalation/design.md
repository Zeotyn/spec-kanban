## Context

- `kanban launch` (spawned by the Board's `startOne`) runs `start.Launch`, then `start.Watch` when `$KANBAN_PROJECT` is set. `Watch` knows the Main pane from `$KANBAN_MAIN_PANE`, which the Board passes on to every spawned child.
- `Watch` waits for a settle, records the session, runs `check`, and then either hands off (`handoff`) or shows a "stopped" or "waits for input" notification and waits until the agent works again.
- `kanban resume` runs `start.Resume` and is not watched.
- `internal/cache` keeps per-Project, per-Change files under `sessions/<project hash>/<change>` with atomic writes.
- The worker agent runs in the Change's checkout, usually a linked worktree. It has no `$KANBAN_PROJECT`. Its shell runs commands with the pane's environment.
- herdr offers `agent wait --until working --timeout <ms>`, `agent read` and `agent prompt`.

## Goals / Non-Goals

**Goals:**
- One delivery path for Handoff, Escalation and Stall: Main agent if it can take the prompt, else a notification.
- No new long-running process: Stall lives in the existing watcher, Escalation is a short command.
- Everything the escalating worker needs is in its Start prompt; nothing to install or configure.

**Non-Goals:**
- A config key for the 10 minutes, or for the sentence.
- The plugin reading or judging the worker's output. The Main agent does that with `herdr agent read`.
- Watching `kanban start` from a shell, or Verify and Archive.
- The attention marker on Card and sidebar. That is m13-attention.

## Decisions

1. **Main pane record in the state dir.** `cache.WriteMain(dir, project, change, pane)` and `cache.ReadMain` store the pane id as plain text under `mains/<project hash>/<change>`. An empty pane removes the file.
   - `launch` and `resumeChange` write it before they watch, from `$KANBAN_MAIN_PANE`.
   - Rejected: keying by herdr agent name. The m13 marker is read by the Board and sync per Project, so all records share the Project/Change key.
   - Rejected: passing the Main pane in the sentence. It is not the worker's business, and the pane can change with a later resume.

2. **Escalation record.** `escalate` writes `escalations/<project hash>/<change>` (an empty file). The watcher reads it at an unfinished settle: present means no Stall timer. The watcher removes it when it sees the agent working again after that settle. `launch` and `resumeChange` remove it before watching, so a leftover from an earlier run never hides a Stall.
   - Rejected: consuming it at the settle. m13 keeps the same file until the agent works again, as the attention marker. Same lifetime now means a smaller m13.

3. **`kanban escalate <change> <reason...>`.**
   - The reason is all remaining arguments joined by spaces, so a missing pair of quotes does no harm.
   - Project: `project.Find(cwd)` gives the checkout the worker runs in. The Project is the first of that checkout and the other checkouts from `worktree.List` that holds a Main record for the Change. Without any record it is the checkout itself, and the prompt falls back to a notification. This works for a linked worktree, for `worktree` off, and for a Project that is itself a linked worktree.
   - It checks the Change with the Source's `Open`, loads Progress through the Board's `loader`, and builds the tasks path as `<checkout>/<src.Rel(change)>/tasks.md`. `<where>` names the checkout when it is not the Project.
   - It prints one line for the worker: `Sent to the Main agent. End your turn.` or `No Main agent took it; the user got a notification. End your turn.` Exit status 0 in both cases; usage errors exit 2, other failures 1.
   - Rejected: a `--project` flag in the sentence. Longer sentence, one more thing for the worker to get wrong.

4. **One delivery helper.** `start.Tell(run, main, own, title, text) bool` holds what `handoff` does today: skip when `main` is empty or equals the Change's own pane, `agent get main` must report an agent that is not blocked, then `agent prompt`; otherwise `notification show <title> --body <text>`. It returns whether the Main agent took it. Handoff, Stall and Escalation all use it. m13 swaps the notification branch for the marker in one place.

5. **The sentence.** `start.Sentence(self, change)` returns `If you cannot continue without a human, run <kanban> escalate <change> "<what you need>" and end your turn.` `self` comes from `os.Executable` in the Board. A path with characters other than `[A-Za-z0-9/._-]` is wrapped in single quotes (a single quote inside becomes `'\''`). `startOne` appends it with one space after `cfg.PromptFor`. `config.PromptFor` stays unchanged, so `kanban start` gets no sentence.
   - Rejected: adding it inside `PromptFor`. `kanban start` uses `PromptFor` too and is not watched and has no Main record.

6. **Stall inside `Watch`.** After an unfinished `idle` or `done` settle without an escalation record:
   - `agent wait <agent> --until working --timeout 600000`.
   - Success: the agent works again; loop as now.
   - Error: `agent get`. Gone → return. Otherwise prompt the Stall text through `Tell`, then `agent wait --until working` without timeout, as today.
   - Telling a timeout apart from a gone agent by a second `agent get`, not by parsing herdr's error text.
   - The 10 minutes are a constant, `stallAfter`. Tests use a fake runner that fails the timed wait.
   - A `blocked` settle skips Stall and only waits until working.

7. **Both notifications go in m12.** The "stopped" and the "waits for input" notifications are removed together.
   - "Stopped" is replaced by Stall, so it must go here.
   - "Waits for input" has no replacement in m13 either: herdr's sidebar already shows a blocked agent, and the plugin can do nothing about a permission dialog. Keeping it one more milestone would leave a spec where blocked notifies but stopped does not.
   - What stays: start, run and resume failures, and the `Tell` fallbacks (finished, needs you, stalled), which m13 turns into the marker.

8. **Resume is watched.** `resumeChange` calls `Watch` after `Resume` succeeds, with the same `Watched` as `launch` plus `Resumed: true`. A resumed watch first runs `agent wait --until working` without a timeout, instead of the 60-second wait a Start uses. So a reopened, finished session is not evaluated before the user made it work. The shared `Watched` setup moves into one helper in `cmd/kanban` used by both.

9. **`check` returns Progress, not a notification body.** Its third result becomes the Progress text (`3/7 tasks`, "" when unknown), used for the Stall prompt.

## Risks / Trade-offs

- [The worker ignores the sentence and just stops] → Stall catches it after 10 minutes.
- [The worker escalates for things it could do itself] → The Main agent does the step or answers it. The cost is one extra prompt.
- [A resumed, finished session is handed off after the user talks to it, and its pane closes] → Accepted as decided: Resume is watched like a Start. The Handoff prompt says how to reopen it again.
- [The prompt reaches a Main agent that is working] → herdr queues it, as for the Handoff today.
- [Fewer notifications: with no Main agent, the user only learns of a Stall from the fallback notification] → m13 adds the marker on Card and sidebar.
- [Agents that never call shell commands cannot escalate] → Stall still applies.
- [A resumed session only knows the sentence if its first prompt had it] → Sessions started after this change have it in their history; older ones fall back to Stall.

## Migration Plan

No data migration. Existing session records stay valid. A running watcher from an older binary keeps the old behaviour until its agent goes away.
