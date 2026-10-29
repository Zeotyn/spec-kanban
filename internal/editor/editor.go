// Package editor opens files in the user's $EDITOR, handing it the terminal.
package editor

import (
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Exec runs cmd with the terminal handed over; tests replace it to see the command.
var Exec = tea.ExecProcess

// Command builds the editor command for files: $EDITOR split on spaces (so "code -w" works), vi when unset.
func Command(files ...string) *exec.Cmd {
	args := strings.Fields(os.Getenv("EDITOR"))
	if len(args) == 0 {
		args = []string{"vi"}
	}
	return exec.Command(args[0], append(args[1:], files...)...)
}

// Open edits files and reports the editor's outcome through done.
func Open(done tea.ExecCallback, files ...string) tea.Cmd {
	return Exec(Command(files...), done)
}
