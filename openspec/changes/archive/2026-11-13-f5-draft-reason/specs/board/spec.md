## ADDED Requirements

### Requirement: Draft reason in the footer
While the selected Card is a Draft Card and the footer shows the key footer, the footer SHALL start with the reason the Change is in Draft, followed by ` · ` and the key footer:
- `draft: <artifacts> missing` when planning artifacts are neither done nor skipped, naming them in the order proposal, specs, design, tasks and separated by `, `;
- `draft: no tasks` when every planning artifact is done or skipped and the Change has no task;
- `draft: planning unknown` when the Source reported no planning state.

The key footer SHALL use its short form when the full form does not fit next to the reason, and the whole footer SHALL be cut to the Board's width. Clicks on the footer actions SHALL keep acting like their keys. When Cards are marked, the filter prompt is open, the wave question is asked, a notice is shown or a Change is starting, the footer SHALL show that as before, without the reason. On any other Card, in a column without Cards, and in the detail view the footer SHALL be unchanged.

#### Scenario: Design missing
- **WHEN** the selected Card is `m14-spike-agent-messaging` in Draft, with proposal and tasks done, specs skipped and design missing
- **THEN** the footer starts with `draft: design missing · ` followed by the key footer

#### Scenario: Several artifacts missing
- **WHEN** the selected Draft Card has only a proposal
- **THEN** the footer starts with `draft: specs, design, tasks missing`

#### Scenario: Planned but no task
- **WHEN** the selected Draft Card has every planning artifact done and a `tasks.md` without a checkbox
- **THEN** the footer starts with `draft: no tasks`

#### Scenario: Ready Card selected
- **WHEN** the user moves the selection from a Draft Card to a Ready Card
- **THEN** the footer shows the key footer without a reason

#### Scenario: Marks win
- **WHEN** two Ready Cards are marked and the selection moves to a Draft Card
- **THEN** the footer says `2 marked · s wave · esc clear` without a reason

#### Scenario: Click after the reason
- **WHEN** the footer shows `draft: design missing · ` before the key footer and the user clicks `q close`
- **THEN** the Board closes
