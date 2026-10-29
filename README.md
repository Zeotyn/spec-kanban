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

## Development

`herdr plugin link` does not run the build step, so build the binary yourself after every change:

```sh
herdr plugin link ~/code/openspec-kanban
go build -o bin/kanban ./cmd/kanban
herdr plugin action invoke openspec-kanban-open
herdr plugin log list
```

Tests: `go test ./...` (rewrite golden views with `go test ./... -update`).
