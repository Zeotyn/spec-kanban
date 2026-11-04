## MODIFIED Requirements

### Requirement: Start opens an agent session for a Change
Starting a Change SHALL open a new herdr tab whose working directory is the Change's checkout (see "Start works in the Change's checkout"; the Project when worktrees are off) and that is labelled or whose workspace is labelled with the Change name, let the Project's Source prepare that checkout (OpenSpec needs nothing; for Spec Kit see the speckit-source spec), focus it, start an agent of the chosen kind in that tab as an agent that herdr tracks, and submit the prompt to it. The prompt SHALL be the configured `prompt` with `{change}` replaced by the Change name when one is set; otherwise the Source's command for that kind (`/opsx:apply <change>` for OpenSpec, Spec Kit's implement command for Spec Kit), followed by a space and the configured `instructions` when they are set. When the preparation fails, Start SHALL report it and start no agent. The herdr agent SHALL be named like the Change, prefixed with `f-` when the name does not start with a lowercase letter (herdr agent names must, and have at most 32 characters). When that name has more than 32 characters, the agent SHALL instead be named with its first 25 characters, without a trailing `-`, followed by `-` and the first 6 hexadecimal characters of the SHA-256 of the Change name; the tab, workspace and worktree branch SHALL keep the full Change name. The agent name SHALL depend on the Change name alone, and everything that finds the agent again (Running badge, Watch, resume, Verify, Archive) SHALL look for the agent name of the Change, not derive a Change from an agent name. The kind SHALL be the configured `agent`; without one, the kind herdr reports for the agent running in the invoking pane (the pane the Board was opened from, or the pane `kanban start` runs in); `claude` when that pane runs no agent, herdr cannot tell, or the reported kind is not one Start can launch.

#### Scenario: Start a Ready Change
- **WHEN** the user runs `kanban start m1-board-view` inside a Project that contains that open Change, from a plain shell, and nothing is configured
- **THEN** a new workspace labelled `m1-board-view` opens on the Change's worktree and its tab is focused before the agent is ready, Claude runs in it, and its first prompt is `/opsx:apply m1-board-view`

#### Scenario: Start with a configured prompt
- **WHEN** the configured prompt is `/opsx:apply {change} with ponytail, TDD, subagents and review loops` and the user presses `s` on Card `m1-board-view`
- **THEN** the agent's first prompt in the new tab is `/opsx:apply m1-board-view with ponytail, TDD, subagents and review loops`

#### Scenario: Start with configured instructions
- **WHEN** the configured `instructions` are `with ponytail, TDD, subagents and review loops`, no `prompt` is set, and the user presses `s` on Card `m1-board-view`
- **THEN** the agent's first prompt in the new tab is `/opsx:apply m1-board-view with ponytail, TDD, subagents and review loops`

#### Scenario: Board opened from a Codex pane
- **WHEN** the Board was opened from a pane where herdr reports a `codex` agent, no `agent` is configured, and the user presses `s` on Card `m1-board-view`
- **THEN** Codex is started in the new tab and receives the configured prompt

#### Scenario: Board opened from a plain shell
- **WHEN** the Board was opened from a pane that runs no agent, no `agent` is configured, and the user presses `s`
- **THEN** Claude is started in the new tab

#### Scenario: herdr tracks the session
- **WHEN** a Change has been started
- **THEN** `herdr agent list` shows an agent for that Change, so its status appears in the herdr sidebar

#### Scenario: Start a Spec Kit Feature
- **WHEN** the user presses `s` on Ready Card `001-user-login-email` of a Spec Kit Project whose integration is `claude`, and nothing is configured
- **THEN** the Feature's worktree gets `.specify/feature.json` naming `specs/001-user-login-email` before Claude starts, and Claude's first prompt is `/speckit-implement`

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
