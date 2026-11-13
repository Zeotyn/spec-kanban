## 0. Precondition

- [x] 0.1 Check that m12-escalation is merged into `main` (Main records, `kanban escalate`, watcher with `Main`/`Base`); stop and tell the user if it is not

## 1. Config

- [x] 1.1 Add `wave` with `size` (integer ≥ 2, default 5) to `internal/config`, with validation messages for an unknown key in `wave`, a `size` below 2 and a non-integer `size`; tests for default, valid, too small and wrong type

## 2. Wave records

- [x] 2.1 Add the wave record (number, branch, worktree path, base, coordinator pane, Main pane, members, dropped, finished) to `internal/cache` as one JSON file per wave under the user cache directory, written atomically; read one, read all of a Project, find a wave by number and directory (design D5); tests
- [x] 2.2 Add the next wave number: one more than the highest recorded number and `wave-<number>` branch, starting at 1; tests including a lost record with branch `wave-3` left

## 3. Placement and launch

- [x] 3.1 Move the "create a worktree for a branch from a base" part of `start.place` into a helper, used by Start unchanged (D6); existing start tests stay green
- [x] 3.2 Add a Launch variant that does not focus the tab, keeping the shell-ready probe (D7); test that no `tab focus` is run
- [x] 3.3 Add the coordinator prompt builder: wave, branch, worktree, base, members with proposal paths, the four commands with the binary's absolute path, and the five rules from the spec; test against the spec scenario

## 4. Starting a wave

- [x] 4.1 Add `start.Wave`: check that every Change is open and in Ready, create `wave-<N>` from the base in a workspace labelled `wave-<N>`, launch the coordinator named `wave-<N>` with its prompt, record the wave with the Main pane, focus the coordinator's tab; on failure record nothing and report the step; tests with fake herdr and git
- [x] 4.2 Add the detached `kanban wave-launch` entry (like `launch`) that runs 4.1 and reports a failure as the notification `spec-kanban: wave failed`

## 5. Wave subcommands

- [x] 5.1 `kanban wave start <N> <change>`: refuse an unknown or finished wave, a non-member and a dropped member; place with base `wave-<N>`, write the coordinator as Main record, launch without focus with Start prompt + escalation sentence + the no-archive sentence, watch with `Main` = coordinator and `Base` = `wave-<N>`; print `started <change>`; tests
- [x] 5.2 `kanban wave drop <N> <change>`: record as dropped, touch no agent, tab, worktree or branch; print `dropped <change>`; tests
- [x] 5.3 `kanban wave escalate <N> <reason…>`: prompt the wave's Main pane via `start.Tell` with wave, branch, worktree, coordinator tab and reason, fallback `spec-kanban: wave <N> needs you`; print who received it; tests
- [x] 5.4 `kanban wave finish <N>`: read open Changes of the wave worktree (D8), refuse naming open members, otherwise mark finished and `start.Tell` the Main pane the Handoff text (archived, dropped, merge into base, remove worktree, never `--force`, keep branch, stop if unclean), fallback `spec-kanban: wave <N> finished`; tests
- [x] 5.5 Add the `wave` subcommands to the usage text

## 6. Board

- [x] 6.1 Marks: `space` toggles a mark on Ready Cards only, notices `only Ready cards join a wave` and `a wave holds at most <size> changes`, marks kept by name across filter and reload, lost when a Card leaves Ready; `✓` before the name; footer `<n> marked · s wave · esc clear`; tests and golden views
- [x] 6.2 `Esc` order: filter, then marks, then close; test the three-step scenario
- [x] 6.3 Wave start: `s` with ≥ 2 marks asks `start wave of <n>? y/n`, `y` hands over and closes, any other key cancels and keeps marks; notices `waves need worktree on` and `waves need OpenSpec`; `s` with 0 or 1 marks unchanged; tests
- [x] 6.4 Wave badge `≋<N>` in the dim colour for members of active waves, from the wave records loaded with the Changes; none for dropped members or finished waves; tests and golden views
- [x] 6.5 Key help lists `space` under Board

## 7. Wiring

- [x] 7.1 Wire the Board's wave callback in `showBoard` (spawn `wave-launch` with the marked names and `KANBAN_MAIN_PANE`) and pass the wave records and `wave.size` to the Board
- [x] 7.2 Run `go test ./...` and `go vet ./...`; all green

## 8. Docs

- [x] 8.1 README: section "Waves" (marking, `s`, coordinator, subcommands, `wave.size`, members do not archive)
- [x] 8.2 CONTEXT.md: terms Wave, Coordinator, Member
- [x] 8.3 CHANGELOG entry

## 9. Manual check

- [ ] 9.1 Build `bin/kanban`, mark two independent Ready Changes in a test Project, start a wave, let the coordinator start both members, and check that both hand off to the coordinator, that it merges and archives them on `wave-1`, and that `kanban wave finish 1` reaches the Main agent once
