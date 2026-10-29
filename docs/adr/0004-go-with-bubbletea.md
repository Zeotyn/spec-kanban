# Go with bubbletea

The Board is almost entirely UI: layout, keyboard, parallel CLI calls. Go with bubbletea and lipgloss provides the event loop (Elm architecture) and asynchronous commands out of the box, builds a single binary in seconds, and is the language of magentic, from which we take the Phase derivation.

## Considered Options

- **Rust with ratatui**: more precise layout, but the loop and concurrency are hand-built, builds are slow, and the toolchain is not installed.
- **Python with curses** (fork of herdr-openspec): no dependencies, but column layout in curses is tedious, and its main advantage (a ready-made parser) is gone because of ADR 0002.
