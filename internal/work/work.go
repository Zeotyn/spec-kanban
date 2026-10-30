// Package work holds the domain the Board shows, independent of the spec tool a Source reads it from.
package work

import (
	"cmp"
	"slices"
	"time"
)

// Phase is a Board column; the values are in column order. The cache stores the int, so reordering
// the constants needs a cache version bump.
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

// Item is one Work item, shown as one Card.
type Item struct {
	Name       string            `json:"name"`
	Done       int               `json:"done"`
	Total      int               `json:"total"`
	Modified   time.Time         `json:"modified"`
	Phase      Phase             `json:"phase"`                // decided by the Source
	Artifacts  map[string]string `json:"artifacts,omitempty"`  // planning artifact id → done|ready|blocked|skipped, nil when not read
	ArchivedOn string            `json:"archivedOn,omitempty"` // "2026-10-29"; "" when open or undated
	Stamp      string            `json:"stamp,omitempty"`      // the Source's reuse key, "" is never reused
	Dir        string            `json:"dir,omitempty"`        // Project directory of the checkout the Item is read from; "" for the Project
}

// Group splits items into columns indexed by Phase: open Phases newest first, Archived by archive
// date newest first with undated last; ties by name.
func Group(items []Item) [len(names)][]Item {
	var cols [len(names)][]Item
	for _, it := range items {
		cols[it.Phase] = append(cols[it.Phase], it)
	}
	for p := range cols {
		slices.SortFunc(cols[p], func(a, b Item) int {
			if Phase(p) == Archived { // descending puts undated ("") last
				return cmp.Or(cmp.Compare(b.ArchivedOn, a.ArchivedOn), cmp.Compare(a.Name, b.Name))
			}
			return cmp.Or(b.Modified.Compare(a.Modified), cmp.Compare(a.Name, b.Name))
		})
	}
	return cols
}

// Doc is one tab of the detail view.
type Doc struct {
	Title string
	Body  string   // markdown; "" when the document does not exist
	Files []string // what e opens, the first one also when Body is ""
}

// Docs are an Item's documents as the detail view's tabs.
type Docs struct {
	Tabs  []Doc
	Tasks int    // tab holding the tasks; shown first when Total > 0, prefixed by Next
	First int    // tab shown first otherwise
	Next  string // "next ▸ …", "all tasks done" or ""
}
