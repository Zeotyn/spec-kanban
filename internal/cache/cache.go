// Package cache keeps a Project's last loaded Changes on disk so the Board can show them at once.
package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/zeotyn/openspec-kanban/internal/herdr"
	"github.com/zeotyn/openspec-kanban/internal/openspec"
)

// Version is the cache format version; bump it whenever the stored shape or the reuse rule changes.
const Version = 1

type file struct {
	Version int               `json:"version"`
	Project string            `json:"project"`
	Changes []openspec.Change `json:"changes"`
}

// Dir returns $HERDR_PLUGIN_STATE_DIR when set, else os.UserCacheDir()/openspec-kanban; "" if neither is available.
func Dir() string {
	if d := os.Getenv("HERDR_PLUGIN_STATE_DIR"); d != "" {
		return d
	}
	base, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	return filepath.Join(base, "openspec-kanban")
}

// name spreads Projects over files; the project field inside is the authority.
func name(project string) string { return hash(project) + ".json" }

func hash(project string) string {
	sum := sha256.Sum256([]byte(project))
	return hex.EncodeToString(sum[:])[:16]
}

// Read returns the Changes cached for project in dir, or nil when there is no valid cache:
// missing/unreadable file, invalid JSON, other version, other project path, or an entry with an empty
// name or an unknown planning value.
func Read(dir, project string) []openspec.Change {
	if dir == "" {
		return nil
	}
	data, err := os.ReadFile(filepath.Join(dir, name(project)))
	if err != nil {
		return nil
	}
	var f file
	if json.Unmarshal(data, &f) != nil || f.Version != Version || f.Project != project {
		return nil
	}
	for _, c := range f.Changes {
		if c.Name == "" || c.Planning < openspec.Unknown || c.Planning > openspec.Complete {
			return nil
		}
	}
	if f.Changes == nil {
		return []openspec.Change{} // valid but empty, unlike "no cache"
	}
	return f.Changes
}

// Write stores changes for project in dir atomically (temp file in dir + os.Rename), creating dir if needed.
func Write(dir, project string, changes []openspec.Change) error {
	if dir == "" {
		return errors.New("cache: no directory")
	}
	data, err := json.Marshal(file{Version, project, changes})
	if err != nil {
		return err
	}
	return writeAtomic(dir, name(project), data)
}

// writeAtomic writes data to dir/base atomically (temp file in dir + os.Rename), creating dir if needed.
func writeAtomic(dir, base string, data []byte) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // no-op after a successful rename
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), filepath.Join(dir, base))
}

// sessions is where the recorded sessions of project's Changes live, one file per Change, so
// concurrent watchers never write the same file. The Project hash keeps equally named Changes apart.
func sessions(dir, project string) string { return filepath.Join(dir, "sessions", hash(project)) }

// ReadSession returns the session recorded for change of project in dir, the zero value when there
// is none or it is unreadable.
func ReadSession(dir, project, change string) herdr.Session {
	var s herdr.Session
	if dir == "" {
		return s
	}
	data, err := os.ReadFile(filepath.Join(sessions(dir, project), change))
	if err != nil || json.Unmarshal(data, &s) != nil {
		return herdr.Session{}
	}
	return s
}

// WriteSession records s for change of project in dir atomically, replacing the previous record.
func WriteSession(dir, project, change string, s herdr.Session) error {
	if dir == "" {
		return errors.New("cache: no directory")
	}
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return writeAtomic(sessions(dir, project), change, data)
}
