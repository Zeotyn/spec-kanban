package sidebar

import (
	"testing"
	"time"

	"github.com/zeotyn/openspec-kanban/internal/source/openspec"
)

func TestSummary(t *testing.T) {
	day := func(d int) time.Time { return time.Date(2026, 10, d, 0, 0, 0, 0, time.UTC) }
	ready := openspec.Change{Name: "r", Total: 2, Planning: openspec.Complete, Modified: day(1)}
	for _, c := range []struct {
		name          string
		changes       []openspec.Change
		change, phase string
	}{
		{"in progress",
			[]openspec.Change{{Name: "m1-board-view", Done: 3, Total: 6, Modified: day(2)}},
			"▰▰▰▱▱▱ 3/6 m1-board-view", "1 wip"},
		{"almost done is not full",
			[]openspec.Change{{Name: "a", Done: 29, Total: 30}},
			"▰▰▰▰▰▱ 29/30 a", "1 wip"},
		{"all done is full",
			[]openspec.Change{{Name: "a", Done: 4, Total: 4}},
			"▰▰▰▰▰▰ 4/4 a", "1 review"},
		{"no tasks",
			[]openspec.Change{{Name: "add-dark-mode"}},
			"no tasks add-dark-mode", "1 draft"},
		{"long name shortened to 28",
			[]openspec.Change{{Name: "m3-a-very-long-change-name-for-the-sidebar", Done: 2, Total: 12}},
			"▰▱▱▱▱▱ 2/12 m3-a-ver…sidebar", "1 wip"},
		{"newest modified wins",
			[]openspec.Change{ready, {Name: "new", Done: 1, Total: 2, Modified: day(3)}, {Name: "mid", Modified: day(2)}},
			"▰▰▰▱▱▱ 1/2 new", "1 draft · 1 ready · 1 wip"},
		{"tie broken by name",
			[]openspec.Change{{Name: "b", Modified: day(2)}, {Name: "a", Modified: day(2)}},
			"no tasks a", "2 draft"},
		{"archived ignored",
			[]openspec.Change{ready, {Name: "old", Archived: true, Modified: day(9)}},
			"▱▱▱▱▱▱ 0/2 r", "1 ready"},
		{"no open change",
			[]openspec.Change{{Name: "old", Archived: true}},
			"", "no open changes"},
		{"zeros omitted",
			[]openspec.Change{ready, ready, {Name: "d1"}, {Name: "d2"}, {Name: "d3"}, {Name: "v", Done: 1, Total: 1}},
			"▱▱▱▱▱▱ 0/2 r", "3 draft · 2 ready · 1 review"},
		{"four phases drop draft",
			[]openspec.Change{ready, ready, {Name: "d1"}, {Name: "d2"}, {Name: "d3"}, {Name: "w", Done: 1, Total: 2}, {Name: "v", Done: 1, Total: 1}},
			"▱▱▱▱▱▱ 0/2 r", "2 ready · 1 wip · 1 review"},
	} {
		t.Run(c.name, func(t *testing.T) {
			change, phases := Summary(c.changes)
			if change != c.change || phases != c.phase {
				t.Errorf("Summary\n got %q, %q\nwant %q, %q", change, phases, c.change, c.phase)
			}
			if n := len([]rune(change)); n > width {
				t.Errorf("change line has %d runes, more than %d", n, width)
			}
		})
	}
}
