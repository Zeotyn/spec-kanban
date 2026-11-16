package start

import (
	"regexp"
	"strings"
)

// inputBox finds Claude Code's input box on an ANSI screen from herdr pane read --format ansi: the
// lines between the last two lines made only of ─, the first one starting with ❯. It returns the
// text after ❯, its lines trimmed of spaces and without dim text, which is the placeholder of an
// empty box.
func inputBox(screen string) (text string, found bool) {
	lines := strings.Split(screen, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(undimmed(lines[i]))
	}
	var rules []int
	for i := len(lines) - 1; i >= 0 && len(rules) < 2; i-- {
		if lines[i] != "" && strings.Trim(lines[i], "─") == "" {
			rules = append(rules, i)
		}
	}
	if len(rules) < 2 || rules[1]+1 == rules[0] {
		return "", false
	}
	box := lines[rules[1]+1 : rules[0]]
	first, ok := strings.CutPrefix(box[0], "❯")
	if !ok {
		return "", false
	}
	box[0] = first
	for i := range box {
		box[i] = strings.TrimSpace(box[i])
	}
	return strings.TrimSpace(strings.Join(box, "\n")), true
}

var sgr = regexp.MustCompile(`\x1b\[([0-9;]*)m`)

// undimmed returns line without its SGR sequences and the text they draw dim.
func undimmed(line string) string {
	var b strings.Builder
	dim, at := false, 0
	for _, m := range sgr.FindAllStringSubmatchIndex(line, -1) {
		if !dim {
			b.WriteString(line[at:m[0]])
		}
		at = m[1]
		ps := strings.Split(line[m[2]:m[3]], ";")
		for i := 0; i < len(ps); i++ {
			switch ps[i] {
			case "", "0", "22":
				dim = false
			case "2":
				dim = true
			case "38", "48", "58": // a color: 5;n or 2;r;g;b follows
				if i+1 < len(ps) && ps[i+1] == "5" {
					i += 2
				} else {
					i += 4
				}
			}
		}
	}
	if !dim {
		b.WriteString(line[at:])
	}
	return b.String()
}
