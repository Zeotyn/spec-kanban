## Why

Start, Resume, Verify and Archive type the agent command into a freshly created herdr pane as soon as `tab create`, `worktree create` or `worktree open` returns, but that pane is still running the user's shell rc. herdr's `agent start` has no readiness check of its own: it queues the command text and Enter and waits only for the agent to be detected. Anything in the rc that reads a key, like the oh-my-zsh update prompt (`[Y/n]`), nvm or `p10k configure`, swallows the first character, the shell runs `laude`, and after 60 s the Board shows "claude did not become ready" without saying why. The plugin is MIT open source, so the fix must work for any user's shell and rc, not just one machine's configuration.

## What Changes

- Before every `agent start`, the plugin waits until the pane's shell is at its interactive prompt: it types a probe command whose output differs from its typed echo (`echo kanban-rea""dy-<nonce>`) and waits for that output; a probe that goes unanswered within 2 s is repeated with a fresh nonce, up to 60 s in all.
- The probe starts with a space, so a dialog in the rc that reads one key gets a harmless "no" and the shell keeps the line out of its history.
- When the shell never answers, or `agent start` still fails, the error (and so the `spec-kanban: start failed` / `run failed` / `resume failed` notification) carries the last lines of the pane, so the user sees what the shell was doing.
- Applies to Start, Resume, Verify and Archive alike, since they all start the agent through the same `Launch` and `Resume` functions.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `agent-start`: new requirement "An agent is started only at the shell prompt" (the probe before every `agent start`, for Start, Resume, Verify and Archive); modified requirement "Start reports herdr failures" (the failure names the pane's last lines).

## Impact

- `internal/start/start.go`: `Launch` gains the readiness wait; new helper `ready` (probe loop) and `tail` (pane's last lines for the error).
- `internal/start/watch.go`: `Resume` gains the same wait.
- `internal/start/*_test.go`: the fake herdr runner answers `pane run`, `pane wait-output` and `pane read`.
- herdr commands used newly: `pane run`, `pane wait-output --match --timeout`, `pane read --source recent-unwrapped --lines`.
- README (Start section), CHANGELOG.
- No config key, no new command, no change to the agent-handoff spec: Resume is covered by the new agent-start requirement.
