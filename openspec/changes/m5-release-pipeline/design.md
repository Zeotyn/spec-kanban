## Context

The repository is public at `github.com/zeotyn/openspec-kanban` with no workflows and no tags. `go vet ./...` and `gofmt -l .` are clean. `internal/openspec/contract_test.go` runs the real `openspec` CLI against `testdata/project` and skips when it is not on PATH; the CLI is the npm package `@fission-ai/openspec` (1.8.0 locally). `cmd/kanban/main.go` has `const version = "0.1.0"`, the manifest says `version = "0.1.0"`, and `[[build]]` runs `go build -o bin/kanban ./cmd/kanban`.

herdr (0.9.5) runs build commands only on `plugin install`/`plugin update` from GitHub, in the fresh checkout, without plugin env (no `HERDR_PLUGIN_ROOT`), and aborts the install on a non-zero exit. The checkout is a shallow git clone with a detached HEAD and no tags, so the script cannot ask git which tag it is on, but `git rev-parse HEAD` works. `plugin install --ref <ref>` pins a ref; `plugin update` refetches it.

## Goals / Non-Goals

**Goals:**
- `main` is checked on Linux and macOS on every push and pull request, including the contract test.
- A tag produces a release from which an install of that tag needs no Go.
- An install never runs a binary built from other sources than its checkout, and never ends worse than today's `go build`.

**Non-Goals:**
- **golangci-lint**: another tool, version pin and config for a codebase that is small and already `go vet`/`gofmt` clean. Revisit if reviews start catching what a linter would. Cut.
- **Makefile**: CI and the README need three commands; a Makefile would be a second place to keep them. Cut.
- goreleaser, Homebrew, archives (`.tar.gz`), signing or SBOMs.
- Windows (the manifest supports macOS and Linux only).
- Prebuilt binaries for untagged commits of `main`.
- Automating the version bump or changelog.

## Decisions

### CI: one workflow, two OSes
`.github/workflows/ci.yml` on `push` and `pull_request`, matrix `ubuntu-latest`, `macos-latest`: `actions/checkout`, `actions/setup-go` with `go-version-file: go.mod`, `actions/setup-node`, `npm install -g @fission-ai/openspec@1.8.0` (pinned so a CLI release cannot turn `main` red overnight; bumped on purpose together with the contract test), `openspec --version` (fails the job if the CLI is not on PATH, so the contract test cannot silently skip), `test -z "$(gofmt -l .)"`, `go vet ./...`, `go test -race ./...`, and `sh scripts/build.sh` (exercises the `go build` fallback, since an untagged commit never matches a release). On Ubuntu also `shellcheck scripts/build.sh` (preinstalled on the runner). macOS is in the matrix because the users are on macOS and the Board's process handling (`Setsid`, editor) differs by OS; it costs one matrix entry.

### Release: plain workflow, raw binaries
`.github/workflows/release.yml` on `push: tags: ['v*']`, one Ubuntu job with `permissions: contents: write`:
1. Fail unless the tag equals `v` + the manifest `version`.
2. Extract the tag's section from `CHANGELOG.md` (`## <version> - <date>` up to the next `## `); fail if empty.
3. `CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=<version>"` for `darwin`/`linux` × `amd64`/`arm64` into `dist/kanban-<os>-<arch>`.
4. `dist/checksums.txt` (`sha256sum`) and `dist/commit.txt` (`git rev-parse HEAD`).
5. `gh release create <tag> dist/* --notes-file <section>`.
6. Verify: run `sh scripts/build.sh` in the tagged checkout and require that it reports a download and that `bin/kanban --version` prints the version.

Alternative: goreleaser. It does all of this, but brings its own config file, action and conventions for four `go build` calls. Raw binaries instead of archives keep the install script free of `tar`.

### Install script `scripts/build.sh`
POSIX `sh`, `set -eu`, run from the checkout root (as today's `go build` is). It tries the download path and falls back to `go build -o bin/kanban ./cmd/kanban` at the first problem, printing why to stderr:
1. Version from `herdr-plugin.toml` (`sed` on the `version = "…"` line); OS from `uname -s` (`Darwin`→`darwin`, `Linux`→`linux`), arch from `uname -m` (`x86_64`→`amd64`, `arm64`/`aarch64`→`arm64`); anything else → fallback.
2. `curl -fsSL --proto '=https' --max-time 60` of `commit.txt` from `https://github.com/zeotyn/openspec-kanban/releases/download/v<version>/`; unless it equals `git rev-parse HEAD`, fall back. This is what makes the binary match the checkout: an install of `main` past a release builds from source instead of running an older binary against a newer manifest.
3. Download `kanban-<os>-<arch>` and `checksums.txt` into a temp file under `bin/`; compare the file's SHA-256 (`sha256sum` or `shasum -a 256`, whichever exists; neither → fallback) with its line in `checksums.txt`.
4. `chmod +x`, run it with `--version` and require the expected version, then `mv` it to `bin/kanban` (atomic, no half-written binary) and print `downloaded kanban <version>`.
5. On fallback without Go, `go build` fails and herdr shows that error as today.

The checksum guards against truncated or corrupted downloads, not against a compromised release (it comes from the same place); for that, source builds remain the option. `curl` missing → fallback. The script never deletes anything outside `bin/`.

The manifest becomes `[[build]] command = ["sh", "scripts/build.sh"]`.

### Version
`var version = ""` in `main.go`, set by `-ldflags -X main.version=…` in releases. When empty, `--version` prints `debug.ReadBuildInfo().Main.Version` (Go 1.24 stamps a pseudo-version from git, or `(devel)`), so source builds still say something useful. The manifest `version` is the single source of truth: the release job enforces tag = manifest, and the script downloads the manifest's version.

### Changelog and release procedure
`CHANGELOG.md` with one `## <version> - <YYYY-MM-DD>` section per release, newest first, written by hand. Releasing = bump the manifest `version`, add the changelog section, commit on `main`, tag `v<version>`, push the tag. The first release is `v0.1.0`, the manifest's current version. README documents the procedure briefly and the Go-free install `herdr plugin install zeotyn/openspec-kanban --ref v<version>`.

## Risks / Trade-offs

- [Installs of `main` between releases still need Go] → the README's Go-free path is `--ref v<version>`; `main` right at a release also downloads. See Open Questions.
- [`--ref v0.1.0` installs never move on `plugin update`] → upgrading means installing the new tag; README says so.
- [A release job failing after `gh release create`] → the release exists without all assets; installs fall back to `go build` (step 2 or 3 fails). Fix by deleting the release and re-running the job.
- [GitHub download outages or proxies] → every failure falls back to `go build`, so an install is never worse than today.
- [Pinned openspec CLI drifts from what users run] → bump the pin deliberately when OpenSpec releases; the contract test then shows what changed.
- [CI minutes on macOS] → one job per push; acceptable for a small repo.

## Open Questions

- Is a Go-free install only for `--ref v<version>` (and `main` exactly at a release) acceptable, or should installs of `main` download the latest release binary even when the checkout is newer? The latter risks a binary that does not match the manifest and was rejected here.
