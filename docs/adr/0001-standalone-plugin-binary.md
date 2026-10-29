# Standalone plugin binary instead of a herdr built-in

herdr plugin v1 has no native UI: plugins declare actions, events and panes in their manifest, and everything visible is a terminal program the plugin brings along. The Board is therefore its own binary that herdr launches as a popup or tab pane; it talks to herdr only through the herdr CLI (`HERDR_BIN_PATH`).

## Consequences

Clicking a Space row in the sidebar cannot open the Board, because sidebar tokens are display-only in herdr 0.9.5. The Board opens through a keybinding bound to the plugin action.
