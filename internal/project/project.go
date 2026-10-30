// Package project finds the Project a directory belongs to.
package project

import "path/filepath"

// Find returns the nearest directory at or above start that detect accepts (a Source's Detect),
// with symlinks resolved, or "" if there is none. start must be absolute: a relative one would
// resolve against the process cwd, which for herdr plugin commands is the plugin root.
func Find(start string, detect func(dir string) bool) string {
	if !filepath.IsAbs(start) {
		return ""
	}
	dir, err := filepath.EvalSymlinks(start)
	if err != nil {
		return ""
	}
	for {
		if detect(dir) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
