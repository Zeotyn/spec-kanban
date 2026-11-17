## 1. Pipeline

- [x] 1.1 In the `release` job of `.github/workflows/pipeline.yml`, check out with `fetch-depth: 0` and add a last step "Move branch release" that runs `git push origin "$GITHUB_SHA:refs/heads/release"` without `--force` (design D1)
- [x] 1.2 Run `actionlint` on the workflow if available, else check the YAML by reading it once more

## 2. Docs

- [x] 2.1 README Install: install with `herdr plugin install zeotyn/spec-kanban --ref release`, and `herdr plugin update` brings each new release; keep `--ref v<version>` for pinning one version and no `--ref` for `main`
- [x] 2.2 README release steps: the release job moves branch `release`; if that step fails, fast-forward it by hand
- [x] 2.3 CHANGELOG `## Unreleased`: install from branch `release` to get each new release with `herdr plugin update`

## 3. Verification

- [x] 3.1 `openspec validate c1-release-branch --strict` passes
- [ ] 3.2 Ask the user before creating branch `release` at `v0.4.0` on GitHub (design D2); after the go-ahead run `git push origin 'v0.4.0^{commit}:refs/heads/release'` and check that `git ls-remote origin release` shows the commit of `v0.4.0`

## Smoke test (user, in herdr)

These have no checkboxes: they do not count as tasks, and only the user runs them.

- Install with `herdr plugin install zeotyn/spec-kanban --ref release`, and after the next release run `herdr plugin update spec-kanban`: `kanban --version` prints the new version
