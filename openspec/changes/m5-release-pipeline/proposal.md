## Why

Installing from GitHub currently requires a Go toolchain and nothing verifies changes automatically. A CI pipeline and tagged releases make the plugin installable without Go and keep `main` green.

## What Changes

- GitHub Actions CI on every push and pull request (Linux and macOS): `gofmt -l`, `go vet`, `go test -race ./...`, with the `@fission-ai/openspec` CLI installed from npm so the contract test runs instead of skipping.
- A release workflow on `v*` tags cross-compiles `kanban` for macOS and Linux (amd64, arm64), uploads the binaries with a `checksums.txt` to a GitHub release, and takes the release notes from `CHANGELOG.md`. It refuses a tag that does not match the manifest `version`.
- The `[[build]]` step runs a small `sh` script: on a checkout of a release tag it downloads the matching binary and verifies its checksum; in every other case, or when anything in that path fails, it builds with `go build` as today.
- `kanban --version` prints the release version stamped at build time, or the Go module version from the build info for source builds.
- `CHANGELOG.md`, README install section (Go-free install of a release).
- Cut from the original scope: `golangci-lint` and a `Makefile` (see design.md).

## Capabilities

### New Capabilities

### Modified Capabilities

No behaviour of the Board, sidebar or Start changes; this is tooling only (`skip_specs: true`). Note: the plugin-host requirement "Plugin builds on install" says an install without Go aborts, which stops being true for release installs (see design.md, Open Questions).

## Impact

- New `.github/workflows/ci.yml`, `.github/workflows/release.yml`, `scripts/build.sh`, `CHANGELOG.md`.
- `herdr-plugin.toml`: `[[build]]` runs the script.
- `cmd/kanban/main.go`: `version` becomes a build-time variable.
- README: install and release sections.
