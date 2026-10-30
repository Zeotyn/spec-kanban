package openspec

import (
	"reflect"
	"testing"
	"time"

	"github.com/zeotyn/openspec-kanban/internal/work"
)

func TestPhaseOf(t *testing.T) {
	tests := []struct {
		name string
		c    Change
		want work.Phase
	}{
		{"archived, whatever its tasks", Change{Archived: true, Done: 1, Total: 3}, work.Archived},
		{"no tasks, complete plan", Change{Planning: Complete}, work.Draft},
		{"all tasks done", Change{Done: 5, Total: 5}, work.Review},
		{"more done than total", Change{Done: 6, Total: 5}, work.Review},
		{"negative total", Change{Total: -1, Planning: Complete}, work.Draft},
		{"progress outranks incomplete planning", Change{Done: 2, Total: 5, Planning: Incomplete}, work.InProgress},
		{"fully planned, not started", Change{Total: 4, Planning: Complete}, work.Ready},
		{"planning not finished", Change{Total: 4, Planning: Incomplete}, work.Draft},
		{"planning unknown", Change{Total: 4, Planning: Unknown}, work.Draft},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := phaseOf(tt.c); got != tt.want {
				t.Errorf("phaseOf = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestItems(t *testing.T) {
	mod := time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC)
	arts := map[string]string{"proposal": "done"}
	got := Items([]Change{
		{Name: "a", Done: 1, Total: 3, Modified: mod, Planning: Complete, Artifacts: arts, Files: "f1", Dir: "/wt/a"},
		{Name: "b", Archived: true, ArchivedOn: "2026-10-01"},
	})
	want := []work.Item{
		{Name: "a", Done: 1, Total: 3, Modified: mod, Phase: work.InProgress, Artifacts: arts, Stamp: "f1", Dir: "/wt/a"},
		{Name: "b", Phase: work.Archived, ArchivedOn: "2026-10-01"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Items =\n%#v\nwant\n%#v", got, want)
	}
}
