# Changelog

## 0.1.0 - 2026-10-30

First release.

- Board: a popup with the workspace's OpenSpec changes as Cards in Phase columns (Draft, Ready, In Progress, Review, Archived), vim navigation, and an instant open from a validated cache.
- Details: a Change's documents in tabs, coloured with the terminal's own palette; `e` edits the shown document in `$EDITOR`.
- Sidebar: the project's latest Change with its progress and the open Changes per column in herdr's Space rows.
- Start: `s` on a Ready or In Progress Card starts an agent in a new tab with `/opsx:apply <change>`; a `●` marks Changes with a Running agent; `kanban start <change>` does the same from a shell.
- Filter: `/` filters Cards by name, `n` and `N` walk the matches.
- Settings: `config.json` sets the agent kind, the prompt and hidden Phases; `?` shows the keys.
- Install: a release installs without Go by downloading its verified binary; `kanban --version` prints the version.
