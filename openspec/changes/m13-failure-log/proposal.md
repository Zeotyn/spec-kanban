## Why

A Start, Verify, Archive or resume that fails after the Board closed reports its error only as a herdr notification, and that notification is gone after a few seconds. Nothing else keeps the error. The launch runs detached without stdout or stderr, so `herdr plugin log list` does not show it either. Example: Verify on `s001-plugin-bundles-hooks` showed `spec-kanban: run failed`. The cause (an outdated binary started the agent before the shell had a prompt) had to be reconstructed from the pane and the binary, because the error text was lost. m13-attention keeps exactly these failure notifications. This change makes them readable later.

## What Changes

- **Failure log.** Each failure that shows a `spec-kanban: … failed` notification also appends one record to `failures.log` in the plugin's state directory.
  - A record holds the time, the Project, the Change, the step (`start`, `run`, `resume`) and the full error text, including the pane's last lines when the error carries them.
  - The log keeps the newest 100 records and drops older ones.
  - The notification still shows when the record cannot be written.
- **`kanban failures`.** A new subcommand prints the failures of the Project of the current directory, newest last. With `--all` it prints the failures of every Project. The user and the Main agent read the full text with it.
- **Notification hint.** The notification body ends with `kanban failures`, so the user knows where the full text is.
- **Board.** When the Board opens and its Project has a failure newer than the Board's last opening, the footer shows it once as a notice: step, Change, age and the first line of the error, for example `run failed s001-plugin-bundles-hooks 4m ago: shell did not reach its prompt…`. The next opening does not show it again.

Non-goals: no attention marker `⚑` for failures, no failure tab in the detail view, no logging of successful launches.

## Capabilities

### New Capabilities

- `failure-log`: the record format, the 100-record limit, `kanban failures` with and without `--all`, and that a log problem never breaks a launch or the Board.

### Modified Capabilities

- `agent-start`: a Start, Verify or Archive failure after the Board closed is also recorded, and its notification names `kanban failures`.
- `agent-handoff`: a failed resume is recorded the same way.
- `board`: on opening, the Board shows the newest unseen failure of its Project as a notice.

## Impact

- **Code:**
  - `internal/cache`: append to, trim and read `failures.log`; a per-Project "last seen" time for the Board.
  - `cmd/kanban`: record failures in `launch`, `runCommand` and `resumeChange`; the notification hint; the `failures` subcommand and its usage text.
  - `internal/board`: the notice on opening.
- **Docs:** README (troubleshooting: `kanban failures`), CONTEXT.md (failure log), CHANGELOG.
- Independent of m12-escalation and m13-attention; it can be built before them.
- No new dependencies.
