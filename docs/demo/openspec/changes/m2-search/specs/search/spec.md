## ADDED Requirements

### Requirement: Search notes as you type
The note list SHALL show only notes whose title or body contains every typed word, ignoring case and accents, and SHALL update while the user types.

#### Scenario: Match in the body
- **WHEN** the user types `groceries` and a note's body contains "Groceries for Sunday"
- **THEN** that note is listed and "Groceries" is highlighted

#### Scenario: Accents are ignored
- **WHEN** the user types `cafe`
- **THEN** notes containing "café" are listed

### Requirement: Focus search with a key
Pressing `/` outside a text field SHALL focus the search box.

#### Scenario: Slash from the list
- **WHEN** the note list is focused and the user presses `/`
- **THEN** the search box has the focus
