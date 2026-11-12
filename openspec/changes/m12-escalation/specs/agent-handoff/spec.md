## ADDED Requirements

### Requirement: Main pane recorded per Change
A Start or a resume from the Board SHALL record the pane the Board was opened from as the Change's Main pane, per Project and Change in the user's cache directory (not the plugin's state directory, which the Change's agent cannot find), replacing the previous record. When that pane is unknown, the record SHALL be removed. Escalation SHALL read the Main agent from this record. Losing the record SHALL only make Escalation fall back to a notification.

#### Scenario: Board opened from a Claude pane
- **WHEN** the Board was opened from pane `p1` and the user presses `s` on Card `m1-board-view`
- **THEN** `p1` is recorded as the Main pane of `m1-board-view` in that Project

#### Scenario: Equally named Change in two Projects
- **WHEN** Changes named `m1-board-view` in Projects `alpha` and `beta` were started from Boards opened in panes `p1` and `p2`
- **THEN** `alpha` records `p1` and `beta` records `p2`

### Requirement: Escalation
`kanban escalate <change> <reason>` SHALL hand a Change that cannot go on without a human to the Main agent at once. All arguments after the Change name SHALL form the reason, joined by single spaces. It SHALL run in the Project or in a checkout of the Project's git repository; it SHALL find the Project there, and SHALL fail with a non-zero status and a message, sending nothing, when there is none, when the reason is empty, or when the Change is not an open Change of the Project. Otherwise it SHALL submit this prompt to the Main agent recorded for the Change (see "Main pane recorded per Change"), with Progress read as the Board reads it:

`<change> needs a human (<done>/<total> tasks): <reason>. Its agent <agent> waits in tab <change><where>. 1. Read its open tasks in <tasks file>. 2. If you may do the step yourself, do it and tick its task; otherwise tell the user exactly what to do and wait for them. 3. Then tell its agent to continue: herdr agent prompt <agent> "<what changed>". Do not close its tab; when all tasks are done, the normal Handoff follows.`

`<agent>` SHALL be herdr's agent name of the Change. `<where>` SHALL be `, worktree <path> on branch <change>` when the Change's checkout is a linked worktree, and empty otherwise. `<tasks file>` SHALL be the absolute path of the Change's tasks file in the checkout the agent works in. When the Main agent cannot take the prompt, it SHALL fall back as described in "Main agent falls back to a notification". Escalation SHALL record that the Change escalated, so that the next settle of a watched agent leads to no Stall, and SHALL print whether the Main agent or a notification received it. It SHALL never close, focus or prompt the Change's own agent.

#### Scenario: Worker cannot publish
- **WHEN** the agent of `m1-board-view` works in worktree `~/.herdr/worktrees/spec-kanban/m1-board-view`, has completed 3 of 7 tasks, and runs `kanban escalate m1-board-view needs npm publish and a git tag` while the recorded Main agent is idle
- **THEN** the Main agent receives a prompt starting with `m1-board-view needs a human (3/7 tasks): needs npm publish and a git tag. Its agent m1-board-view waits in tab m1-board-view, worktree ~/.herdr/worktrees/spec-kanban/m1-board-view on branch m1-board-view.`, the worker's tab stays open, and the command prints that the Main agent received it

#### Scenario: Without a worktree
- **WHEN** `worktree` is off and the agent of `m1-board-view` escalates with 3 of 7 tasks completed
- **THEN** the prompt says `Its agent m1-board-view waits in tab m1-board-view.` and names `<project>/openspec/changes/m1-board-view/tasks.md`

#### Scenario: No Main agent
- **WHEN** the Board that started `m1-board-view` was opened from a shell pane and its agent escalates
- **THEN** nothing is typed into any pane, herdr shows a notification `spec-kanban: m1-board-view needs you` with the prompt text, and the command prints that a notification received it

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

`<agent>` SHALL be herdr's agent name of the Change. When the Progress cannot be read, `at <done>/<total> tasks` SHALL be left out. The 10 minutes SHALL be fixed. When the agent works again within them, nothing SHALL be sent. When the Main agent cannot take the prompt, the plugin SHALL fall back as described in "Main agent falls back to a notification". A blocked settle SHALL lead to no Stall.

#### Scenario: Worker stops silently
- **WHEN** the watched agent of `m1-board-view` settles idle at 3 of 7 tasks without escalating and stays idle for 10 minutes
- **THEN** the Main agent receives `m1-board-view stopped at 3/7 tasks 10 minutes ago without asking for help.` followed by the rest of the Stall prompt, once

#### Scenario: Worker goes on by itself
- **WHEN** the watched agent of `m1-board-view` settles idle at 3 of 7 tasks and starts working again after 2 minutes
- **THEN** no prompt and no notification is sent

#### Scenario: User answers in time
- **WHEN** the watched agent of `m1-board-view` settles idle at 3 of 7 tasks and the user answers it in its tab after 5 minutes
- **THEN** no Stall prompt is sent for that settle

#### Scenario: Second silent stop
- **WHEN** after a Stall prompt the agent of `m1-board-view` is told to continue, works, and settles again unfinished for 10 minutes
- **THEN** the Main agent receives a second Stall prompt

## MODIFIED Requirements

### Requirement: A Start from the Board is watched until its agent goes away
After a Start from the Board has sent its prompt, or a resume from the Board has started the recorded session, the plugin SHALL watch the Change's agent in the background and evaluate it each time it settles (idle, done or blocked) after working. A resumed agent SHALL be evaluated only after it has worked once, so that reopening a session leads to no Handoff and no Stall by itself. Watching SHALL end with a Handoff, or silently when the agent no longer exists (the agent exited, the tab or pane was closed, herdr stopped). A Start through `kanban start`, a Start or resume that only focused a running agent, and Verify and Archive SHALL not be watched.

#### Scenario: Agent works through several turns
- **WHEN** the agent of `m1-board-view` settles with tasks left, the user answers it within 10 minutes, and it later settles with all tasks completed
- **THEN** the first settle shows no notification and prompts no Main agent, and the second one leads to a Handoff

#### Scenario: User closes the tab
- **WHEN** the user closes the agent tab of a watched Change before it is finished
- **THEN** watching ends without a notification and nothing else is closed or prompted

#### Scenario: Start from a shell
- **WHEN** the user runs `kanban start m1-board-view` and the agent later finishes all tasks
- **THEN** the tab stays open and no Main agent is prompted

#### Scenario: Resumed session finishes the Change
- **WHEN** the user resumes In Progress Change `m1-board-view` from the Board, tells its agent to go on, and it settles with all tasks completed
- **THEN** a Handoff follows as after a Start

#### Scenario: Resumed session of a finished Change
- **WHEN** the user resumes Review Change `m1-board-view` from the Board and does not type anything for 15 minutes
- **THEN** the tab stays open, no Handoff happens and no Main agent is prompted

### Requirement: An unfinished settle leaves the tab open
When the watched agent settles blocked, or with the Change not finished, the plugin SHALL leave its tab open, show no notification, prompt no Main agent at that moment (see "Stall" for later), and keep watching.

#### Scenario: Agent stops with tasks left
- **WHEN** the agent of `m1-board-view` settles idle with 3 of 7 tasks completed
- **THEN** the tab stays open, no notification is shown and no Main agent is prompted at once

#### Scenario: Agent asks for approval
- **WHEN** the agent of `m1-board-view` settles blocked, even with all tasks completed
- **THEN** the tab stays open, no notification is shown, no Main agent is prompted, and herdr's own sidebar shows the agent as blocked

### Requirement: Main agent falls back to a notification
For a Handoff and a Stall, the Main agent SHALL be the agent, of any kind herdr detects, in the pane the Board was opened from; for an Escalation, the agent in the Main pane recorded for the Change. When there is no such pane, it no longer exists, it is the Change's own agent pane, herdr detects no agent in it, or herdr rejects the prompt (for example because it is blocked), the plugin SHALL show the prompt text as a herdr notification instead: titled `spec-kanban: <change> finished` for a Handoff, `spec-kanban: <change> needs you` for an Escalation and `spec-kanban: <change> stalled` for a Stall. The Handoff itself (session recorded, pane closed) SHALL still happen.

#### Scenario: Board opened from a shell pane
- **WHEN** the Board was opened from a pane running only a shell (no agent) and a started Change finishes cleanly
- **THEN** the agent's pane closes and herdr shows a notification `spec-kanban: m1-board-view finished` with the Handoff text

#### Scenario: Main agent blocked
- **WHEN** a Change finishes cleanly while the Main agent waits at an approval dialog
- **THEN** no text is typed into the Main agent and herdr shows the Handoff text as a notification

#### Scenario: Stall without a Main agent
- **WHEN** the Board that started `m1-board-view` was opened from a shell pane and its agent stalls for 10 minutes
- **THEN** herdr shows a notification `spec-kanban: m1-board-view stalled` with the Stall text

### Requirement: Resume a recorded session
Whether a recorded session is reopenable SHALL be decided by the agent kind herdr reported with the session: a session id is reopenable for `claude` (`--resume <id>`), `copilot` (`--resume=<id>`), `codex` (`resume <id>`), `opencode` (`--session <id>`), `cursor` (`--resume <id>`) and `gemini` (`--resume <id>`), the arguments following the agent's executable; every other kind, and a session reported as a path, SHALL be not reopenable. Resuming a Change SHALL choose its tab the way Start does (a running agent is only focused; the exited agent's tab, the Change's worktree reopened or recreated from its branch, or a new tab labelled with the Change), focus it, and start the recorded agent kind in it as the Change's agent with its resume argument, without submitting a prompt. A resume from the Board SHALL then be watched (see "A Start from the Board is watched until its agent goes away"). A failure after the Board closed SHALL be shown as a herdr notification `spec-kanban: resume failed`.

#### Scenario: Resume a finished Change
- **WHEN** the Change `m1-board-view` has the recorded `claude` session `abc`, no agent, and no open workspace for its worktree, and is resumed
- **THEN** its worktree opens as a workspace labelled `m1-board-view` and is focused, Claude runs in it on session `abc`, no prompt is sent, and the tab is not closed before the agent has worked

#### Scenario: Resume after the worktree was removed
- **WHEN** the Main agent removed the worktree of `m1-board-view` but its branch and recorded session remain, and the Change is resumed
- **THEN** herdr recreates the worktree from branch `m1-board-view` and the session is resumed in it

#### Scenario: Resume a Copilot session
- **WHEN** the Change `m1-board-view` has the recorded `copilot` session `abc` and is resumed
- **THEN** Copilot is started in the Change's tab with `--resume=abc`

#### Scenario: Recorded session no longer exists
- **WHEN** the recorded session of `m1-board-view` was deleted from Claude's history and the Change is resumed
- **THEN** the tab stays open with Claude's own error, or herdr shows `spec-kanban: resume failed` when Claude does not become ready
