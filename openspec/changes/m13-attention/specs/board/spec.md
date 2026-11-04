## ADDED Requirements

### Requirement: Attention marker on the Card
An open Card SHALL show `⚑` in ANSI red (`1`) followed by a space before its name while its Change waits for the user (see "Attention marker"). It SHALL take the place of the Running agent badge, so a Card shows at most one of `⚑` and `●`, and the name and milestone badge keep the same width as with the Running agent badge. Archived Cards SHALL show no `⚑`. The Board SHALL look the markers up together with the Running agents: the Cards SHALL appear without `⚑` and gain it when herdr answers, and `r` SHALL look them up again. When herdr cannot be asked, only Handoff markers SHALL be shown, and no error.

#### Scenario: Escalated Change
- **WHEN** the agent of In Progress Change `m1-board-view` escalated and is idle, and the Board opens
- **THEN** Card `m1-board-view` shows a red `⚑` before its name and no `●`, and stays in In Progress

#### Scenario: Agent works again
- **WHEN** Card `m1-board-view` shows `⚑`, its agent starts working, and the user presses `r`
- **THEN** the Card shows the green `●` instead of `⚑`

#### Scenario: Handoff without a Main agent
- **WHEN** `m1-board-view` was handed off without a Main agent, its pane is closed, and the Board opens
- **THEN** Review Card `m1-board-view` shows `⚑`

#### Scenario: Long name
- **WHEN** Card `m1-board-view` shows `⚑` in a narrow column
- **THEN** its name is shortened exactly as it would be next to `●`

#### Scenario: herdr unavailable
- **WHEN** the agent lookup fails and `m1-board-view` has an Escalation marker
- **THEN** Card `m1-board-view` shows no `⚑` and the Board shows no error

## MODIFIED Requirements

### Requirement: Running agent badge
An open Card SHALL show a Running agent badge when herdr reports an agent named after the Change whose working directory lies in the Project or in another checkout of the Project's git repository. The badge SHALL use a terminal colour, not change the Card's Phase or column, and not appear on Archived Cards. A Card that shows the attention marker SHALL show it in the badge's place instead of the badge (see "Attention marker on the Card"). Looking up Running agents SHALL not delay showing the Board: the Cards SHALL appear without badges and gain them when herdr answers. `r` SHALL look them up again. When herdr cannot be asked, the Board SHALL show no badges and no error.

#### Scenario: Agent working on a Change
- **WHEN** herdr reports an agent named `m1-board-view` working in the Project and the Board opens
- **THEN** Card `m1-board-view` shows the Running agent badge and stays in its Phase column

#### Scenario: Agent working in a worktree
- **WHEN** herdr reports an agent named `m1-board-view` working in a linked worktree of the Project's repository
- **THEN** Card `m1-board-view` shows the Running agent badge

#### Scenario: Same name in another Project
- **WHEN** herdr reports an agent named `m1-board-view` whose working directory is another Project
- **THEN** Card `m1-board-view` shows no badge

#### Scenario: Cached Board opens at once
- **WHEN** the Board opens with a cached snapshot and herdr answers later
- **THEN** the cached Cards show immediately and the badges appear once herdr has answered

#### Scenario: herdr unavailable
- **WHEN** the agent lookup fails
- **THEN** the Board shows its Cards without badges and no error
