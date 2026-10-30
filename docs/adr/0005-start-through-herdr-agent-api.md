---
status: accepted
---

# Start through the herdr agent API

Start creates a new herdr tab in the Project (`herdr tab create --cwd <project> --label <change> --focus`) _(where the tab opens: superseded by [ADR 0007](0007-start-in-a-git-worktree.md))_, launches Claude in it with `herdr agent start <name> --kind claude --pane <pane>`, and submits `/opsx:apply <change>` with `herdr agent prompt`. That way herdr knows the session as an agent (status in the sidebar, `agent wait`) instead of us typing text into a shell.

Confirmed by the M0 spike on herdr 0.9.5: start to ready takes about 4 s. herdr keeps an agent registered after Claude exits, so Start checks the pane's foreground process (`pane process-info`): a running agent is focused, an exited one is restarted in its existing tab, and only an unknown name gets a new tab.

## Considered Options

- **`pane send-text "claude '/opsx:apply …'"`**: simpler, but herdr only recognises the agent heuristically, and typos end up in the shell.
- **Popup or split instead of a tab**: an implementation run takes long; its own tab stays visible without changing the layout of the current tab.
