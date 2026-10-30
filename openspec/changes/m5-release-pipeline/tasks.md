## 1. Version (TDD)

- [x] 1.1 Write a failing test for the version string: a stamped `version` is printed as is; an empty one falls back to the build info's main module version
- [x] 1.2 Make `version` a `var` set by `-ldflags -X main.version=…` with the build-info fallback until the test passes; check `go build -ldflags "-X main.version=9.9.9" -o /tmp/k ./cmd/kanban && /tmp/k --version` prints `9.9.9`

## 2. Install script

- [x] 2.1 Write `scripts/build.sh` as in design.md (version from the manifest, OS/arch mapping, `commit.txt` must equal `git rev-parse HEAD`, checksum check with `sha256sum` or `shasum -a 256`, `--version` check, atomic `mv`, `go build` fallback with a reason on stderr); `shellcheck` clean
- [x] 2.2 Check the fallback paths by hand with PATH shims for `uname` and `curl` in a scratch copy of the checkout: unknown arch, no release for the version (curl fails), commit mismatch, checksum mismatch, binary printing the wrong version; each ends in `go build` with its reason and leaves no temp file in `bin/`
- [x] 2.3 Point `[[build]]` in `herdr-plugin.toml` at `sh scripts/build.sh`

## 3. CI

- [x] 3.1 Add `.github/workflows/ci.yml`: push and pull_request, matrix ubuntu/macos, setup-go from `go.mod`, setup-node, `npm install -g @fission-ai/openspec@1.8.0`, `openspec --version`, `gofmt -l` empty, `go vet ./...`, `go test -race ./...`, `sh scripts/build.sh`; `shellcheck scripts/build.sh` on Ubuntu
- [ ] 3.2 Push to `main` (no PRs while the user is the only contributor) and confirm both jobs pass and the contract test ran (`go test -v -run TestContract ./internal/openspec` output in the log shows PASS, not SKIP)

## 4. Release

- [x] 4.1 Add `.github/workflows/release.yml` as in design.md: tag = `v` + manifest version check, changelog section as notes, four cross-compiled binaries with `-X main.version`, `checksums.txt`, `commit.txt`, `gh release create`, verification by running `scripts/build.sh` in the tagged checkout
- [x] 4.2 Add `CHANGELOG.md` with a `## 0.1.0 - <date>` section summarising M0–M5

## 5. Docs

- [x] 5.1 README: requirements (Go only for unreleased commits), install of a release with `--ref v<version>` and upgrading by installing the new tag, install of `main` with Go, a short release procedure (bump manifest, changelog, tag, push)

## 6. Verification

- [x] 6.1 `go test -race ./...`, `go vet ./...`, `gofmt -l .` and `shellcheck scripts/build.sh` pass locally
- [ ] 6.2 Release smoke test by the user: tag `v0.1.0` on `main` and push it; the release workflow passes and the release has four binaries, `checksums.txt`, `commit.txt` and the changelog notes; `herdr plugin uninstall`/`unlink` the local plugin, `herdr plugin install zeotyn/openspec-kanban --ref v0.1.0` with Go hidden from PATH succeeds by download and `kanban --version` prints `0.1.0`; the Board opens; then re-link the development checkout
