// Package sidebar writes the Project summary into herdr's Space rows.
package sidebar

import (
	"cmp"
	"fmt"
	"strings"

	"github.com/zeotyn/openspec-kanban/internal/board"
	"github.com/zeotyn/openspec-kanban/internal/work"
)

// width is the longest value a token may have; it fits sidebar_width = 34.
const width = 28

var labels = [...]string{work.Draft: "draft", work.Ready: "ready", work.InProgress: "wip", work.Review: "review"}

// Summary returns the Most recently modified Change with its Progress ("" when there is no open
// Change) and the open Changes per Phase.
func Summary(changes []work.Item) (change, phases string) {
	var latest *work.Item
	var counts [len(labels)]int
	for i, c := range changes {
		if c.Phase == work.Archived {
			continue
		}
		counts[c.Phase]++
		if latest == nil || cmp.Or(c.Modified.Compare(latest.Modified), cmp.Compare(latest.Name, c.Name)) > 0 {
			latest = &changes[i]
		}
	}
	if latest == nil {
		return "", "no open changes"
	}

	head := "no tasks "
	if latest.Total > 0 {
		filled := min(latest.Done, latest.Total) * 6 / latest.Total // floor: full only when all are done
		head = strings.Repeat("▰", filled) + strings.Repeat("▱", 6-filled) + fmt.Sprintf(" %d/%d ", latest.Done, latest.Total)
	}
	change = head + board.FitName(latest.Name, width-len([]rune(head)))

	var parts []string
	for p, n := range counts {
		if n > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", n, labels[p]))
		}
	}
	for len(parts) > 1 && len([]rune(strings.Join(parts, " · "))) > width {
		parts = parts[1:] // Draft is the least urgent, and the Board shows it anyway
	}
	return change, strings.Join(parts, " · ")
}
