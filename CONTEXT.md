# openspec-kanban

A herdr plugin that shows a Project's OpenSpec Changes as a kanban Board and summarises them in the sidebar.

## Projects and Changes

**Project**:
A directory containing `openspec/changes` in which the panes of a herdr workspace work.
_Avoid_: Repo, workspace

**Change**:
An OpenSpec change: a piece of planned work with proposal, specs, design and tasks, open or archived.
_Avoid_: Spec, ticket, story

**Spec**:
A main specification under `openspec/specs`; a Change carries spec deltas but is not a Spec itself.
_Avoid_: Change

**Progress**:
The number of completed tasks of a Change relative to all its tasks.
_Avoid_: Status, percentage

**Most recently modified Change**:
The open Change with the latest modification time as reported by OpenSpec.
_Avoid_: Active change, last spec

## Board

**Board**:
The view of all Changes of a Project, arranged in columns by Phase.
_Avoid_: List, dashboard

**Card**:
The representation of one Change on the Board.
_Avoid_: Ticket, item

**Phase**:
The state of a Change derived from its artifacts and tasks; each Phase is one column.
_Avoid_: Status, lane

**Draft**:
Phase of a Change whose planning artifacts are not yet complete, or that has no tasks at all.

**Ready**:
Phase of a fully planned Change with at least one task, none of which is completed.

**In Progress**:
Phase of a Change some, but not all, of whose tasks are completed.

**Review**:
Phase of an open Change all of whose tasks are completed, even while an agent is still working on it.
_Avoid_: Done, finished

**Archived**:
Phase of an archived Change.
_Avoid_: Done

**Running agent**:
A herdr agent session currently working on a Change; shown on its Card without changing its Phase.
_Avoid_: Live session, worker

**Start**:
The Board's only action: hands a Change to a new agent session in its own tab.
_Avoid_: Move, apply button

## Sidebar

**Space row**:
The lines of a herdr workspace in the sidebar into which the plugin writes a summary.
_Avoid_: Status line
