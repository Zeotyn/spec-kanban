package phase

import (
	"reflect"
	"testing"
	"time"

	"github.com/zeotyn/openspec-kanban/internal/openspec"
)

func TestOf(t *testing.T) {
	tests := []struct {
		name string
		c    openspec.Change
		want Phase
	}{
		{"archived, whatever its tasks", openspec.Change{Archived: true, Done: 1, Total: 3}, Archived},
		{"no tasks, complete plan", openspec.Change{Planning: openspec.Complete}, Draft},
		{"all tasks done", openspec.Change{Done: 5, Total: 5}, Review},
		{"more done than total", openspec.Change{Done: 6, Total: 5}, Review},
		{"negative total", openspec.Change{Total: -1, Planning: openspec.Complete}, Draft},
		{"progress outranks incomplete planning", openspec.Change{Done: 2, Total: 5, Planning: openspec.Incomplete}, InProgress},
		{"fully planned, not started", openspec.Change{Total: 4, Planning: openspec.Complete}, Ready},
		{"planning not finished", openspec.Change{Total: 4, Planning: openspec.Incomplete}, Draft},
		{"planning unknown", openspec.Change{Total: 4, Planning: openspec.Unknown}, Draft},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Of(tt.c); got != tt.want {
				t.Errorf("Of = %v, want %v", got, tt.want)
			}
		})
	}
}

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
	ready := func(name string, modified time.Time) openspec.Change {
		return openspec.Change{Name: name, Total: 3, Planning: openspec.Complete, Modified: modified}
	}
	archived := func(name, on string) openspec.Change {
		return openspec.Change{Name: name, Archived: true, ArchivedOn: on}
	}
	cols := Group([]openspec.Change{
		ready("older", day(27)),
		archived("undated", ""),
		ready("newest", day(29)),
		archived("m0-spike-agent-start", "2026-10-29"),
		{Name: "doing", Done: 1, Total: 3},
		ready("tie-b", day(28)),
		archived("old", "2026-10-01"),
		ready("tie-a", day(28)),
		archived("m0-plugin-skeleton", "2026-10-29"),
	})
	names := func(p Phase) []string {
		var n []string
		for _, c := range cols[p] {
			n = append(n, c.Name)
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
