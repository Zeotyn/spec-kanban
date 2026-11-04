package work

import (
	"reflect"
	"testing"
	"time"
)

func TestString(t *testing.T) {
	var got []string
	for p := Draft; p <= Archived; p++ {
		got = append(got, p.String())
	}
	if want := []string{"Draft", "Ready", "In Progress", "Review", "Archived"}; !reflect.DeepEqual(got, want) {
		t.Errorf("names %q, want %q", got, want)
	}
}

func TestGroup(t *testing.T) {
	day := func(d int) time.Time { return time.Date(2026, 10, d, 12, 0, 0, 0, time.UTC) }
	item := func(p Phase, name string, modified time.Time) Item {
		return Item{Name: name, Total: 3, Phase: p, Modified: modified}
	}
	archived := func(name, on string) Item {
		return Item{Name: name, Phase: Archived, ArchivedOn: on}
	}
	names := func(items []Item) []string {
		var n []string
		for _, it := range items {
			n = append(n, it.Name)
		}
		return n
	}
	same := day(28)
	for _, c := range []struct {
		name  string
		phase Phase
		items []Item
		want  []string
	}{
		{"two Draft Changes", Draft, []Item{item(Draft, "a", day(27)), item(Draft, "b", day(28))}, []string{"b", "a"}},
		{"Draft newest first, ties by name", Draft,
			[]Item{item(Draft, "older", day(27)), item(Draft, "newest", day(29)), item(Draft, "tie-b", same), item(Draft, "tie-a", same)},
			[]string{"newest", "tie-a", "tie-b", "older"}},
		{"In Progress newest first", InProgress, []Item{item(InProgress, "a", day(27)), item(InProgress, "b", day(28))}, []string{"b", "a"}},
		{"Review newest first", Review, []Item{item(Review, "a", day(27)), item(Review, "b", day(28))}, []string{"b", "a"}},
		{"two Ready Changes", Ready, []Item{item(Ready, "m3-filter", day(28)), item(Ready, "m2-board-view", day(27))}, []string{"m2-board-view", "m3-filter"}},
		{"fixes before milestones", Ready,
			[]Item{item(Ready, "m1-board-view", same), item(Ready, "f2-cache-race", same), item(Ready, "f1-instant-board-open", same)},
			[]string{"f1-instant-board-open", "f2-cache-race", "m1-board-view"}},
		{"milestones by number", Ready,
			[]Item{item(Ready, "m10-card-actions", same), item(Ready, "m9-mouse", same), item(Ready, "m11-ready-queue", same)},
			[]string{"m9-mouse", "m10-card-actions", "m11-ready-queue"}},
		{"other Changes last in natural order", Ready,
			[]Item{item(Ready, "m1-board-view", day(27)), item(Ready, "c10-cleanup", day(29)), item(Ready, "fix-login", same), item(Ready, "c2-logo", same), item(Ready, "add-dark-mode", same)},
			[]string{"m1-board-view", "add-dark-mode", "c2-logo", "c10-cleanup", "fix-login"}},
		{"Spec Kit Features in Ready", Ready,
			[]Item{item(Ready, "010-export", same), item(Ready, "002-search", same), item(Ready, "001-user-login", same)},
			[]string{"001-user-login", "002-search", "010-export"}},
		{"Ready ties by name", Ready, []Item{item(Ready, "m1-b", same), item(Ready, "m01-a", same), item(Ready, "m1-a", same)}, []string{"m01-a", "m1-a", "m1-b"}},
		{"archived on the same day", Archived,
			[]Item{archived("undated", ""), archived("m0-spike-agent-start", "2026-10-29"), archived("old", "2026-10-01"), archived("m0-plugin-skeleton", "2026-10-29")},
			[]string{"m0-plugin-skeleton", "m0-spike-agent-start", "old", "undated"}},
	} {
		cols := Group(c.items)
		if got := names(cols[c.phase]); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: %s %q, want %q", c.name, c.phase, got, c.want)
		}
		if n := len(cols[c.phase]); n != len(c.items) {
			t.Errorf("%s: %d of %d items in %s", c.name, n, len(c.items), c.phase)
		}
	}
}

func TestQueueRank(t *testing.T) {
	for name, want := range map[string]int{
		"f1-instant-board-open": 0, "f12-cache-race": 0,
		"m9-mouse": 1, "m10-card-actions": 1,
		"fix-login": 2, "m-foo": 2, "c2-logo": 2, "001-user-login": 2, "add-dark-mode": 2, "m12": 2,
		"": 2, "f": 2, "f-": 2, "f1": 2,
	} {
		if got := queueRank(name); got != want {
			t.Errorf("queueRank(%q) = %d, want %d", name, got, want)
		}
	}
}

func TestNatural(t *testing.T) {
	for _, c := range []struct {
		a, b string
		want int
	}{
		{"c2-logo", "c10-cleanup", -1},
		{"001-user-login", "002-search", -1},
		{"002-search", "010-export", -1},
		{"10", "2", 1},
		{"m01", "m1", 0},
		{"a", "a", 0},
		{"a", "ab", -1},
		{"a1", "ab", -1},
		{"99999999999999999999999999-x", "100000000000000000000000000-x", -1},
	} {
		if got := natural(c.a, c.b); got != c.want {
			t.Errorf("natural(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}
