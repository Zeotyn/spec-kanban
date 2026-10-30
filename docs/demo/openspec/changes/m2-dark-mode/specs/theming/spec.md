## ADDED Requirements

### Requirement: Follow the system theme
The app SHALL use the dark theme when the system prefers a dark colour scheme and the user has not chosen a theme.

#### Scenario: System in dark mode
- **WHEN** the system prefers dark and no theme is chosen in Settings
- **THEN** the app renders with the dark theme

### Requirement: Manual override
The user SHALL be able to choose Light, Dark or System in Settings, and the choice SHALL survive a restart.

#### Scenario: Force light
- **WHEN** the user picks Light while the system prefers dark
- **THEN** the app renders with the light theme after the next start too
