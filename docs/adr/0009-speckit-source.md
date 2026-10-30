---
status: accepted
---

# Spec Kit Source

Spec Kit (`internal/source/speckit`) is the second Source behind the interface of ADR 0008. Its Work items are Features, the folders `specs/<name>/` with a `spec.md`. The Phase derives from the files, as for OpenSpec: Draft, Ready, In Progress, Review.

Spec Kit has no archive, so a merged branch counts as Archived: a Feature whose tasks are all done and whose branch, named like the Feature, is merged into the base moves to Archived. Requiring all tasks done keeps a branch just created from the base from archiving a Ready Feature, and requiring the branch to be checked out nowhere keeps an agent's finished but uncommitted work (its branch tip still the base) in Review until the worktree is removed after the merge. One `git branch --merged` call per load, and only while a Feature is in Review; when git fails, nothing is archived. A Feature without a branch stays in Review.

A Project has one Source. Detection takes the nearest directory at or above the working directory that either tool recognises. When both do, OpenSpec wins, because that is today's behaviour; the config key `source` prefers Spec Kit instead. It is only a preference, so it is harmless as a global setting for Projects that only one tool recognises. The Board never mixes the two.

Start writes `.specify/feature.json` into the agent's checkout, since a fresh worktree has no current feature and the pane's environment cannot be set. The Handoff asks to delete it before `git worktree remove`, which refuses an untracked file. The default prompt follows the agent kind and `.specify/integration.json`. The config key `instructions` appends text to the Source's command; `prompt` stays a complete override.

## Considered Options

- **Review as the final column, no Archived for Spec Kit**: no git call, but Features pile up in Review and never leave the Board's open columns.
- **Rename Archived to Done**: neutral, but a visible change (columns, `hide` values, sidebar, goldens) for a gain only this Source would use.
- **Mix both Sources on one Board**: needs a Source per Item and a unified Phase and cache story; a Project that uses both tools is rare, and `source` covers choosing.

## Consequences

`Prepare`, a prompt per agent kind and the checkout in `Finish` were added to the interface. A Project's cache may show the other Source's Items until the first refresh after `source` changes. Spec Kit state other than `feature.json` is never written.
