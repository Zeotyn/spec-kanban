## Purpose

Defines the user settings of openspec-kanban: where the config file lives, which settings it holds with their defaults, and how a missing or invalid config is handled.

## ADDED Requirements

### Requirement: Config file location
The plugin SHALL read its settings from `config.json` in herdr's config directory for the plugin, as given to plugin commands in `HERDR_PLUGIN_CONFIG_DIR`. When that variable is not set, as for `kanban start` run from a shell, the plugin SHALL use the directory that `herdr plugin config-dir openspec-kanban` prints. A missing file, or a config directory that cannot be determined, SHALL mean all defaults without a warning.

#### Scenario: Board opened through herdr
- **WHEN** `HERDR_PLUGIN_CONFIG_DIR` is `~/.config/herdr/plugins/config/openspec-kanban` and that directory contains `config.json`
- **THEN** the Board uses the settings from that file

#### Scenario: Start from a shell
- **WHEN** the user runs `kanban start m1-board-view` in a shell without `HERDR_PLUGIN_CONFIG_DIR`, and `config.json` exists in the directory `herdr plugin config-dir openspec-kanban` prints
- **THEN** Start uses the settings from that file

#### Scenario: No config file
- **WHEN** the config directory has no `config.json`
- **THEN** all settings have their defaults and no warning is shown

### Requirement: Settings and defaults
The config file SHALL be a JSON object with these optional keys:
- `prompt`: a string, the text Start sends to the agent, in which every `{change}` is replaced by the Change name. It SHALL contain `{change}` at least once. Default: `/opsx:apply {change}`.
- `hide`: a list of Phase names (`Draft`, `Ready`, `In Progress`, `Review`, `Archived`, ignoring case) whose columns the Board does not show. It SHALL leave at least one Phase shown. Default: empty.

A key that is left out SHALL have its default.

#### Scenario: Custom prompt
- **WHEN** `config.json` is `{"prompt": "/opsx:apply {change} with TDD and review loops"}`
- **THEN** starting `m1-board-view` sends `/opsx:apply m1-board-view with TDD and review loops`, and no Phase is hidden

#### Scenario: Hide Archived
- **WHEN** `config.json` is `{"hide": ["archived"]}`
- **THEN** the Board shows no Archived column, and Start sends `/opsx:apply <change>`

### Requirement: Invalid config falls back to defaults with a warning
When `config.json` cannot be read, is not valid JSON, contains an unknown key, or a setting breaks its rule (a `prompt` without `{change}`, an unknown Phase in `hide`, all Phases hidden, a value of the wrong type), all settings SHALL have their defaults. The Board SHALL show a warning naming the file and the problem in place of its footer when it opens, until the next key; `kanban start` SHALL print the warning to stderr and continue.

#### Scenario: Typo in a key
- **WHEN** `config.json` is `{"promt": "/opsx:apply {change} with TDD"}` and the Board opens
- **THEN** the footer warns that `config.json` has an unknown key `promt` and defaults are used, and Start would send `/opsx:apply <change>`

#### Scenario: Prompt without the Change
- **WHEN** `config.json` is `{"prompt": "go on"}` and the user runs `kanban start m1-board-view`
- **THEN** stderr shows a warning that the prompt lacks `{change}`, and the agent receives `/opsx:apply m1-board-view`

#### Scenario: Broken JSON
- **WHEN** `config.json` is not valid JSON and the Board opens
- **THEN** the Board shows all five columns and the footer shows the warning; the next key restores the key help
