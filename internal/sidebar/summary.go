// Package sidebar writes the Project summary into herdr's Space rows.
package sidebar

import (
	"cmp"
	"fmt"
	"strings"

	"github.com/zeotyn/openspec-kanban/internal/openspec"
	"github.com/zeotyn/openspec-kanban/internal/phase"
)

// width is the longest value a token may have; it fits sidebar_width = 34.
const width = 28

var labels = [...]string{phase.Draft: "draft", phase.Ready: "ready", phase.InProgress: "wip", phase.Review: "review"}

// Summary returns the Most recently modified Change with its Progress ("" when there is no open
// Change) and the open Changes per Phase.
func Summary(changes []openspec.Change) (change, phases string) {
	var latest *openspec.Change
	var counts [len(labels)]int
	for i, c := range changes {
		if c.Archived {
			continue
		}
		counts[phase.Of(c)]++
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
	change = head + fitName(latest.Name, width-len([]rune(head)))

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

// fitName shortens name in the middle with "…" to at most width runes, keeping more of the start.
// Copied from board.
func fitName(name string, width int) string {
	r := []rune(name)
	if len(r) <= width {
		return name
	}
	if width < 2 {
		return string(r[:max(width, 0)])
	}
	keep := width - 1
	return string(r[:keep-keep/2]) + "…" + string(r[len(r)-keep/2:])
}
