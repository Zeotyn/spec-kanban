# openspec CLI as the data source

The Board reads Changes through `openspec … --json` (`list`, `status`, `instructions apply`, `show`) instead of parsing `openspec/changes/` itself. The CLI understands schemas, stores and nested specs; a parser of our own would have to re-implement every new OpenSpec layout, which herdr-openspec (Python) and magentic (Go) already get wrong today. We read the filesystem only where the CLI offers nothing: archived Changes and the markdown content of proposal, design and specs.

## Considered Options

- **Own parser** (as in herdr-openspec and magentic): fast, but drifts away from the OpenSpec format.
- **Import openspec as an npm library**: the package only exports config and store functions; the logic behind `list`/`status` is internal.

## Consequences

Every CLI call costs about 0.65 s (Node startup). `list` is enough for all Cards; `status` per Change runs in parallel in the background. Measured on magentic (9 Changes, openspec 1.8.0, 2026-10-29): `list` 0.65 s, `instructions apply` 0.65 s, `status` for all 9 Changes 5.95 s sequentially and 1.45 s with 8 in parallel. The JSON format is not versioned, so all calls sit behind one adapter with contract tests against the real CLI.
