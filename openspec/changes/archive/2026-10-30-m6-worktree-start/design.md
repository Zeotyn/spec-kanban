## Context

See proposal.md for the motivation. Today `start.Prepare` (run in the Board, or by `kanban start`) checks `herdr agent get <change>`: a running agent in the Project is focused, an exited one's pane and tab are reused, an agent outside the Project is refused; otherwise `herdr tab create --cwd <project> --label <change> --no-focus` returns the root pane and tab. The Board then spawns the detached `kanban launch <change> <pane> <tab> <kind> <prompt>`, whose `start.Launch` runs `tab focus`, `agent start`, `agent prompt`. `start.Running` gives the Board's badges: named agents whose `cwd` lies in the Project. `openspec.Load` runs `openspec list --json` (and `status` for Changes without completed tasks) in the Project; the Board and the sidebar sync share it through the cache in `cmd/kanban/main.go`'s `loader`.

herdr 0.9.5 facts (`herdr worktree <cmd> --help`, `herdr api schema --json`, docs `cli-reference.mdx`, `socket-api.mdx`, `configuration.mdx` at tag v0.9.5):
- `worktree create [--workspace ID | --cwd PATH] [--branch NAME] [--base REF] [--path PATH] [--label TEXT] [--focus|--no-focus] [--trust-repository]` creates a git worktree and opens it as a workspace grouped under the source workspace. An existing local branch is checked out; otherwise the branch is created from `--base` or `HEAD`. Without `--path` the checkout goes to `<worktrees.directory>/<repo>/<branch-slug>` (herdr config `[worktrees] directory`, default `~/.herdr/worktrees`). Result `worktree_created`: `workspace` (`workspace_id`, `label`, …), `tab` (`tab_id`), `root_pane` (`pane_id`, `tab_id`), `worktree` (`path`, `branch`, …).
- `worktree open [--workspace ID | --cwd PATH] (--path PATH | --branch NAME) [--label] [--focus|--no-focus]` opens an existing checkout, or returns the already open workspace. Result `worktree_opened`: same records plus `already_open`.
- `worktree list [--workspace ID | --cwd PATH]` returns `source` (`repo_root`, `source_checkout_path`) and `worktrees` (`path`, `branch`, `is_linked_worktree`, `is_prunable`, `open_workspace_id`, …).
- `worktree remove --workspace ID [--force]` runs `git worktree remove` on an open linked workspace and never deletes the branch.
- `--trust-repository` only grants git's `safe.directory` trust for a repository owned by another user, per request; herdr's agent skill says to pass it only after the user verified the repository.
- Tab ids carry their workspace (`w2:tF`).

## Goals / Non-Goals

**Goals:**
- One git checkout per Started Change, created and opened by herdr, reused on every later Start.
- The Board and the sidebar show the agent's Progress from its checkout without a merge.
- Everything testable with fake runners for herdr, git and openspec; stdlib only.

**Non-Goals:**
- Removing worktrees, deleting branches, merging (m6-agent-handoff).
- Projects in a subdirectory of a checkout (monorepos): they keep today's tab in the Project.
- Changes created in a checkout whose branch is not named like them (for example a `plan-<change>` branch): not shown until merged.
- A plugin setting for the worktree location; herdr's `[worktrees] directory` decides.
- `kanban start` for a Change that exists only in its worktree (it still requires an open Change of the Project).
- Copying uncommitted planning from the Project into the new worktree.

## Decisions

### Checkouts come from `git worktree list --porcelain`
New package `internal/worktree`: `List(project, git) ([]Checkout, error)` with `Checkout{Path, Branch string}`, parsed from `git worktree list --porcelain` (skip bare, detached and prunable entries; `branch refs/heads/<name>` → `<name>`). It returns nothing (no error) unless `git rev-parse --show-prefix` in the Project prints an empty prefix, which is what "the Project is the top directory of a git checkout" means; outside git both calls fail and the list is empty. The git runner is a `herdr.Runner` built like `openspecCLI` (`herdr.Run("git", proj, …)`), so tests use the same fake. Alternative: `herdr worktree list` (gives `open_workspace_id` too, but the sidebar sync and the loader should not need herdr to read Changes, and `worktree open` reports `already_open` anyway).

### Prepare chooses the checkout
`Prepare(project, change string, wt Worktree, run, git herdr.Runner)` with `Worktree{On bool; Base string}` from the config:
1. `checkouts := worktree.List` (only when `wt.On`).
2. `agent get <change>` as today; the "other Project" check accepts a `cwd` inside the Project or inside any checkout's path.
3. No agent: pick the checkout whose branch equals the Change.
   - none and `wt.On` with checkouts → choose the start ref: `<change>` when `git rev-parse --verify --quiet refs/heads/<change>` succeeds, else `worktree.Base`. Check `git cat-file -e <ref>:openspec/changes/<change>`; on failure return `<change> is not on <ref>: commit it there or set base`. Then `herdr worktree create --cwd <project> --branch <change> --base <base> --label <change> --no-focus` → `root_pane.pane_id`, `tab.tab_id`.
   - a linked checkout → `herdr worktree open --cwd <project> --path <path> --label <change> --no-focus`; when `already_open`, `herdr tab create --workspace <workspace.workspace_id> --cwd <path> --label <change> --no-focus` (the open workspace's root pane may be in use); otherwise its root pane and tab.
   - the Project's own checkout, `wt.On` false, or no checkouts → today's `tab create --cwd <project>`.
`Launch` and the `kanban launch` arguments do not change: focus still moves only in the background launch, so `--no-focus` everywhere. No `--path`: herdr's configured location is where the user already looks for worktrees, and the plugin needs no setting for it. No `--trust-repository`: a repository the user works in is owned by them; for one that is not, herdr's error ("dubious ownership") is shown and the user decides (`worktree: false`, or trusting it in git). Alternatives: `--path <project>-<change>` next to the Project (clutters the parent dir, duplicates a herdr setting); plain `git worktree add` plus `tab create` (herdr would not group the workspace and we would parse paths ourselves).

### Base detection
`worktree.Base(configured, git)`: `configured` when set; else `git symbolic-ref --short refs/remotes/origin/HEAD` (`origin/main`) → the local branch of that name when `refs/heads/<name>` exists, else the remote ref itself; else the first of `main`, `master`, `develop` that exists locally; else `git branch --show-current`; `HEAD` when detached. The local branch wins over `origin/<name>` because the user plans and merges locally first: a Change planned and merged into local `main` but not pushed must still be on the base. The Change-on-base check turns the remaining gap (planning not committed, or only pushed elsewhere) into a clear error instead of an agent that cannot find its Change. Alternative: `origin/main` as is (the user's spec wording; fails whenever local `main` is ahead).

### Progress comes from the Change's checkout
`openspec.Change` gains `Dir string` (`json:"dir,omitempty"`): the checkout the Change's files are read from, `""` for the Project. `openspec.Load(project, cli func(dir string) herdr.Runner, others []worktree.Checkout, known)`:
- `others` are the checkouts except the Project's own; for each whose `<path>/openspec/changes/<branch>` is a directory (one `os.Stat`, no process), `openspec list --json` runs in that path, in parallel with the Project's list.
- From that list only the Change named like the branch is taken, with `Dir = path`. After the archive is read, it replaces the Project's open Change of that name, or is appended when the Project has no Change (open or archived) of that name. A failed list or a missing entry leaves the Project's Changes.
- `status` and the file fingerprint run in `Dir` when set (`setup` stamps that checkout's config); the cache reuse also requires an equal `Dir`.
- `docs.Dir` uses `c.Dir` instead of the Project when set, so the detail view and `e` show and edit the worktree's files.
The loader in `main.go` computes `others` with `worktree.List` on every load, so the sidebar sync and the Board get it through the shared cache without further changes; a worktree workspace is a Project of its own for the sidebar and is read as before. Cost: one fast git call per load, plus one `openspec list` per worktree that holds its Change (parallel to the Project's list). Alternatives: read `tasks.md` of the worktree directly (duplicates OpenSpec's counting, against ADR 0002); `herdr worktree list` (needs herdr for a read path); overlay every Change of every worktree (stale copies of other Changes would win over the Project).

### Running agent badge
`start.Running(dirs []string, run)` counts a named agent whose `cwd` lies in any of `dirs`; the Board's lookup passes the Project and its checkouts' paths, read with `worktree.List` at lookup time (a new worktree appears without reopening the Board). The same `inside`-any check serves Prepare's other-Project refusal. Alternative: map the agent's `cwd` to a branch and require branch == name (stricter, more code, no observed need).

### Config
`worktree` (`*bool`, default `true`) and `base` (`*string`, non-empty) in `config.Parse`, carried as `Config.Worktree bool` and `Config.Base string`. Unknown or wrongly typed values already fall back to defaults with the warning.

### Docs and glossary
ADR 0007 "Start in a git worktree" (supersedes the tab location of ADR 0005, keeps its agent API). CONTEXT.md: **Start** "hands a Change to a new agent session in its own tab, in the Change's checkout"; new term **Change checkout**: the git checkout whose branch is named like a Change, where Start runs its agent, usually a linked worktree (_Avoid_: sandbox, clone). README: worktree Start, where checkouts go (herdr's `[worktrees] directory`), `worktree` and `base`, that nothing is removed automatically yet.

### What m6-agent-handoff must adapt
m6-agent-handoff is planned against today's tab-in-Project Start; after this change it must:
1. Handoff text: after review the Main agent merges branch `<change>` into the base in the Project, runs `/opsx:archive <change>` there unless the agent archived it in the worktree, and removes the checkout. The text needs the checkout path and base, so `Watch` gets them (from `agent get` `cwd` and `worktree.Base`).
2. Cleanup command: `herdr worktree remove --workspace <id>` needs the worktree workspace open, but the Handoff's `pane close <agent pane>` closes a single-tab worktree workspace. Either keep the workspace (close nothing, or close only after removal) and put `herdr worktree remove --workspace <id>` in the text, or name `git worktree remove <path>` (works without a workspace). Never `--force`, never delete the branch automatically.
3. Clean check: `openspec list` in the Project shows stale Progress; use `openspec.Load` with the checkouts (this change) so Progress comes from the worktree. A Change the agent archived in the worktree is not in the Project's archive until merged: look for it under `<checkout>/openspec/changes/archive/` too.
4. Resume (`s` on Review) calls the new `Prepare` signature, so it opens the Change's checkout; `kanban resume` needs no other change.
5. Its MODIFIED board "Start a Change from the Board" must be rebased on this change's version (scenario "Start a Ready Card" now opens a worktree workspace).

## Risks / Trade-offs

- [`tab focus` on a tab of the new worktree workspace might not switch the workspace] → smoke test; if it does not, `Launch` first runs `herdr workspace focus <id>` with the id taken from the tab id prefix (`w2:tF` → `w2`), no argument change.
- [Planning edited in the Project but not committed is missing in the new worktree] → the Change-on-base check catches a missing Change; edits to an existing one are not detected (accepted; README says to commit planning before `s`).
- [The worktree's checkout is not ready for the agent (dependencies, build caches, `.env`)] → accepted; herdr's `worktree.created` event hooks can bootstrap it, outside this plugin.
- [A stale worktree (branch already merged, not removed) keeps overriding the Project's open Change] → only until archived in the Project; m6-agent-handoff removes checkouts.
- [Extra `openspec list` per active worktree slows each load] → bounded by active worktrees and run in parallel with the Project's list.
- [Two Projects in one repository share change names] → not supported (Projects must be the checkout's top directory).
- [`worktree.List` fails for a real git problem and Start silently uses the Project] → accepted; the tab's location shows it, and git errors in create/open are reported.

## Migration Plan

Rebuild the plugin. Running agents started in the Project keep working: their tabs are reused by `agent get` as before. The next Start of a Change without an agent creates its worktree from the base. `{"worktree": false}` restores the old behaviour. Rollback: revert; created worktrees stay as normal git worktrees and branches.

## Open Questions

- Does `herdr tab focus` switch to the tab's workspace? (Smoke test; fallback in Risks.)
- Does `worktree open --path` of an already open workspace keep focus where it is with `--no-focus`? (Smoke test.)
