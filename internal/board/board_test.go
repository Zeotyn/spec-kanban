package board

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

var update = flag.Bool("update", false, "rewrite golden files")

func golden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", "golden", name+".txt")
	if *update {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v (run go test -update)", err)
	}
	if got != string(want) {
		t.Errorf("view differs from %s\n got:\n%s\nwant:\n%s", path, got, want)
	}
}

func TestQuitKeys(t *testing.T) {
	for _, key := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune("q")},
		{Type: tea.KeyEsc},
		{Type: tea.KeyCtrlC},
	} {
		_, cmd := New("/x/proj").Update(key)
		if cmd == nil || cmd() != tea.Quit() {
			t.Errorf("%q does not quit", key)
		}
	}
	if _, cmd := New("/x/proj").Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}); cmd != nil {
		t.Error("j must not quit")
	}
}

func TestView(t *testing.T) {
	golden(t, "project", New("/x/proj").View())
	golden(t, "no-project", New("").View())
}
