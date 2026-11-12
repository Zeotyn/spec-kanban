## 1. Readiness probe in `internal/start`

- [ ] 1.1 Tests first with the fake herdr runner of `start_test.go` and a counting `nonce`: the first probe answered → calls are `pane run <pane>  echo kanban-rea""dy-1`, `pane wait-output <pane> --match kanban-ready-1 --timeout 2000`, then `agent start` ("Shell already at its prompt"); a first wait that fails → a second probe with nonce 2 and no `agent start` before it is answered ("Shell rc reads a whole line", "Earlier marker in the same tab"); 30 failed waits → error containing `shell did not reach its prompt in tab "<change>"`, no `agent start` ("Shell never reaches its prompt"); a failed `pane run` → that error at once, one probe only ("Pane gone during the wait")
- [ ] 1.2 Implement `ready(pane string, run herdr.Runner) error` and the `nonce` package variable (design.md decisions 1 and 2); `crypto/rand`, no new dependency
- [ ] 1.3 Tests first for `tail`: 20 lines read, empty lines dropped, the last 8 joined by ` | `; a failed `pane read` yields `pane could not be read: <err>` ("Pane cannot be read for the report")
- [ ] 1.4 Implement `tail` (design.md decision 4)

## 2. `Launch` and `Resume`

- [ ] 2.1 Update `TestLaunch`, `TestLaunchKindAndPrompt` and `TestLaunchPromptWithLeadingDash`: the probe calls come between `tab focus` and `agent start`; "claude not ready" and the probe timeout read the pane and carry `pane shows: ...` ("Claude does not become ready", "Shell did not reach its prompt after the Board closed")
- [ ] 2.2 Make `Launch` call `ready` after `tab focus`, and append `tail` to the probe and `agent start` errors
- [ ] 2.3 Update `TestResume`, `TestResumeFails` and `TestWatchAndResumeFeatureName`: the probe comes before `agent start ... -- --resume <id>` ("Resume waits too"); the not-ready error carries the pane's lines
- [ ] 2.4 Make `Resume` call `ready` and append `tail` the same way
- [ ] 2.5 `go test ./...` and `go vet ./...` green; `cmd/kanban` needs no change (the notifications already carry `err.Error()`)

## 3. Docs

- [ ] 3.1 README, `s` paragraph under Keys: before the agent starts, the plugin waits for the shell prompt by typing a probe line (`kanban-ready-…` in the tab is that), so an rc that asks a question no longer swallows the agent command; a start failure shows the pane's last lines; known limit: a shell that does not read `""` as an empty string (nushell) fails the probe
- [ ] 3.2 CHANGELOG `## Unreleased`: Fixed: starting an agent in a new tab whose shell rc asks a question (oh-my-zsh update prompt) typed `laude`; the plugin now waits for the shell prompt with a probe line before every start, resume, verify and archive, and a start that still fails shows the pane's last lines in the notification

## 4. Verification

- [ ] 4.1 `openspec validate f3-shell-ready --strict`
- [ ] 4.2 Build the branch binary from the worktree: `go build -ldflags "-X main.version=$(git describe --tags --abbrev=0)" -o bin/kanban ./cmd/kanban`

## Smoke test (user, in herdr)

Build the branch binary first (task 4.2), then with the Board open:

1. Press `s` on a Ready Card: the new tab shows one `kanban-ready-…` line, then Claude starts and receives the prompt; the Card shows `●`.
2. Add `read -k 1 "?Update? [Y/n] "` as the last line of `~/.zshrc`, open a new tab by hand to see the question appear, then press `s` on another Ready Card: the question gets answered, no `laude`, Claude starts and receives the prompt. Remove the line afterwards.
3. Press `s` on the first Card again after its agent exited (type `/exit` in its tab): the same tab is reused, one more `kanban-ready-…` line, Claude starts.
4. `s` on a Review Card with a recorded session: the probe line, then `claude --resume …`.
5. Add `claude() { echo nope; }` as the last line of `~/.zshrc` and press `s` on a Ready Card: after 60 s the notification `spec-kanban: start failed` says `claude did not become ready` and shows the pane's last lines, among them `nope`. Remove the line afterwards.
