## Context

`board.Model` keeps every Change by Phase in `all [columns][]openspec.Change` and derives `cols` (what `nav` and rendering see) in `refilter`; `nav` works on per-column counts and already skips empty columns, stops at the ends and moves focus off a column that empties. Keys go through `Update` in this order: Start running (only `ctrl+c`), filter prompt, detail view (keys forwarded to `detail.Model`), Board keys. `notice` replaces the footer until the next key. Start's agent kind (`--kind claude`) and prompt are hard-coded in `start.Launch`, which runs in the detached `kanban launch <change> <pane> <tab>` child spawned by the Board, or inline for `kanban start`.

`start.Prepare` tells a running agent from an exited one with `pane process-info`: running when a foreground process has `argv0 == "claude"`, exited when the shell holds the foreground, an error otherwise. herdr keeps an exited agent registered with `agent_status: "done"` (M0 spike), and `done` also means "finished a turn" (M4), so `agent get`/`agent list` cannot tell exit apart. `agent start --kind` accepts 24 kinds (`claude`, `codex`, `gemini`, `opencode`, …) and calls the value the "canonical executable"; `agent get` reports the detected kind as `agent`, and accepts a pane id as target: `herdr agent get <pane_id>` on a pane running Claude returns `"agent":"claude"`. `kanban open` gets `focused_pane_id` in `HERDR_PLUGIN_CONTEXT_JSON`, the pane the Board was opened from; shells inside herdr have `HERDR_PANE_ID`.

herdr (0.9.5) gives every plugin command `HERDR_PLUGIN_CONFIG_DIR`, creates that directory on install/link, and leaves the file format to the plugin; `herdr plugin config-dir <id>` prints it. `herdr plugin pane open` accepts `--placement overlay|split|tab|zoomed` but no width or height; popup size exists only in the manifest or the raw socket request.

## Goals / Non-Goals

**Goals:**
- The prompt template, the setting users asked for, the agent kind, so the plugin is not Claude-only, and hiding Phases, with no new dependency.
- One config load per process, pure and table-tested validation.
- Help that cannot drift far from the keys: one static text in the board package, golden-tested.

**Non-Goals:**
- **Placement (`popup`/`tab`)**: possible through `--placement tab`, but no user has asked for it, and a tab breaks assumptions the Board relies on (it closes after Start, `q` returns focus to the previous pane, each open would create another tab). Cut.
- **Popup size**: the CLI cannot pass it; it would need a socket client or editing the manifest of a managed checkout, which `plugin update` overwrites. Cut.
- Live reload of the config; it is read when the Board opens or Start runs.
- Hiding Phases in the sidebar summary; it keeps counting all open Changes.
- A settings UI or a `kanban config` command; the README documents the file.
- Scrolling inside the help; what does not fit the height is cut off (the README has the full table).

## Decisions

### Format: JSON with `encoding/json`
`config.json`, decoded with `json.Decoder.DisallowUnknownFields`, so a typo in a key is reported instead of silently ignored. Alternatives:
- TOML via `github.com/BurntSushi/toml`: matches herdr's own config and allows comments, but adds a dependency for three keys, and unknown keys need an extra `MetaData.Undecoded` check.
- A hand-written `key = value` format: no dependency and comment-friendly, but looks like TOML without being it (quotes, arrays for `hide`), so users would write TOML and get confusing errors.
JSON costs no comments; with three documented keys that is acceptable. Revisit if the settings grow.

### Package `internal/config`
`config.Config{Agent, Prompt string; Hide []phase.Phase}` (`Agent` empty means detect) with `Default()`, `Parse([]byte) (Config, error)` (validation: unknown keys, wrong types, `agent` not in `config.Kinds`, `prompt` without `{change}`, unknown Phase names matched case-insensitively against `phase.Phase` names, all five hidden) and `Load(dir string) (Config, warning error)`: missing file → defaults and nil; any other problem → defaults and an error naming the file. `Dir(run herdr.Runner) string` returns `HERDR_PLUGIN_CONFIG_DIR` or the trimmed output of `herdr plugin config-dir openspec-kanban`, "" on failure (→ defaults, no warning). `Config.PromptFor(change)` is `strings.ReplaceAll(Prompt, "{change}", change)`. Whole-file fallback rather than per-key: one rule, easy to explain, and the warning tells the user to fix the file.

### Agent kind: the invoking pane's agent, config overrides
Start launches the kind of the first of these that gives one:
1. `agent` in the config, when set.
2. The agent running in the invoking pane: `herdr agent get <pane>`, its `agent` field. For the Board the pane is `focused_pane_id` from the plugin context, which `kanban open` passes to the Board as `--env KANBAN_MAIN_PANE=<id>` (m6 reuses this variable for the Main agent); for `kanban start` it is `HERDR_PANE_ID`.
3. `claude`, when there is no pane, the pane runs no agent (plain shell, `agent_not_found`), herdr fails, or the reported kind is not in `config.Kinds`.

So a user working with Codex who opens the Board from that pane gets Codex without any config. `start.Kind(configured, pane string, run herdr.Runner) string` does this; it never fails. The Board calls it when `s` is pressed, not when it opens, so opening stays one process spawn faster (f1) and the kind is current; it passes the result to `kanban launch`. Alternatives: config only (every non-Claude user must write a file); resolve when the Board opens (extra `agent get` on every open, for a Start that may not happen).

### Allowed kinds: a fixed list, matched exactly
`config.Kinds` is the list from `herdr agent start --help` (herdr 0.9.5). A configured `agent` must be one of them, matched exactly (herdr's values are lowercase); an unknown value is an invalid config like any other: all defaults, warning. A detected kind outside the list falls back to `claude` silently, since the user did not write it. Alternatives: pass any string through and let `agent start` fail (the error comes after the Board closed, as a notification, and the tab is already open); ask herdr at runtime (`--help` parsing, brittle). The list needs a touch when herdr adds kinds; a value herdr knows but the list does not is reported, not silently dropped.

### The prompt stays `/opsx:apply {change}`
OpenSpec installs its commands per tool (`openspec init --tools claude,codex,…`), and their names and syntax differ by tool. The plugin does not map kinds to commands: the default prompt stays the Claude command, and the README says that with another agent kind, detected or configured, the `prompt` has to name the command OpenSpec installed for that tool. Alternative: a built-in kind→command table; it would go stale with every OpenSpec release and still be wrong for users with their own prompt. No warning for a non-Claude kind with the default prompt either: the plugin cannot know which commands a tool has.

### Still-running check: is the shell idle
`agentRunning(run, pane)` keeps the `pane process-info` call but ignores `argv0`: the agent counts as running whenever the foreground process group is not the shell (`foreground_process_group_id != shell_pid`), and Start restarts only when the shell is idle, which is also what `herdr agent start` requires. Node-based kinds (`gemini`, `copilot`, `opencode`, …) show `argv0` `node`, so comparing `argv0` with the kind would call a running agent "another program". `Prepare` no longer needs a kind; the Board resolves it with `start.Kind` only when something has to be launched. Alternatives:
- `agent_status` / `interactive_ready` from `agent get` or `agent list`: no extra call, but herdr keeps an exited agent registered with status `done`, which also means "finished a turn", and `interactive_ready` is not documented as cleared on exit. Not reliable.
- `agent explain`: screen-detection rules per kind; it reports a state, not whether the process lives. Not reliable either.
- `argv0` against the kind: fails for node-based kinds, and a per-kind alias table would need verifying every kind.
Trade-off: a program other than the agent in the foreground (e.g. `vim`) is treated like a running agent: `s` focuses its tab and starts nothing, so nothing is typed into that program.

### Kind and prompt flow as arguments
`start.Launch(change, kind, prompt, pane, tab, run)` runs `agent start <change> --kind <kind>` and sends `prompt` verbatim (herdr reads the text positionally, even with a leading `-`, and rejects `--`); `start.Start` takes `kind` too. The Board process loads the config once and passes kind and rendered prompt to the child: `kanban launch <change> <pane> <tab> <kind> <prompt>` (exec argv, no shell quoting issues). Error texts say "the agent" or the kind instead of "Claude". The child does not read the config again, so the kind and prompt configured when the Board opened are the ones used, and a config warning has exactly one place to appear. `kanban start` loads the config itself and prints a warning to stderr.

### Running badge stays as is
`start.Running` matches agents by `name` and `cwd` only, never by kind, so the badge works for every kind without change.

### Hidden Phases
`board.New` receives the hidden Phases (a `[columns]bool`) and an initial notice (the config warning, "" otherwise). `refilter` leaves hidden columns empty, so `nav` skips them, `1`–`5` on them do nothing and the focus fallback avoids them without any change to `nav`. `columns()` lays out only shown columns: the width is divided by their number, and the narrow-mode threshold uses it too. The narrow indicator keeps Phase numbers (`◀ 2/5 Ready ▶`) to match the digit keys. The header's Progress reads `all`, so hidden open Changes still count.

### Help overlay
`help bool` in the Model. `?` sets it on the Board and while the detail view is open; the check sits after the Start and prompt checks (so `?` is typed in the prompt and ignored while starting) and before forwarding keys to the detail view. While `help` is set, `ctrl+c` quits and any other key only clears it; the detail model is untouched, so its scroll position survives. `View` renders the help in place of everything when `help` is set: a title `Keys`, the three groups stacked (Board, Filter prompt, Details) as two aligned columns key/description in the default foreground, group titles bold, and a last line `any key closes`, cut to the view height. The text is one constant next to the footer strings; the footers gain `? help` (narrow variant `?`).

## Risks / Trade-offs

- [Another program holds the agent's tab] → `s` only focuses that tab; the user quits the program and presses `s` again.
- [herdr adds or renames kinds] → `config.Kinds` rejects new ones with a warning naming the value; updating the list is a one-line change.
- [Another kind with the default `/opsx:apply` prompt] → that tool may not know the command; the README says so next to `agent`.

- [Help text and actual keys drift apart] → golden test of the help; the keys table in the README and the help are updated in the same task.
- [Help taller than a small popup] → cut at the bottom; Board group comes first since it is used most.
- [A config warning disappears with the first key] → it names the file and the problem; `kanban start` repeats it on stderr. Persisting it in the header was considered and rejected as clutter for a one-time fix.
- [No comments in JSON] → README shows a complete example; unknown keys are rejected, so commented-out keys like `"_prompt"` fail loudly rather than silently.
- [Focus on a hidden Draft when nothing is shown] → when no column has Cards, nav focuses Draft; rendering shows the first shown column in narrow mode instead.
- [`herdr plugin config-dir` fallback costs a process spawn] → only for `kanban start` from a shell; plugin commands have the env var.
