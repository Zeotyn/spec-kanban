## Context

m7 put every OpenSpec specific behind `work.Source` (ADR 0008). `cmd/kanban` builds one `openspec.New(openspecCLI)` and hands it to `project.Find`, `loader`/`syncLoad`, `start.Start`/`start.Prepare`, `start.Watched` and `cfg.PromptFor(change, src.Prompt())`. The interface today:

```go
Detect(dir string) bool
Load(project string, others []worktree.Checkout, known map[string]Item) ([]Item, error)
Open(project string) ([]string, error)
ArchivedIn(dir, name string) bool
Docs(project string, it Item) Docs
Rel(name string) string
Prompt() string
Finish(name string) string
```

What the spike against `specify` 0.12.14 (`~/code/speckit-spike`) found, and what decides the shape here:

| Spec Kit fact | Consequence |
|---|---|
| Unit `specs/NNN-slug/` (or `YYYYMMDD-HHMMSS-slug`), `spec.md` first, later `plan.md`, `research.md`, `data-model.md`, `quickstart.md`, `contracts/`, `checklists/`, `tasks.md` | An Item is a folder with `spec.md`, read from disk |
| `tasks.md`: `- [ ] T001 [P] [US1] …`, checked `[x]`/`[X]`, under `## Phase N: …` | Progress = checkbox lines, as OpenSpec's `next ▸` regex already reads them |
| `spec.md` `**Status**: Draft` is never updated | Ignored |
| No state CLI, no archive | Phase from files; Archived needs another signal |
| Current feature from `SPECIFY_FEATURE_DIRECTORY`, `SPECIFY_FEATURE` or the untracked `.specify/feature.json` (not from the branch, since 0.12) | A fresh worktree has no current feature, so the Source must write `feature.json` before the agent starts |
| `herdr worktree create` has no `--env` (only `tab create` has) | An environment variable cannot be set for the new worktree's root pane |
| `.specify/integration.json`: `integration`, `installed_integrations`, `integration_settings.<name>.invoke_separator`; `specify_cli` renders `$speckit-<cmd>` for `codex`/`zcode` skills, else `/speckit<sep><cmd>` | The default prompt depends on the agent kind and the integration |
| `git worktree remove` refuses an untracked `feature.json` (verified), but not an ignored one | The Handoff has to say to delete it |

Config is global (one `config.json` for all Projects), and `kanban sync` serves every herdr workspace, so its Projects can use different tools.

## Goals / Non-Goals

**Goals:**
- A Spec Kit Project gets Board, detail, sidebar, Start and Handoff through the existing seam, with the smallest interface additions it needs.
- Every OpenSpec text, golden file, and herdr/git/openspec call sequence stays as it is. The one exception is the no-project notice, which has to name both markers.
- Rely only on what `specify` 0.12.14 and 1.2.5 share: the `specs/<name>/` layout, `tasks.md` checkboxes, `feature.json`, `integration.json`. Everything else falls back.

**Non-Goals:**
- A Board that mixes both tools in one Project, or a Source per subdirectory.
- A prompt or instructions per Source in config (`prompt` and `instructions` stay global), or per-Source `hide`. Per-Source `instructions` can be added later without breaking the string form.
- Reading `checklists/`, `contracts/`, `data-model.md`, `quickstart.md` (no tabs, no Phase effect), and the `**Status**` line.
- Renaming the Archived column, or "change" in UI texts and `{change}`.
- Writing any Spec Kit state other than `feature.json`. Running `specify`.
- BMAD.

## Decisions

### One Source per Project, chosen in `cmd/kanban`

`cmd/kanban` builds both Sources and a chooser:

```go
// detect accepts a dir any Source recognises; sourceFor picks proj's Source:
// the preferred one (config source) when it recognises proj, else OpenSpec, else Spec Kit.
func sources(cfg config.Config) (detect func(dir string) bool, sourceFor func(proj string) work.Source)
```

`project.Find(cwd, detect)` stays unchanged, and every command calls `sourceFor(proj)` once it has the Project. `sidebar.Sync` gets `detect` and a `syncLoad` that calls `sourceFor(proj)` for each Project. `open` needs only `detect` (the preference only matters once the Project is found). `sync` now loads the config too, for `source` and `base`, and ignores a broken file silently, as `launch` does.

Alternatives: a composite Source that dispatches per call (every method would need the Project, but `Rel`, `Finish` and `ArchivedIn` don't get one); forcing the configured `source` even where that tool is absent (a global setting would then break every other Project); adding `Name()` to the interface (only the chooser needs names, and it has them as map keys).

OpenSpec wins a tie because it is today's behaviour, so nothing changes for this repo or any existing user.

### Interface: one method added, two signatures widened

```go
Prepare(dir, name string) error            // new: make the checkout dir ready for name's agent; OpenSpec: nil
Prompt(project, kind string) string        // was Prompt(); OpenSpec ignores both
Finish(name, checkout string) string       // was Finish(name); checkout "" = no linked worktree; "" = no step
```

- **`Prepare(dir, name)`** is the smallest hook for `feature.json`. It takes the name rather than the Item because `start.Prepare` knows only the name and the Feature folder follows from it (`specs/<name>`). It writes `{"feature_directory":"specs/<name>"}` with `os.MkdirAll` + `os.WriteFile`.
- **`Prompt(project, kind)`**: the implement command depends on the integration installed in the Project and on the kind being started. ADR 0008 already expected "the prompt per agent kind". Both arguments are known where the prompt is built (`startChange`, `showBoard.startOne`).
- **`Finish(name, checkout)`**: Spec Kit has no archive, but in a worktree it must ask to delete the `feature.json` the Board wrote, or `git worktree remove` refuses. Without a worktree there is nothing to say. Only the Source knows that, and only with the checkout. OpenSpec still returns `archive it with /opsx:archive <name>` for both cases, so its texts stay byte-identical.

`Watched.text` drops the clause when `Finish` returns `""`: `X is finished (p). Please review it.` without a worktree, and `…, merge branch X into B, then remove its worktree …` with one.

Alternatives for `feature.json`:
- Setting `SPECIFY_FEATURE_DIRECTORY` in the agent's pane: `worktree create` has no `--env`, and an existing pane's shell cannot be changed.
- Writing the feature into the prompt: Spec Kit's prerequisite script fails before the agent reads arguments.
- Adding `feature.json` to `.git/info/exclude`: that file is shared by all worktrees, so it would silently change the user's main checkout too.
- Doing it in `Launch`: the spawned process knows only pane and tab, not the checkout.

### Where Start calls `Prepare`

`start.Prepare` already decides the checkout. It calls `src.Prepare(dir, change)` on every path that returns a pane, so on a new tab, a reused worktree, a created worktree, and an idle agent's tab, but not on a busy agent, which is only focused:

| Path | `dir` |
|---|---|
| no checkouts / Project on the branch | `project` |
| existing linked worktree (`worktree open`) | its `Path` |
| new worktree (`worktree create`) | `result.worktree.path` from herdr's output (present in the existing fixture) |
| idle agent's tab | the entry of `dirs(project, checkouts)` that contains the agent's `cwd` (longest match), else `project` |

An error returns `prepare <change> in <dir>: <err>` and no pane, so no agent starts. The tab may already exist; the agent-start spec allows a reported failure after a tab was created. Resume also goes through `start.Prepare`, so a resumed Spec Kit session finds its feature too.

### Spec Kit Load: files only, one git call at most

```go
func New(git func(dir string) herdr.Runner, base string) *Source
```

- `Detect`: `.specify` and `specs` are both directories.
- `Load(project, others, known)`: every `specs/<d>/spec.md` in the Project, then every checkout in `others` (not the Project) whose branch `b` has `specs/<b>/spec.md` replaces or adds `b` with `Dir` set. Same rule as OpenSpec, without its "archived in Project" guard, since Spec Kit has no archive folder. `known` is ignored and `Stamp` stays `""`: reading a handful of files costs less than validating a cache, and the Board still opens from the cache snapshot.
- Per Feature: `tasks.md` checkboxes → `Done`/`Total`; newest mtime under the folder → `Modified`; `Artifacts` = `{"proposal":"skipped"}` plus `"specs"`/`"design"`/`"tasks"` = `"done"` for `spec.md`/`plan.md`/`tasks.md`, which `board.planMarks` renders as `- S D T` without a board change.
- Phase: `phaseOf(done, total)` = Draft / Ready / In Progress / Review, then the merged rule.
- Merged rule: only when at least one Feature is in Review, `base := worktree.Base(s.base, git(project))`, then one `git branch --merged <base> --format=%(refname:short)`. A Review Feature whose name is in that set becomes Archived. A git failure leaves everything in Review. Requiring all tasks done is what keeps a just-created branch (tip = base) from archiving a Ready Feature.
- `Open(project)`: Feature names in the Project without git (a merged Feature can still be started from a shell; harmless, the branch is reused).
- `ArchivedIn` → `false`. `Rel(name)` → `specs/<name>`.
- `Docs`: Tasks, Spec, then Plan and Research only when present; `Tasks: 0`, `First: 1`; `Next` from the same checkbox regex. The regex is copied from the OpenSpec package (six lines) rather than shared. Moving it into `work` would put a file format into the domain package.

Cost of the merged rule: `worktree.Base` (1–4 git calls) plus `git branch --merged` (~5 ms) per load, only while a Feature is in Review. Not cached: the cache would need the branch state as its key, which costs the same call.

Alternatives: `git merge-base --is-ancestor <name> <base>` per Feature (one call each, same answer); a merge date for `ArchivedOn` (`git log` per Feature, no user asked for it); "all tasks done → Archived" (then Review would never be reached).

### Default prompt from `integration.json`

`Prompt(project, kind)` reads `<project>/.specify/integration.json`. It uses the integration named `kind` when it is in `installed_integrations`, else `integration`. For `codex`/`zcode` it returns `$speckit-implement`. Otherwise it returns `/speckit` + `integration_settings[name].invoke_separator` + `implement`. A missing file, a missing name or a missing separator gives `/speckit-implement`. The prompt carries no `{change}`: `feature.json` names the Feature, and Spec Kit treats any argument as extra user instructions. herdr kind names equal Spec Kit integration keys except `cursor` (`cursor-agent`), which falls back to the default integration. `kimi`'s `/skill:` form is left to the `prompt` setting.

### Config `instructions`: the Source's command plus free text

The common wish is "the usual command, with these extra words". A global `prompt` cannot express that once a Project can be Spec Kit, because the command differs per Source and per integration. So:

```go
Config.Instructions string   // "" = unset
// PromptFor(change, def): Prompt set → Prompt with {change} replaced (Instructions ignored);
// else def with {change} replaced, + " " + Instructions when Instructions != "".
```

- `def` is the Source's `Prompt(project, kind)` as before, so the append happens in one place for every Source. `instructions` is passed verbatim: no `{change}` replacement, no trimming.
- `prompt` stays a complete override, byte-identical to today, so existing configs keep working.
- Both set: `Parse` keeps the settings, clears `Instructions`, and returns a warning (`instructions is ignored because prompt is set`). `Load` then returns that config with the warning instead of `Default()`. The Board and `kanban start` show it where they show any config warning. Unlike every other problem, this one does not reset settings, because the user's intent (`prompt`) is unambiguous.
- Not a string → wrong type, defaults and the usual warning. `""` is unset and silent.
- One text for all Sources. A per-Source form (e.g. an object keyed by Source) is a later extension if a user mixes tools with different wishes.

Alternative: a `{apply}` placeholder in `prompt` that expands to the Source's command. Rejected: users would have to learn the placeholder, and nothing stops them from still hard-coding `/opsx:apply {change}`, which breaks Spec Kit Starts silently. `instructions` cannot name a command at all, so it cannot be wrong for a Source.

### Config `source`

`Config.Source string`, `""` | `openspec` | `speckit`, validated like `agent`. It is only a preference among Sources that recognise the Project, so it is safe as a global setting.

### Small visible texts

- No-project notice: `No project here.` / `Open the board from a pane inside a project (a directory containing openspec/changes, or .specify and specs).` The `no-project` golden changes.
- `kanban start` outside a Project: `no OpenSpec or Spec Kit project at or above <cwd>`.

### Docs

README gets a "Spec Kit" section: detection and `source`, Phases and the merged-branch rule, planning marks mapping, tabs, `feature.json`, prompt and the global `prompt` caveat, and committing `specs/` before `s`. CONTEXT.md gets **Feature**, and its Source, Project and Archived entries are widened. ADR 0009 records two decisions that are hard to reverse: a merged branch counts as Archived for a Source without an archive, and one Source per Project with OpenSpec winning a tie.

## Risks / Trade-offs

- [`feature.json` in the Project overwrites the user's current feature when `worktree` is off] → that is what `/speckit-implement` needs anyway. It is documented in the README.
- [An untracked `feature.json` blocks `git worktree remove`] → the Handoff asks the Main agent to delete it first. If it is forgotten, the removal fails cleanly and the existing "stop and tell the user" rule applies.
- [The implement agent may commit `feature.json`, so it reaches the base by merge] → harmless (Spec Kit rewrites it on the next `specify`). The Handoff says "unless git tracks it".
- [A Feature finished without a branch, or whose branch was deleted, never reaches Archived] → it stays in Review, as agreed. `hide` cannot hide just those. The README states the fallback.
- [A stale merged branch named like a Feature that is finished on the base archives it] → that is the correct reading of "merged".
- [Global `prompt` applies to Spec Kit too, so an `/opsx:apply {change}` prompt breaks Spec Kit Starts] → `instructions` covers the usual reason to set `prompt`, and the README recommends it; README caveat for `prompt`. A per-Source prompt waits for a user who needs both.
- [kind ↔ integration name mismatch (`cursor`)] → the default integration's command is used. `prompt` overrides it.
- [Spec Kit 1.2.5 may differ from the spiked 0.12.14] → only shared files are read, and every read has a fallback. The smoke test runs on 0.12.14, and 1.2.5 is checked in the open questions below.
- [The cache may hold another Source's Items after `source` changes] → shown until the first refresh. OpenSpec's reuse needs a matching `Stamp`, which Spec Kit Items never have.

## Migration Plan

Nothing for users: config gains two optional keys (`source`, `instructions`), existing `prompt` configs behave as before, the cache format is unchanged, and OpenSpec Projects behave identically. Rollback is a revert. An older binary rejects a config with `source` or `instructions` as an unknown key and falls back to defaults with a warning.

## Open Questions

- Does `specify` 1.2.5 still write `.specify/integration.json` with `invoke_separator`, and still read `.specify/feature.json`? The fallbacks cover a missing `integration.json`. If 1.2.5 dropped `feature.json` resolution, only `Prepare`'s body changes, not the interface. Check when upgrading the spike.
