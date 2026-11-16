## Context

Every prompt to a Main agent goes through `start.Tell(run, main, own, text)`:
- the watcher's Handoff (`watch.go`, after the Change's pane closed) and Stall (`watch.go`, inside the watch loop);
- `kanban escalate` (`cmd/kanban/main.go`);
- `kanban wave escalate` and `kanban wave finish` through `tell` (`cmd/kanban/wave.go`).

`Tell` checks the agent (`herdr agent get`: an agent, not `blocked`) and runs `herdr agent prompt`. The watcher runs detached. The three subcommands run inside an agent's turn. `spawn` already starts detached entries (`launch`, `run`, `resume`, `wave-launch`).

`herdr pane read <pane> --source visible` returns the screen as text. In Claude Code the input box sits between two rules of `─`, with the input after `❯`, and only a status line or two below it. `herdr agent get` reports `focused`.

See proposal.md for why.

## Goals / Non-Goals

**Goals:**
- Never type into input the user has started, and never delay a prompt the user cannot collide with.
- Keep every caller's fallback unchanged: a prompt that waited too long is a prompt "not taken".

**Non-Goals:**
- No input-box detection for kinds other than `claude`.
- No change to prompts into a Change's own pane.

## Decisions

### D1: A pure parser finds Claude's input box
`inputBox(screen string) (text string, found bool)` reads the screen from `herdr pane read <pane> --source visible --format ansi` and scans the lines from the bottom:
- it finds the last two lines made only of `─` (after trimming spaces);
- it takes the lines between them;
- it requires the first one to start with `❯`;
- it returns the text after `❯` joined with the following lines, trimmed of spaces and non-breaking spaces, and without dim text.

A fresh Claude Code session shows a dimmed placeholder such as `Try "edit <filepath> to..."` in its empty box. Plain text cannot tell it from typed input, so the parser reads the ANSI screen and drops text drawn dim (SGR 2).

Without such a box it returns `found == false`. It is table-tested with screens captured from Claude Code 2.1.325: empty, the placeholder, one line, several lines, and no box.
Alternative: ask herdr for the input line. That lost because herdr has no such query.

### D2: `Tell` waits through an injected clock
`Tell` keeps its signature and behaviour for callers, with a wait added before the prompt. Before it runs `herdr agent prompt`, it calls `clear(run, pane, kind)`, which reports whether the prompt may go out now:
- for `claude` with a found box: the box is empty;
- otherwise: the pane is not focused.

`Tell` polls `clear` every 2 seconds for at most 10 minutes. The sleep and the clock are package variables, so tests run without waiting. After 10 minutes `Tell` returns false without prompting.
A second function `TellNow(run, main, own, text) (taken, waiting bool)` does one check and no waiting. The subcommands use it.

### D3: Subcommands hand a waiting prompt to a detached `tell`
When `TellNow` reports `waiting`, the subcommand spawns `kanban tell <pane> <title> <text>` with the same `spawn` as `launch`. Then it prints `Queued for the Main agent. End your turn.` and returns. The detached `tell` runs `Tell`.
- When `Tell` reports false and `title` is not empty, it shows `title` as a herdr notification. That is the wave fallback.
- `escalate` passes an empty title, because it has already set the attention marker before spawning.
- `wave finish` marks the wave finished before spawning.

Alternative: block the subcommand. That lost because an agent's tool call would hang for up to 10 minutes and the agent could not end its turn.

### D4: The watcher waits inline
The watcher already runs detached, so its Handoff and Stall call the waiting `Tell` directly.
- A Handoff waits after the Change's pane closed. Nothing else depends on it then.
- A Stall waits inside the watch loop, so the loop misses settles during that wait. That is acceptable, because the next poll after the wait reads the current state again.

## Risks / Trade-offs

- [Claude Code changes its input box] → The parser then finds no box and falls back to focus. That is safe but slower. The parser tests name the version the screens come from.
- [The screen shows an old box above the real one] → The parser uses the last two rules on the screen, which is the live box.
- [Text in the box that the user abandoned] → The prompt waits 10 minutes, then the marker or notification shows it. The user's text is never touched.
- [Two prompts queue for the same pane] → Each waits on its own and goes out when the box is empty. herdr submits them one after the other, and Claude Code queues the second while it works on the first.
- [A Stall prompt waits while its worker resumes] → The prompt can arrive after the worker already works again. The marker is set before the wait and clears when the worker works again, and the prompt itself tells the Main agent to do nothing when the worker continues by itself.
