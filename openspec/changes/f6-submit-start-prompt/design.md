## Context

`start.LaunchUnfocused` runs the shell-ready probe, then `herdr agent start … --timeout 60000`, then `herdr agent prompt <name> <prompt>`. It returns an error only when one of these fails. Every launch of a new agent goes through it: Start (`launch`), Verify and Archive in a new tab (`runCommand`), and a wave's coordinator and members. f4 added `inputBox` (Claude Code's input box from `herdr pane read --format ansi`, dim placeholder removed) and `getAgent` (status and kind) in the same package. The callers already report a launch error as `spec-kanban: <step> failed` and record it in the failure log (m13-failure-log).

See proposal.md for why.

## Goals / Non-Goals

**Goals:**
- Catch a first prompt that was typed but not submitted, fix it with Enter, and report it when Enter does not help.
- Never submit text the user typed.

**Non-Goals:**
- No check for prompts to running agents, and none for kinds other than `claude`.

## Decisions

### D1: `submitted` checks after the prompt in `LaunchUnfocused`
After `herdr agent prompt` succeeds and the kind is `claude`, `LaunchUnfocused` calls `submitted(run, name, pane, prompt)`:
1. Poll every second for up to 5 seconds. The prompt is taken when `getAgent` reports `working`, or the box is empty. It counts as taken right away when there is no box, the read fails, or the box holds text that does not start with the prompt's first 20 characters.
2. Otherwise send `herdr pane send-keys <pane> enter` and poll again for 2 seconds. Repeat up to 3 times.
3. When it is still not taken, return the error `the prompt was typed but not submitted in tab <change>` plus `tail(pane, run)`.

Sleep and clock are the package variables f4 added, so tests do not wait.
Alternatives:
- Always send one Enter after the prompt. That lost because a working agent would get a stray Enter.
- Retype the whole prompt. That lost because it would duplicate the text in the box.

### D2: The prompt's first 20 characters identify it
The box text is compared with the prompt's start, after trimming. Claude Code wraps long input over several lines. The start of the first line is stable, while the rest depends on the width. 20 characters cover `/opsx:apply ` plus the start of the Change name, which is distinct from anything the user would type in a new tab within seconds.

## Risks / Trade-offs

- [A slow agent is not working after 5 seconds although it took the prompt] → Then the box is empty, which counts as taken. An Enter goes out only while the prompt still sits in the box.
- [Claude Code changes its box] → `inputBox` finds nothing, the prompt counts as taken, and the behaviour is the same as today.
- [Up to 11 seconds more per launch in the worst case] → This only happens while the prompt is stuck. A normal launch is done at the first poll, about 1 second.
