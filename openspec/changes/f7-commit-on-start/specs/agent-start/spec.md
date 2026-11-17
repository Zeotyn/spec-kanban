## ADDED Requirements

### Requirement: Start commits the plan when configured
When the config sets `commit` to `true` and a Start or a wave would create a new worktree from the base, Start SHALL first commit the uncommitted files of each Change it starts, when all of these hold: the Project is a git checkout, its current branch is the base (a local branch of that name), and the Change's directory in the checkout has uncommitted changes, including new files. The commit SHALL hold only the files under that Change's directory, whatever else is staged or changed, SHALL be one commit per Change with the message `docs: plan <change>`, and SHALL run the repository's hooks. When the commit fails, Start SHALL fail with git's error and create nothing. In every other case, and when `commit` is `false`, Start SHALL commit nothing and check the base as before. When a Change is not on the base and nothing was committed, the error SHALL say how to fix it: `<change> is not on <ref>: commit it there, set base, or set "commit": true`. A Change that already has its own branch SHALL never be committed by Start.

#### Scenario: New Change committed on Start
- **WHEN** `commit` is `true`, the Project is on `main`, the base is `main`, and `openspec/changes/m15-sort-name-date` exists only uncommitted, while `README.md` is staged too
- **THEN** Start commits only `openspec/changes/m15-sort-name-date` as `docs: plan m15-sort-name-date`, `README.md` stays staged and uncommitted, and the worktree is created from `main` with the Change in it

#### Scenario: Edited plan committed
- **WHEN** `commit` is `true`, the Change is on `main`, and its `tasks.md` has uncommitted edits
- **THEN** Start commits the edit before creating the worktree, so the worktree holds the edited `tasks.md`

#### Scenario: Checkout on another branch
- **WHEN** `commit` is `true`, the base is `main`, the Project's checkout is on `feature-x`, and the Change is not on `main`
- **THEN** Start commits nothing and fails with `m15-sort-name-date is not on main: commit it there, set base, or set "commit": true`

#### Scenario: Commit off
- **WHEN** `commit` is `false` or not set and the Change exists only uncommitted
- **THEN** Start commits nothing and fails with the same message

#### Scenario: Hook refuses
- **WHEN** `commit` is `true` and the repository's pre-commit hook refuses the commit
- **THEN** Start fails with git's error and creates no worktree, tab or agent

#### Scenario: Wave
- **WHEN** `commit` is `true`, the Project is on `main`, and a wave is started with `m15-a` and `m15-b`, both only uncommitted
- **THEN** `m15-a` and `m15-b` are committed as two commits before branch `wave-<N>` is created from `main`
