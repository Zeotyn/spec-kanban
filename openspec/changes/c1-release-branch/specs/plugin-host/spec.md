## ADDED Requirements

### Requirement: Branch release follows the newest release
The repository SHALL keep a branch `release` that points at the commit of the newest published release. After a `v<version>` release is published, the release pipeline SHALL move `release` to that tag's commit by fast-forward only; when the tag's commit is not a descendant of `release`, the pipeline SHALL leave the branch unchanged and fail that step. The README SHALL document installing with `--ref release` so that `herdr plugin update` brings each new release.

#### Scenario: Update brings the newest release
- **WHEN** a user installed with `herdr plugin install zeotyn/spec-kanban --ref release` at `v0.4.0`, `v0.5.0` was published since, and the user runs `herdr plugin update spec-kanban`
- **THEN** the plugin is at the commit of `v0.5.0` with its downloaded, checksum-verified binary, and `kanban --version` prints `0.5.0`

#### Scenario: No new release
- **WHEN** the user runs `herdr plugin update spec-kanban` and no release was published since the install
- **THEN** herdr skips the plugin as unchanged, even when `main` has new commits

#### Scenario: Tag behind the branch
- **WHEN** a tag `v0.4.1` is pushed on a commit that is not a descendant of `release`, which points at `v0.5.0`
- **THEN** the release `v0.4.1` is published, `release` still points at `v0.5.0`, and the pipeline shows the failed branch step
