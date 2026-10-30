# Changelog

## Unreleased

- Spec Kit Source: a Project with `.specify` and `specs` shows its Features on the Board, in the details and in the sidebar, and `s` starts the implement command of the agent kind in a worktree after writing `.specify/feature.json`; the Handoff asks to delete that file before removing the worktree. A Feature whose tasks are done and whose branch is merged is Archived. A Project with both tools is read as OpenSpec unless `source` says otherwise. The no-project notice and the `kanban start` error name both tools.
- New config keys `source` (`openspec` or `speckit`) and `instructions` (text appended to the Source's command; `prompt` still replaces everything and wins).
- Renamed from openspec-kanban to spec-kanban, since more spec tools than OpenSpec are planned: the repository is github.com/zeotyn/spec-kanban, the plugin id is `spec-kanban` and the action `spec-kanban-open`. Re-link or re-install the plugin, update the keybinding's `command`, and move `config.json` from `plugins/config/openspec-kanban` to `plugins/config/spec-kanban`.

## 0.2.0 - 2026-10-30

- Worktrees: `s` starts the agent in its own git worktree on a branch named after the Change, based on the repository's default branch, so work in the main checkout goes on in parallel; the Board and the sidebar read Progress from the worktree. The `worktree` and `base` settings turn this off or pick the base.
- Handoff: once an agent started from the Board has finished all tasks, its pane closes and the agent in the pane the Board was opened from is asked to review, merge, archive and remove the worktree; otherwise a herdr notification shows the Progress.
- Resume: `s` on a Review Card reopens the recorded session of claude, copilot, codex, opencode, cursor or gemini.

## 0.1.0 - 2026-10-30

First release.

- Board: a popup with the workspace's OpenSpec changes as Cards in Phase columns (Draft, Ready, In Progress, Review, Archived), vim navigation, and an instant open from a validated cache.
- Details: a Change's documents in tabs, coloured with the terminal's own palette; `e` edits the shown document in `$EDITOR`.
- Sidebar: the project's latest Change with its progress and the open Changes per column in herdr's Space rows.
- Start: `s` on a Ready or In Progress Card starts an agent in a new tab with `/opsx:apply <change>`; a `●` marks Changes with a Running agent; `kanban start <change>` does the same from a shell.
- Filter: `/` filters Cards by name, `n` and `N` walk the matches.
- Settings: `config.json` sets the agent kind, the prompt and hidden Phases; `?` shows the keys.
- Install: a release installs without Go by downloading its verified binary; `kanban --version` prints the version.
