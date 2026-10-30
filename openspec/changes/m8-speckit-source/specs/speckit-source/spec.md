## Purpose

Defines how the plugin reads a Spec Kit Project: which Source a Project gets, which Features it has with their Progress and Phase, which documents the detail view shows, and what Start and the Handoff do differently for a Feature.

## ADDED Requirements

### Requirement: One Source per Project
A directory SHALL be a Spec Kit Project when it contains both a `.specify` directory and a `specs` directory, and an OpenSpec Project when it contains `openspec/changes`. The Project of a working directory SHALL be the nearest directory at or above it that is a Project of either kind. A Project SHALL be read by exactly one Source: the only one that recognises it, or, when it is both kinds, OpenSpec unless the config key `source` is `speckit`. A `source` naming a Source that does not recognise the Project SHALL be ignored for that Project. The Board SHALL never mix Work items of two Sources.

#### Scenario: Spec Kit Project
- **WHEN** the Board is opened from a pane in `<repo>/src`, and `<repo>` has `.specify/` and `specs/` but no `openspec/changes`
- **THEN** the Board shows `<repo>` with its Spec Kit Features

#### Scenario: Both tools, no preference
- **WHEN** `<repo>` has `openspec/changes`, `.specify/` and `specs/`, and no `source` is configured
- **THEN** the Board shows `<repo>`'s OpenSpec Changes only

#### Scenario: Both tools, Spec Kit preferred
- **WHEN** `<repo>` has both markers and the config sets `source` to `speckit`
- **THEN** the Board shows `<repo>`'s Spec Kit Features only

#### Scenario: Preference for a missing tool
- **WHEN** the config sets `source` to `speckit` and the Project has only `openspec/changes`
- **THEN** the Board shows its OpenSpec Changes

#### Scenario: Spec Kit initialised without features
- **WHEN** a directory has `.specify/` but no `specs/` directory
- **THEN** it is not a Spec Kit Project

### Requirement: Features come from the specs directory
Each directory directly below the Project's `specs/` that contains a `spec.md` SHALL be one open Work item, a Feature, named like its directory (`001-user-login-email`, `20261030-142501-user-login`). A directory without `spec.md` SHALL be no Feature. A Feature's Progress SHALL count the checkbox lines of its `tasks.md` (`- [ ] …`, `- [x] …` or `- [X] …`, under any heading), the checked ones as completed. Its modification time SHALL be that of the newest file in its directory, at any depth. The `**Status**` line of `spec.md` SHALL be ignored. When the Project's `specs/` directory cannot be read, the load SHALL fail with that error.

#### Scenario: Feature with tasks
- **WHEN** `specs/001-user-login-email/tasks.md` has 12 task lines under `## Phase 1: Setup` and `## Phase 2: Foundational`, 5 of them checked with `[X]` or `[x]`
- **THEN** the Card `001-user-login-email` shows 5/12

#### Scenario: Folder without a spec
- **WHEN** `specs/002-draft/` holds only `plan.md`
- **THEN** no Card `002-draft` is shown

#### Scenario: Status line ignored
- **WHEN** `spec.md` says `**Status**: Draft` and all tasks are checked
- **THEN** the Feature is in Review, not Draft

### Requirement: Phase of a Feature
A Feature's Phase SHALL follow these rules in order: Archived when all of its at least one tasks are completed and its branch is merged into the base (see "Merged branch archives a Feature"); Review when it has at least one task and all are completed; In Progress when some but not all are completed; Ready when it has at least one task and none is completed; Draft otherwise (only `spec.md`, `plan.md` without `tasks.md`, or a `tasks.md` without task lines). Open checklists in `checklists/` SHALL not affect the Phase. A Feature SHALL show no milestone badge; the name filter SHALL match it by substring as any Work item.

#### Scenario: Specified only
- **WHEN** a Feature has only `spec.md`
- **THEN** it is in Draft

#### Scenario: Planned without tasks
- **WHEN** a Feature has `spec.md`, `plan.md` and `research.md` but no `tasks.md`
- **THEN** it is in Draft

#### Scenario: Tasks generated
- **WHEN** a Feature's `tasks.md` has 9 unchecked task lines
- **THEN** it is in Ready

#### Scenario: Implementation under way
- **WHEN** 4 of its 9 tasks are checked
- **THEN** it is in In Progress

#### Scenario: All tasks done
- **WHEN** all 9 tasks are checked and no branch named like the Feature exists
- **THEN** it is in Review

#### Scenario: Open checklist
- **WHEN** a Feature with 9 unchecked tasks has `checklists/requirements.md` with unchecked items
- **THEN** it is in Ready

### Requirement: Merged branch archives a Feature
A Feature whose tasks are all completed SHALL be Archived when a local branch named exactly like the Feature exists and is merged into the base (the configured `base`, else the default branch as Start determines it), that is, its tip is reachable from the base. An Archived Feature SHALL show no archive date. A branch that is not merged, or no such branch, SHALL leave the Feature in Review. A Feature with tasks left SHALL never be Archived, so a branch just created from the base does not archive it. The merge state SHALL be read with at most one git call per load, and only when some Feature has all its tasks completed; when git fails, no Feature SHALL be Archived.

#### Scenario: Merged after review
- **WHEN** Feature `001-user-login-email` has all 9 tasks checked in the Project and branch `001-user-login-email` is merged into `main`, the base
- **THEN** the Card `001-user-login-email` is in Archived and shows `archived` without a date

#### Scenario: Fresh branch
- **WHEN** Start just created branch `002-export` from `main` and the Feature has 0 of 6 tasks checked
- **THEN** it is in Ready

#### Scenario: Done but not merged
- **WHEN** a Feature has all tasks checked on branch `003-search`, which has commits not in `main`
- **THEN** it is in Review

#### Scenario: Worked on the base without a branch
- **WHEN** a Feature has all tasks checked and no branch of its name exists
- **THEN** it stays in Review

#### Scenario: Nothing to check
- **WHEN** no Feature has all its tasks completed
- **THEN** the load runs no git call for merge state

### Requirement: Features worked on in another checkout
A git checkout of the Project's repository other than the Project itself whose branch is named like a Feature directory it holds (`specs/<branch>/spec.md`) SHALL supply that Feature: it SHALL be read in that checkout, replacing the Project's Feature of the same name or added when the Project has none. Its documents and Progress SHALL come from that checkout.

#### Scenario: Progress from the worktree
- **WHEN** Feature `001-user-login-email` has 0 of 9 tasks checked in the Project and 4 of 9 in the worktree on branch `001-user-login-email`
- **THEN** the Card shows 4/9 in In Progress

#### Scenario: Feature created in a worktree
- **WHEN** a worktree on branch `004-reports` holds `specs/004-reports/spec.md` and the Project has no such folder
- **THEN** the Board shows `004-reports` in Draft

### Requirement: Planning marks of a Feature
Draft and Ready Cards of a Feature SHALL show the planning marks `P S D T` as follows: `P` always as skipped (`-`), `S` done when `spec.md` exists, `D` done when `plan.md` exists, `T` done when `tasks.md` exists; a missing file SHALL be marked missing.

#### Scenario: Planned Feature
- **WHEN** a Draft Feature has `spec.md` and `plan.md`
- **THEN** its Card shows `- S D ·`

### Requirement: Detail tabs of a Feature
The detail view of a Feature SHALL show the tabs Tasks (`tasks.md`) and Spec (`spec.md`), followed by Plan (`plan.md`) and Research (`research.md`) only when the file exists, in this order. It SHALL open on Tasks when the Feature has tasks and on Spec otherwise. The Tasks tab SHALL start with the next open task line or a note that all tasks are done. `e` SHALL edit the shown tab's file.

#### Scenario: Feature with plan and research
- **WHEN** the user opens the details of a Feature with `spec.md`, `plan.md`, `research.md` and `tasks.md` (3 of 9 tasks checked)
- **THEN** the tabs are `1 Tasks`, `2 Spec`, `3 Plan`, `4 Research`, Tasks is shown and starts with `next ▸` and the fourth task line

#### Scenario: Specified only
- **WHEN** the user opens the details of a Feature with only `spec.md`
- **THEN** the tabs are Tasks and Spec, Spec is shown, and the Tasks tab says there are no tasks

#### Scenario: Edit the plan
- **WHEN** the Plan tab is shown and the user presses `e`
- **THEN** `plan.md` of the Feature (in its checkout, when it is read from one) opens in the editor

### Requirement: Start prepares the Feature's checkout
Before the agent of a Feature starts, Start SHALL write `.specify/feature.json` in the checkout the agent works in (the Feature's worktree, or the Project) with the content `{"feature_directory":"specs/<name>"}`, replacing an existing one and creating `.specify/` when needed, so Spec Kit's commands find the Feature without a branch or environment variable. When it cannot be written, Start SHALL fail with a message naming the file, start no agent and send no prompt. Start SHALL check that `specs/<name>` exists on the ref a new worktree starts from.

#### Scenario: First Start in a worktree
- **WHEN** the user presses `s` on Ready Card `001-user-login-email` and `specs/001-user-login-email` is committed on `main`
- **THEN** herdr creates the worktree on branch `001-user-login-email`, its `.specify/feature.json` reads `{"feature_directory":"specs/001-user-login-email"}` before the agent starts, and the agent receives the implement command

#### Scenario: Worktree off
- **WHEN** `worktree` is off, the Project's `.specify/feature.json` names `specs/002-export`, and the user Starts `001-user-login-email`
- **THEN** the Project's `.specify/feature.json` names `specs/001-user-login-email` before the agent starts

#### Scenario: Feature not committed
- **WHEN** `specs/001-user-login-email` exists only uncommitted in the Project
- **THEN** no worktree, tab or agent is created and Start reports that `001-user-login-email` is not on `main`

### Requirement: Default prompt of a Feature
Without a configured `prompt`, Start SHALL send Spec Kit's implement command for the agent kind it starts, read from the Project's `.specify/integration.json`: the integration named like the kind when it is installed there, else the file's `integration`. For the integrations `codex` and `zcode` the command SHALL be `$speckit-implement`; for any other it SHALL be `/speckit` followed by the integration's `invoke_separator` and `implement`. When the file is missing or unreadable, or names no integration or no separator, the command SHALL be `/speckit-implement`. A configured `prompt` SHALL replace it, `{change}` becoming the Feature name.

#### Scenario: Claude integration
- **WHEN** `.specify/integration.json` names integration `claude` with `invoke_separator` `-` and the user Starts a Feature with Claude
- **THEN** the agent receives `/speckit-implement`

#### Scenario: Codex
- **WHEN** `codex` is an installed integration and the agent kind is `codex`
- **THEN** the agent receives `$speckit-implement`

#### Scenario: Dot separator
- **WHEN** the integration is `gemini` with `invoke_separator` `.`
- **THEN** the agent receives `/speckit.implement`

#### Scenario: Older layout
- **WHEN** the Project has no `.specify/integration.json`
- **THEN** the agent receives `/speckit-implement`

### Requirement: Handoff of a finished Feature
A Feature SHALL count as finished when all its at least one tasks are completed; it has no archive. The Handoff prompt of a finished Feature SHALL ask for no archive step. Without a linked worktree it SHALL read `<name> is finished (<done>/<total> tasks). Please review it.`; with one it SHALL ask to review it, merge its branch into the base, delete `<path>/.specify/feature.json` unless git tracks it, then remove the worktree as for any Work item.

#### Scenario: Finish without a worktree
- **WHEN** `worktree` is off and the agent of `001-user-login-email` settles with 9/9 tasks
- **THEN** the Main agent's prompt starts with `001-user-login-email is finished (9/9 tasks). Please review it.`

#### Scenario: Finish in a worktree
- **WHEN** the agent of `001-user-login-email` settles with 9/9 tasks in worktree `~/.herdr/worktrees/app/001-user-login-email`
- **THEN** the Main agent's prompt asks to review it, merge branch `001-user-login-email` into `main`, delete `~/.herdr/worktrees/app/001-user-login-email/.specify/feature.json` unless git tracks it, then run `git worktree remove ~/.herdr/worktrees/app/001-user-login-email`, and names no archive command
