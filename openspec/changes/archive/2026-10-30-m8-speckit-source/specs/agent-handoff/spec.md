## MODIFIED Requirements

### Requirement: Handoff of a finished Change
When the watched agent settles in idle or done and the Change is finished, the plugin SHALL record the agent's session for the Change, close the agent's pane (which closes its tab when it is the only pane), and submit a prompt to the Main agent that names the Change and its Progress and asks to review it. The finishing step SHALL come from the Project's Source: for OpenSpec, archiving the Change with `/opsx:archive <change>`; for Spec Kit, none without a worktree and deleting the worktree's `.specify/feature.json` unless git tracks it with one (see the speckit-source spec); a Source without a step adds nothing to the prompt. Without a linked worktree on a branch named like the Change, the prompt SHALL ask for the finishing step unless the Change is already archived. With such a worktree, the prompt SHALL name its branch and path and ask, after the review, to merge that branch into the base (the configured `base`, else the default branch as Start determines it), to take the finishing step unless the Change is already archived, and then to remove the worktree with `git worktree remove <path>`; it SHALL say never to use `--force`, to keep the branch, and to stop when the merge or the removal does not go through cleanly. The plugin itself SHALL never merge, remove a worktree or delete a branch. When the recorded session is reopenable, the prompt SHALL also say how to reopen it (`s` on its Card, or the agent's resume command in the Change's checkout). The Handoff SHALL work the same for every agent kind.

#### Scenario: Clean finish without a worktree
- **WHEN** `worktree` is off, the Claude agent of `m1-board-view` settles idle, all 7 tasks are completed, and the Board was opened from a pane running an agent
- **THEN** the agent's tab closes and the agent in that pane receives a prompt starting with `m1-board-view is finished (7/7 tasks). Please review it and archive it with /opsx:archive m1-board-view.` that names `claude --resume <session>`

#### Scenario: Clean finish in a worktree
- **WHEN** the agent of `m1-board-view` works in the worktree `~/.herdr/worktrees/spec-kanban/m1-board-view` on branch `m1-board-view`, the base is `main`, and it settles idle with all 7 tasks completed
- **THEN** the agent's pane closes and the Main agent's prompt asks to review the Change, merge branch `m1-board-view` into `main`, archive it with `/opsx:archive m1-board-view`, then run `git worktree remove ~/.herdr/worktrees/spec-kanban/m1-board-view`, never with `--force` and keeping the branch; the worktree and the branch still exist

#### Scenario: Agent archived the Change in its worktree
- **WHEN** the agent of `m1-board-view` archives the Change inside its worktree and settles idle
- **THEN** the agent's pane closes and the Main agent's prompt starts with `m1-board-view is finished and archived` and asks to merge branch `m1-board-view` and remove the worktree, without asking to archive it

#### Scenario: Agent archived the Change without a worktree
- **WHEN** `worktree` is off and the agent of `m1-board-view` archives the Change and settles idle
- **THEN** the agent's tab closes and the Main agent receives a prompt starting with `m1-board-view is finished and archived. Please review it.`

#### Scenario: Clean finish of another reopenable kind
- **WHEN** the `codex` agent of `m1-board-view` finishes it cleanly with session `abc`
- **THEN** the Main agent's prompt names `codex resume abc`

#### Scenario: Clean finish of a kind without resume
- **WHEN** the configured agent kind is one without a known resume argument (for example `pi`) and its agent finishes `m1-board-view` cleanly
- **THEN** the pane closes and the Main agent receives the same prompt without a way to reopen the session

#### Scenario: Other panes in the agent tab
- **WHEN** the user split the agent tab and a Handoff happens
- **THEN** only the agent's pane closes and the user's other pane stays

#### Scenario: Source without a finishing step
- **WHEN** `worktree` is off and the agent of Spec Kit Feature `001-user-login-email` settles idle with all 9 tasks completed
- **THEN** the Main agent's prompt starts with `001-user-login-email is finished (9/9 tasks). Please review it.` and asks for no archive
