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
	ready := func(name string, modified time.Time) Item {
		return Item{Name: name, Total: 3, Phase: Ready, Modified: modified}
	}
	archived := func(name, on string) Item {
		return Item{Name: name, Phase: Archived, ArchivedOn: on}
	}
	cols := Group([]Item{
		ready("older", day(27)),
		archived("undated", ""),
		ready("newest", day(29)),
		archived("m0-spike-agent-start", "2026-10-29"),
		{Name: "doing", Done: 1, Total: 3, Phase: InProgress},
		ready("tie-b", day(28)),
		archived("old", "2026-10-01"),
		ready("tie-a", day(28)),
		archived("m0-plugin-skeleton", "2026-10-29"),
	})
	names := func(p Phase) []string {
		var n []string
		for _, it := range cols[p] {
			n = append(n, it.Name)
		}
		return n
	}
	if got, want := names(Ready), []string{"newest", "tie-a", "tie-b", "older"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Ready %q, want %q", got, want)
	}
	if got, want := names(InProgress), []string{"doing"}; !reflect.DeepEqual(got, want) {
		t.Errorf("In Progress %q, want %q", got, want)
	}
	if got, want := names(Archived), []string{"m0-plugin-skeleton", "m0-spike-agent-start", "old", "undated"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Archived %q, want %q", got, want)
	}
	if len(cols[Draft]) != 0 || len(cols[Review]) != 0 {
		t.Errorf("Draft and Review must be empty, got %v", cols)
	}
}
