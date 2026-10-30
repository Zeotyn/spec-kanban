## MODIFIED Requirements

### Requirement: Overdue todos
A todo SHALL be overdue when its due date is before the current date in the device's time zone.

#### Scenario: Evening west of UTC
- **WHEN** a todo is due today and it is 20:00 in New York
- **THEN** the todo is not overdue
