package openspec

import "github.com/zeotyn/openspec-kanban/internal/work"

// phaseOf applies the change-phases rules in order; progress outranks planning.
func phaseOf(c Change) work.Phase {
	switch {
	case c.Archived:
		return work.Archived
	case c.Total <= 0: // also guards against bad CLI counts
		return work.Draft
	case c.Done >= c.Total:
		return work.Review
	case c.Done > 0:
		return work.InProgress
	case c.Planning == Complete:
		return work.Ready
	default:
		return work.Draft
	}
}

// Items converts changes to Items with their Phase decided.
func Items(changes []Change) []work.Item {
	items := make([]work.Item, len(changes))
	for i, c := range changes {
		items[i] = work.Item{Name: c.Name, Done: c.Done, Total: c.Total, Modified: c.Modified, Phase: phaseOf(c),
			Artifacts: c.Artifacts, ArchivedOn: c.ArchivedOn, Stamp: c.Files, Dir: c.Dir}
	}
	return items
}
