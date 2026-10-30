package openspec

import "github.com/zeotyn/openspec-kanban/internal/work"

// phaseOf applies the change-phases rules to an open Change in order; progress outranks planning.
func phaseOf(done, total int, planned bool) work.Phase {
	switch {
	case total <= 0: // also guards against bad CLI counts
		return work.Draft
	case done >= total:
		return work.Review
	case done > 0:
		return work.InProgress
	case planned:
		return work.Ready
	default:
		return work.Draft
	}
}
