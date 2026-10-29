---
status: accepted
---

# openspec-kanban owns the sidebar summary

openspec-kanban reports its own Space tokens (`kanban_change`, `kanban`, source `openspec-kanban`) and replaces the sidebar part of herdr-openspec, which the user has disabled. The summary reads Changes the same way the Board does (ADR 0002) and shares its cache, so sidebar and Board always agree on Phases and Progress, and every sidebar refresh warms the Board.

## Considered Options

- **Keep herdr-openspec for the sidebar**: no work, but it parses `openspec/changes` itself and derives a different notion of state (local git edits instead of Phases), so sidebar and Board would disagree.
- **Report into herdr-openspec's token names** (`openspec`, `openspec_change`): users keep their rows, but two plugins writing the same names fight whenever both are enabled.

## Consequences

Users replace herdr-openspec's `$openspec` rows with this plugin's snippet; plugins cannot change the user's config. Going back means re-enabling herdr-openspec and restoring its rows; the token names differ, so both can be installed at once. Every refresh costs one `openspec list` per Project (~0.65 s), triggered by herdr events rather than a timer.
