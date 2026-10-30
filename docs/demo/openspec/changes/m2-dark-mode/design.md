## Context

Colours are hard-coded in about forty components.

## Decisions

### Theme tokens as CSS custom properties
All colours become `--color-*` variables on `:root`; the dark theme redefines them under
`[data-theme="dark"]`. No component needs to know which theme is active.

### System setting first
`prefers-color-scheme` picks the theme unless the user chose one in Settings; the choice
is stored per device.

## Risks

- Third-party date picker ignores the tokens → wrap it and override its two colours
