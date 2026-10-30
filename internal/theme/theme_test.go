package theme

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestPalette(t *testing.T) {
	for got, want := range map[lipgloss.Color]lipgloss.Color{Accent: "4", Dim: "8", Milestone: "5", Running: "2"} {
		if got != want {
			t.Errorf("colour %q, want %q", got, want)
		}
	}
	if !Faint.GetFaint() {
		t.Error("Faint must be faint")
	}
	if !AccentBold.GetBold() || AccentBold.GetForeground() != Accent {
		t.Errorf("AccentBold: bold %v, foreground %v", AccentBold.GetBold(), AccentBold.GetForeground())
	}
}
