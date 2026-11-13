## Why

A Change stays in Draft until its planning is complete, and the Card gives only the planning marks `P S D T` as a clue. The user does not read them as a reason. m14-spike-agent-messaging had proposal and tasks, specs skipped and no design. It stayed in Draft, and nobody saw why until the code was read: OpenSpec counts the design as missing. The Board should say why a Change is in Draft, right where the user looks.

## What Changes

- While the selected Card is a Draft Card, the Board footer starts with the reason, followed by the key footer:
  - `draft: design missing`, or several at once, `draft: specs, design missing`, when planning artifacts are neither done nor skipped;
  - `draft: no tasks` when all planning artifacts are done or skipped but `tasks.md` holds no task;
  - `draft: planning unknown` when the planning state could not be read.
- The reason uses the same artifacts as the marks `P S D T`, in that order: proposal, specs, design, tasks. It works for OpenSpec and Spec Kit alike.
- Marks, the filter prompt, the wave question, notices and the starting message keep taking the footer's place as today. The reason is only added in front of the key footer.
- No change to when a Change is in Draft or Ready.

Non-goals: no reason on the Card itself, no reason in the detail view, no change to the Ready rule.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `board`: a new requirement for the Draft reason at the start of the footer.

## Impact

- **Code:** `internal/board` (reason from the Card's Phase, Artifacts and Progress; footer prefix with shifted click positions). The Sources stay unchanged.
- **Docs:** README (Board footer), CHANGELOG.
- No new dependencies.
