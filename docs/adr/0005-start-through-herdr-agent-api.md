---
status: proposed
---

# Start through the herdr agent API

Start creates a new herdr tab in the Project (`herdr tab create --cwd <project> --label <change> --focus`), launches Claude in it with `herdr agent start <name> --kind claude --pane <pane>`, and submits `/opsx:apply <change>` with `herdr agent prompt`. That way herdr knows the session as an agent (status in the sidebar, `agent wait`) instead of us typing text into a shell.

Status stays `proposed` until the M0 spike confirms the flow on herdr 0.9.5.

## Considered Options

- **`pane send-text "claude '/opsx:apply …'"`**: simpler, but herdr only recognises the agent heuristically, and typos end up in the shell.
- **Popup or split instead of a tab**: an implementation run takes long; its own tab stays visible without changing the layout of the current tab.
