package board

import (
	"testing"
	"time"
)

func TestFitName(t *testing.T) {
	for _, c := range []struct {
		name  string
		width int
		want  string
	}{
		{"m1-board-view", 13, "m1-board-view"},
		{"m1-board-view", 20, "m1-board-view"},
		{"m1-board-view", 12, "m1-boa…-view"},
		{"abcdefghij", 6, "abc…ij"},
		{"abcdefghij", 5, "ab…ij"},
		{"abcdefghij", 2, "a…"},
		{"abcdefghij", 1, "a"},
		{"abcdefghij", 0, ""},
		{"add-a-very-long-change-name-that-never-ends-in-any-column", 20, "add-a-very…ny-column"},
	} {
		if got := FitName(c.name, c.width); got != c.want {
			t.Errorf("FitName(%q, %d) = %q, want %q", c.name, c.width, got, c.want)
		}
	}
}

func TestAge(t *testing.T) {
	for _, c := range []struct {
		d    time.Duration
		want string
	}{
		{-time.Minute, "0m"},
		{30 * time.Second, "0m"},
		{59 * time.Minute, "59m"},
		{60 * time.Minute, "1h"},
		{23*time.Hour + 59*time.Minute, "23h"},
		{24 * time.Hour, "1d"},
		{13*24*time.Hour + 23*time.Hour, "13d"},
		{14 * 24 * time.Hour, "2w"},
		{29 * 24 * time.Hour, "4w"},
	} {
		if got := age(c.d); got != c.want {
			t.Errorf("age(%v) = %q, want %q", c.d, got, c.want)
		}
	}
}

func TestMilestone(t *testing.T) {
	for name, want := range map[string]string{
		"m1-board-view": "M1",
		"m12-x":         "M12",
		"add-dark-mode": "",
		"m-x":           "",
		"mx-y":          "",
		"m1":            "",
		"m1x-y":         "",
		"":              "",
	} {
		if got := milestone(name); got != want {
			t.Errorf("milestone(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestPlanMarks(t *testing.T) {
	for _, c := range []struct {
		artifacts map[string]string
		want      string
	}{
		{nil, "· · · ·"},
		{map[string]string{"proposal": "done", "specs": "skipped"}, "P - · ·"},
		{map[string]string{"proposal": "done", "specs": "done", "design": "done", "tasks": "done"}, "P S D T"},
		{map[string]string{"proposal": "ready", "specs": "blocked", "design": "skipped", "tasks": "done"}, "· · - T"},
	} {
		if got := planMarks(c.artifacts); got != c.want {
			t.Errorf("planMarks(%v) = %q, want %q", c.artifacts, got, c.want)
		}
	}
}

func TestProgressBar(t *testing.T) {
	for _, c := range []struct {
		done, total, width int
		want               string
	}{
		{0, 5, 5, "░░░░░ 0/5"},
		{5, 5, 5, "█████ 5/5"},
		{6, 5, 5, "█████ 6/5"},
		{1, 3, 10, "███░░░░░░░ 1/3"},
		{2, 3, 10, "███████░░░ 2/3"},
		{5, 11, 10, "█████░░░░░ 5/11"},
		{0, 0, 10, "no tasks"},
	} {
		if got := progressBar(c.done, c.total, c.width); got != c.want {
			t.Errorf("progressBar(%d, %d, %d) = %q, want %q", c.done, c.total, c.width, got, c.want)
		}
	}
}
