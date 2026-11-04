// Package work holds the domain the Board shows, independent of the spec tool a Source reads it from.
package work

import (
	"cmp"
	"slices"
	"strings"
	"time"

	"github.com/zeotyn/spec-kanban/internal/worktree"
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

// Group splits items into columns indexed by Phase: Ready as a queue (fixes, then milestones, both
// by number, then the rest in natural name order), the other open Phases newest first, Archived by
// archive date newest first with undated last; ties by name.
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
			if Phase(p) == Ready {
				return cmp.Or(cmp.Compare(queueRank(a.Name), queueRank(b.Name)), natural(a.Name, b.Name), cmp.Compare(a.Name, b.Name))
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

// Docs are an Item's documents as the detail view's tabs: 1 to 9 of them, Tasks and First among
// them, the tasks tab with at least one file.
type Docs struct {
	Tabs  []Doc
	Tasks int    // tab holding the tasks; shown first when Total > 0, prefixed by Next
	First int    // tab shown first otherwise
	Next  string // "next ▸ …", "all tasks done" or ""
}

// Action is a step a Source may name a command for, after a Work item is implemented.
type Action string

const (
	Verify  Action = "verify"
	Archive Action = "archive"
)

// Source reads the Work items of one spec tool and knows its conventions.
type Source interface {
	Detect(dir string) bool                                                                 // dir holds this tool's work
	Load(project string, others []worktree.Checkout, known map[string]Item) ([]Item, error) // others: checkouts that may supply an Item; known: the last load, by name, for reuse
	Open(project string) ([]string, error)                                                  // names of the open Items, without planning reads
	ArchivedIn(dir, name string) bool                                                       // archived inside the checkout dir
	Docs(project string, it Item) Docs                                                      // the detail view's tabs
	Rel(name string) string                                                                 // an open Item's path below a checkout
	Prepare(dir, name string) error                                                         // make the checkout dir ready for name's agent
	Prompt(project, kind string) string                                                     // default Start prompt for an agent of kind, may hold {change}
	Finish(name, checkout string) string                                                    // Handoff step after review; checkout "" = no linked worktree; "" = no step
	Command(a Action, name string) string                                                   // the command that does a for name; "" = none
}

// queueRank places a name in the Ready queue: 0 for a fix "f<digits>-…", 1 for a milestone
// "m<digits>-…", 2 for everything else. Within a rank, natural order on the name sorts by number.
func queueRank(name string) int {
	if name == "" {
		return 2
	}
	rest := strings.TrimLeft(name[1:], "0123456789")
	rank := strings.IndexByte("fm", name[0])
	if rank < 0 || len(rest) == len(name)-1 || !strings.HasPrefix(rest, "-") {
		return 2
	}
	return rank
}

// natural compares a and b with runs of digits compared by value, so "c2" comes before "c10".
func natural(a, b string) int {
	for a != "" && b != "" {
		if isDigit(a[0]) && isDigit(b[0]) {
			var x, y string
			x, a = digitRun(a)
			y, b = digitRun(b)
			if c := cmp.Or(cmp.Compare(len(x), len(y)), cmp.Compare(x, y)); c != 0 {
				return c
			}
			continue
		}
		if c := cmp.Compare(a[0], b[0]); c != 0 {
			return c
		}
		a, b = a[1:], b[1:]
	}
	return cmp.Compare(len(a), len(b))
}

func isDigit(c byte) bool { return '0' <= c && c <= '9' }

// digitRun splits s after its leading digits and returns them without leading zeros.
func digitRun(s string) (run, rest string) {
	rest = strings.TrimLeft(s, "0123456789")
	return strings.TrimLeft(s[:len(s)-len(rest)], "0"), rest
}
