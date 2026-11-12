## Context

See proposal.md for the bug. What herdr 0.9.5 offers, checked against its `--help`, `--skill` text and the v0.9.5 docs:

- `tab create`, `pane split`, `worktree create` and `worktree open` always start the user's shell; none takes a command to run instead of it.
- `agent start` types `<kind>` and Enter into the pane and waits for the agent to be detected. Its own text says the pane must be at its interactive shell prompt, and that is the caller's job: nothing checks it.
- There is no shell-ready field, flag or socket event. `pane process-info` cannot tell an rc that waits in a builtin `read` from the prompt: both show the shell in the foreground.
- `pane run <pane> <text>` queues text and Enter; `pane wait-output <pane> --match <text> --timeout <ms>` searches the recent snapshot at once and then polls; `pane read --source recent-unwrapped --lines N` returns the last lines without soft wraps.

All starts go through two functions: `start.Launch` (Start, Verify, Archive) and `start.Resume` (Resume), both in `internal/start`, both already called with the pane and tab to use. `cmd/kanban` turns their errors into the notifications `start failed`, `run failed` and `resume failed`.

## Goals / Non-Goals

**Goals:**
- Never type the agent command into a shell that is not at its prompt, for any shell and rc the plugin cannot know about.
- When the start still fails, show what the pane was doing.

**Non-Goals:**
- Fixing the user's rc, or detecting which dialog is in the way.
- A config key to turn the probe off or tune its timing.
- Changing herdr; an upstream request for a shell-ready signal is a separate step.

## Decisions

### 1. Probe the prompt with a command whose output differs from its echo

Before `agent start`, `Launch` and `Resume` type ` echo kanban-rea""dy-<nonce>` with `pane run` and wait with `pane wait-output --match kanban-ready-<nonce> --timeout 2000`. The marker appears only after the shell has read and run a complete line, which is exactly the state `agent start` needs.

- `rea""dy`: the typed line shows `kanban-rea""dy-<nonce>`, never the marker, so the echo of the input cannot end the wait early. That echo is what fooled a naive "is `claude` visible" check in the bug report. `""` is an empty string in sh, bash, zsh and fish.
- Leading space: a dialog that reads one key gets a space, the "no" answer of `[Y/n]` style questions; the shell ignores it, and zsh with `HIST_IGNORE_SPACE` keeps the line out of its history.
- Nonce: six hex characters from `crypto/rand`, new for every probe, so the recent snapshot, which `wait-output` searches first and which may still hold an earlier Start's marker in a reused tab, cannot answer. A package variable `nonce func() string` is swapped in tests for a counter.
- Alternatives rejected: waiting for the pane to go quiet (a waiting dialog looks quiet too); `pane process-info` (see Context); a shell-integration prompt mark (OSC 133; the plugin does not control the user's shell); matching the typed echo (the bug).

### 2. Repeat every 2 s, give up after 60 s, stop at once when herdr refuses

`ready(pane, run)` loops: `pane run`, `pane wait-output --timeout 2000`; on a wait timeout a new probe with a new nonce, up to 30 probes (60 s, the same budget `agent start` gets). A failed `pane run` returns its error at once, since the pane is gone or herdr is. 2 s is long enough for the shell to execute a queued line once it is at the prompt, and short enough that a dialog which swallows a whole line costs one retry. The cap is a probe count, not a clock, so tests need no time.

What the user sees in the tab: one `kanban-ready-<nonce>` line per answered probe, and, per dialog that ate the first character, one `command not found: cho` line. An agent tab starts with output anyway; the lines are cheap and say what they are.

### 3. Probe before every start, also in a reused tab

`Launch` and `Resume` probe unconditionally after `tab focus`. The exited-agent tab is at its prompt, so the probe costs one round trip and one line there. Knowing in `place` whether the pane is fresh would thread a flag through three return paths for no user-visible gain.

### 4. The failure carries the pane's last lines

`tail(pane, run)` runs `pane read --source recent-unwrapped --lines 20`, drops empty lines and keeps the last 8; the error from a probe timeout (`shell did not reach its prompt in tab "<change>"`) and from `agent start` (`<kind> did not become ready in tab "<change>", the prompt was not sent`) gets `; pane shows: <lines joined by " | ">` appended, or `; pane could not be read: <err>` when `pane read` fails. The joined form keeps a herdr notification body on a few lines. `cmd/kanban` needs no change: it already puts `err.Error()` into the notification.

## Risks / Trade-offs

- [A shell that does not read `""` as an empty string, like nushell] → every probe times out, the start fails after 60 s with the pane's lines showing the shell's complaint; the user sees why. Named in the README as a known limit.
- [A wizard in the rc that takes the probes as navigation] → same failure after 60 s; the pane's lines show the wizard. The space is the least harmful key to send.
- [herdr changes `pane run` or `wait-output` semantics] → the herdr skill text documents both as stable, versions are pinned in the README requirement line.
- [Two probes answered at once after a slow rc] → both markers print; the wait for the latest nonce ends; harmless.
- [macOS canonical input limit] → the probe is under 40 characters, far below the 1024-byte limit herdr checks.
