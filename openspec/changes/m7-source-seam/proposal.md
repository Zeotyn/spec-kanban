## Why

The plugin should later show work from more spec tools than OpenSpec (Spec Kit next, then BMAD v6). Today OpenSpec is spread over the whole code: `openspec.Change` is the domain type of the Board, cache, detail, sidebar and Start, and the project marker, document layout, Phase rules, start check and agent prompts each sit in a different package. Adding a second tool now would mean touching every package; drawing one seam first keeps that later change small.

## What Changes

- A neutral package `internal/work` holds the domain: `Item` (one Work item, shown as one Card), `Phase` with the column model and the Card order within a column, `Docs` (the detail view's tabs) and the `Source` interface.
- `internal/openspec` and `internal/docs` move into `internal/source/openspec`, the only `Source` for now. It decides each Item's Phase itself (the change-phases rules move there unchanged) and owns everything OpenSpec-specific: the `openspec/changes` marker, the CLI and archive reads, the document tabs, the path checked on the start ref, the default Start prompt `/opsx:apply {change}` and the `/opsx:archive` instruction in the Handoff text.
- Board, detail, sidebar, cache, Start, Handoff and config use only `work` types; `internal/phase` goes away (its column model and ordering move to `work`).
- The cache stores Items; its format version goes from 1 to 2, so the first load after the upgrade starts without a cache, as after any version change.
- The Archived column keeps its name; a neutral name waits for a Source that needs one.
- No visible behaviour changes: same columns, Cards, detail tabs, sidebar rows, Start, worktree handling, Handoff text and config. Hence no spec deltas (`skip_specs: true`).
- New ADR 0008 (sources behind adapters); ADR 0002 gets a note that it is now the OpenSpec Source's decision; CONTEXT.md gains **Work item** and **Source**.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

None. Every requirement in `openspec/specs/` still holds word for word; the change only moves where it is implemented.

## Impact

- New: `internal/work/`, `internal/source/openspec/` (from `internal/openspec/` and `internal/docs/`).
- Removed: `internal/openspec/`, `internal/docs/`, `internal/phase/`.
- Changed: `internal/board`, `internal/detail`, `internal/sidebar`, `internal/cache` (version 2), `internal/start` (`Start`, `Prepare`, `Watched`), `internal/config` (prompt default from the Source, `Hide` indexed by `work.Phase`), `internal/project` (`Find` asks the Source), `cmd/kanban/main.go` (one Source built and passed everywhere).
- Unchanged: all golden files, README, specs, config file format, herdr calls.
- Docs: `docs/adr/0008-sources-behind-adapters.md`, a note in `docs/adr/0002-openspec-cli-as-data-source.md`, `CONTEXT.md`.
