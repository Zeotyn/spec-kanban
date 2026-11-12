# agent-start Specification

## Purpose

Defines how a Change is handed to a new coding-agent session in its own herdr tab, so implementation can begin from the Board with a single action.
## Requirements
### Requirement: Start opens an agent session for a Change
Starting a Change SHALL open a new herdr tab whose working directory is the Change's checkout (see "Start works in the Change's checkout"; the Project when worktrees are off) and that is labelled or whose workspace is labelled with the Change name, let the Project's Source prepare that checkout (OpenSpec needs nothing; for Spec Kit see the speckit-source spec), focus it, start an agent of the chosen kind in that tab as an agent that herdr tracks, and submit the prompt to it. The prompt SHALL be the configured `prompt` with `{change}` replaced by the Change name when one is set; otherwise the Source's command for that kind (`/opsx:apply <change>` for OpenSpec, Spec Kit's implement command for Spec Kit), followed by a space and the configured `instructions` when they are set. A Start from the Board SHALL append, after a space, the escalation sentence `If you cannot continue without a human, run <kanban> escalate <change> "<what you need>" and end your turn.`, also when a `prompt` is configured; `<kanban>` SHALL be the absolute path of the running plugin binary, in single quotes when it holds characters other than letters, digits, `/`, `.`, `_` and `-` (see "Escalation"). A Start through `kanban start` SHALL send no escalation sentence. When the preparation fails, Start SHALL report it and start no agent. The herdr agent SHALL be named like the Change, prefixed with `f-` when the name does not start with a lowercase letter (herdr agent names must, and have at most 32 characters). When that name has more than 32 characters, the agent SHALL instead be named with its first 25 characters, without a trailing `-`, followed by `-` and the first 6 hexadecimal characters of the SHA-256 of the Change name; the tab, workspace and worktree branch SHALL keep the full Change name. The agent name SHALL depend on the Change name alone, and everything that finds the agent again (Running badge, Watch, resume, Verify, Archive) SHALL look for the agent name of the Change, not derive a Change from an agent name. The kind SHALL be the configured `agent`; without one, the kind herdr reports for the agent running in the invoking pane (the pane the Board was opened from, or the pane `kanban start` runs in); `claude` when that pane runs no agent, herdr cannot tell, or the reported kind is not one Start can launch.

#### Scenario: Start a Ready Change
- **WHEN** the user runs `kanban start m1-board-view` inside a Project that contains that open Change, from a plain shell, and nothing is configured
- **THEN** a new workspace labelled `m1-board-view` opens on the Change's worktree and its tab is focused before the agent is ready, Claude runs in it, and its first prompt is `/opsx:apply m1-board-view`

#### Scenario: Start with a configured prompt
- **WHEN** the configured prompt is `/opsx:apply {change} with ponytail, TDD, subagents and review loops` and the user presses `s` on Card `m1-board-view`
- **THEN** the agent's first prompt in the new tab is `/opsx:apply m1-board-view with ponytail, TDD, subagents and review loops`, followed by a space and the escalation sentence

#### Scenario: Start with configured instructions
- **WHEN** the configured `instructions` are `with ponytail, TDD, subagents and review loops`, no `prompt` is set, and the user presses `s` on Card `m1-board-view`
- **THEN** the agent's first prompt in the new tab is `/opsx:apply m1-board-view with ponytail, TDD, subagents and review loops`, followed by a space and the escalation sentence

#### Scenario: Escalation sentence from the Board
- **WHEN** the plugin binary is `/opt/herdr/plugins/spec-kanban/kanban`, nothing is configured, and the user presses `s` on Card `m1-board-view`
- **THEN** the agent's first prompt is `/opsx:apply m1-board-view If you cannot continue without a human, run /opt/herdr/plugins/spec-kanban/kanban escalate m1-board-view "<what you need>" and end your turn.`

#### Scenario: Binary path with a space
- **WHEN** the plugin binary is `/Users/me/Library/Application Support/herdr/kanban` and the user presses `s` on Card `m1-board-view`
- **THEN** the escalation sentence names `'/Users/me/Library/Application Support/herdr/kanban' escalate m1-board-view`

#### Scenario: Board opened from a Codex pane
- **WHEN** the Board was opened from a pane where herdr reports a `codex` agent, no `agent` is configured, and the user presses `s` on Card `m1-board-view`
- **THEN** Codex is started in the new tab and receives the configured prompt with the escalation sentence

#### Scenario: Board opened from a plain shell
- **WHEN** the Board was opened from a pane that runs no agent, no `agent` is configured, and the user presses `s`
- **THEN** Claude is started in the new tab

#### Scenario: herdr tracks the session
- **WHEN** a Change has been started
- **THEN** `herdr agent list` shows an agent for that Change, so its status appears in the herdr sidebar

#### Scenario: Start a Spec Kit Feature
- **WHEN** the user presses `s` on Ready Card `001-user-login-email` of a Spec Kit Project whose integration is `claude`, and nothing is configured
- **THEN** the Feature's worktree gets `.specify/feature.json` naming `specs/001-user-login-email` before Claude starts, and Claude's first prompt is `/speckit-implement` followed by a space and the escalation sentence for `001-user-login-email`

#### Scenario: Feature name starting with a digit
- **WHEN** the user presses `s` on Ready Card `001-user-login-email`
- **THEN** herdr's agent is named `f-001-user-login-email`, and the Card shows the Running badge while it works

#### Scenario: Name too long for herdr
- **WHEN** the user starts the Change `update-a-change-from-its-base-branch` (36 characters)
- **THEN** the tab or workspace is labelled `update-a-change-from-its-base-branch`, the worktree branch is `update-a-change-from-its-base-branch`, and herdr's agent is named `update-a-change-from-its-302ef9`

#### Scenario: Prefixed Feature name too long for herdr
- **WHEN** the user starts the Spec Kit Feature `001-a-feature-name-far-too-long-for-herdr`
- **THEN** herdr's agent is named `f-001-a-feature-name-far-8b5dbb`

#### Scenario: Name that fits stays unchanged
- **WHEN** the user starts the Change `add-a-change-of-thirty-two-chars` (32 characters)
- **THEN** herdr's agent is named `add-a-change-of-thirty-two-chars`, as before this change

#### Scenario: Same long name, same agent
- **WHEN** the user starts `update-a-change-from-its-base-branch` while its agent `update-a-change-from-its-302ef9` is still running
- **THEN** no second tab or agent is created and that agent's tab is focused

#### Scenario: A shortened agent is found again
- **WHEN** herdr reports an agent named `update-a-change-from-its-302ef9` working in the Project
- **THEN** the Card `update-a-change-from-its-base-branch` shows the Running badge, and Resume, Verify and Archive on that Card reach that agent

#### Scenario: Preparation fails
- **WHEN** the Source cannot prepare the Change's checkout
- **THEN** no agent is started, no prompt is sent, and Start reports the failure

### Requirement: Start does not duplicate a running session
When an agent for the Change is already running, Start SHALL focus that agent instead of creating another tab or agent. The agent SHALL count as running while anything but the shell holds the foreground of its pane, whatever its kind and whatever kind Start would launch now. Only when the shell is idle again SHALL Start start the chosen kind again in that tab. A program other than the agent in the foreground SHALL be treated like a running agent: its tab is focused and nothing is started.

#### Scenario: Start twice
- **WHEN** the user starts `m1-board-view` and then starts it again while the first agent is still running
- **THEN** no second tab is created and the existing agent's tab is focused

#### Scenario: Running agent of another kind
- **WHEN** a Codex agent `m1-board-view` is running and the user starts `m1-board-view` from a Board opened in a Claude pane
- **THEN** no second tab is created and the Codex agent's tab is focused

#### Scenario: Start after the agent exited
- **WHEN** the agent for `m1-board-view` has exited but its tab is still open, and the user starts the Change again
- **THEN** no second tab is created; the tab is focused, then the chosen agent kind is started again in it and receives the configured prompt

#### Scenario: Agent shown as node
- **WHEN** a Gemini agent `m1-board-view` is running and its process shows as `node`
- **THEN** no second tab is created and the Gemini agent's tab is focused

#### Scenario: Another program in the tab
- **WHEN** the agent for `m1-board-view` has exited and `vim` runs in its tab
- **THEN** Start creates no tab, starts no agent, and focuses that tab

### Requirement: Start rejects unknown Changes
Start SHALL check that the Change exists as an open Change of the Project before calling herdr, and SHALL exit with a non-zero status and a message naming the Change otherwise.

#### Scenario: Typo in the Change name
- **WHEN** the user runs `kanban start m1-bord-view` and no such open Change exists
- **THEN** no tab is created and the command prints that `m1-bord-view` is not an open Change of the Project

#### Scenario: Archived Change
- **WHEN** the user runs `kanban start` with the name of a Change that exists only in the archive
- **THEN** no tab is created and the command prints that the Change is not open

### Requirement: Start reports herdr failures
When any herdr step fails, Start SHALL stop, report which step failed with herdr's error, and leave no agent half-started without saying so. When the shell did not reach its prompt or the agent did not become ready, the report SHALL also show the last lines of the pane (up to 8 non-empty lines of recent output, without escape sequences), so the user sees what the shell was doing; when the pane cannot be read, the report says so instead. A failure in the part that runs after the Board closed SHALL be reported as a herdr notification.

#### Scenario: Claude does not become ready
- **WHEN** the agent does not reach readiness within herdr's start timeout
- **THEN** the command exits non-zero, names the tab it created, says the prompt was not sent, and shows the pane's last lines

#### Scenario: Claude does not become ready after the Board closed
- **WHEN** a Start from the Board has closed the Board and the agent does not reach readiness
- **THEN** herdr shows the notification `spec-kanban: start failed` with that error and the pane's last lines

#### Scenario: Shell did not reach its prompt after the Board closed
- **WHEN** a Start from the Board has closed the Board and the shell of the new tab answers no probe within 60 seconds
- **THEN** herdr shows the notification `spec-kanban: start failed` saying the shell did not reach its prompt in tab `m1-board-view`, with the pane's last lines

#### Scenario: Pane cannot be read for the report
- **WHEN** the agent does not become ready and `pane read` fails as well
- **THEN** the report carries the start error and says the pane could not be read, with that error

### Requirement: Start does not take another Project's agent
herdr agent names are unique across all Projects. When an agent named after the Change exists but works in a directory outside the Project and outside every checkout of the Project's git repository, Start SHALL neither focus nor restart it, create no tab, and fail with a message naming the Change and the other agent's directory.

#### Scenario: Equally named Change in two Projects
- **WHEN** an agent `m1-board-view` works in Project `alpha` and the user starts `m1-board-view` in Project `beta`
- **THEN** no tab is created, the `alpha` agent is not focused, and Start reports that `m1-board-view` already runs in `alpha`

#### Scenario: Agent in the Change's worktree
- **WHEN** an agent `m1-board-view` is running in a linked worktree of the Project's repository and the user starts `m1-board-view`
- **THEN** no tab or worktree is created and that agent's tab is focused

### Requirement: Start works in the Change's checkout
When the `worktree` setting is on (the default) and the Project is the top directory of a git checkout, Start SHALL run the agent in the Change's checkout: the checkout of the Project's repository whose branch is named exactly like the Change. When such a checkout exists, Start SHALL reuse it: a linked worktree that is not open in herdr SHALL be opened as a herdr workspace whose root pane becomes the agent's tab; a checkout that is already open, or the Project's own checkout when it is on that branch, SHALL get a new tab labelled with the Change. Otherwise Start SHALL have herdr create a linked worktree on a new branch named like the Change from the base, at herdr's configured worktree location, opened as a new herdr workspace labelled with the Change; an existing local branch of that name without a checkout SHALL be checked out instead of created. Start SHALL never remove a worktree or delete a branch. When `worktree` is off, or the Project is not the top directory of a git checkout (outside git, or a subdirectory of a checkout), Start SHALL work in the Project as without this setting.

#### Scenario: First Start of a Change
- **WHEN** the Project `spec-kanban` is a git repository with default branch `main`, no branch `m1-board-view` exists, and the user presses `s` on Card `m1-board-view`
- **THEN** herdr creates a worktree of the repository on new branch `m1-board-view` from `main`, opens it as a workspace labelled `m1-board-view`, and the agent starts in its root pane and receives the configured prompt

#### Scenario: Rerun with a closed worktree
- **WHEN** a linked worktree on branch `m1-board-view` exists, herdr has no workspace open for it, and no agent `m1-board-view` exists
- **THEN** no worktree or branch is created; the existing worktree is opened as a workspace and the agent starts in it

#### Scenario: Worktree workspace still open
- **WHEN** the worktree on branch `m1-board-view` is open as a herdr workspace and its agent has gone
- **THEN** a new tab labelled `m1-board-view` opens in that worktree and the agent starts in it

#### Scenario: Branch exists without a checkout
- **WHEN** a local branch `m1-board-view` exists but no checkout has it
- **THEN** herdr creates a worktree that checks out the existing branch, without creating a branch

#### Scenario: Project's own checkout is on the branch
- **WHEN** the Project's checkout itself is on branch `m1-board-view` and the user Starts `m1-board-view`
- **THEN** no worktree is created and a new tab labelled `m1-board-view` opens in the Project

#### Scenario: Worktree off
- **WHEN** the config sets `worktree` to `false` and the user Starts `m1-board-view`
- **THEN** no worktree is created and a new tab labelled `m1-board-view` opens in the Project

#### Scenario: Project outside git
- **WHEN** the Project lies in no git repository, or in a subdirectory of a git checkout, and `worktree` is on
- **THEN** Start opens a new tab in the Project as with `worktree` off

### Requirement: Base of a new worktree
A new worktree's branch SHALL start from the configured `base`. Without one, the base SHALL be the branch `origin/HEAD` points at, using the local branch of the same name when it exists and the remote branch otherwise; when `origin/HEAD` is not set, the first existing local branch of `main`, `master` and `develop`; otherwise the branch the Project's checkout is on. Before creating anything, Start SHALL check that the Change's directory exists on the ref the new worktree starts from (the existing branch named like the Change, otherwise the base), and SHALL otherwise fail with a message naming the Change and that ref.

#### Scenario: Default branch from origin
- **WHEN** `origin/HEAD` points at `origin/develop`, a local `develop` exists, and no `base` is configured
- **THEN** the new worktree's branch starts from `develop`

#### Scenario: No local copy of the default branch
- **WHEN** `origin/HEAD` points at `origin/main` and no local `main` exists
- **THEN** the new worktree's branch starts from `origin/main`

#### Scenario: No origin
- **WHEN** the repository has no `origin/HEAD` and has local branches `master` and `develop`
- **THEN** the new worktree's branch starts from `master`

#### Scenario: Configured base
- **WHEN** the config sets `base` to `release/2.x`
- **THEN** the new worktree's branch starts from `release/2.x`, whatever `origin/HEAD` says

#### Scenario: Change not committed on the base
- **WHEN** Change `m1-board-view` exists in the Project's checkout but not on the base `main`
- **THEN** no worktree, branch, tab or agent is created and Start reports that `m1-board-view` is not on `main`

### Requirement: Verify and Archive run the Source's command
Verify and Archive SHALL send the Source's verify or archive command for the Change (see "Verify and archive commands"), never the configured `prompt` or `instructions`. When a herdr agent of the Change exists in the Project or one of its checkouts and its agent runs in the foreground of its pane:
- with agent status `idle`, or `done` (its turn ended and nobody has looked yet), the command SHALL be sent to it as a prompt and its tab focused, without opening a tab;
- with any other status nothing SHALL be sent, and the caller reports the status.

Otherwise the command SHALL go to a new agent of the configured or detected kind in the Change's checkout, opened as for Start, reusing the Change's tab when its agent has exited. No watcher SHALL follow Verify or Archive: when the agent finishes, no Handoff prompt and no notification is sent. An agent of the Change's name outside the Project and its checkouts SHALL be refused as for Start.

#### Scenario: New agent for Verify
- **WHEN** no agent runs for Review Change `m1-board-view`, whose worktree exists, and Verify runs
- **THEN** a tab in that worktree is created, Claude starts in it and receives `/opsx:verify m1-board-view`, and nothing watches it

#### Scenario: Idle agent receives Archive
- **WHEN** the agent `m1-board-view` is idle in its tab and Archive runs
- **THEN** no tab is created, the agent receives `/opsx:archive m1-board-view` and its tab is focused

#### Scenario: Blocked agent
- **WHEN** the agent `m1-board-view` waits for a permission answer (status `blocked`) and Verify runs
- **THEN** nothing is sent and the status `blocked` is reported

#### Scenario: Instructions are not appended
- **WHEN** the config sets `instructions` to `with ponytail` and Verify runs for `m1-board-view`
- **THEN** the agent receives exactly `/opsx:verify m1-board-view`

### Requirement: An agent is started only at the shell prompt
Before the plugin starts an agent in a pane (Start, Resume, Verify and Archive alike, in a new tab or in the Change's reused tab), it SHALL wait until the pane's shell is at its interactive prompt. herdr reports no such state, so the plugin SHALL probe: it types into the pane, as one line, a space followed by a command that prints a marker which does not appear in the typed line itself (`echo kanban-rea""dy-<nonce>` prints `kanban-ready-<nonce>`), and waits until the pane shows that marker. A `<nonce>` SHALL be unique per probe, so that neither the typed echo nor an earlier probe's output in the same pane can satisfy the wait. A probe that is not answered within 2 seconds SHALL be followed by a new probe with a fresh nonce; the plugin SHALL give up after 60 seconds in all, and SHALL then start no agent and report the failure as in "Start reports herdr failures". A pane that cannot take the probe (herdr refuses it, for example because the pane is gone) SHALL end the wait at once with that error. Only after the marker appeared SHALL the agent command be typed.

#### Scenario: Shell rc asks a question
- **WHEN** the user's `.zshrc` ends in a prompt that reads one key, like oh-my-zsh's `[Y/n]` update question, and the user presses `s` on Card `m1-board-view`
- **THEN** the first probe's space answers that question, the shell runs the rest of the probe at its prompt and prints `kanban-ready-<nonce>`, then Claude is started and receives `/opsx:apply m1-board-view`; `laude` is never typed

#### Scenario: Shell rc reads a whole line
- **WHEN** the user's rc reads a whole line before continuing, and the user presses `s` on Card `m1-board-view`
- **THEN** the first probe is swallowed as that line, a second probe with another nonce is typed after 2 seconds and answered at the prompt, and Claude is started once

#### Scenario: Slow rc without a dialog
- **WHEN** the user's rc takes 3 seconds (nvm, conda) and the user presses `s`
- **THEN** the probes typed meanwhile run as soon as the prompt appears, the wait ends with the first marker the pane shows, and Claude is started

#### Scenario: Shell already at its prompt
- **WHEN** the agent of `m1-board-view` has exited, its shell is at its prompt, and the user starts the Change again
- **THEN** the first probe is answered at once and Claude starts in that tab without a noticeable delay

#### Scenario: Earlier marker in the same tab
- **WHEN** the tab of `m1-board-view` still shows `kanban-ready-a1b2c3` from an earlier Start, and the Change is started again
- **THEN** the new probe uses another nonce and only its own marker ends the wait

#### Scenario: Resume waits too
- **WHEN** the Change `m1-board-view` with the recorded `claude` session `abc` is resumed in a new tab
- **THEN** the probe runs before `claude --resume abc` is typed

#### Scenario: Shell never reaches its prompt
- **WHEN** the shell of the new tab shows a wizard that takes the probes as input and never reaches a prompt, and the user presses `s`
- **THEN** after 60 seconds no agent is started, and the failure says the shell did not reach its prompt in the Change's tab and shows the pane's last lines

#### Scenario: Pane gone during the wait
- **WHEN** the user closes the new tab while the plugin is still probing
- **THEN** the wait ends at once with herdr's error and no further probe is typed
