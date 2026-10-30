## 1. Config (TDD)

- [ ] 1.1 Write failing table tests for `config.Parse`: empty object gives defaults; `prompt` alone keeps `hide` empty; `hide` alone keeps the default prompt; Phase names case-insensitive (`archived`, `IN PROGRESS`); errors for invalid JSON, unknown key (`promt`), wrong type (`"hide": "Archived"`), `prompt` without `{change}`, empty `prompt`, unknown Phase, all five Phases hidden
- [ ] 1.2 Write failing tests for `config.Load` and `Config.PromptFor`: missing file gives defaults and no warning; unreadable file and invalid content give defaults and a warning naming the file; every `{change}` is replaced
- [ ] 1.3 Write failing tests for `config.Dir`: `HERDR_PLUGIN_CONFIG_DIR` wins; without it the trimmed output of `herdr plugin config-dir openspec-kanban` (fake runner); runner error gives ""
- [ ] 1.4 Implement `internal/config` until the tests pass

## 2. Start takes the prompt (TDD)

- [ ] 2.1 Change `start` tests: `Launch` and `Start` send the given prompt verbatim (default and custom); main's `launch` test passes the fifth argument through
- [ ] 2.2 Implement the prompt parameter in `start.Launch`/`start.Start`, `kanban launch <change> <pane> <tab> <prompt>` and `spawnLaunch`; update the usage text

## 3. Hidden Phases on the Board (TDD)

- [ ] 3.1 Write failing Board tests: hidden Archived shows four columns using the full width (golden `board-hidden`); `5` does nothing; `l` from Review stays; hidden Draft with open Changes keeps the header Progress; focus fallback after a load skips hidden columns; filter and reload keep hidden columns empty; narrow mode never shows a hidden column
- [ ] 3.2 Implement the hidden Phases in `board.New`, `refilter` and `columns()` until the tests pass

## 4. Key help (TDD)

- [ ] 4.1 Write failing Board tests: `?` shows the help (golden `help-120`, and a short popup cutting it off); `q`, `j`, `s`, `esc` close the help without other effect; `ctrl+c` quits from the help; `?` in the detail view shows the help and the next key returns to the detail view at the same scroll position; `?` in the filter prompt is typed; `?` while starting is ignored; Board and detail footers show `? help`
- [ ] 4.2 Implement the help text, `help` state, key handling and footers until the tests pass; rewrite affected goldens with `-update` and review the diff

## 5. Wiring

- [ ] 5.1 `showBoard` loads the config, passes hidden Phases and the warning as initial notice to `board.New`, and the rendered prompt to `spawnLaunch`; `startChange` loads it, prints a warning to stderr and passes the prompt; test the stderr warning with a broken config in a temp dir

## 6. Docs

- [ ] 6.1 README: `?` in the keys table and the Keys section; a Settings section with the path (`herdr plugin config-dir openspec-kanban`), a full `config.json` example, both keys with defaults, and the invalid-config behaviour

## 7. Verification

- [ ] 7.1 `go test -race ./...`, `go vet ./...` and `gofmt -l .` pass
- [ ] 7.2 Smoke test in herdr: rebuild; `?` on the Board and in the detail view shows the help, any key closes it; write `config.json` with `"hide": ["Archived"]` and a custom prompt, reopen the Board: four columns, `s` on a Ready Card sends the custom prompt; break the JSON, reopen: five columns and a footer warning; `kanban start` from a shell prints the warning; remove the file again
