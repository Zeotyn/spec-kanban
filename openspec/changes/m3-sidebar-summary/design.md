## Context

herdr 0.9.5 shows plugin data in the sidebar only through display-only tokens: `herdr workspace report-metadata <ws> --source <id> --token name=value` (or `--clear-token name`), rendered wherever the user's `[ui.sidebar.spaces] rows` name `$name`. Plugins cannot edit the user's config; styling, including `rules` on the token value, lives there. herdr runs `[[startup]]` once after restore and `[[events]]` hooks once per event, each as a separate process with `HERDR_PLUGIN_STATE_DIR` set.

herdr-openspec (Python) solved the same problem: whole-world sync on startup and on `pane.agent_status_changed`, `pane.focused`, `workspace.created`, `pane.created`, a `flock` plus pending file to coalesce bursts, and `herdr pane list` (`foreground_cwd`, focused pane wins) to map workspaces to Projects. Its parsing is replaced by ADR 0002; its process shape works and is kept.

`cmd/kanban.loader(dir, proj, run)` already loads a Project through `openspec.Load` with cached planning reuse and writes the cache; with a warm cache a load is one `openspec list` (~0.65 s) plus a parallel `openspec --version`.

## Goals / Non-Goals

**Goals:**
- Sidebar up to date within a few seconds of an agent turn ending.
- A pass costs one `openspec list` per distinct Project, run in parallel.
- Pure, table-tested formatting; herdr and openspec behind fakes.

**Non-Goals:**
- Showing Running agents, milestones or ages in the sidebar.
- Refreshing on file changes without a herdr event (no watcher, no daemon).
- Making the row width configurable; 28 characters matches `sidebar_width = 34`.
- Clicking a Space row to open the Board (ADR 0001).

## Decisions

### Tokens and source
Source id `openspec-kanban`; tokens `kanban_change` (Most recently modified Change) and `kanban` (Phase overview). Short names that cannot collide with herdr-openspec's `openspec`/`openspec_change`, so both plugins' rows can coexist while the user migrates. Values are plain text; styling stays in the config (herdr's rule).

### Row content

```
▰▰▰▱▱▱ 3/6 m1-board-view
3 draft · 2 ready · 1 review
```

- `kanban_change`: six-cell bar, `done/total`, name fitted with the Board's middle-ellipsis rule to the remaining width of 28. The bar uses floor (`done*6/total`), so it is full only at `done == total` (Review); that makes `starts_with = "▰▰▰▰▰▰"` an exact Review rule without a marker character. No tasks → `no tasks <name>`.
- `kanban`: counts for Draft, Ready, In Progress (`wip`), Review in Board order, zeros omitted, ` · ` separator. If longer than 28, drop Phases from the left until it fits: Draft is the least urgent and the Board shows it anyway. `in progress` would push Review out too often; `wip` is the one abbreviation, explained in the README.
- Most recently modified Change: the open Change with the newest `Modified`, ties by name; Phases come from `phase.Of`.

The Board's `progressBar` stays in `board`; the sidebar has its own six-cell bar (different glyphs and rounding). The sidebar copies the 10-line `fitName` instead of exporting it from `board` or adding a shared package.

### Workspace → Project
`herdr workspace list` gives every workspace id (needed to clear tokens for workspaces without panes in a Project). `herdr pane list` gives `workspace_id`, `foreground_cwd` (fallback `cwd`) and `focused`; `project.Find` maps a directory to its Project. Per workspace the focused pane's Project wins, otherwise the first listed pane with a Project. herdr marks only the globally focused pane, so for other workspaces "first listed" decides — same as herdr-openspec; good enough.

### Coalescing
`kanban sync` takes `flock(LOCK_EX|LOCK_NB)` on `<state>/sync.lock` (stdlib `syscall.Flock`, macOS and Linux only, as the manifest declares). If the lock is taken it touches `<state>/sync.pending` and exits 0. The holder loops: remove `pending`, run one pass, repeat while `pending` exists. The lock dies with the process, so a crash never leaves a stale lock. `<state>` is `cache.Dir()`; cache files are `<hash>.json`, no clash. Alternatives: a long-running daemon (herdr does not supervise startup hooks), debounce by sleeping (adds latency to every refresh).

### Events
`[[startup]]` plus `[[events]]` for `pane.agent_status_changed` (an agent turn ending is when tasks get ticked), `pane.focused` (user moves between Projects; also catches edits made outside agents), `pane.created` and `workspace.created` (new workspace or pane may enter a Project). Same set as herdr-openspec, proven on 0.9.5. Focus events are frequent, but coalescing caps them at one pass in flight plus one queued.

### One pass
1. List workspaces and panes; on error return it (nothing reported).
2. Map workspaces to Projects; collect distinct Projects.
3. Load each Project once through the existing `loader` (cache read, `openspec.Load`, cache write), at most 4 Projects at a time. Each load may itself start up to 8 `status` calls, but with a warm cache it is `list` + `--version`; 4 keeps the worst case (cold caches) at 32 Node processes.
4. Per workspace: no Project → `--clear-token kanban --clear-token kanban_change`; load failed → no call, error to stderr (herdr's plugin log); otherwise one `report-metadata` call with both tokens, or `--token kanban=… --clear-token kanban_change` when there is no open Change.

`internal/sidebar` exposes `Summary(changes) (change, phases string)` (pure) and `Sync(run herdr.Runner, load func(project string) ([]openspec.Change, error)) error`; `Coalesce(dir string, pass func() error) error` wraps it. `cmd/kanban` wires `herdr.Exec` and a `load` built from `loader(cache.Dir(), proj, openspecRunner(proj))`. Package name `sidebar`, not `sync`, to avoid shadowing the stdlib package.

### Failure behaviour
Keep, not clear, on a failed Project load: a transient CLI error (Node update, lock contention) should not blank the sidebar, and herdr drops tokens on restart anyway, so stale values cannot outlive the session. If the CLI is missing for good, the startup pass fails and the rows stay empty after a restart, which is honest. herdr unreachable → exit 1, nothing touched. No `--ttl-ms` and no `--seq`: only one pass runs at a time, so reports cannot overtake each other.

### Config snippet

```toml
[ui]
sidebar_width = 34

[ui.sidebar.spaces]
rows = [
  ["state_icon", "workspace"],
  ["branch", "git_status"],
  [{ token = "$kanban_change", fg = "#89b4fa", rules = [{ starts_with = "▰▰▰▰▰▰", fg = "#a6e3a1" }] }],
  [{ token = "$kanban", dim = true }],
]
```

## Risks / Trade-offs

- [Edits without a herdr event (editor in a plain shell, `git pull`) stay invisible until the next focus change] → acceptable; focus changes are frequent.
- [Many Projects with cold caches make the first pass slow (~1.5 s per Project, 4 at a time)] → runs in the background after startup; later passes are warm.
- [Board and sync write the same cache file at once] → atomic rename, both writes are valid loads (change-cache spec).
- [Unfocused workspace with panes in two Projects shows the first listed one] → rare; the focused workspace is always right.
- [Users keep the old `$openspec` rows] → they render nothing once herdr-openspec is disabled; README says to replace them.
- [Glyph width: `▰▱…·` are single-width in common fonts but East Asian Ambiguous] → same glyphs herdr-openspec used in this sidebar without issues.

## Migration Plan

1. Install or rebuild the plugin; the startup hook runs at the next herdr server start, the event hooks right away (or run `bin/kanban sync` by hand from a herdr pane).
2. Replace the `$openspec`/`$openspec_change` rows in `~/.config/herdr/config.toml` with the snippet; `herdr server reload-config`.
3. Rollback: `herdr plugin enable openspec` and restore its rows; the tokens have different names, so nothing conflicts.
