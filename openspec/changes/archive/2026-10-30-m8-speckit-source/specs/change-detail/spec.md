## MODIFIED Requirements

### Requirement: Tabs for a Change's documents
The detail view SHALL show one Change with the tabs its Source supplies, in the Source's order: for an OpenSpec Change four tabs, Tasks, Proposal, Design, Specs; for a Spec Kit Feature the tabs of the speckit-source spec. A header SHALL show the Change name, its milestone badge, its Progress and its Phase. The view SHALL open on Tasks when the Change has tasks and otherwise on the Source's first document tab (Proposal for OpenSpec). `1` up to the number of tabs SHALL switch to that tab, `Tab` and `shift+Tab` to the next and previous tab, wrapping around.

#### Scenario: Open a Change with tasks
- **WHEN** the user opens the details of `m1-board-view`, which has 5 of 11 tasks completed
- **THEN** the header shows `m1-board-view`, `M1`, `5/11` and `In Progress`, and the Tasks tab is shown

#### Scenario: Open a Draft without tasks
- **WHEN** the user opens the details of a Change that has only a proposal
- **THEN** the Proposal tab is shown

#### Scenario: Switch tabs
- **WHEN** the Tasks tab is shown and the user presses `Tab` four times
- **THEN** Proposal, Design, Specs and then Tasks are shown in turn

#### Scenario: Feature with two tabs
- **WHEN** the details of a Spec Kit Feature with only `spec.md` are shown and the user presses `3`
- **THEN** the Spec tab stays shown, since the Feature has only Tasks and Spec
