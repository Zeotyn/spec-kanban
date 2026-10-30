## Why

m7 drew the `Source` seam so a second spec tool could be added without touching every package. Spec Kit is the next one: its projects keep their Work items as `specs/NNN-slug/` folders with `spec.md`, `plan.md` and `tasks.md`, and its users want the same Board, sidebar, Start and Handoff that OpenSpec projects have. The spike against `specify` 0.12.14 showed what the seam is still missing: Spec Kit has no state CLI and no archive, its implement command depends on the installed agent integration, and its scripts find the current feature through an untracked `.specify/feature.json`, which a fresh worktree does not have.

## What Changes

- New Source `internal/source/speckit`. A Spec Kit Project is a directory with `.specify/` and `specs/`. Each `specs/<name>/` holding a `spec.md` is one Work item, a **Feature**. Progress comes from the checkbox lines in `tasks.md` (`[x]` or `[X]` counts as done); the modification time is the newest file in the folder.
- Phases of a Feature: Draft without tasks (only `spec.md`, or `plan.md` without `tasks.md`), Ready with tasks and none done, In Progress, Review when all tasks are done. A Feature is **Archived** when it would be in Review and its branch `<name>` is merged into the base. Without such a branch it stays in Review. `spec.md`'s `**Status**` line and `checklists/` are ignored.
- A Feature worked on in another checkout (on branch `<name>`) is read from there, as OpenSpec Changes are.
- Detail tabs of a Feature: Tasks and Spec always, Plan and Research when their file exists. `e` edits the shown file. Planning marks reuse `P S D T`: `P` skipped, `S` = `spec.md`, `D` = `plan.md`, `T` = `tasks.md`.
- Start: before the agent starts, the Source prepares the Feature's checkout. Spec Kit writes `.specify/feature.json` with `{"feature_directory":"specs/<name>"}`, and OpenSpec does nothing. The default prompt is Spec Kit's implement command, built from `.specify/integration.json` for the agent kind being started (`/speckit-implement`, `$speckit-implement`, `/speckit.implement`), with `/speckit-implement` as the fallback. A configured `prompt` still overrides it.
- New config key `instructions` (string, optional): text appended to the Source's command after one space, verbatim, so one setting works for OpenSpec (`/opsx:apply <change> <instructions>`) and Spec Kit (`/speckit-implement <instructions>`). Empty means unset. `prompt` stays a complete override: when it is set, it wins as today and `instructions` is ignored; setting both shows the config warning but keeps the settings. One text for all Sources; per-Source instructions are a possible later extension.
- Handoff: a Spec Kit Feature has no archive step. In a worktree the Main agent is asked to delete the `.specify/feature.json` the Board wrote (unless git tracks it) before `git worktree remove`.
- One Source per Project. `project.Find` accepts either marker. When a directory holds both, OpenSpec wins unless the new config key `source` (`openspec` | `speckit`) prefers Spec Kit. A board that mixes both tools is out of scope.
- `work.Source` changes: new `Prepare(dir, name string) error`, `Prompt(project, kind string) string` (was `Prompt()`), and `Finish(name, checkout string) string` (was `Finish(name)`, and `""` now means no step). OpenSpec's texts and calls stay byte-identical.
- The no-project notice and the `kanban start` error stop saying "OpenSpec" only.
- Docs: README section "Spec Kit", CONTEXT.md (**Feature**, Source, Project, Archived), ADR 0009 (merged branch counts as Archived; one Source per Project).

## Capabilities

### New Capabilities
- `speckit-source`: how Spec Kit Features are found, read, phased (with the merged-branch rule for Archived), shown as tabs, prepared for Start (`feature.json`, default prompt) and handed off, and which Source a Project gets.

### Modified Capabilities
- `agent-start`: Start lets the Project's Source prepare the checkout before the agent starts, and the default prompt is the Source's for the agent kind, followed by `instructions` when set.
- `agent-handoff`: the finishing step in the Handoff text comes from the Source and may be empty (Spec Kit has no archive).
- `kanban-config`: new keys `source` and `instructions`. Without `prompt`, Start sends the Source's command plus `instructions`.
- `plugin-host`: the open action finds a Project by any Source's marker, and the notice outside a Project names both.
- `change-detail`: the tabs come from the Work item's Source; OpenSpec keeps its four.

## Impact

- New: `internal/source/speckit/` (Source, load, phase, docs, prompt, prepare, with tests and fixtures), `docs/adr/0009-speckit-source.md`.
- Changed: `internal/work` (Source interface), `internal/source/openspec` (no-op `Prepare`, new `Prompt`/`Finish` signatures), `internal/start` (`Prepare` calls the Source's `Prepare` with the checkout dir, and `Watched.text` drops an empty finishing step), `internal/config` (`source`, `instructions`, `PromptFor` appends `instructions`, a kept-settings warning when both `prompt` and `instructions` are set), `internal/board` (no-project notice + golden), `cmd/kanban` (Source choice per Project, `detect` for any Source), README (Settings example uses `instructions`), CONTEXT.md, CHANGELOG.
- Unchanged: cache format (version 2), Board columns and Card layout, sidebar, every OpenSpec text, golden and herdr/git call sequence other than the notice.
- Cost: at most one `git branch --merged` per Spec Kit load (only when a Feature has all tasks done), plus the base detection's git calls. No new dependency.
