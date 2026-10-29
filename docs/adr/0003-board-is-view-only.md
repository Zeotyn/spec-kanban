# The Board is view-only

Cards cannot be moved. A Change's Phase is derived from the state of its artifacts and tasks on every load and is stored nowhere; the truth lives in the OpenSpec files that agents and people edit anyway. The only action is **Start**: it hands a Change to a new agent session (see ADR 0005), which then edits the files, and the Card moves on by itself on the next load.

## Considered Options

- **Moving a Card triggers an action** (Ready → In Progress starts apply, Review → Archived archives): rejected, because the Board would then need write access and confirmations, and archiving should happen deliberately inside the agent session (`/opsx:archive`).
