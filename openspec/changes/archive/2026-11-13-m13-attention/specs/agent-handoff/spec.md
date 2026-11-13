## ADDED Requirements

### Requirement: Attention marker
The plugin SHALL keep an attention marker per Project and Change in its state directory while the Change waits for the user, with its reason: Escalation, Stall or Handoff. Escalation and Stall SHALL set it, and a Handoff that no Main agent took SHALL set it (see "Main agent falls back to the attention marker"). The watcher SHALL clear it when the Change's agent starts working again after a settle; a Handoff the Main agent took, a Start and a resume from the Board SHALL clear it too. Readers SHALL show a marker only while it still holds, so that a marker left by a watcher that went away does not stay:
- an Escalation or Stall marker while herdr reports the Change's agent in the Project or one of its checkouts and its status is not `working`;
- a Handoff marker while the Change is open in the Project.

When herdr cannot be asked for its agents, Escalation and Stall markers SHALL not be shown. Losing the state directory SHALL only lose the markers.

#### Scenario: Worker escalates
- **WHEN** the agent of `m1-board-view` escalates and then ends its turn
- **THEN** `m1-board-view` waits for the user

#### Scenario: Worker continues
- **WHEN** the Main agent tells the escalated agent of `m1-board-view` to continue and it starts working
- **THEN** the marker of `m1-board-view` is cleared and it no longer waits

#### Scenario: Watcher gone
- **WHEN** `m1-board-view` has an Escalation marker, its watcher has ended, and the user then makes its agent work in its tab
- **THEN** `m1-board-view` does not wait while its agent works, although the marker file is still there

#### Scenario: Agent tab closed
- **WHEN** `m1-board-view` has a Stall marker and the user closes its agent's tab
- **THEN** `m1-board-view` no longer waits

#### Scenario: Handoff without a Main agent
- **WHEN** `m1-board-view` was handed off with no Main agent and is still open
- **THEN** `m1-board-view` waits until it is archived in the Project, started or resumed again

#### Scenario: New Start
- **WHEN** a Change with a marker is started or resumed from the Board
- **THEN** its marker is cleared

## RENAMED Requirements

- FROM: `### Requirement: Main agent falls back to a notification`
- TO: `### Requirement: Main agent falls back to the attention marker`

## MODIFIED Requirements

### Requirement: Main agent falls back to the attention marker
For a Handoff and a Stall, the Main agent SHALL be the agent, of any kind herdr detects, in the pane the Board was opened from; for an Escalation, the agent in the Main pane recorded for the Change. When there is no such pane, it no longer exists, it is the Change's own agent pane, herdr detects no agent in it, or herdr rejects the prompt (for example because it is blocked), the plugin SHALL type nothing and show no notification. Instead, a Handoff SHALL set the Change's attention marker for a Handoff; an Escalation and a Stall set their marker anyway. The Handoff itself (session recorded, pane closed) SHALL still happen. A Handoff the Main agent took SHALL clear the Change's attention marker.

#### Scenario: Board opened from a shell pane
- **WHEN** the Board was opened from a pane running only a shell (no agent) and a started Change finishes cleanly
- **THEN** the agent's pane closes, no notification is shown, and the attention marker of `m1-board-view` is set

#### Scenario: Main agent blocked
- **WHEN** a Change finishes cleanly while the Main agent waits at an approval dialog
- **THEN** no text is typed into the Main agent, no notification is shown, and the attention marker is set

#### Scenario: Stall without a Main agent
- **WHEN** the Board that started `m1-board-view` was opened from a shell pane and its agent stalls for 10 minutes
- **THEN** no notification is shown and the attention marker of `m1-board-view` is set

#### Scenario: Handoff clears an older marker
- **WHEN** `m1-board-view` escalated earlier, its agent then finished all tasks, and the Main agent takes the Handoff prompt
- **THEN** the attention marker of `m1-board-view` is cleared

### Requirement: Escalation
`kanban escalate <change> <reason>` SHALL hand a Change that cannot go on without a human to the Main agent at once. All arguments after the Change name SHALL form the reason, joined by single spaces. It SHALL run in the Project or in a checkout of the Project's git repository; it SHALL find the Project there, and SHALL fail with a non-zero status and a message, sending nothing, when there is none, when the reason is empty, or when the Change is not an open Change of the Project. Otherwise it SHALL submit this prompt to the Main agent recorded for the Change (see "Main pane recorded per Change"), with Progress read as the Board reads it:

`<change> needs a human (<done>/<total> tasks): <reason>. Its agent <agent> waits in tab <change><where>. 1. Read its open tasks in <tasks file>. 2. If you may do the step yourself, do it and tick its task; otherwise tell the user exactly what to do and wait for them. 3. Then tell its agent to continue: herdr agent prompt <agent> "<what changed>". Do not close its tab; when all tasks are done, the normal Handoff follows.`

`<agent>` SHALL be herdr's agent name of the Change. `<where>` SHALL be `, worktree <path> on branch <change>` when the Change's checkout is a linked worktree, and empty otherwise. `<tasks file>` SHALL be the absolute path of the Change's tasks file in the checkout the agent works in. Whether or not the Main agent takes the prompt (see "Main agent falls back to the attention marker"), Escalation SHALL set the Change's attention marker for an Escalation, so that the next settle of a watched agent leads to no Stall, and SHALL print whether the Main agent received it or only the marker shows it. It SHALL never close, focus or prompt the Change's own agent.

#### Scenario: Worker cannot publish
- **WHEN** the agent of `m1-board-view` works in worktree `~/.herdr/worktrees/spec-kanban/m1-board-view`, has completed 3 of 7 tasks, and runs `kanban escalate m1-board-view needs npm publish and a git tag` while the recorded Main agent is idle
- **THEN** the Main agent receives a prompt starting with `m1-board-view needs a human (3/7 tasks): needs npm publish and a git tag. Its agent m1-board-view waits in tab m1-board-view, worktree ~/.herdr/worktrees/spec-kanban/m1-board-view on branch m1-board-view.`, the worker's tab stays open, the attention marker of `m1-board-view` is set, and the command prints that the Main agent received it

#### Scenario: Without a worktree
- **WHEN** `worktree` is off and the agent of `m1-board-view` escalates with 3 of 7 tasks completed
- **THEN** the prompt says `Its agent m1-board-view waits in tab m1-board-view.` and names `<project>/openspec/changes/m1-board-view/tasks.md`

#### Scenario: No Main agent
- **WHEN** the Board that started `m1-board-view` was opened from a shell pane and its agent escalates
- **THEN** nothing is typed into any pane, no notification is shown, the attention marker of `m1-board-view` is set, and the command prints that only the marker shows it

#### Scenario: Unknown Change
- **WHEN** `kanban escalate m1-bord-view needs a tag` runs and no such open Change exists
- **THEN** nothing is sent, and the command exits non-zero naming `m1-bord-view`

#### Scenario: No reason
- **WHEN** `kanban escalate m1-board-view` runs without a reason
- **THEN** nothing is sent and the command exits non-zero with its usage

#### Scenario: Escalation is followed by no Stall
- **WHEN** the watched agent of `m1-board-view` escalates, ends its turn, and stays idle for more than 10 minutes
- **THEN** the Main agent receives only the Escalation prompt, no Stall prompt

### Requirement: Stall
When the watched agent settles idle or done, the Change is not finished, the agent did not escalate since it last started working, and it does not start working again within 10 minutes, the plugin SHALL submit this prompt to the Main agent once for that settle:

`<change> stopped at <done>/<total> tasks 10 minutes ago without asking for help. Read its output with: herdr agent read <agent> --lines 40. If it waits for something that finishes by itself, do nothing. If it needs a human, tell the user what it needs. To let it continue, run: herdr agent prompt <agent> "<what to do>". Do not close its tab.`

`<agent>` SHALL be herdr's agent name of the Change. When the Progress cannot be read, `at <done>/<total> tasks` SHALL be left out. The 10 minutes SHALL be fixed. When the agent works again within them, nothing SHALL be sent. With the prompt, whether or not the Main agent takes it (see "Main agent falls back to the attention marker"), the plugin SHALL set the Change's attention marker for a Stall. A blocked settle SHALL lead to no Stall.

#### Scenario: Worker stops silently
- **WHEN** the watched agent of `m1-board-view` settles idle at 3 of 7 tasks without escalating and stays idle for 10 minutes
- **THEN** the Main agent receives `m1-board-view stopped at 3/7 tasks 10 minutes ago without asking for help.` followed by the rest of the Stall prompt, once, and the attention marker of `m1-board-view` is set

#### Scenario: Worker goes on by itself
- **WHEN** the watched agent of `m1-board-view` settles idle at 3 of 7 tasks and starts working again after 2 minutes
- **THEN** no prompt and no notification is sent

#### Scenario: User answers in time
- **WHEN** the watched agent of `m1-board-view` settles idle at 3 of 7 tasks and the user answers it in its tab after 5 minutes
- **THEN** no Stall prompt is sent for that settle

#### Scenario: Second silent stop
- **WHEN** after a Stall prompt the agent of `m1-board-view` is told to continue, works, and settles again unfinished for 10 minutes
- **THEN** the Main agent receives a second Stall prompt
