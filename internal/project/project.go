// Package project finds the OpenSpec Project a directory belongs to.
package project

import (
	"os"
	"path/filepath"
)

// Find returns the nearest directory at or above start that contains openspec/changes,
// with symlinks resolved, or "" if there is none. start must be absolute: a relative one would
// resolve against the process cwd, which for herdr plugin commands is the plugin root.
func Find(start string) string {
	if !filepath.IsAbs(start) {
		return ""
	}
	dir, err := filepath.EvalSymlinks(start)
	if err != nil {
		return ""
	}
	for {
		if info, err := os.Stat(filepath.Join(dir, "openspec", "changes")); err == nil && info.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
