## Context

The Board starts work in two ways: Start (`startCard`, then `startOne` in `cmd/kanban`: `start.Prepare`, then a detached `kanban launch` that runs `start.Launch` and, from the Board, `start.Watch`) and Resume (`kanban resume`).
- `start.Prepare` finds the Change's agent with `agent get`. When its agent process holds the pane's foreground (`agentRunning`), it only focuses it; otherwise it returns its pane or opens the Change's checkout.
- The Board learns which Changes have an agent from `start.Running`, which ignores the agent status.
- herdr reports `agent_status` (`idle`, `working`, `blocked`, …) on `agent get`.
- The Source interface (`internal/work`) already gives the Start prompt (`Prompt`) and the Handoff step (`Finish`).

## Goals / Non-Goals

**Goals:**
- Verify and Archive reuse Start's placement: the same checkout, tab reuse and refusal of foreign agents.
- One status check, done when the action runs, so the Board never sends to a busy agent.

**Non-Goals:**
- A Spec Kit equivalent. `/speckit-analyze` checks before implementing, not after.
- Watching Verify or Archive, or reading their result.
- Footer entries for `v` and `a`. The Card menu covers mouse users.
- Merging after Archive. When the Change has a worktree, its branch still needs a merge, which the user or the Main agent does as before.

## Decisions

1. **Source interface: `Command(action work.Action, name string) string`**, with `work.Verify` and `work.Archive`, where "" means the Source has none.
   - One method rather than two keeps the interface small, and BMAD can later answer per action.
   - OpenSpec returns `/opsx:verify <name>` and `/opsx:archive <name>`. Spec Kit returns "".
   - Rejected: a list of actions per Source. Nothing needs more than these two yet.

2. **`start.Run(project, change, action, kind, wt, run, src, git)`** next to `Prepare`.
   - It reads the agent with `agent get`, the same lookup `Prepare` does, refusing a foreign one in the same way.
   - If the agent process holds the foreground: status `idle` → `agent prompt <agent> <command>` and `tab focus`. Any other status → an error that carries the status.
   - Otherwise → `Prepare`, and the caller spawns a detached `kanban run <change> <pane> <tab> <kind> <command>`: `Launch` without `Watch`.
   - Rejected: a flag on `kanban launch`. Its fixed argument count is part of the spawn contract, and a separate subcommand reads clearer than a positional flag.
   - Rejected: reading the status in `start.Running` for the menu's faint state. That status is only as fresh as the last lookup, while `Run` checks at the moment of the action anyway. The menu shows an entry faint only for what the Board knows for sure (Phase, Source); a busy agent shows as a notice on click.


3. **The Board's callbacks.** `board.New` gets one more callback, `act(change string, a work.Action) error`, plus `has(a work.Action) bool` for the Source's support.
   - `v`/`a` check the Phase and the support first (notices), then set `starting` with a verb ("verifying", "archiving") and call `act` in a command, exactly like `startCard`.
   - A returned busy error becomes the notice `<name> is <status>: wait until it is idle`.

4. **Prompt and instructions do not apply.** `act` passes the Source's command unchanged. `config.PromptFor` is only for Start.

5. **Card menu with up to three entries.**
   - The `menu` box renders a list of entries, each with its key. A click maps the entry line to its key and feeds it through `press`. Pressing an entry's key while the menu is open does the same.
   - Spec Kit Cards keep the single-entry menu.

6. **Smoke tests without checkboxes.** In `tasks.md` the user smoke test is a plain list under its own heading. The task count, and so Review and the Handoff, does not depend on the user.

## Risks / Trade-offs

- [The agent's status changes between the check and the prompt] → herdr queues a prompt to an agent that has just started working, and the user sees it in the tab. Accepted.
- [Archive in a worktree leaves an unmerged branch without a Handoff] → That is what the user asked for: the action runs without a Handoff. The Card shows Archived, read from the worktree, and the README says that merging stays manual.
- [A new agent for Verify has none of the implementing agent's context] → Intended: a fresh read checks the code against the artifacts without bias.
