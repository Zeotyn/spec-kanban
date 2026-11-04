# Changelog

## Unreleased

- Mouse: click a Card to select it and again to open its details, scroll a column or the details with the wheel, click `◀`/`▶` in a narrow popup, tabs in the details and the actions in both footers. A right click on a Card, or in its details, opens the Card menu, whose entry starts or resumes the Change like `s`. Hold `Shift` (or your terminal's modifier) while dragging to select text.
- Verify and Archive: `v` and `a` on a Review Card, or the Card menu's new entries, send `/opsx:verify <change>` or `/opsx:archive <change>` to the Change's idle agent, or start a new one in its worktree; without a watcher or Handoff. Spec Kit has neither.
- Fixed: a Change or Feature whose agent name would exceed herdr's 32 characters starts with a shortened agent name, like `update-a-change-from-its-302ef9`, instead of being refused. Names that fit keep their agent name.

## 0.3.0 - 2026-11-02

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
