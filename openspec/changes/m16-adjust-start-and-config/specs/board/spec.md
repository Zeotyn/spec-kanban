## ADDED Requirements

### Requirement: Start with adjusted instructions
`S` on a Card that `s` would start (Ready or In Progress, without a running agent) SHALL replace the footer with an input line `start> ` pre-filled with the configured `instructions`, or with the whole configured `prompt` (its `{change}` replaced by the Change name) when one is set. The user SHALL be able to edit the line as in the filter prompt. `enter` SHALL start the Change as `s` does, with the edited text in place of the configured one: after the Source's command, separated by one space, when no `prompt` is set, or as the whole text when a `prompt` is set; a blank line SHALL mean the Source's command alone. The Escalation sentence SHALL be appended as for `s`. Nothing SHALL be saved, and the next `s` SHALL send the configured text again. `esc` SHALL close the line without starting. On a Card with a running agent, `S` SHALL focus it as `s` does. On a Card `s` cannot start, and while 2 or more Cards are marked for a wave, `S` SHALL show a short notice (`S starts one change: clear the marks first` with marks) and open no line. The Card menu SHALL offer "start with…" on the Cards where `S` works, doing the same. The key help SHALL list `S`.

#### Scenario: Different instructions for one start
- **WHEN** the config sets `instructions` to `with TDD`, the user presses `S` on Ready Card `m16-adjust-start-and-config`, changes the line to `with TDD, no new dependencies` and presses `enter`
- **THEN** the agent gets `/opsx:apply m16-adjust-start-and-config with TDD, no new dependencies` followed by the Escalation sentence, and `config.json` is unchanged

#### Scenario: Next start uses the config again
- **WHEN** the user started a Change with `S` and adjusted instructions, and later presses `s` on another Ready Card
- **THEN** that agent gets the configured `instructions`

#### Scenario: Empty line
- **WHEN** no `prompt` is set, the user presses `S`, clears the line and presses `enter`
- **THEN** the agent gets `/opsx:apply <change>` and the Escalation sentence

#### Scenario: Prompt configured
- **WHEN** the config sets `prompt` to `/opsx:apply {change} quickly` and the user presses `S` on `m1-board-view`
- **THEN** the line holds `/opsx:apply m1-board-view quickly`, and `enter` sends the edited line as the whole prompt

#### Scenario: Cancel
- **WHEN** the line is open and the user presses `esc`
- **THEN** the line closes, nothing starts and the Board is as before

#### Scenario: Marks for a wave
- **WHEN** 2 Ready Cards are marked and the user presses `S`
- **THEN** no line opens and the footer says `S starts one change: clear the marks first`

### Requirement: Edit the config from the Board
`c` on the Board SHALL open the plugin's `config.json` (see the kanban-config spec) in the user's editor (`$EDITOR`, `vi` when unset) and suspend the Board while the editor runs, as `e` does. When the file does not exist, it SHALL be created with `{}` first, together with its directory. When the editor exits, the Board SHALL read the config again: hidden Phases, the wave size and every later Start, Verify and Archive SHALL use the new settings (except `source` and the `base` Spec Kit reads merged Features against, which apply when the Board opens next), marks beyond a lowered wave size SHALL be cleared, and a config that is broken SHALL show the config warning in the footer until the next key and mean the defaults, as at opening. When the config directory cannot be determined, `c` SHALL show a short notice and open no editor. The selection and the filter SHALL stay. The key help SHALL list `c`.

#### Scenario: Change instructions for good
- **WHEN** the user presses `c`, sets `"instructions": "with TDD"` in the editor and quits it
- **THEN** the Board is shown again, and the next `s` sends `/opsx:apply <change> with TDD`

#### Scenario: Hide a Phase
- **WHEN** the user presses `c`, adds `"hide": ["archived"]` and quits the editor
- **THEN** the Board no longer shows the Archived column

#### Scenario: No config yet
- **WHEN** no `config.json` exists and the user presses `c`
- **THEN** the editor opens a new `config.json` holding `{}`

#### Scenario: Broken config
- **WHEN** the user leaves invalid JSON in the editor and quits it
- **THEN** the footer shows the config warning naming the file and the problem, and the Board uses the defaults
