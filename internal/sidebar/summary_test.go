package sidebar

import (
	"testing"
	"time"

	"github.com/zeotyn/spec-kanban/internal/work"
)

func TestSummary(t *testing.T) {
	day := func(d int) time.Time { return time.Date(2026, 10, d, 0, 0, 0, 0, time.UTC) }
	ready := work.Item{Name: "r", Total: 2, Phase: work.Ready, Modified: day(1)}
	for _, c := range []struct {
		name          string
		changes       []work.Item
		change, phase string
		waiting       map[string]bool
	}{
		{"in progress",
			[]work.Item{{Name: "m1-board-view", Done: 3, Total: 6, Phase: work.InProgress, Modified: day(2)}},
			"▰▰▰▱▱▱ 3/6 m1-board-view", "1 wip", nil},
		{"almost done is not full",
			[]work.Item{{Name: "a", Done: 29, Total: 30, Phase: work.InProgress}},
			"▰▰▰▰▰▱ 29/30 a", "1 wip", nil},
		{"all done is full",
			[]work.Item{{Name: "a", Done: 4, Total: 4, Phase: work.Review}},
			"▰▰▰▰▰▰ 4/4 a", "1 review", nil},
		{"no tasks",
			[]work.Item{{Name: "add-dark-mode"}},
			"no tasks add-dark-mode", "1 draft", nil},
		{"long name shortened to 28",
			[]work.Item{{Name: "m3-a-very-long-change-name-for-the-sidebar", Done: 2, Total: 12, Phase: work.InProgress}},
			"▰▱▱▱▱▱ 2/12 m3-a-ver…sidebar", "1 wip", nil},
		{"newest modified wins",
			[]work.Item{ready, {Name: "new", Done: 1, Total: 2, Phase: work.InProgress, Modified: day(3)}, {Name: "mid", Modified: day(2)}},
			"▰▰▰▱▱▱ 1/2 new", "1 draft · 1 ready · 1 wip", nil},
		{"tie broken by name",
			[]work.Item{{Name: "b", Modified: day(2)}, {Name: "a", Modified: day(2)}},
			"no tasks a", "2 draft", nil},
		{"archived ignored",
			[]work.Item{ready, {Name: "old", Phase: work.Archived, Modified: day(9)}},
			"▱▱▱▱▱▱ 0/2 r", "1 ready", nil},
		{"no open change",
			[]work.Item{{Name: "old", Phase: work.Archived}},
			"", "no open changes", nil},
		{"zeros omitted",
			[]work.Item{ready, ready, {Name: "d1"}, {Name: "d2"}, {Name: "d3"}, {Name: "v", Done: 1, Total: 1, Phase: work.Review}},
			"▱▱▱▱▱▱ 0/2 r", "3 draft · 2 ready · 1 review", nil},
		{"four phases drop draft",
			[]work.Item{ready, ready, {Name: "d1"}, {Name: "d2"}, {Name: "d3"}, {Name: "w", Done: 1, Total: 2, Phase: work.InProgress}, {Name: "v", Done: 1, Total: 1, Phase: work.Review}},
			"▱▱▱▱▱▱ 0/2 r", "2 ready · 1 wip · 1 review", nil},
		{"One Change waits",
			[]work.Item{ready, {Name: "v", Done: 1, Total: 1, Phase: work.Review}},
			"▱▱▱▱▱▱ 0/2 r", "⚑1 1 ready · 1 review", map[string]bool{"v": true}},
		{"Prefix keeps its place",
			[]work.Item{ready, ready, {Name: "w", Done: 1, Total: 2, Phase: work.InProgress}, {Name: "v", Done: 1, Total: 1, Phase: work.Review}},
			"▱▱▱▱▱▱ 0/2 r", "⚑1 1 wip · 1 review", map[string]bool{"w": true}},
		{"archived Changes do not count",
			[]work.Item{ready, {Name: "old", Phase: work.Archived}},
			"▱▱▱▱▱▱ 0/2 r", "1 ready", map[string]bool{"old": true}},
	} {
		t.Run(c.name, func(t *testing.T) {
			change, phases := Summary(c.changes, c.waiting)
			if change != c.change || phases != c.phase {
				t.Errorf("Summary\n got %q, %q\nwant %q, %q", change, phases, c.change, c.phase)
			}
			if n := max(len([]rune(change)), len([]rune(phases))); n > width {
				t.Errorf("change line has %d runes, more than %d", n, width)
			}
		})
	}
}
