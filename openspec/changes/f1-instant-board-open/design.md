## Context

Measured on 2026-10-29: `openspec.Load` takes 1.45 s in this repo (12 Changes, 7 `status` calls) and 1.36 s in magentic (10 Changes, 3 `status` calls); herdr needs ~0.8 s to show the popup before our binary runs. Each openspec CLI call is a Node start (~0.65 s); there is no batch `status`. ADR 0002 keeps the CLI as the data source, so reading planning state from disk is out.

## Goals / Non-Goals

**Goals:**
- Board visible with data ~0.8 s after `prefix+o` for any Project opened before.
- Refresh usually costs one `openspec list`.
- Never trust cached data that fails validation.

**Non-Goals:**
- Speeding up the first open of a Project, herdr's popup start, or `openspec list` itself.
- Sharing the cache with the sidebar (M3 may reuse it later).

## Decisions

### Location and key
`$HERDR_PLUGIN_STATE_DIR` when set (herdr creates it per plugin), else `os.UserCacheDir()/openspec-kanban`. One file per Project, named by the first 16 hex characters of the SHA-256 of the absolute Project path, plus `.json`. The path inside the file is the authority; the hash only spreads files.

### File format

```json
{"version": 1, "project": "/abs/path", "changes": [ <openspec.Change as JSON> ]}
```

`openspec.Change` gets JSON tags; `Planning` is stored as its integer. A version bump is the escape hatch whenever `Change` or the reuse rule changes shape.

### Validation (all-or-nothing per file)
`cache.Read(dir, project)` returns the cached Changes or nothing. It returns nothing when the file is missing or unreadable, JSON fails to parse, `version` differs, `project` differs, or any entry has an empty name. Per-entry trust for planning reuse is decided in `openspec.Load` (next decision), because only `list` knows the current modification time.

### Reuse inside `openspec.Load`
`Load(project, run, known map[string]Change)`: for an open Change with `Done == 0`, if `known[name]` exists, has `Planning != Unknown`, and `Modified.Equal(listed.Modified)`, and `known[name].Files` equals the Change's current file fingerprint, copy its `Planning` and `Artifacts` instead of calling `status`. `list` reports the newest *file* mtime (openspec `core/list.js` `getLastModified`), so edits and new files change it, but deleting or renaming an older file does not. `Files` closes that gap: a short SHA-256 of the sorted relative file paths under `openspec/changes/<name>/`, computed in Go for Changes without completed tasks (a directory walk, microseconds). The fingerprint also covers what `status` depends on outside the Change: size and mtime of `openspec/config.yaml`, of every file under `openspec/schemas/`, and the openspec CLI version (`openspec --version`, run in parallel with `list`, so it adds no wall time; the CLI is often behind a version-manager shim whose file never changes), computed once per load. A zero `Modified` or an empty fingerprint never matches. `nil` keeps today's behaviour, so existing callers and tests stay valid.

### Stale-while-revalidate in the Board
`board.New` gets an optional initial snapshot. With one, the Model starts `loaded` and `loading` (header "refreshing…") and still runs the loader in `Init`; the fresh result goes through the existing `loadedMsg` path, so `nav.keep` keeps the selection by name. Without one, the Board behaves as today.

### Writing
After a successful load, `kanban board` writes the result: marshal → temp file in the same directory → `os.Rename`. Errors are ignored (spec: cache problems never break the Board). The write happens in the loader function, not in the Board, keeping the Board free of I/O.

### Wiring

```
kanban board
  dir      := cache.Dir()
  snapshot := cache.Read(dir, proj)                       // validated or nil
  load     := func() { cs, err := openspec.Load(proj, run, byName(snapshot)); if ok { cache.Write(dir, proj, cs) }; return cs, err }
  board.New(proj, load, time.Now, snapshot)
```

The loader reads the newest snapshot each time (`cache.Read` again on `r`) so that a refresh after a refresh also skips unchanged `status` calls.

## Risks / Trade-offs

- [Stale view for ~0.7 s] → marked "refreshing…" in the header; Cards cannot be acted on in M1 anyway, Start (M4) must wait for fresh data — noted for `m4-start-action`.
- [A Change deleted or archived since the last load shows briefly] → replaced on swap; acceptable for under a second.
- [Two Boards writing at once] → atomic rename, last writer wins; both results are valid loads.
- [`openspec --version` fails] → no fingerprint, so nothing is reused; loads then cost as much as without a cache.
- [Global or user-level openspec config and schemas outside the Project] → not fingerprinted; after changing them, bump `cache.Version` or delete the cache files.
- [Content edited while deliberately keeping mtimes (`touch -r`, `cp -p`, tar extraction)] → not detected; `status` only checks which files exist, so only a renamed schema or a changed `.openspec.yaml` with a restored mtime could slip through. Accepted.
- [Reading the file list touches the filesystem, which ADR 0002 avoids] → it only validates the cache; planning itself still comes from the CLI.
