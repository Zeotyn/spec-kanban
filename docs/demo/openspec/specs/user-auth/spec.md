# user-auth Specification

## Purpose
Sign users in with email and password and keep them signed in.

## Requirements

### Requirement: Email login
The app SHALL sign a user in with a registered email address and the matching password.

#### Scenario: Valid credentials
- **WHEN** the user enters a registered email and the correct password
- **THEN** the app opens their notes
