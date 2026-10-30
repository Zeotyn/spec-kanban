# openspec-kanban

A herdr plugin that shows a Project's OpenSpec Changes as a kanban Board and summarises them in the sidebar.

## Projects and Changes

**Project**:
A directory a Source recognises (for OpenSpec, one containing `openspec/changes`) in which the panes of a herdr workspace work.
_Avoid_: Repo, workspace

**Source**:
The spec tool a Project's Work items are read from, behind one adapter; today only OpenSpec.
_Avoid_: Backend, provider, data source

**Work item**:
One piece of planned work a Source reports, shown as one Card; its kind depends on the Source.
_Avoid_: Ticket, task, item

**Change**:
The OpenSpec kind of Work item: a piece of planned work with proposal, specs, design and tasks, open or archived.
_Avoid_: Spec, ticket, story

**Spec**:
A main specification under `openspec/specs`; a Change carries spec deltas but is not a Spec itself.
_Avoid_: Change

**Progress**:
The number of completed tasks of a Work item relative to all its tasks.
_Avoid_: Status, percentage

**Most recently modified Change**:
The open Change with the latest modification time as reported by OpenSpec.
_Avoid_: Active change, last spec

## Board

**Board**:
The view of all Work items of a Project, arranged in columns by Phase.
_Avoid_: List, dashboard

**Card**:
The representation of one Work item on the Board.
_Avoid_: Ticket, item

**Phase**:
The state of a Work item as its Source decides it; each Phase is one column. For a Change it is derived from its artifacts and tasks by the rules below.
_Avoid_: Status, lane

**Draft**:
Phase of an open Change that has no tasks, or none of whose tasks is completed while its planning is incomplete or unknown; once a task is completed, progress decides the Phase.

**Ready**:
Phase of a fully planned Change with at least one task, none of which is completed.

**In Progress**:
Phase of a Change some, but not all, of whose tasks are completed.

**Review**:
Phase of an open Change with at least one task, all of which are completed, even while an agent is still working on it.
_Avoid_: Done, finished

**Archived**:
Phase of an archived Change.
_Avoid_: Done

**Running agent**:
A herdr agent session currently working on a Change; shown on its Card without changing its Phase.
_Avoid_: Live session, worker

**Start**:
The Board's only action: hands a Change to a new agent session in its own tab, in the Change's checkout.
_Avoid_: Move, apply button

**Change checkout**:
The git checkout whose branch is named like a Change, where Start runs its agent; usually a linked worktree.
_Avoid_: Sandbox, clone

## Sidebar

**Space row**:
The lines of a herdr workspace in the sidebar into which the plugin writes a summary.
_Avoid_: Status line
