## Why

A Start, Verify, Archive, resume or wave that fails after the Board closed reports its error only as a herdr notification, and that notification is gone after a few seconds. Nothing else keeps the error. The launch runs detached without stdout or stderr, so `herdr plugin log list` does not show it either. Two examples:
- Verify on `s001-plugin-bundles-hooks` showed `spec-kanban: run failed`. The cause (an outdated binary started the agent before the shell had a prompt) had to be reconstructed from the pane and the binary, because the error text was lost.
- The manual check of m14-waves started a wave that failed and left branch `wave-1` behind. Its error was never read, so the cause is still unknown.

m13-attention kept exactly these failure notifications. This change makes them readable later.

## What Changes

- **Failure log.** Each failure that shows a `spec-kanban: <step> failed` notification also appends one record to `failures.log` in the user's cache directory, next to the Main records. Agent panes and plain shells can read it there, unlike the plugin's state directory.
  - A record holds the time, the Project, the Change (for a wave: its Changes), the step (`start`, `run`, `resume`, `wave`) and the full error text, including the pane's last lines when the error carries them.
  - The log keeps the newest 100 records and drops older ones.
  - The notification still shows when the record cannot be written.
- **`kanban failures`.** A new subcommand prints the failures of the Project of the current directory, oldest first. With `--all` it prints the failures of every Project. The user and the Main agent read the full text with it.
- **Notification hint.** The notification body ends with `Details: kanban failures`, so the user knows where the full text is.
- **Board.** When the Board opens and its Project has a failure newer than the Board's last opening, the footer shows the newest one once as a notice: step, Change, age and the first line of the error, for example `run failed s001-plugin-bundles-hooks 4m ago: shell did not reach its prompt…`. The next opening does not show it again.

Non-goals: no attention marker `⚑` for failures, no failure tab in the detail view, no logging of successful launches, no cleanup of what a failed launch left behind.

## Capabilities

### New Capabilities

- `failure-log`: which failures are recorded, the record, where the log lives, the 100-record limit, the notification hint, `kanban failures` with and without `--all`, and that a log problem never breaks a launch or the Board.

### Modified Capabilities

- `board`: on opening, the Board shows the newest unseen failure of its Project as a notice.

The Start, Verify, Archive, resume and wave requirements keep their notifications unchanged. The failure-log spec adds the record and the hint to every one of them.

## Impact

- **Code:**
  - `internal/cache`: append to, trim and read `failures.log`, and a per-Project "last opened" time for the Board, both in the user cache directory.
  - `cmd/kanban`: record failures where `launch`, `runCommand`, `resumeChange`, `waveLaunch` and `waveMember` show their notification, add the hint, and add the `failures` subcommand and its usage text.
  - `internal/board`: the notice on opening.
- **Docs:** README (troubleshooting: `kanban failures`), CONTEXT.md (failure log), CHANGELOG.
- No new dependencies.
