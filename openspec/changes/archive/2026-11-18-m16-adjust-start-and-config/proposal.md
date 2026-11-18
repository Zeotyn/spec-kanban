## Why

The text Start sends comes from `config.json`: the Source's command plus `instructions`, or the whole `prompt`. To start one Change with different instructions, or to change them for good, the user has to find the plugin's config directory, edit the file and come back to the Board. There is no way to adjust a single start, and no way to reach the config from the Board.

## What Changes

- **`S` (capital) starts with adjusted instructions:** on a Card that `s` would start, `S` opens an input line `start> ` pre-filled with the configured `instructions` (or the whole `prompt` when one is set). `enter` starts the Change with the edited text in place of the configured one, for this start only; nothing is saved. `esc` closes the line. The Source's command (`/opsx:apply <change>`) and the Escalation sentence are added as for `s` and are not part of the line.
- **`s` is unchanged.**
- **`c` edits the config:** it opens `config.json` in the user's editor (`$EDITOR`, `vi` when unset), as `e` does with `tasks.md`. A missing file is created with `{}` first. When the editor exits, the Board reads the config again: hidden Phases, wave size and every later Start use the new settings, and a broken config shows the usual warning.
- The Card menu gets an entry "start with…" that does what `S` does. Key help and README list `S` and `c`.

Non-goals: no saving of the adjusted text; no per-Change or per-Project config; no `S` for a wave (with marks, `S` says that it starts one Change).

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `board`: new requirements for `S` and `c`.

## Impact

- **Code:** `internal/board` (input line for `S`, key `c`, menu entry, help), `cmd/kanban` (Start with an override of the instructions, the config path for `c`, reloading the config after `c`), `internal/config` (creating a missing file).
- **Docs:** README (keys, Settings), CHANGELOG.
- No new dependencies.
