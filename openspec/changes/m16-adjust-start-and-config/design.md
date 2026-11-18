## Context

`showBoard` in `cmd/kanban` loads the config once with `config.Load(config.Dir(…))` and builds the Board's callbacks from it: `startOne(change)` prepares the tab and spawns `kanban launch` with `boardPrompt(cfg, src, proj, change, kind, self)`, which joins the Source's command, `instructions` (or the whole `prompt`) and the Escalation sentence. The Board gets `cfg.Hide`, the wave size and a notice, and handles `e` by suspending itself for `$EDITOR` and reloading afterwards. The filter prompt is the Board's existing input line.

See proposal.md for why.

## Goals / Non-Goals

**Goals:**
- Adjust the instructions of one start without touching the config.
- Reach the whole config from the Board.

**Non-Goals:**
- No saving from `S`, no config per Project, no `S` for waves.

## Decisions

### D1: `startOne(change, text *string)` with an override
`boardPrompt` takes an optional override of the configured text: `nil` means the config as today (`s`); otherwise the text replaces `instructions`, or `prompt` when one is set. The Board calls `startOne(change, nil)` for `s` and `startOne(change, &line)` for `S`. Kind detection, worktree, Escalation sentence and launch stay shared, so `S` cannot drift from `s`.
Alternative: a second callback for `S`. That lost because both would have to build the prompt the same way.

### D2: The `S` line reuses the filter prompt's input
It is the same text input with its own prefix `start> ` and its own `enter`/`esc` handling, and only one input line is open at a time. The pre-filled value comes from a Board callback `startText(change) string`, so the Board does not need the config.

### D3: `c` goes through a config holder in `cmd/kanban`
`showBoard` keeps the config in a small holder with `reload() (hide, waveSize, warning)`. All callbacks read the current config from it, not a copy taken at opening. The Board calls `editConfig()` for the path (creating `{}` and its directory when missing), runs `$EDITOR` the way `e` does, then calls `reload` and applies hidden Phases, the wave size and the warning. A missing config directory makes `editConfig` return an error, which the Board shows as a notice.
Alternative: close and reopen the Board after `c`. That lost because it loses the selection and the filter.

## Risks / Trade-offs

- [The user saves broken JSON] → The same fallback as at opening: defaults and the warning. The file stays as the user left it, so the next `c` opens it for a fix.
- [Settings that Starts already running use] → A running Start or wave keeps what it was spawned with. Only later actions use the new config.
- [`S` text with a line break] → The input is one line, so this cannot happen.
