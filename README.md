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

## Development

`herdr plugin link` does not run the build step, so build the binary yourself after every change:

```sh
herdr plugin link ~/code/openspec-kanban
go build -o bin/kanban ./cmd/kanban
herdr plugin action invoke openspec-kanban-open
herdr plugin log list
```

Tests: `go test ./...` (rewrite golden views with `go test ./... -update`).
