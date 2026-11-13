## Context

Five places report a failure after the Board closed as `spec-kanban: <step> failed`:
- `launch` (start)
- `runCommand` (run)
- `resumeChange` (resume)
- `waveLaunch` (wave)
- `waveMember` (start)

All five run detached, spawned by `spawn` without stdout or stderr. Each knows its Project:
- `launch`, `runCommand` and `resumeChange` from `$KANBAN_PROJECT`;
- `waveLaunch` and `waveMember` from their arguments and the wave record.

`internal/cache` already keeps per-Project records in two places:
- `cache.Dir()`: the plugin's state directory when `$HERDR_PLUGIN_STATE_DIR` is set;
- `cache.UserDir()`: the user's cache directory, for records that agent panes and shells read.

It writes them atomically with `writeAtomic`.

`showBoard` turns a config warning into the Board's initial `notice`, which the footer shows until the next key. `kanban escalate` finds its Project from the current directory with `project.Find` and the repository's checkouts from `worktree.List`.

See proposal.md for why.

## Goals / Non-Goals

**Goals:**
- Keep the full error text of every detached failure, readable from any shell, and point the notification at it.
- Never let the log change whether a launch or the Board works.

**Non-Goals:**
- No record of failures the Board shows itself in its footer: the user saw them.
- No locking between processes beyond what a single append gives (see Risks).
- No rotation by age, no per-Project files.

## Decisions

### D1: One JSON-lines file in the user cache directory
The log lives at `cache.UserDir()/failures.log`, one JSON object per line: `time` (RFC 3339), `project`, `change`, `step`, `error`.
- Why the user cache directory: `kanban failures` runs in shells and agent panes without `$HERDR_PLUGIN_STATE_DIR`, the same reason m12 moved the Main records there.
- Why one file: the 100-record limit is global, and `--all` reads everything at once.

Alternatives:
- One file per Project. That lost because it needs a global limit across files and a directory scan for `--all`.
- Plain text. That lost because multi-line errors with the pane's last lines need a safe record boundary.

### D2: Append, then trim when over the limit
A record is written with one `O_APPEND` write of a single line. When the file then holds more than 100 valid lines, the writer rewrites the newest 100 valid lines through `writeAtomic`. Invalid lines are skipped on read and dropped on that rewrite.
Alternatives: always rewrite atomically. That lost because two failing processes would each overwrite the other's record. An append keeps both, and only a concurrent trim can lose one (see Risks).

### D3: One helper reports a failure
`cmd/kanban` gets `fail(run, step, proj, change, err)`. It writes the record, ignoring a write error, and shows the notification with body `err.Error() + "\n\nDetails: kanban failures"`. The five call sites call it instead of `run("notification", "show", …)`. Their return values stay unchanged. The Board's own notices are untouched.

### D4: `kanban failures` finds the Project like `kanban escalate`
It runs `project.Find(cwd, detect)`, then tries that checkout and every other checkout of the repository (`worktree.List`), with resolved symlinks. It takes the first one that has records. Without records it uses the found checkout and prints `no failures`. `--all` skips the lookup. The output is plain text so the Main agent can read it: a header line `<local time>  <step>  <change>  <project>`, then the error indented by two spaces, then a blank line.

### D5: The Board's notice is computed in `showBoard`
`showBoard` already sets the initial `notice` from the config warning. Without a warning, it reads the log and the Project's last opening, and formats the newest newer failure as `<step> failed <change> <age> ago: <first line>`. Then it writes the new opening time. With a warning, it reads and writes neither.
- The opening record is a small file under `cache.UserDir()/opened/<project hash>`, holding an RFC 3339 time and written atomically.
- Without a record, the cutoff is 24 hours ago.
- The `board` package needs no change: the notice already shows until the next key and is cut to the width.

Alternatives: a "seen" flag in each log record. That lost because it rewrites the log on every Board opening and races with appends.

## Risks / Trade-offs

- [A trim and an append race, and one record is lost] → Both need two failures within milliseconds and a log over 100 records. The notification still showed, so the loss is only the later copy.
- [Error texts can hold paths and pane output] → The file stays in the user's own cache directory with mode `0600`, like the other records. Nothing leaves the machine.
- [The Board reads one more file on opening] → The file is at most 100 short lines, read before the Board shows. That is well below the cached snapshot's cost.
- [A failure in a Project the user never opens the Board for] → It stays readable through `kanban failures --all`. No notice is needed.
