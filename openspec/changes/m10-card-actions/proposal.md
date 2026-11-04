## Why

A Change in Review has two steps left, verify and archive. The Board can start and resume a Change, but these two steps still need someone to type `/opsx:verify <change>` or `/opsx:archive <change>` into an agent by hand. This hurts most when no Handoff happened: with `worktree` off, when the agent ended without a clean finish, or when the user wants to run the step themselves. The Board should start both steps the same way it starts a Change.

## What Changes

- Two new Card actions, **Verify** (`v`) and **Archive** (`a`), on the Board and in the Card menu. In the menu they come below Start/Resume, each with its key.
- **When they are allowed:** both work only on a Review Card whose Change has no busy agent.
  - On any other Phase they do nothing except show a notice.
  - When the Change's agent is not idle (working, blocked, …) they also only show a notice.
  - In the Card menu an unavailable entry is faint, like Start on a Draft Card.
- **Where the command runs:**
  - The Change's agent exists and is idle: the command goes to it as a prompt and its tab is focused. No new tab is opened.
  - No agent runs for the Change: a new agent opens in the Change's checkout, the same place Start uses, and receives the command.
  - Either way the Board closes. No watcher follows and no Handoff is sent. The user reads the result in the agent's tab.
- **What command is sent:** the Source's own command with the Change name: `/opsx:verify <change>` and `/opsx:archive <change>` for OpenSpec. The configured `prompt` and `instructions` apply to Start only, not to these commands.
- **Spec Kit:** It has no verify or archive command for a finished Feature, so its Cards show neither entry, and `v` and `a` explain that.
- **Key help:** `v` and `a` appear under Board. The footer stays as it is; mouse users reach both actions through the Card menu.
- **Smoke tests:** From this change on, the user smoke tests in `tasks.md` are a plain list without checkboxes. Agents cannot tick them, and they do not hold a finished Change out of Review.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `board`: new requirement for the `v` and `a` keys. The Card menu requirement changes from one entry to Start/Resume plus Verify and Archive.
- `agent-start`: new requirement for how Verify and Archive reach an idle agent or start a new one, without a watcher or Handoff.
- `change-source`: new requirement that a Source names its verify and archive commands, or none.

## Impact

- **Code:**
  - `internal/work`: the Source interface gets the verify and archive commands.
  - `internal/source/openspec`, `internal/source/speckit`: their commands, or none.
  - `internal/start`: run a command for a Change, with the agent status check and no watcher.
  - `internal/board`: keys, Card menu entries, notices and help.
  - `cmd/kanban`: wiring and a new `kanban` subcommand for the detached launch without a watcher.
- **Docs:** README, CONTEXT.md (Verify, Archive as Card actions), CHANGELOG.
- No new dependencies.
