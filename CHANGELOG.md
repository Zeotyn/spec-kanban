# Changelog

## Unreleased

- Waves: mark Ready Cards with `space` and press `s` to start them as a wave on branch `wave-<N>`. A coordinator agent starts the members with `kanban wave start`, merges and archives each on the wave branch, and hands the wave to the Main agent once with `kanban wave finish`; `kanban wave drop` and `kanban wave escalate` cover the rest. Members show `≋<N>` and do not archive themselves. New config key `wave` with `size` (default 5). `Esc` clears marks before it closes the Board.
- Attention marker: a red `⚑` before a Card's name, in place of the green `●`, and a `⚑<n>` prefix on the sidebar's `kanban` line show the Changes that wait for you after an Escalation, a Stall or a Handoff no Main agent took. It goes away when the agent works again or its tab is closed, or, after a Handoff, when the Change is archived, started or resumed.
- Removed: the `finished`, `needs you` and `stalled` notifications; the attention marker replaces them. herdr notifications now only report failures (start, run and resume).
- Escalation: a worker that cannot continue without a human runs `kanban escalate <change> <reason>` and the Main agent gets a prompt at once. A Start from the Board appends a sentence to its prompt that tells the agent how.
- Stall: a watched agent that stopped unfinished and does not work again within 10 minutes, without having escalated, gets one prompt to the Main agent.
- Resume from the Board is watched like a Start: Handoff, Stall and Escalation apply. The plugin evaluates a resumed agent only after its first turn.
- Removed: the "stopped at x/y tasks" and "waits for input" notifications. Stall replaces the first; herdr's sidebar shows a blocked agent.
- Mouse: click a Card to select it and again to open its details, scroll a column or the details with the wheel, click `◀`/`▶` in a narrow popup, tabs in the details and the actions in both footers. A right click on a Card, or in its details, opens the Card menu, whose entry starts or resumes the Change like `s`. Hold `Shift` (or your terminal's modifier) while dragging to select text.
- Verify and Archive: `v` and `a` on a Review Card, or the Card menu's new entries, send `/opsx:verify <change>` or `/opsx:archive <change>` to the Change's idle agent, or start a new one in its worktree; without a watcher or Handoff. Spec Kit has neither.
- Ready is a queue: fixes first, then milestones by number (`m9` before `m10`), then the rest in natural name order (`c2` before `c10`), so editing a Ready Change no longer moves its Card. Draft, In Progress and Review keep newest first.
- Fixed: a Change or Feature whose agent name would exceed herdr's 32 characters starts with a shortened agent name, like `update-a-change-from-its-302ef9`, instead of being refused. Names that fit keep their agent name.
- Fixed: starting an agent in a new tab whose shell rc asks a question (like the oh-my-zsh update prompt) typed `laude`. The plugin now waits for the shell prompt with a probe line before every start, resume, verify and archive, and a start that still fails shows the pane's last lines in the notification.

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
