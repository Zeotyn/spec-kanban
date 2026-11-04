// Package herdr calls the herdr CLI.
package herdr

import (
	"crypto/sha256"
	"encoding/hex"
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

// AgentName is the herdr agent of the Work item name. herdr names start with a lowercase letter and
// have at most 32 characters, so a name that does not start with one, like Spec Kit's 001-…, gets the
// prefix f-, and a longer one is cut to 25 characters and ends in 6 hex characters of its SHA-256.
func AgentName(item string) string {
	a := item
	if a == "" || a[0] < 'a' || a[0] > 'z' {
		a = "f-" + a
	}
	if len(a) <= 32 {
		return a
	}
	sum := sha256.Sum256([]byte(item))
	return strings.TrimSuffix(a[:25], "-") + "-" + hex.EncodeToString(sum[:3])
}

// Session is the agent_session herdr reports for an agent: its kind, "id" or "path", and value.
type Session struct {
	Agent string `json:"agent"`
	Kind  string `json:"kind"`
	Value string `json:"value"`
}
