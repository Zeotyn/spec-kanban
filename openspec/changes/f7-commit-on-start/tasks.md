## 1. Config

- [x] 1.1 Add `commit` (boolean, default `false`) to `internal/config`; a wrong type falls back to the defaults with the usual warning; tests

## 2. Start

- [x] 2.1 Add `commitPlan` in `internal/start` (design D1): only on the base branch itself, only when `git status --porcelain -- <rel>` lists something, `git add -- <rel>` and `git commit -m "docs: plan <change>" -- <rel>`; table tests with a fake git (new Change, edited plan, other branch, remote base, nothing to commit, commit fails)
- [x] 2.2 Call it in Start before the base check when `commit` is on and no Change branch exists; call it in `Wave` for each member (design D3); change the refusal to `<change> is not on <ref>: commit it there, set base, or set "commit": true` in both places; adjust the tests
- [x] 2.3 One integration test with a real temporary git repository: an uncommitted Change plus a staged other file; after `commitPlan` only the Change is committed and the other file is still staged

## 3. Docs

- [x] 3.1 README: `commit` under Settings, and in the worktree section that `"commit": true` commits the plan on Start
- [x] 3.2 CHANGELOG `## Unreleased`: new key `commit`, and the new refusal text

## 4. Verification

- [x] 4.1 `go test -race ./...`, `go vet ./...`, `gofmt -l .` pass
- [x] 4.2 `openspec validate f7-commit-on-start --strict` passes

## Smoke test (user, in herdr)

These have no checkboxes: they do not count as tasks, and only the user runs them.

- Set `"commit": true`, plan a Change without committing it, and press `s` on its Card: `git log -1` shows `docs: plan <change>` and the agent starts in the worktree
