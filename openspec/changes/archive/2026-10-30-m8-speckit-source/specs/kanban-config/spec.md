## MODIFIED Requirements

### Requirement: Settings and defaults
The config file SHALL be a JSON object with these optional keys:
- `agent`: a string, the agent kind Start launches, overriding the kind detected from the invoking pane. It SHALL be one of the kinds herdr's `agent start --kind` accepts (`claude`, `codex`, `gemini`, `opencode`, …), matched exactly. Default: unset.
- `prompt`: a string, the complete text Start sends to the agent, in which every `{change}` is replaced by the Change name. It SHALL contain `{change}` at least once. It applies to every Project, whatever its Source. Default: unset, meaning the Source's command followed by `instructions` (see below).
- `instructions`: a string, extra text Start appends to the Source's command, separated by one space and passed verbatim (no `{change}` replacement). The Source's command is `/opsx:apply <change>` for OpenSpec and Spec Kit's implement command for Spec Kit. One text applies to every Source. An empty string counts as unset. It is ignored when `prompt` is set. Default: unset, so Start sends the Source's command alone.
- `hide`: a list of Phase names (`Draft`, `Ready`, `In Progress`, `Review`, `Archived`, ignoring case) whose columns the Board does not show. It SHALL leave at least one Phase shown. Default: empty.
- `worktree`: a boolean, whether Start runs the agent in the Change's git checkout rather than in the Project. Default: `true`.
- `base`: a non-empty string, the git ref a new worktree's branch starts from, overriding the detected base. Default: unset.
- `source`: `openspec` or `speckit`, the Source that reads a Project recognised by both. Default: unset, meaning OpenSpec.

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

### Requirement: Invalid config falls back to defaults with a warning
When `config.json` cannot be read, is not valid JSON, contains an unknown key, or a setting breaks its rule (an `agent` herdr does not know, a `prompt` without `{change}`, an unknown Phase in `hide`, all Phases hidden, an empty `base`, a `source` other than `openspec` or `speckit`, an `instructions` that is not a string, a value of the wrong type), all settings SHALL have their defaults. The Board SHALL show a warning naming the file and the problem in place of its footer when it opens, until the next key; `kanban start` SHALL print the warning to stderr and continue.

#### Scenario: Typo in a key
- **WHEN** `config.json` is `{"promt": "/opsx:apply {change} with TDD"}` and the Board opens
- **THEN** the footer warns that `config.json` has an unknown key `promt` and defaults are used, and Start would send `/opsx:apply <change>`

#### Scenario: Prompt without the Change
- **WHEN** `config.json` is `{"prompt": "go on"}` and the user runs `kanban start m1-board-view`
- **THEN** stderr shows a warning that the prompt lacks `{change}`, and the agent receives `/opsx:apply m1-board-view`

#### Scenario: Unknown agent kind
- **WHEN** `config.json` is `{"agent": "Codex"}` and the Board opens
- **THEN** the footer warns that `agent` `Codex` is not a herdr agent kind and defaults are used, so Start launches the kind detected from the invoking pane

#### Scenario: Broken JSON
- **WHEN** `config.json` is not valid JSON and the Board opens
- **THEN** the Board shows all five columns and the footer shows the warning; the next key restores the key help

#### Scenario: Worktree as a string
- **WHEN** `config.json` is `{"worktree": "no"}` and the Board opens
- **THEN** the footer warns about the wrong type of `worktree` and defaults are used, so Start creates worktrees

#### Scenario: Unknown source
- **WHEN** `config.json` is `{"source": "spec-kit"}` and the Board opens
- **THEN** the footer warns that `source` `spec-kit` is not `openspec` or `speckit` and defaults are used

#### Scenario: Instructions not a string
- **WHEN** `config.json` is `{"instructions": ["with ponytail"]}` and the Board opens
- **THEN** the footer warns about the wrong type of `instructions` and defaults are used, so Start sends `/opsx:apply <change>`
