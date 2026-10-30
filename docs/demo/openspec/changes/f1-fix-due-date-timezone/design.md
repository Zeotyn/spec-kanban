## Decisions

### Calendar date, not timestamp
A due date has no time of day, so it is stored as `YYYY-MM-DD` and compared with the
local date.
