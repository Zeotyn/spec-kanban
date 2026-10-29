# openspec-kanban

A [herdr](https://herdr.dev) plugin that shows a project's [OpenSpec](https://github.com/Fission-AI/OpenSpec) changes as a kanban board.

## Requirements

- herdr ≥ 0.9.3
- Go (the plugin is built on install)
- macOS or Linux

## Install

```sh
herdr plugin install zeotyn/openspec-kanban
```

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
| `q` `Esc` | close | back to the Board |

`e` opens `$EDITOR` (`vi` when unset) inside the popup. GUI editors need their wait flag, for example `EDITOR="code -w"`, or the Board reloads before you have saved.

## Sidebar

The plugin writes two lines into each workspace's Space rows in the sidebar, for the project its panes work in (the focused pane's project wins):

```
▰▰▰▱▱▱ 3/6 m1-board-view
3 draft · 2 ready · 1 review
```

The first line is the most recently modified change with its progress; it turns green once all its tasks are done (the bar is full only then). The second counts the open changes per column; `wip` stands for In Progress. When the line gets too long, Draft is dropped first. Workspaces outside a project show nothing.

herdr refreshes the lines on startup, when an agent's status changes and when a pane gets focus or a pane or workspace is created. Copy the `[ui]` and `[ui.sidebar.spaces]` part of [`herdr-config.toml`](herdr-config.toml) into `~/.config/herdr/config.toml` and run `herdr server reload-config`. If you used herdr-openspec, replace its `$openspec` and `$openspec_change` rows; they stay empty once that plugin is disabled.

## Development

`herdr plugin link` does not run the build step, so build the binary yourself after every change:

```sh
herdr plugin link ~/code/openspec-kanban
go build -o bin/kanban ./cmd/kanban
herdr plugin action invoke openspec-kanban-open
herdr plugin log list
```

Tests: `go test ./...` (rewrite golden views with `go test ./... -update`).
