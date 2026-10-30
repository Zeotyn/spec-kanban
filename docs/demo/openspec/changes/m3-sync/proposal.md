## Why

Notes only live on the device they were written on. Users with a phone and a laptop
copy notes by hand and lose edits.

## What Changes

- Sync notes and todos between devices through the account backend
- Resolve concurrent edits per field, last writer wins
- Show a small sync status in the header

## Capabilities

### New Capabilities
- `sync`: replicate notes between a user's devices

## Impact

- New sync endpoint in the API, a change log table in the database
