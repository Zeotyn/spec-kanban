## Context

Notes are stored in SQLite on the device; the largest test account has 4,000 notes.

## Decisions

### SQLite FTS5 instead of an in-memory index
FTS5 ships with SQLite, survives restarts and answers prefix queries in a few
milliseconds. An in-memory index would have to be rebuilt on every start.

### Search as you type, debounced
Queries run 120 ms after the last keystroke; older results are dropped when a newer
query returns first.

### Accent folding in the tokenizer
`unicode61 remove_diacritics 2` makes `cafe` find `café` without a second column.

## Risks

- Index migration on first start for large accounts → build it in the background and
  show a hint until it is ready
