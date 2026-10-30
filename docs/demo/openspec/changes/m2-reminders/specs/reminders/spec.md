## ADDED Requirements

### Requirement: Remind at a chosen time
The app SHALL show a notification for a todo at its reminder time unless the todo is done.

#### Scenario: Reminder fires
- **WHEN** a todo has a reminder at 09:00 and is still open at 09:00
- **THEN** a notification with the todo's title appears
