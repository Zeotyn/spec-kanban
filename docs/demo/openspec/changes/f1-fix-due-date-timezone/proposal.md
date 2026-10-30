## Why

Todos due "today" show up as overdue for users west of UTC after 17:00, because due
dates are compared in UTC.

## What Changes

- Store due dates as calendar dates, compare them in the device's time zone

## Capabilities

### Modified Capabilities
- `notes`: due dates are local calendar dates

## Impact

- One migration that converts existing due timestamps to dates
