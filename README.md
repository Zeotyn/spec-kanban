# spec-kanban

A [herdr](https://herdr.dev) plugin that shows a project's [OpenSpec](https://github.com/Fission-AI/OpenSpec) changes as a kanban board.

## Requirements

- herdr ≥ 0.9.3
- macOS or Linux
- Go, only to install a commit that is not a release

## Install

Install a release (see [Releases](https://github.com/zeotyn/spec-kanban/releases)); this downloads its checksum-verified binary and needs no Go:

```sh
herdr plugin install zeotyn/spec-kanban --ref v<version>
```

`herdr plugin update` keeps a release pinned; to upgrade, install the new tag. Installing without `--ref` takes `main`, which builds from source with Go unless `main` is exactly a release.

Then copy the keybinding from [`herdr-config.toml`](herdr-config.toml) into `~/.config/herdr/config.toml` and run `herdr server reload-config`. `prefix+o` opens the board for the project of the focused pane; `q` or `Esc` closes it.

## Keys

| Key | Board | Details |
| --- | --- | --- |
| `h` `l` / `j` `k` | column / Card | – / scroll |
| `gg` `G`, `ctrl+d` `ctrl+u` | first/last Card, half page | top/end, half page |
| `1`–`5` / `1`–`4`, `Tab` | jump to column | switch tab |
| `Enter` | open details | – |
| `e` | edit `tasks.md` | edit the shown document |
| `r` | reload | – |
| `s` | start the Change in a new agent tab in its worktree | – |
| `/` | filter Cards by name | – |
| `n` `N` | next/previous match | – |
| `?` | key help | key help |
| `q` `Esc` | close (`Esc` first clears a filter) | back to the Board |

`?` shows all keys and the version; any key closes the help.

`e` opens `$EDITOR` (`vi` when unset) inside the popup. GUI editors need their wait flag, for example `EDITOR="code -w"`, or the Board reloads before you have saved.

`/` opens a prompt in the footer; every key typed there is text and filters the columns at once, ignoring case. `Enter` keeps the filter, `Esc` clears it, `Backspace` on an empty prompt closes it. `m1` (or `M1`) matches only the Changes with milestone badge M1, not `m10-…` or `fix-m1-…`. The filter shows in the header and stays across reloads.

`s` works on Ready and In Progress Cards. It opens a new tab named after the Change in a git worktree of the project and closes the Board at once; the tab then gets focus, the agent starts in it and gets the prompt (`/opsx:apply <change>` unless configured, see [Settings](#settings)). If that Change already has an agent, `s` focuses it instead. Errors before the Board closes show in its footer, later ones (agent not ready, prompt not sent) as a herdr notification. On a Review Card, `s` reopens the Change's recorded session (see [Handoff](#handoff)) in its worktree without a prompt, or focuses its running agent; without a recorded session, or for an agent kind that cannot resume, the footer says why not, as it does on Draft and Archived Cards. A green `●` before a Card's name marks a Change with a Running agent. From a shell, `kanban start <change>` does the same without the Phase check.

The worktree is a checkout on a branch named like the Change, created by herdr at `<dir>/<repo>/<change>`, where `<dir>` is `[worktrees] directory` in herdr's config (default `~/.herdr/worktrees`). New branches start from the project's default branch (`origin/HEAD`, else `main`, `master` or `develop`, else the current branch), or from `base` when set. An existing checkout on that branch is reused. The Change must be committed on the start branch before you press `s`; otherwise Start reports that the Change is not on the base. The Board and the sidebar read the Change's Progress from its worktree. The plugin never removes a worktree or deletes a branch; after a [Handoff](#handoff) the Main agent removes the worktree. Without a git repository, `s` opens the tab in the project.

### Handoff

After `s` the plugin watches the agent in the background. Each time it stops after working, the plugin checks the Change's Progress (in its worktree):

- **Finished** (all tasks completed, or the agent archived the Change, in the project or in its worktree) and not waiting for input: the agent's pane closes, and with it its tab when nothing else is open there. The agent in the pane you opened the Board from, the **Main agent** (any kind), gets a prompt like `m1-board-view is finished (7/7 tasks). Please review it and archive it with /opsx:archive m1-board-view.` For a Change in a worktree it asks instead to review it, merge branch `m1-board-view` into the base, archive it, then run `git worktree remove <path>`, never with `--force`, keeping the branch, and to stop if the merge or the removal does not go through cleanly. The Main agent merges and removes; `git worktree remove` refuses a checkout with uncommitted work, so nothing is lost. The prompt also names the command that reopens the session (for example `claude --resume <id>`) when the agent kind supports it.
- **Not finished** (tasks left) or waiting for input: the tab stays open, herdr shows a notification with the Progress, and watching goes on.

When the Main agent cannot take the prompt (the Board was opened from a shell pane, the pane is gone, or the Main agent waits at a dialog), herdr shows the same text as a notification `spec-kanban: <change> finished`. Watching ends with the Handoff or when the agent's tab is closed. `kanban start` from a shell is not watched.

The agent kind and session herdr reports are recorded per project and Change under `sessions/` in the plugin's state directory (`$HERDR_PLUGIN_STATE_DIR`, else the user cache directory's `spec-kanban`). Sessions of `claude`, `copilot`, `codex`, `opencode`, `cursor` and `gemini` can be reopened with `s` on the Review Card; the worktree is reopened, or recreated from its branch after it was removed.

## Settings

Settings live in `config.json` in the plugin's config directory, which `herdr plugin config-dir spec-kanban` prints. The file is optional; every key is too:

```json
{
  "agent": "codex",
  "prompt": "/opsx:apply {change} with TDD and review loops",
  "hide": ["Archived"]
}
```

| Key | Default | Meaning |
| --- | --- | --- |
| `agent` | unset | Agent kind `s` starts, one of `herdr agent start --help`'s `--kind` values (`claude`, `codex`, `gemini`, `opencode`, …) |
| `prompt` | `/opsx:apply {change}` | Text sent to the agent; every `{change}` becomes the Change name, and it must contain one |
| `worktree` | `true` | `false` opens the agent's tab in the project instead of a git worktree |
| `base` | detected | Non-empty branch or ref new worktrees start from; overrides the detected default branch |
| `hide` | `[]` | Phases whose columns the Board leaves out, ignoring case; at least one must stay |

The kind `s` starts is `agent` when set, else the agent running in the pane the Board was opened from (for `kanban start`, the pane it runs in), else `claude`. OpenSpec names its commands per tool, so with a kind other than `claude` set `prompt` to the command OpenSpec installed for that tool (`openspec init --tools <tool>`).

The file is read when the Board opens or `kanban start` runs. If it is broken (invalid JSON, an unknown key, a value breaking the rules above), all settings fall back to their defaults: the Board shows the problem in its footer until the next key, and `kanban start` prints it to stderr.

## Sidebar

The plugin writes two lines into each workspace's Space rows in the sidebar, for the project its panes work in (the focused pane's project wins):

```
▰▰▰▱▱▱ 3/6 m1-board-view
3 draft · 2 ready · 1 review
```

The first line is the most recently modified change with its progress; it turns bold once all its tasks are done (the bar is full only then); both lines use the terminal's own colours, since herdr token rows only take hex. The second counts the open changes per column; `wip` stands for In Progress. When the line gets too long, Draft is dropped first. Workspaces outside a project show nothing.

herdr refreshes the lines on startup, when an agent's status changes and when a pane gets focus or a pane or workspace is created. To show them, edit `~/.config/herdr/config.toml` as in [`herdr-config.toml`](herdr-config.toml): add `sidebar_width = 34` and `sidebar_max_width = 36` to your existing `[ui]` table (a second `[ui]` header is a TOML error), copy the `[ui.sidebar.spaces]` table, and run `herdr server reload-config`. If you used herdr-openspec, replace its `$openspec` and `$openspec_change` rows; they stay empty once that plugin is disabled.

## Development

`herdr plugin link` does not run the build step, so build the binary yourself after every change:

```sh
herdr plugin link ~/code/spec-kanban
go build -o bin/kanban ./cmd/kanban
herdr plugin action invoke spec-kanban-open
herdr plugin log list
```

Tests: `go test ./...` (rewrite golden views with `go test ./... -update`).

To release: bump `version` in `herdr-plugin.toml`, add a `## <version> - <YYYY-MM-DD>` section to [`CHANGELOG.md`](CHANGELOG.md), commit on `main`, then tag and push:

```sh
git tag v<version> && git push origin v<version>
```

The release workflow refuses a tag that does not match the manifest, builds the binaries and publishes them with the changelog section as notes.

## Credits

- [herdr-openspec](https://github.com/FlorisKr/herdr-openspec) by FlorisKr: the original OpenSpec plugin for herdr. Its board popup and sidebar rows were the starting point for this plugin.
- [magentic](https://github.com/MarDonhauser/magentic) by MarDonhauser: the Phase derivation and the starting of agents from the board follow its approach.
