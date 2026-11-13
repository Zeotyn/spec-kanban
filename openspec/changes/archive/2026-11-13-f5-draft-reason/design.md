## Context

`internal/board/mouse.go` builds the footer in `Model.footer()`:
- with marks, it returns the marks summary without actions;
- otherwise it returns the full key footer, or the short one when the full one is wider than the Board, together with the click positions (`hit`) of the actions.

`view()` then puts the filter prompt, the wave question, a notice or the starting message in its place.

Cards carry `Artifacts` (artifact id → `done|ready|blocked|skipped`, nil when the Source did not read it) and `Total`. The OpenSpec Source fills `Artifacts` from `openspec status`. The Spec Kit Source fills them from the files: a missing file is an absent key. `planMarks` draws `P S D T` from the same map.

## Goals / Non-Goals

**Goals:**
- Name the Draft reason from data the Board already has, without another CLI call.
- Keep the footer's clicks right.

**Non-Goals:**
- No change to the Phase rule or to the Sources.

## Decisions

### D1: The reason is a pure function of the Card
`draftReason(c work.Item) string` returns `""` unless `c.Phase == work.Draft`. Otherwise:
- with `c.Artifacts == nil`, it returns `draft: planning unknown`;
- it lists the ids `proposal, specs, design, tasks` whose status is neither `done` nor `skipped` (an absent key counts as missing), as `draft: <ids> missing`;
- with none missing, it returns `draft: no tasks`.

It is tested as a table without a Model.
Alternative: let the Source report the reason. That lost because both Sources already deliver everything the reason needs.

### D2: The prefix goes into `footer()`, shifting the hits
When there are no marks and the selected Card has a reason, `footer()` writes `<reason> · ` first. It measures the full key footer against `width - prefix width` to choose full or short, and adds the prefix width to every `hit`.
Alternatives:
- Prepend the reason in `view()`. That lost because the mouse code reads `footer()`'s hits and would click beside the actions.
- Show the reason instead of the keys. That lost because the user would lose the clickable `q close`.

## Risks / Trade-offs

- [A long reason pushes the short footer off a narrow Board] → The whole footer is cut to the width as before. The reason comes first because it is the new information, and every key still works from the keyboard.
- [A Source reports artifact ids other than the four] → Unknown ids are ignored, the same as the marks do.
