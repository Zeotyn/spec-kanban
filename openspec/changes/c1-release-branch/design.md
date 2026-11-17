## Context

The pipeline's `release` job runs for `v*` tags after the tests pass. It checks the tag against the manifest version, builds the binaries, writes `checksums.txt` and `commit.txt`, creates the GitHub release and installs from it with `scripts/build.sh`. `build.sh` downloads the binary only when `git rev-parse HEAD` equals the release's `commit.txt`. herdr's `plugin update` refetches the ref recorded at install and skips unchanged commits; its docs do not say whether it reruns `[[build]]`, but an update creates a new installation directory.

See proposal.md for why.

## Goals / Non-Goals

**Goals:**
- `herdr plugin update` on an install from `--ref release` brings the newest release without Go.
- Never rewrite `release` history.

**Non-Goals:**
- No support for fix releases of older versions on `release`.
- No change to `build.sh` or to herdr.

## Decisions

### D1: The release job moves `release` after the release is published
A last step in the `release` job runs `git push origin "$GITHUB_SHA:refs/heads/release"` with the job's `GITHUB_TOKEN` (`contents: write` is already granted). Without `--force`, git refuses a push that is not a fast-forward, which gives the "tag behind the branch" failure for free. The step runs after "Install from the release", so the branch never points at a release whose binary failed to install. The checkout needs `fetch-depth: 0`: from a shallow clone, the push of a new branch can be refused.
Alternatives:
- A separate workflow on `release: published`. That lost because events from `GITHUB_TOKEN` do not start other workflows, and the job already has the token and the commit.
- Force-moving the branch. That lost because an update from a rewritten branch would go backwards.

### D2: Create `release` once by hand at `v0.4.0`
`git push origin v0.4.0^{commit}:refs/heads/release`, done by the user or with the user's go-ahead during apply. Then users can switch now, and the first pipeline push is already a fast-forward.

### D3: Pushing `release` runs the test job once
The pipeline runs on every branch push, so the branch move triggers a test run of the release commit. `GITHUB_TOKEN` pushes do not start workflows, so this only happens for the manual creation in D2. That is harmless and needs no filter.

## Risks / Trade-offs

- [herdr's `update` does not rerun `[[build]]`] → `bin/kanban` would be missing or stale. Check once with a real update from `v0.4.0` to the next release (smoke test). If it fails, the README falls back to `install --ref release` and an issue goes to herdr.
- [A release is published but the branch step fails] → The release exists, and `update` users just do not see it yet. The failed step shows in the pipeline, and the fix is a manual fast-forward push.
