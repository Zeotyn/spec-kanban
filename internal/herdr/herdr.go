// Package herdr calls the herdr CLI.
package herdr

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Runner runs a herdr CLI command and returns its stdout. Tests pass a recording func.
type Runner func(args ...string) ([]byte, error)

// Exec runs the herdr binary herdr points plugins at, falling back to herdr on PATH.
func Exec(args ...string) ([]byte, error) {
	bin := os.Getenv("HERDR_BIN_PATH")
	if bin == "" {
		bin = "herdr"
	}
	return Run(bin, "", args...)
}

// Run runs bin in dir ("" for the current one) and folds its stderr into the error.
func Run(bin, dir string, args ...string) ([]byte, error) {
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		err = fmt.Errorf("%s %s: %w: %s", filepath.Base(bin), strings.Join(args, " "), err, strings.TrimSpace(string(exit.Stderr)))
	}
	return out, err
}
