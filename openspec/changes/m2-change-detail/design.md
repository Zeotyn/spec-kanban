## Context

The Board (`internal/board`) is a single bubbletea Model with columns, `nav` and loading. A Change on the Board knows its name, Progress, planning artifacts and archive date, but not its documents. ADR 0002 keeps the openspec CLI as the data source for state; document content is read from disk, as already planned there. herdr-openspec's browser (Python) did the same view with tabs and its own markdown renderer.

## Goals / Non-Goals

**Goals:**
- Read any Change's documents in the popup with the same vim keys as the Board.
- No extra openspec CLI call when opening details (instant).

**Non-Goals:**
- Editing (`m2-editor-open`), an inline task list on the Board (dropped), Mermaid diagrams, a browser export.

## Decisions

### Detail as a sub-model inside the Board
`internal/detail.Model` handles its own keys and view. `board.Model` holds an optional `*detail.Model`; when set, keys go to it and `View` shows it. `q`/`Esc` in the detail clears it (a `detail.CloseMsg`), so the Board's `nav` is untouched and the selection survives. Alternative: one Model with a mode flag; rejected because detail state (tabs, scroll per tab) would bloat the Board.

### Change directory without a new field
Open Changes live in `openspec/changes/<name>`, archived ones in `openspec/changes/archive/<ArchivedOn>-<name>` (or `<name>` when undated). Both follow from `openspec.Change`, so no `Dir` field and no cache version bump.

### Reading documents
On open, the detail reads `tasks.md`, `proposal.md`, `design.md` and every `specs/**/spec.md` (`filepath.WalkDir`, lexical order). The Specs tab joins them as `# <capability path>` + content. Files are read once per open; `r` is not offered inside the detail (close and reopen to refresh).

### Rendering with glamour
glamour renders markdown with headings, lists, checkboxes and code blocks at a given word-wrap width; writing that by hand is far more code. Runtime uses an embedded style (`internal/detail/terminal.json`) shaped like glamour's `dark` but coloured only with the 16 ANSI colours and the terminal's default foreground, so the user's terminal theme decides the look, light or dark (glamour's built-in styles hard-code 256 colours, and its auto style queries the terminal, which can stall in a popup); tests use the `notty` style so goldens contain no escape codes. Rendered lines are cached per tab and recomputed on `WindowSizeMsg`. The version is the newest glamour that works with lipgloss v1.1.0 already in go.mod.

### Next task
The first line matching `^\s*- \[ \] ` in `tasks.md` gives `next ▸ <text>`; none and at least one `- [x]` gives `all tasks done`. Shown above the rendered tasks.

### Scrolling
Per tab an offset into the rendered lines; `j`/`k`/`ctrl-d`/`ctrl-u`/`gg`/`G` clamp to `[0, len(lines)-height]`. No bubbles/viewport: ~20 lines of code do the same without a dependency.

### Layout

```
m1-board-view  M1  5/11  In Progress
 1 Tasks │ 2 Proposal │ 3 Design │ 4 Specs        (current tab bold and marked ▌)
──────────────────────────────────────────────────
next ▸ 2.3 Implement the navigation state

  1. Model and loading
  ✓ 1.1 …
  …
──────────────────────────────────────────────────
1-4/tab tab · j/k scroll · gg/G top/end · q back
```

Missing documents get a `·` before their tab name.

## Risks / Trade-offs

- [glamour pulls in several packages (goldmark, chroma)] → accepted for correct markdown; binary grows by a few MB.
- [Large specs render slowly on every resize] → rendered once per tab and width, cached.
- [Goldens depend on glamour's output] → pinned version in go.mod; regenerate with `-update` on upgrades.
