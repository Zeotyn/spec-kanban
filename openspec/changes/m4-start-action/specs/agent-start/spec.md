## ADDED Requirements

### Requirement: Start does not take another Project's agent
herdr agent names are unique across all Projects. When an agent named after the Change exists but works in a directory outside the Project, Start SHALL neither focus nor restart it, create no tab, and fail with a message naming the Change and the other agent's directory.

#### Scenario: Equally named Change in two Projects
- **WHEN** an agent `m1-board-view` works in Project `alpha` and the user starts `m1-board-view` in Project `beta`
- **THEN** no tab is created, the `alpha` agent is not focused, and Start reports that `m1-board-view` already runs in `alpha`
