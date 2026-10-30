## 1. Config (TDD)

- [x] 1.1 Write failing table tests for `config.Parse`: empty object gives defaults (`agent` unset); `agent` `codex` is kept; errors for `agent` `Codex`, `bash` and `""`; `prompt` alone keeps `hide` empty; `hide` alone keeps the default prompt; Phase names case-insensitive (`archived`, `IN PROGRESS`); errors for invalid JSON, unknown key (`promt`), wrong type (`"hide": "Archived"`), `prompt` without `{change}`, empty `prompt`, unknown Phase, all five Phases hidden
- [x] 1.2 Write failing tests for `config.Load` and `Config.PromptFor`: missing file gives defaults and no warning; unreadable file and invalid content give defaults and a warning naming the file; every `{change}` is replaced
- [x] 1.3 Write failing tests for `config.Dir`: `HERDR_PLUGIN_CONFIG_DIR` wins; without it the trimmed output of `herdr plugin config-dir openspec-kanban` (fake runner); runner error gives ""
- [x] 1.4 Implement `internal/config` until the tests pass

## 2. Start takes kind and prompt (TDD)

- [x] 2.1 Write failing tests for `start.Kind` (fake runner): configured kind wins without calling herdr; empty pane gives `claude`; `agent get <pane>` reporting `codex` gives `codex`; `agent_not_found`, other herdr error, unparsable output and a kind outside `config.Kinds` give `claude`
- [x] 2.2 Change `start` tests: `Launch` runs `agent start … --kind <kind>` and sends the given prompt verbatim (default and custom); `Prepare` treats a pane whose foreground is not the shell as running and focuses it, also for a `node` process or `vim`; main's `launch` test passes kind and prompt through
- [x] 2.3 Implement `start.Kind`, the kind and prompt parameters in `start.Launch`/`start.Start`, `agentRunning(run, pane)` checking only whether the shell is idle, `kanban launch <change> <pane> <tab> <kind> <prompt>` and `spawnLaunch`; update the usage text and error texts that say Claude
- [x] 2.4 Write a failing test for `open`: a context with `focused_pane_id` adds `--env KANBAN_MAIN_PANE=<id>`, none without it; implement it

## 3. Hidden Phases on the Board (TDD)

- [x] 3.1 Write failing Board tests: hidden Archived shows four columns using the full width (golden `board-hidden`); `5` does nothing; `l` from Review stays; hidden Draft with open Changes keeps the header Progress; focus fallback after a load skips hidden columns; filter and reload keep hidden columns empty; narrow mode never shows a hidden column
- [x] 3.2 Implement the hidden Phases in `board.New`, `refilter` and `columns()` until the tests pass

## 4. Key help (TDD)

- [x] 4.1 Write failing Board tests: `?` shows the help (golden `help-120`, and a short popup cutting it off); `q`, `j`, `s`, `esc` close the help without other effect; `ctrl+c` quits from the help; `?` in the detail view shows the help and the next key returns to the detail view at the same scroll position; `?` in the filter prompt is typed; `?` while starting is ignored; Board and detail footers show `? help`
- [x] 4.2 Implement the help text, `help` state, key handling and footers until the tests pass; rewrite affected goldens with `-update` and review the diff

## 5. Wiring

- [x] 5.1 `showBoard` loads the config, passes hidden Phases and the warning as initial notice to `board.New`; on `s` the Board resolves the kind with `start.Kind(cfg.Agent, KANBAN_MAIN_PANE)` and passes it with the rendered prompt to `spawnLaunch`; `startChange` loads the config, prints a warning to stderr and uses `start.Kind(cfg.Agent, HERDR_PANE_ID)` and the prompt; test the stderr warning with a broken config in a temp dir

## 6. Docs

- [x] 6.1 README: `?` in the keys table and the Keys section; a Settings section with the path (`herdr plugin config-dir openspec-kanban`), a full `config.json` example, all three keys with defaults, the invalid-config behaviour, how the kind is chosen (config, invoking pane, `claude`), and that another kind needs a `prompt` naming the command OpenSpec installed for that tool (`openspec init --tools <tool>`); texts that say Claude where any kind is meant say "the agent"

## 7. Verification

- [x] 7.1 `go test -race ./...`, `go vet ./...` and `gofmt -l .` pass
- [x] 7.2 Smoke test in herdr: rebuild; `?` on the Board and in the detail view shows the help, any key closes it; write `config.json` with `"hide": ["Archived"]` and a custom prompt, reopen the Board: four columns, `s` on a Ready Card sends the custom prompt; break the JSON, reopen: five columns and a footer warning; `kanban start` from a shell prints the warning; open the Board from a plain shell pane and start: Claude; set `"agent": "gemini"` with a matching prompt, start, then `s` again on the same Card: the running Gemini agent is focused; remove the file again
