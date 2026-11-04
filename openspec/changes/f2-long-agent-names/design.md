## Context

- `internal/start` names the agent with `agentName(name)`: the name, or `f-` + name when it does not start with a lowercase letter. `prepare` refuses a name whose agent name has more than 32 characters before any herdr call.
- Every herdr call that targets the agent uses `agentName(change)`: `place` (`agent get`), `Prepare` (`agent focus`), `Run` (`agent prompt`), `Launch` and `Resume` (`agent start`, `agent prompt`), and `Watch` (`agent wait`, `agent get`).
- `Running` goes the other way. It reads `agent list` and returns names for the Board: the agent name and `itemName(agent)`, which cuts `f-`. Both are kept because `f-2fa` may be the Change `f-2fa` or the Feature `2fa`.
- The Board (`internal/board`) reads `m.running[c.Name]` for the Running badge and for Resume on `s`. `cmd/kanban` passes `start.Running` to it as the `running` callback.
- herdr 0.9.5 checks agent names with this message (from the binary): `agent name must start with a lowercase letter and contain only lowercase letters, digits, '-' or '_' (1-32 characters)`.

## Goals / Non-Goals

**Goals:**
- Every Change gets a valid agent name, and the same one every time.
- Names that fit today keep their agent name, so running agents and recorded sessions stay reachable after an update.
- One function decides the agent name; nothing reverses it.

**Non-Goals:**
- Other characters herdr rejects, such as upper case or `.`. OpenSpec and Spec Kit names are lower-case kebab-case; such a name still fails at `agent start` with herdr's own error.
- Shortening the tab label, the workspace label or the branch. They keep the full Change name (see "herdr limits" below).
- A configurable length or hash.

## Decisions

1. **Shortening rule.** When the (prefixed) name has more than 32 characters: its first 25 characters, a trailing `-` trimmed, then `-` and the first 6 hex characters of `sha256(change)`.
   - At most 32 characters, and 31 when the cut ends on `-`. Trimming avoids a `--` in the name; it only depends on the name, so it stays deterministic.
   - The hash is of the full Work item name, not the prefixed one. Either is deterministic; the item name is what the user sees.
   - 25 characters keep the readable part of the name, which herdr shows in its sidebar. 6 hex characters (24 bits) make a clash between two long names with the same 25-character start very unlikely.
   - The cut is by byte. Names are ASCII by herdr's rule, so this never splits a character of a valid name.
   - Rejected: a hash for every name. It would rename every running agent and lose the readable name.
   - Rejected: cutting without a hash. Two Changes with the same first 32 characters would share one agent.

2. **`herdr.AgentName(item string) string` in `internal/herdr`.**
   - It is herdr's naming rule, and both `internal/start` and `internal/board` need it. `internal/herdr` is small and has no dependencies of its own, so the Board can import it.
   - Rejected: exporting it from `internal/start`. The Board would then import start, with config, worktree and the Source code behind it.
   - Rejected: `internal/work`. Work items do not know about herdr.

3. **`Running` returns agent names, the Board looks up `AgentName(c.Name)`.**
   - `Running` keeps its signature and drops `itemName`. The set holds exactly what `agent list` reported for the Project.
   - The Board's two lookups (`View` badge, Resume on `s`) become `m.running[herdr.AgentName(c.Name)]`.
   - `itemName` and its test go away.
   - Rejected: passing the item names into `Running` and mapping there. `cmd/kanban` would need `src.Open`, which costs about 0.65 s, and the Board already has its Cards.

4. **No refusal left.** The length check in `prepare` goes away, with `TestPrepareNameTooLong`. Its place is taken by tests of the shortened name in `herdr` and one Start test that sends the shortened name to herdr.

## herdr limits

Checked read-only in herdr 0.9.5 (`--help`, error strings in the binary, CHANGELOG):
- **Agent name:** lowercase letter first, then lowercase letters, digits, `-` or `_`, 1-32 characters. The shortened name meets it.
- **Tab and workspace label (`--label`):** no length limit in the help or the error strings. A long label may be cut in the tab strip, but that is display only.
- **Worktree branch (`worktree create --branch`):** herdr passes it to git, which allows long branch names. No herdr limit found.
- **Worktree path:** herdr derives it from the branch. No limit found; long paths are a filesystem matter far above 36 characters.

## Risks / Trade-offs

- [Two long Changes share the first 25 characters and the 6 hex characters] → One agent would serve both. Very unlikely at 24 bits for the handful of open Changes in a Project. Accepted.
- [The Change `f-2fa` and the Feature `2fa` share the agent name `f-2fa`] → Already true today, and rare: a Project has one Source, and the Board badge just follows the shared agent. Unchanged.
- [An agent started before this change under a refused name] → None exists: such a Change could not be started.

## Open Questions

- Does herdr show a long tab label without trouble, and does `worktree create` accept a 36-character branch? Nothing in the help says otherwise; the smoke test confirms it.
