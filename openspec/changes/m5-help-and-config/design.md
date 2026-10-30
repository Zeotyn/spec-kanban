## Context

`board.Model` keeps every Change by Phase in `all [columns][]openspec.Change` and derives `cols` (what `nav` and rendering see) in `refilter`; `nav` works on per-column counts and already skips empty columns, stops at the ends and moves focus off a column that empties. Keys go through `Update` in this order: Start running (only `ctrl+c`), filter prompt, detail view (keys forwarded to `detail.Model`), Board keys. `notice` replaces the footer until the next key. Start's prompt is hard-coded in `start.Launch`, which runs in the detached `kanban launch <change> <pane> <tab>` child spawned by the Board, or inline for `kanban start`.

herdr (0.9.5) gives every plugin command `HERDR_PLUGIN_CONFIG_DIR`, creates that directory on install/link, and leaves the file format to the plugin; `herdr plugin config-dir <id>` prints it. `herdr plugin pane open` accepts `--placement overlay|split|tab|zoomed` but no width or height; popup size exists only in the manifest or the raw socket request.

## Goals / Non-Goals

**Goals:**
- The prompt template, the setting users asked for, and hiding Phases, with no new dependency.
- One config load per process, pure and table-tested validation.
- Help that cannot drift far from the keys: one static text in the board package, golden-tested.

**Non-Goals:**
- **Placement (`popup`/`tab`)**: possible through `--placement tab`, but no user has asked for it, and a tab breaks assumptions the Board relies on (it closes after Start, `q` returns focus to the previous pane, each open would create another tab). Cut.
- **Popup size**: the CLI cannot pass it; it would need a socket client or editing the manifest of a managed checkout, which `plugin update` overwrites. Cut.
- **Agent kind**: herdr knows many kinds, but `/opsx:apply` is a Claude command, the running-agent check matches the `claude` executable, and UI texts and specs say Claude. Supporting other agents is its own change if someone needs it. Cut.
- Live reload of the config; it is read when the Board opens or Start runs.
- Hiding Phases in the sidebar summary; it keeps counting all open Changes.
- A settings UI or a `kanban config` command; the README documents the file.
- Scrolling inside the help; what does not fit the height is cut off (the README has the full table).

## Decisions

### Format: JSON with `encoding/json`
`config.json`, decoded with `json.Decoder.DisallowUnknownFields`, so a typo in a key is reported instead of silently ignored. Alternatives:
- TOML via `github.com/BurntSushi/toml`: matches herdr's own config and allows comments, but adds a dependency for two keys, and unknown keys need an extra `MetaData.Undecoded` check.
- A hand-written `key = value` format: no dependency and comment-friendly, but looks like TOML without being it (quotes, arrays for `hide`), so users would write TOML and get confusing errors.
JSON costs no comments; with two documented keys that is acceptable. Revisit if the settings grow.

### Package `internal/config`
`config.Config{Prompt string; Hide []phase.Phase}` with `Default()`, `Parse([]byte) (Config, error)` (validation: unknown keys, wrong types, `prompt` without `{change}`, unknown Phase names matched case-insensitively against `phase.Phase` names, all five hidden) and `Load(dir string) (Config, warning error)`: missing file → defaults and nil; any other problem → defaults and an error naming the file. `Dir(run herdr.Runner) string` returns `HERDR_PLUGIN_CONFIG_DIR` or the trimmed output of `herdr plugin config-dir openspec-kanban`, "" on failure (→ defaults, no warning). `Config.PromptFor(change)` is `strings.ReplaceAll(Prompt, "{change}", change)`. Whole-file fallback rather than per-key: one rule, easy to explain, and the warning tells the user to fix the file.

### Prompt flows as an argument
`start.Launch(change, prompt, pane, tab, run)` sends `prompt` verbatim; `start.Start` takes it too. The Board process loads the config once and passes the rendered prompt to the child as a fifth argument: `kanban launch <change> <pane> <tab> <prompt>` (exec argv, no shell quoting issues). The child does not read the config again, so the prompt the user saw configured when the Board opened is the one sent, and a config warning has exactly one place to appear. `kanban start` loads the config itself and prints a warning to stderr.

### Hidden Phases
`board.New` receives the hidden Phases (a `[columns]bool`) and an initial notice (the config warning, "" otherwise). `refilter` leaves hidden columns empty, so `nav` skips them, `1`–`5` on them do nothing and the focus fallback avoids them without any change to `nav`. `columns()` lays out only shown columns: the width is divided by their number, and the narrow-mode threshold uses it too. The narrow indicator keeps Phase numbers (`◀ 2/5 Ready ▶`) to match the digit keys. The header's Progress reads `all`, so hidden open Changes still count.

### Help overlay
`help bool` in the Model. `?` sets it on the Board and while the detail view is open; the check sits after the Start and prompt checks (so `?` is typed in the prompt and ignored while starting) and before forwarding keys to the detail view. While `help` is set, `ctrl+c` quits and any other key only clears it; the detail model is untouched, so its scroll position survives. `View` renders the help in place of everything when `help` is set: a title `Keys`, the three groups stacked (Board, Filter prompt, Details) as two aligned columns key/description in the default foreground, group titles bold, and a last line `any key closes`, cut to the view height. The text is one constant next to the footer strings; the footers gain `? help` (narrow variant `?`).

## Risks / Trade-offs

- [Help text and actual keys drift apart] → golden test of the help; the keys table in the README and the help are updated in the same task.
- [Help taller than a small popup] → cut at the bottom; Board group comes first since it is used most.
- [A config warning disappears with the first key] → it names the file and the problem; `kanban start` repeats it on stderr. Persisting it in the header was considered and rejected as clutter for a one-time fix.
- [No comments in JSON] → README shows a complete example; unknown keys are rejected, so commented-out keys like `"_prompt"` fail loudly rather than silently.
- [Focus on a hidden Draft when nothing is shown] → when no column has Cards, nav focuses Draft; rendering shows the first shown column in narrow mode instead.
- [`herdr plugin config-dir` fallback costs a process spawn] → only for `kanban start` from a shell; plugin commands have the env var.
