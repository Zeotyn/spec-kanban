package start

import (
	"os"
	"path/filepath"
	"testing"
)

// The pane-read-claude-*.txt screens come from Claude Code 2.1.325 through
// herdr pane read <pane> --source visible --format ansi.
func TestInputBox(t *testing.T) {
	const rule = "\x1b[0m\x1b[38;2;136;136;136m────────\x1b[0m"
	tests := []struct {
		name, screen, text string
		found              bool
	}{
		{"empty box below older boxes", "pane-read-claude-empty.txt", "", true},
		{"dim placeholder", "pane-read-claude-placeholder.txt", "", true},
		{"one line", "pane-read-claude-one-line.txt", "/opsx:propose wenn", true},
		{"several lines", "pane-read-claude-lines.txt", "/opsx:propose wenn noch\neine zweite Zeile", true},
		{"trust dialog: no box", "pane-read-claude-dialog.txt", "", false},
		{"a shell: no box", "pane-read.txt", "", false},
		{"non-breaking spaces after ❯", "\n" + rule + "\n❯   \n" + rule + "\n  status", "", true},
		{"a truecolor 2 is not dim", rule + "\n❯ \x1b[38;2;2;2;2mhi\x1b[0m\n" + rule, "hi", true},
		{"dim ended by 22", rule + "\n❯ \x1b[2mTry\x1b[22m x\n" + rule, "x", true},
		{"first line without ❯", rule + "\n> hi\n" + rule, "", false},
		{"one rule only", "❯ hi\n" + rule, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			screen := tt.screen
			if filepath.Ext(screen) == ".txt" {
				b, err := os.ReadFile(filepath.Join("testdata", "herdr", screen))
				if err != nil {
					t.Fatal(err)
				}
				screen = string(b)
			}
			text, found := inputBox(screen)
			if text != tt.text || found != tt.found {
				t.Errorf("inputBox = %q, %v, want %q, %v", text, found, tt.text, tt.found)
			}
		})
	}
}
