# waves Specification

## Purpose
Lets the user start several Ready Changes of one Project at once as a wave. A coordinator agent merges the Changes on a shared branch and hands the result to the Main agent once, asking the user only when no agent can go on.
## Requirements
### Requirement: A wave is plugin state
A wave SHALL be recorded per Project in the user's cache directory, next to the Main pane records. The record SHALL hold the wave's number, branch, worktree path, base, the coordinator's pane, the Main pane (the pane the Board was opened from), the member Changes and the dropped ones, and whether the wave is finished. A wave SHALL not be an OpenSpec Change and SHALL write nothing below `openspec/`. A wave is active from its start until it is finished. N SHALL be one more than the highest number among the Project's recorded waves and its git branches named `wave-<number>`, starting at 1. Losing the record SHALL make the `kanban wave` subcommands for that wave fail with `unknown wave <N>` and remove its badges. It SHALL not touch any branch, worktree or agent.

#### Scenario: First wave
- **WHEN** a Project has no recorded wave and no branch `wave-<number>`
- **THEN** its first wave has number 1 and branch `wave-1`

#### Scenario: Branch left from an earlier wave
- **WHEN** a Project's records were lost and branch `wave-3` still exists
- **THEN** the next wave has number 4

### Requirement: Starting a wave
Starting a wave of Changes from the Board SHALL first check that each Change is open and in Ready. Then it SHALL create branch `wave-<N>` from the base a new worktree would use (the configured `base`, else the detected default branch), in a new worktree workspace labelled `wave-<N>`. In it, it SHALL start the coordinator, an agent of the kind Start would launch, named `wave-<N>`, only at the shell prompt as for Start, and submit the coordinator prompt (see "Coordinator prompt"). It SHALL record the wave and focus the coordinator's tab. It SHALL start no member itself. When a step fails, it SHALL stop, record no wave, and report the step and herdr's or git's error, as Start does.

#### Scenario: Wave of three
- **WHEN** the user starts a wave of `f5-bar`, `m12-escalation` and `c4-foo` from a Board opened in pane `p1` of a Project whose base is `main`
- **THEN** branch `wave-1` exists from `main` in a worktree workspace labelled `wave-1`, an agent `wave-1` runs there and receives the coordinator prompt naming the three Changes, the wave records `p1` as its Main pane, and none of the three is started

#### Scenario: A member is no longer Ready
- **WHEN** `c4-foo` was started by another Board after it was marked, and the user starts the wave
- **THEN** no wave is started and the error says `c4-foo` is not in Ready

### Requirement: Coordinator prompt
The coordinator prompt SHALL name the wave, its branch and worktree, the base, and each member with the path of its proposal. It SHALL give these commands with the absolute path of the running plugin binary: `wave start <N> <change>`, `wave drop <N> <change>`, `wave escalate <N> "<what you need>"` and `wave finish <N>`. It SHALL tell the coordinator to:
1. read the proposals and decide the order, starting a member only after every member it builds on is merged into `wave-<N>`, and starting independent members right away;
2. when a member hands off, review it, merge its branch into `wave-<N>`, resolve every conflict itself, run the Project's tests, archive the Change on `wave-<N>` with `/opsx:archive <change>`, and remove the member's worktree, never with `--force` and keeping its branch;
3. drop a member it cannot make work, saying why in its final report;
4. answer a member's escalation itself when it can, and escalate to the Main agent only when it cannot decide;
5. finish the wave when every member is archived on `wave-<N>` or dropped.

#### Scenario: Commands in the prompt
- **WHEN** wave 1 is started and the plugin binary is `/opt/kanban/bin/kanban`
- **THEN** the coordinator prompt contains `/opt/kanban/bin/kanban wave start 1 <change>` and `/opt/kanban/bin/kanban wave finish 1`

### Requirement: Start a member
`kanban wave start <N> <change>` SHALL start the member as a Start from the Board would. The differences are:
- a new worktree's branch SHALL start from `wave-<N>` instead of the base;
- the member's Main pane SHALL be the coordinator's pane;
- the member's tab SHALL not be focused;
- the prompt SHALL be the Start prompt, then the escalation sentence, then `Do not archive <change>: the coordinator of wave <N> merges and archives it.`

The member SHALL be watched as a Start from the Board is. When the Change already has a checkout, the member SHALL work there as Start would, and its branch keeps its own base. The command SHALL refuse an unknown or finished wave, a Change that is not a member, and a dropped one, naming the reason, and SHALL print `started <change>` on success. It SHALL find the Project from the wave records, whatever directory it runs in.

#### Scenario: Member branches off the wave
- **WHEN** the coordinator of wave 1 runs `kanban wave start 1 c4-foo` and `c4-foo` has no checkout
- **THEN** a worktree on branch `c4-foo` from `wave-1` opens without taking focus, its agent receives `/opsx:apply c4-foo`, the escalation sentence and `Do not archive c4-foo: the coordinator of wave 1 merges and archives it.`, and the coordinator's pane is recorded as its Main pane

#### Scenario: Not a member
- **WHEN** the coordinator of wave 1 runs `kanban wave start 1 f9-other`
- **THEN** nothing is started and the error says `f9-other` is not a member of wave 1

### Requirement: The coordinator is the members' Main agent
For a member of an active wave, the coordinator SHALL take the Main agent's place for Handoff, Stall and Escalation. The Handoff SHALL name `wave-<N>` as the branch to merge into. When the coordinator does not take a prompt, the fallback SHALL be the same as for a missing Main agent.

#### Scenario: Member finishes
- **WHEN** member `c4-foo` of wave 1 settles idle with all tasks done in its worktree
- **THEN** its pane closes and the coordinator's prompt asks to review `c4-foo`, merge branch `c4-foo` into `wave-1`, archive it with `/opsx:archive c4-foo` and remove its worktree

#### Scenario: Member escalates
- **WHEN** the agent of member `c4-foo` runs `kanban escalate c4-foo "need the API key"`
- **THEN** the coordinator receives the escalation prompt and the Main agent receives nothing

### Requirement: Drop a member
`kanban wave drop <N> <change>` SHALL record the member as dropped and print `dropped <change>`. It SHALL not stop its agent, close its tab or remove its worktree or branch. A dropped Change SHALL be an ordinary Change again, with its Main pane unchanged until it is started again. The command SHALL refuse an unknown or finished wave and a Change that is not a member.

#### Scenario: Drop a failing member
- **WHEN** the coordinator of wave 1 runs `kanban wave drop 1 c4-foo`
- **THEN** wave 1 records `c4-foo` as dropped, its Card loses the badge `≋1`, and its worktree and branch stay

### Requirement: Coordinator escalation
`kanban wave escalate <N> <reason…>` SHALL prompt the wave's Main agent with the wave, its branch and worktree, the coordinator's tab and the reason. The prompt SHALL tell it to answer the coordinator with `herdr agent prompt wave-<N> "<answer>"` or to tell the user exactly what to do. When no Main agent takes the prompt, herdr SHALL show it as the notification `spec-kanban: wave <N> needs you`. The command SHALL print who received it and refuse an unknown or finished wave.

#### Scenario: Coordinator is stuck
- **WHEN** the coordinator of wave 1 runs `kanban wave escalate 1 "c4-foo and f5-bar change the same API in opposite ways"`
- **THEN** the Main agent receives a prompt naming wave 1, branch `wave-1`, its worktree, tab `wave-1` and that reason

### Requirement: Finish a wave
`kanban wave finish <N>` SHALL refuse while a member is neither dropped nor archived on `wave-<N>`, naming those members. Otherwise it SHALL mark the wave finished and hand it to the Main agent once. The prompt SHALL name the archived members and the dropped ones. It SHALL ask the Main agent to:
- review branch `wave-<N>` in its worktree;
- merge it into the base;
- remove the worktree with `git worktree remove <path>`, never with `--force`, keeping the branch;
- stop and tell the user when the merge or the removal does not go through cleanly.

When no Main agent takes the prompt, herdr SHALL show it as the notification `spec-kanban: wave <N> finished`. The coordinator's tab SHALL stay open. The plugin itself SHALL never merge, remove a worktree or delete a branch.

#### Scenario: All members through
- **WHEN** `f5-bar` and `m12-escalation` are archived on `wave-1`, `c4-foo` was dropped, and the coordinator runs `kanban wave finish 1`
- **THEN** the Main agent receives one prompt naming `f5-bar` and `m12-escalation` as archived and `c4-foo` as dropped, asking to merge `wave-1` into `main` and remove the wave worktree, and the wave's badges disappear

#### Scenario: A member is still open
- **WHEN** `c4-foo` is neither archived on `wave-1` nor dropped and the coordinator runs `kanban wave finish 1`
- **THEN** the wave stays active, nothing is sent and the error names `c4-foo`

