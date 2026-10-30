package openspec

import (
	"testing"

	"github.com/zeotyn/openspec-kanban/internal/work"
)

func TestPhaseOf(t *testing.T) {
	tests := []struct {
		name        string
		done, total int
		planned     bool
		want        work.Phase
	}{
		{"no tasks, complete plan", 0, 0, true, work.Draft},
		{"all tasks done", 5, 5, false, work.Review},
		{"more done than total", 6, 5, false, work.Review},
		{"negative total", 0, -1, true, work.Draft},
		{"progress outranks incomplete planning", 2, 5, false, work.InProgress},
		{"fully planned, not started", 0, 4, true, work.Ready},
		{"planning not finished or unknown", 0, 4, false, work.Draft},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := phaseOf(tt.done, tt.total, tt.planned); got != tt.want {
				t.Errorf("phaseOf = %v, want %v", got, tt.want)
			}
		})
	}
}
