## Why

`herdr plugin update` refetches the ref recorded at install and nothing else. A user who installed a release with `--ref v0.4.0` never gets a newer release from `update`; they must run `herdr plugin install … --ref v<new>` again. A user who installed without `--ref` follows `main`, which between releases holds unreleased commits and builds from source, so it needs Go. herdr has no option to follow the newest release.

A branch that always points at the newest release gives `update` exactly that: install once with `--ref release`, and every `herdr plugin update` brings the newest release as its downloaded, checksum-verified binary.

## What Changes

- **`release` branch:** after the release job published a `v<version>` release, it moves branch `release` to the tag's commit, fast-forward only. A tag that is not ahead of `release` (for example a fix release of an older version) leaves the branch alone and fails that step visibly.
- **Install docs:** the README installs with `--ref release` and says that `herdr plugin update` then brings each new release; `--ref v<version>` stays documented for pinning one version, and no `--ref` for following `main`.
- **First setup:** branch `release` is created once at `v0.4.0`, so users can switch before the next release.
- No change to `scripts/build.sh`: it already downloads the release binary when the checkout is exactly the commit of the manifest's release.

Non-goals: no change to herdr; no release of anything but tags; no automatic switch of existing installs.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `plugin-host`: a new requirement that installing from branch `release` follows the newest release on `herdr plugin update`.

## Impact

- **CI:** `.github/workflows/pipeline.yml`, release job (full checkout history, one push step).
- **Docs:** README (Install, release steps), CHANGELOG.
- **Repository:** new branch `release` on GitHub.
- No code change, no new dependencies.
