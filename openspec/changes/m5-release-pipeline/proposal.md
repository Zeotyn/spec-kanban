## Why

Installing from GitHub currently requires a Go toolchain and nothing verifies changes automatically. A CI pipeline and tagged releases make the plugin installable and keep `main` green.

## What Changes

- GitHub Actions: `go test ./...`, `go vet`, `golangci-lint`, and contract tests with `openspec` installed from npm.
- Tagged releases build macOS and Linux binaries; the `[[build]]` step downloads the matching release binary and only falls back to `go build` when none exists.
- Changelog and version shown by `kanban --version`.

## Capabilities

### New Capabilities

### Modified Capabilities

No behaviour of the plugin changes; this is tooling only (`skip_specs: true`).

## Impact

- `.github/workflows/`, `herdr-plugin.toml` build step, `Makefile`, README install section.
