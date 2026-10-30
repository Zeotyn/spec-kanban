---
status: accepted
---

# Sources behind adapters

The plugin reads its Work items through a `Source` interface (`internal/work`) instead of knowing OpenSpec everywhere. A Source detects its Project, loads the Items with their Phase, Progress and documents, names the path Start checks on the start ref, and supplies the default Start prompt and the instruction the Handoff asks for after review. Board, detail, sidebar, cache, Start and Handoff use only the neutral `work` types. OpenSpec (`internal/source/openspec`) is the only Source for now; Spec Kit and BMAD v6 are meant to follow as further adapters. ADR 0002 is from now on the decision of the OpenSpec Source.

The Source decides each Item's Phase: BMAD stores a status per story, OpenSpec and Spec Kit derive one from their files, so no central rule fits all. The columns, their order and the Card order stay in `work`.

The interface has only the methods today's code calls. A prompt per agent kind (Spec Kit), a group field for epics (BMAD) and the choice among several detected Sources are added with the Source that needs them. The Archived column keeps its name until a Source without an archive needs a neutral one.

## Considered Options

- **Keep OpenSpec everywhere, branch per tool later**: no work now, but every package would grow a switch over tools.
- **A generic file-pattern Source configured per tool**: one adapter for all, but BMAD's status file and OpenSpec's CLI do not reduce to patterns.
- **Rename Archived to Done now**: neutral from the start, but a visible change (columns, `hide` values, sidebar, goldens) for a Source that does not exist yet.

## Consequences

`cmd/kanban` builds one Source and hands it to every command. The cache stores neutral Items (format version 2). Adding a tool means a new package under `internal/source/` plus Source selection; no consumer package should need to change except for features the new tool brings.
