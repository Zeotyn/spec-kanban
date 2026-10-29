package editor

import (
	"slices"
	"testing"
)

func TestCommand(t *testing.T) {
	tests := []struct {
		editor string
		files  []string
		want   []string
	}{
		{"", []string{"/c/tasks.md"}, []string{"vi", "/c/tasks.md"}},
		{"nvim", []string{"/c/tasks.md"}, []string{"nvim", "/c/tasks.md"}},
		{"code -w", []string{"/c/proposal.md"}, []string{"code", "-w", "/c/proposal.md"}},
		{"nvim -p", []string{"/c/specs/a/spec.md", "/c/specs/b/spec.md"}, []string{"nvim", "-p", "/c/specs/a/spec.md", "/c/specs/b/spec.md"}},
	}
	for _, tt := range tests {
		t.Setenv("EDITOR", tt.editor)
		if got := Command(tt.files...).Args; !slices.Equal(got, tt.want) {
			t.Errorf("EDITOR=%q: args %q, want %q", tt.editor, got, tt.want)
		}
	}
}
