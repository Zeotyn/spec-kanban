// Package phase places Changes in Phases (Board columns) and orders the Cards within them.
package phase

import (
	"cmp"
	"slices"

	"github.com/zeotyn/openspec-kanban/internal/source/openspec"
)

// Phase is a Board column; the values are in column order.
type Phase int

const (
	Draft Phase = iota
	Ready
	InProgress
	Review
	Archived
)

var names = [...]string{"Draft", "Ready", "In Progress", "Review", "Archived"}

func (p Phase) String() string { return names[p] }

// Of applies the change-phases rules in order; progress outranks planning.
func Of(c openspec.Change) Phase {
	switch {
	case c.Archived:
		return Archived
	case c.Total <= 0: // also guards against bad CLI counts
		return Draft
	case c.Done >= c.Total:
		return Review
	case c.Done > 0:
		return InProgress
	case c.Planning == openspec.Complete:
		return Ready
	default:
		return Draft
	}
}

// Group splits changes into columns indexed by Phase: open Phases newest first, Archived by archive
// date newest first with undated last; ties by name.
func Group(changes []openspec.Change) [len(names)][]openspec.Change {
	var cols [len(names)][]openspec.Change
	for _, c := range changes {
		p := Of(c)
		cols[p] = append(cols[p], c)
	}
	for p := range cols {
		slices.SortFunc(cols[p], func(a, b openspec.Change) int {
			if Phase(p) == Archived { // descending puts undated ("") last
				return cmp.Or(cmp.Compare(b.ArchivedOn, a.ArchivedOn), cmp.Compare(a.Name, b.Name))
			}
			return cmp.Or(b.Modified.Compare(a.Modified), cmp.Compare(a.Name, b.Name))
		})
	}
	return cols
}
