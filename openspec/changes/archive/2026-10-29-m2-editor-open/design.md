## Context

Builds on `m2-change-detail` (detail sub-model, document paths per tab) and the Board's reload with selection by name. The popup is a real terminal, so a terminal editor can take it over.

## Goals / Non-Goals

**Goals:**
- `e` edits the file the user is looking at, with no setup beyond `$EDITOR`.

**Non-Goals:**
- Creating missing documents (that is `/opsx:continue`'s job), GUI editors that return immediately (they work, but the reload happens before the edit).

## Decisions

### Run the editor with `tea.ExecProcess`
bubbletea suspends the program, hands the terminal to the editor and resumes with a message when it exits. The message carries the error, if any, which becomes the notice. Alternative: `herdr pane split` with the editor next to the popup; rejected because the popup is modal and the edit is short.

### Editor command
`$EDITOR` split on spaces (so `code -w` or `nvim -p` work), `vi` when unset, file paths appended. On the Specs tab all spec paths are appended, so vim shows them as buffers.

### After the editor
Detail: re-read the Change's documents (same function as on open) and keep the tab; scroll offset clamps. Board: send the existing reload, which also refreshes the cache; `nav.keep` keeps the selection.

### Notices
A one-line message in the footer area, cleared by the next key. Used for "no design to edit", "archived Changes are read-only", "no tasks to edit" and editor start failures.

## Risks / Trade-offs

- [GUI editors without a wait flag return at once] → the reload shows the old content; documented in the README (`code -w`).
- [Editing an archived Change in place would desync the archive] → `e` is refused for archived Changes.
