## Context

`work.Item` holds `Name`, `Modified`, `ArchivedOn` and the planning state. `work.Group` splits Items into Phase columns and sorts them: Ready as a queue (`queueRank`, `natural`), Draft, In Progress and Review by `Modified` newest first, Archived by `ArchivedOn`. The Board renders the Card's last line from the planning marks (Draft, Ready) and `age(now - Modified)`. Both Sources build Items, and the cache stores them as JSON under `cache.Version` (2). The filter and the wave marks already keep the selection on its Card across reloads.

See proposal.md for why.

## Goals / Non-Goals

**Goals:**
- Sort the open columns by name or creation date with one key.
- Show the creation date on every open Card.

**Non-Goals:**
- No priority, no dependency order.
- No sort kept across openings, and no config key for the default sort.

## Decisions

### D1: `Created string` on `work.Item`, `YYYY-MM-DD` or ""
It is the same shape as `ArchivedOn`, so a descending string compare sorts newest first with undated last. The Sources set it while they build the Item. OpenSpec reads `.openspec.yaml` with a line scan for `created:` instead of a YAML library, as the file is flat; the value is checked with `time.Parse("2006-01-02", …)`. Spec Kit takes the first `**Created**:` line of `spec.md` that holds a valid `YYYY-MM-DD`, so the template's `[DATE]` is skipped. Both read from the checkout the Item comes from. `cache.Version` goes to 3, so old caches without `created` are dropped once.
Alternative: `time.Time`. That lost because a calendar date has no time zone, and an empty string is the natural "undated".

### D2: `work.Group(items, sort)` with a `Sort` type
`Sort` is `Default`, `ByName` or `ByDate`. `ByName` uses `natural` then the plain compare; `ByDate` compares `Created` descending, then the name. Archived ignores the sort. The Board keeps the sort in its model, `o` cycles it and regroups the loaded Items, and the selection is found again by name, as the filter does.

### D3: Card line `MM-DD · <age>`
The date is shown before the age in the dim style the age has. The year is left out for dates in the current year, judged by the Board's clock (`m.now()`), so the line stays short. When the line is too wide for the Card, the date is dropped before the age is cut.
Alternative: the date only while sorting by date. That lost because the user wants it visible without toggling.

## Risks / Trade-offs

- [OpenSpec changes its metadata format] → No date, so the Card shows the age only and the date sort puts the Change last. Nothing breaks.
- [Changes created before OpenSpec wrote `created`] → The same: they have no date and sort last.
- [Narrow columns] → D3 drops the date first, so the age stays visible as today.
