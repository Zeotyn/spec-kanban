## MODIFIED Requirements

### Requirement: Settings and defaults
The config file SHALL be a JSON object with these optional keys:
- `agent`: a string, the agent kind Start launches, overriding the kind detected from the invoking pane. It SHALL be one of the kinds herdr's `agent start --kind` accepts (`claude`, `codex`, `gemini`, `opencode`, …), matched exactly. Default: unset.
- `prompt`: a string, the complete text Start sends to the agent, in which every `{change}` is replaced by the Change name. It SHALL contain `{change}` at least once. It applies to every Project, whatever its Source. Default: unset, meaning the Source's command followed by `instructions` (see below).
- `instructions`: a string, extra text Start appends to the Source's command, separated by one space and passed verbatim (no `{change}` replacement). The Source's command is `/opsx:apply <change>` for OpenSpec and Spec Kit's implement command for Spec Kit. One text applies to every Source. An empty string counts as unset. It is ignored when `prompt` is set. Default: unset, so Start sends the Source's command alone.
- `hide`: a list of Phase names (`Draft`, `Ready`, `In Progress`, `Review`, `Archived`, ignoring case) whose columns the Board does not show. It SHALL leave at least one Phase shown. Default: empty.
- `worktree`: a boolean, whether Start runs the agent in the Change's git checkout rather than in the Project. Default: `true`.
- `base`: a non-empty string, the git ref a new worktree's branch starts from, overriding the detected base. Default: unset.
- `commit`: a boolean, whether Start commits a Change's uncommitted planning files on the base before it creates the Change's worktree (see the agent-start spec). Default: `false`.
- `source`: `openspec` or `speckit`, the Source that reads a Project recognised by both. Default: unset, meaning OpenSpec.
- `wave`: an object with the optional key `size`, the maximum number of Changes in a wave (see the waves spec), an integer of at least 2. Default: `{"size": 5}`; a `wave` without `size` has size 5.

A key that is left out SHALL have its default.

When both `prompt` and `instructions` are set, the settings SHALL be kept, `prompt` SHALL be used and `instructions` ignored, and the Board and `kanban start` SHALL show a warning that `instructions` is ignored because `prompt` is set, in the same places as the config warning.

#### Scenario: Custom prompt
- **WHEN** `config.json` is `{"prompt": "/opsx:apply {change} with TDD and review loops"}`
- **THEN** starting `m1-board-view` sends `/opsx:apply m1-board-view with TDD and review loops`, and no Phase is hidden

#### Scenario: Hide Archived
- **WHEN** `config.json` is `{"hide": ["archived"]}`
- **THEN** the Board shows no Archived column, and Start sends `/opsx:apply <change>`

#### Scenario: Agent override
- **WHEN** `config.json` is `{"agent": "codex", "prompt": "/opsx-apply {change}"}` and the Board was opened from a pane running Claude
- **THEN** Start launches Codex and sends `/opsx-apply <change>`

#### Scenario: Worktrees off
- **WHEN** `config.json` is `{"worktree": false}`
- **THEN** Start opens the agent's tab in the Project and creates no worktree

#### Scenario: Base override
- **WHEN** `config.json` is `{"base": "develop"}`
- **THEN** Start creates new worktree branches from `develop`

#### Scenario: Prefer Spec Kit
- **WHEN** `config.json` is `{"source": "speckit"}` and the Project has both `openspec/changes` and `.specify/` with `specs/`
- **THEN** the Board shows the Project's Spec Kit Features

#### Scenario: Spec Kit default prompt
- **WHEN** no `prompt` is configured and the user starts a Feature of a Spec Kit Project whose integration is `claude`
- **THEN** Start sends `/speckit-implement`

#### Scenario: Instructions for OpenSpec
- **WHEN** `config.json` is `{"instructions": "with ponytail, TDD and review loops"}` and the user starts Change `m1-board-view`
- **THEN** Start sends `/opsx:apply m1-board-view with ponytail, TDD and review loops`

#### Scenario: Instructions for Spec Kit
- **WHEN** `config.json` is `{"instructions": "with ponytail, TDD and review loops"}` and the user starts a Feature of a Spec Kit Project whose integration is `gemini` with `invoke_separator` `.`
- **THEN** Start sends `/speckit.implement with ponytail, TDD and review loops`

#### Scenario: Empty instructions
- **WHEN** `config.json` is `{"instructions": ""}` and the user starts Change `m1-board-view`
- **THEN** Start sends `/opsx:apply m1-board-view`, with no trailing space, and no warning is shown

#### Scenario: Prompt wins over instructions
- **WHEN** `config.json` is `{"prompt": "/opsx:apply {change} quickly", "instructions": "with ponytail"}` and the Board opens
- **THEN** the footer warns that `instructions` is ignored because `prompt` is set, and starting `m1-board-view` sends `/opsx:apply m1-board-view quickly`

#### Scenario: Wave size
- **WHEN** `config.json` is `{"wave": {"size": 3}}` and the user marks a fourth Card on the Board
- **THEN** the fourth Card is not marked and the footer says `a wave holds at most 3 changes`

#### Scenario: Default wave size
- **WHEN** `config.json` has no `wave` key
- **THEN** a wave holds at most 5 Changes


#### Scenario: Commit on Start
- **WHEN** `config.json` is `{"commit": true}`
- **THEN** Start commits a Change's uncommitted planning files on the base before it creates the worktree

#### Scenario: Commit off by default
- **WHEN** `config.json` has no `commit` key
- **THEN** Start commits nothing and refuses a Change that is not on the base, as before
