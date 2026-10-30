## Context

`internal/openspec` defines `Change` (name, Progress, `Modified`, `Planning`, `Artifacts`, `Archived`, `ArchivedOn`, `Files`, `Dir`) and `Load(project, cli, others, known)`, plus `OpenNames` and `ArchivedIn`. `Change` is imported by `board`, `cache`, `detail`, `docs`, `phase`, `sidebar`, `start` and `cmd/kanban`. The OpenSpec specifics outside that package:

| Where | What |
|---|---|
| `project.Find` | marker `openspec/changes` |
| `docs.Dir`, `docs.Read` | `openspec/changes[/archive/<date>-]<name>`, tabs Tasks/Proposal/Design/Specs, `next ▸` line, spec files for `e` |
| `phase.Of` | change-phases rules over `Done`/`Total`/`Planning`/`Archived` |
| `board.edit`, `detail.edit` | `tasks.md`, `<tab>.md` in the Change folder |
| `start.Prepare` | `git cat-file -e <ref>:openspec/changes/<change>` |
| `start.checkOpen` | `openspec.OpenNames` |
| `start.Watched.check`/`text` | `openspec.ArchivedIn`, `/opsx:archive <change>` |
| `config.Default` | prompt `/opsx:apply {change}` |
| `cmd/kanban` | `openspecCLI`, `openspec.Load` in `loader` |

`config.Hide` is indexed by `phase.Phase` and `hide` accepts the Phase names, `Archived` among them. Planning marks on Draft/Ready Cards (`P S D T`) read `Artifacts` by the OpenSpec artifact ids. The milestone badge and the `m<N>` filter come from the Change name (`board.milestone`), a naming convention of this repo, not of OpenSpec.

## Goals / Non-Goals

**Goals:**
- One package boundary: nothing outside `internal/source/openspec` knows an OpenSpec path, CLI call, artifact rule or slash command, except the Card's planning marks (below) and user-facing text that already says "change".
- Every golden file, every spec scenario and every herdr/git/openspec call sequence stays as it is.
- Each task group leaves `go test ./...` green.

**Non-Goals:**
- A second Source, Source selection or auto-detection among several tools.
- Renaming the Archived column, "change" in UI texts, `{change}` in the prompt, or the `kanban start <change>` argument.
- A `Group` field for milestones/epics: the badge stays name-derived until BMAD needs epics.
- Per-agent-kind prompts (Spec Kit needs them later, see Future sources).
- Moving the planning marks out of `board`.

## Decisions

### `internal/work` holds the domain; `internal/phase` goes away

```go
package work

type Phase int // Draft, Ready, InProgress, Review, Archived; String() as today
func Group(items []Item) [Phases][]Item // column order and Card order, unchanged

type Item struct {
	Name       string            `json:"name"`
	Done       int               `json:"done"`
	Total      int               `json:"total"`
	Modified   time.Time         `json:"modified"`
	Phase      Phase             `json:"phase"`               // decided by the Source
	Artifacts  map[string]string `json:"artifacts,omitempty"` // planning artifact id → done|ready|blocked|skipped, nil when not read
	ArchivedOn string            `json:"archivedOn,omitempty"`
	Stamp      string            `json:"stamp,omitempty"`     // the Source's reuse key, "" is never reused
	Dir        string            `json:"dir,omitempty"`       // checkout the Item is read from, "" for the Project
}
```

`Item.Phase` would import `phase` and `phase.Group` would import `work`; moving the Phase type and `Group` into `work` avoids the cycle and leaves `phase` empty, so it is deleted. `Archived bool` is dropped: `Phase == work.Archived` says the same. `Planning` and `Files` are OpenSpec internals and leave the shared type (see cache below).

Alternatives: keep `phase` and put `Item` there (a "phase" package owning the domain type reads wrong); keep `Archived` next to `Phase` (two fields that can disagree).

### The Source decides the Phase

`phase.Of` becomes an unexported `phaseOf` in `internal/source/openspec`, with its table tests, and `Load` sets `Item.Phase`. BMAD reports a status per story and Spec Kit derives one from its own files, so the rule cannot stay central. The change-phases spec is unchanged: it describes OpenSpec Changes, which is exactly what the rule still applies to.

### `Source` has only the methods the current code calls

```go
type Source interface {
	Detect(dir string) bool                                                           // project.Find: dir holds this tool's work
	Load(project string, others []worktree.Checkout, known map[string]Item) ([]Item, error) // loader
	Open(project string) ([]string, error)                                            // kanban start: names of the open Items, without planning reads
	ArchivedIn(dir, name string) bool                                                 // Handoff: archived inside a checkout
	Docs(project string, it Item) Docs                                                // detail tabs, board e
	Rel(name string) string                                                           // start ref check: an open Item's path below a checkout
	Prompt() string                                                                   // default Start prompt, with {change}
	Finish(name string) string                                                        // Handoff instruction after review
}
```

`openspec.New(cli func(dir string) herdr.Runner) *Source` implements it; the CLI runner moves from `Load`'s parameters into the constructor. `cmd/kanban` builds one Source and passes it to `project.Find`, `loader`, `start.Start`, `start.Prepare`, `start.Watched` and the config prompt. A compile-time assertion in the openspec tests keeps the type on the interface.

- `Open` stays separate from `Load` because `kanban start` checks only names and must not pay for `openspec status` calls.
- `Rel` replaces a richer `Exists(git, ref, item)`: start keeps its git call, the Source only names the path (`openspec/changes/<name>`), which `Load` also uses for its checkout check.
- `Finish` returns `archive it with /opsx:archive <name>`; `Watched.text` keeps every other word, so the Handoff text is byte-identical.
- `Prompt` returns `/opsx:apply {change}`. `config.Default` gets an empty `Prompt` meaning "the Source's", and `PromptFor(change, def string)` falls back to `def`. Validation of a configured prompt is unchanged.

Alternatives: no interface until the second Source (one adapter is a hypothetical seam, but the user wants the seam drawn now and the fakes in `start` tests benefit); a `Kind`/`Name` method, `Prompt(kind)`, status writes, or grouping (not needed by any caller today).

### Docs as tabs

```go
type Doc struct { Title, Body string; Files []string } // Files: what e opens, the first one also when Body is ""
type Docs struct {
	Tabs  []Doc
	Tasks int    // tab holding the tasks; shown first when Total > 0, prefixed by Next
	First int    // tab shown first otherwise
	Next  string // "next ▸ …", "all tasks done" or ""
}
```

OpenSpec returns Tasks, Proposal, Design, Specs (Tasks 0, First 1) from the folder `docs.Dir` computes today. `detail.New` takes a `read func() work.Docs` (re-called after the editor) instead of a directory; the tab bar, `1-4` footer and "_This change has no …_" texts derive from `Tabs` and come out identical for four tabs. `board.edit` opens `Tabs[Tasks].Files[0]` with the existing `os.Stat` check. Spec Kit (spec, plan, tasks, research) fits without touching `detail`.

### Cache version 2

The stored type changes from `openspec.Change` to `work.Item` (`phase`, `stamp` in, `planning`, `files`, `archived` out), so `cache.Version` becomes 2 and a version 1 file is ignored, as the change-cache spec already requires. Validation checks a non-empty name and a Phase from Draft to Archived (the "unknown planning value" check of the spec, now on the value that carries planning).

Planning reuse in `openspec.Load` keeps the same rule with the new fields: reuse when the known Item has the same `Dir`, the same non-zero `Modified`, the same non-empty `Stamp` (the old `Files` fingerprint) and non-nil `Artifacts` (the old `Planning != Unknown`: `status` returns no artifacts exactly when it fails); then `Phase` and `Artifacts` are copied. Copying `Phase` is safe because an equal `Modified` and fingerprint mean equal tasks, hence equal `Done`/`Total`.

Alternative: an opaque `json.RawMessage` per Item for Source state (more general, but nothing needs it yet).

### Archived keeps its name

Renaming the column to "Done" now would change the Board, the sidebar, the `hide` values, card text, README and goldens, for a Source that does not exist yet; Spec Kit and BMAD have no archive and will map their finished state to Review or a renamed column when they arrive. Recorded in ADR 0008.

### Steps that keep the tests green

Move first (`internal/openspec`, `internal/docs` → `internal/source/openspec`), then add `work` next to the old types with a temporary `openspec.Items([]Change) []work.Item` used in `cmd/kanban`, switch consumers one at a time (sidebar, board+config, detail, start/watch), then switch `Load` and the cache together and delete `Change`, `Items` and `phase`, then put the remaining specifics behind `Source`. Goldens are the regression net: a task that changes one is wrong.

## Future sources

Not built here; listed so the seam fits them.

- **Spec Kit 1.2.5**: an Item is `specs/NNN-slug/` (`spec.md`, `plan.md`, `tasks.md`, `research.md`, `contracts/`, `checklists/`). No state CLI: Progress from `- [ ] T012 [P] [US1] …` lines in `tasks.md`, grouped under `## Phase N:`. No archive or done concept, so nothing reaches Archived. Implement command differs by agent kind: `/speckit-implement` (Claude, Copilot), `$speckit-implement` (Codex), `/speckit.implement` (Gemini) → `Prompt` will need the kind. Branch `{number}-{slug}` only with the optional git extension, so worktree reads fit when it is used.
- **BMAD v6.12**: `_bmad-output/implementation-artifacts/sprint-status.yaml`, `development_status`: stories `N-M-slug` with `backlog|ready-for-dev|in-progress|review|done`, epics `backlog|in-progress|done`. The Phase is explicit (stories map one to one onto columns), status also sits in the story file's frontmatter (two places to reconcile). Epics become the group the filter uses → the `Group` field arrives then. Implement command `/bmad-build`. No archive. Only v6; v4 and the unreleased v7 lay files out differently.
- Consequences already in the seam: the Source sets Phase; `Docs` is a tab list; `Rel` names any path; `Finish` and `Prompt` are per Source. Still to come with them: the prompt per agent kind, a group field, and choosing among several detected Sources.

## Risks / Trade-offs

- [Phase stored as an int in the cache] → reordering the constants would misplace cached Cards; the existing rule "bump the version when the stored shape changes" covers it, noted on the `Phase` type.
- [First Board open after the upgrade has no snapshot] → one load as without a cache; the next sync writes version 2.
- [Detail tabs generalised without a second tab set to test against] → table test with a three-tab fake `Docs` in `detail`, besides the unchanged goldens.
- [Interface with one implementation] → every method has a caller today; the fakes in `start` tests use it.
- [Planning marks still name OpenSpec artifacts in `board`] → unchanged for OpenSpec Items; a Source without those ids shows `· · · ·`, to revisit with Spec Kit.

## Migration Plan

Nothing for users: cache version 2 replaces the old file on the next load; config files stay valid. Rollback is a revert; an older binary ignores a version 2 cache the same way.
