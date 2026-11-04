## ADDED Requirements

### Requirement: Verify and archive commands
Each Source SHALL name its command to verify and its command to archive a finished Work item, or none. OpenSpec SHALL name `/opsx:verify <change>` and `/opsx:archive <change>`. Spec Kit SHALL name none of the two, since it has no step after implementing that checks or archives a Feature.

#### Scenario: OpenSpec commands
- **WHEN** the verify and archive commands of OpenSpec Change `m1-board-view` are asked for
- **THEN** they are `/opsx:verify m1-board-view` and `/opsx:archive m1-board-view`

#### Scenario: Spec Kit has none
- **WHEN** the verify and archive commands of Spec Kit Feature `003-search` are asked for
- **THEN** there are none
