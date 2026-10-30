## ADDED Requirements

### Requirement: Changes worked on in another checkout
When the Project is the top directory of a git checkout, the source SHALL read each other checkout of that repository (a linked worktree, or the main checkout when the Project is a linked worktree) whose branch is named exactly like a Change and which contains that Change's directory, through `openspec list` run in that checkout. The Change of that name as reported there, with its Progress, last modification time and planning state read there, SHALL replace the Project's open Change of that name, or be added as an open Change when the Project has no Change of that name. The source SHALL report for every open Change the directory its files are read from, which the Board's detail view and `e` use. Other Changes of that checkout, a Change the Project has archived, and a checkout whose read fails or that does not report the Change SHALL leave the Project's Changes as they are, without an error. A Project that is not the top directory of a git checkout SHALL be read as before.

#### Scenario: Agent ticks tasks in its worktree
- **WHEN** Change `m1-board-view` has 0 of 6 tasks done in the Project and 3 of 6 in the linked worktree on branch `m1-board-view`
- **THEN** the source reports `m1-board-view` with 3 of 6 tasks done and the worktree's last modification time, and its tasks file is the worktree's

#### Scenario: Change created in its worktree
- **WHEN** the linked worktree on branch `m2-export` contains open Change `m2-export` and the Project has no Change `m2-export`
- **THEN** the source reports the open Change `m2-export` as read in that worktree

#### Scenario: Other Changes of a worktree
- **WHEN** the worktree on branch `m1-board-view` also contains an open Change `m9-other` that the Project does not have
- **THEN** the source does not report `m9-other`

#### Scenario: Branch without the Change
- **WHEN** a linked worktree is on branch `fix-typo` and no Change `fix-typo` exists in it
- **THEN** the source does not run `openspec list` there and reports the Project's Changes unchanged

#### Scenario: Archived after merge
- **WHEN** the Project has archived `m1-board-view` and a stale worktree on branch `m1-board-view` still reports it open
- **THEN** the source reports `m1-board-view` only as archived

#### Scenario: Worktree read fails
- **WHEN** `openspec list` fails in the worktree on branch `m1-board-view`
- **THEN** the source reports the Project's own `m1-board-view` and no error
