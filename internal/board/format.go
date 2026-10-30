package board

import (
	"fmt"
	"strings"
	"time"
)

// FitName shortens name in the middle with "…" to at most width runes, keeping more of the start.
func FitName(name string, width int) string {
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

// age formats d rounded down as minutes, hours, days (under two weeks) or weeks.
func age(d time.Duration) string {
	switch {
	case d < time.Hour:
		return fmt.Sprintf("%dm", max(int(d/time.Minute), 0))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d/time.Hour))
	case d < 14*24*time.Hour:
		return fmt.Sprintf("%dd", int(d/(24*time.Hour)))
	}
	return fmt.Sprintf("%dw", int(d/(7*24*time.Hour)))
}

// milestone returns "M<N>" for names starting with "m<digits>-", else "".
func milestone(name string) string {
	if !strings.HasPrefix(name, "m") {
		return ""
	}
	n := strings.IndexFunc(name[1:], func(r rune) bool { return r < '0' || r > '9' })
	if n < 1 || name[1+n] != '-' {
		return ""
	}
	return "M" + name[1:1+n]
}

// matches reports whether name contains filter, ignoring case; a filter "m<N>" (either case) instead
// matches exactly the names with milestone badge "M<N>".
func matches(name, filter string) bool {
	if b := strings.ToUpper(filter); b != "" && milestone(strings.ToLower(filter)+"-") == b {
		return milestone(name) == b
	}
	return strings.Contains(strings.ToLower(name), strings.ToLower(filter))
}

// planMarks renders proposal, specs, design, tasks as "P S D T", "-" for skipped, "·" otherwise.
func planMarks(artifacts map[string]string) string {
	marks := make([]string, 4)
	for i, a := range []string{"proposal", "specs", "design", "tasks"} {
		switch artifacts[a] {
		case "done":
			marks[i] = strings.ToUpper(a[:1])
		case "skipped":
			marks[i] = "-"
		default:
			marks[i] = "·"
		}
	}
	return strings.Join(marks, " ")
}

// progressBar renders done/total as width cells followed by " done/total".
func progressBar(done, total, width int) string {
	if total <= 0 {
		return "no tasks"
	}
	filled := (min(done, total)*width*2 + total) / (2 * total) // rounded to nearest
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled) + fmt.Sprintf(" %d/%d", done, total)
}
