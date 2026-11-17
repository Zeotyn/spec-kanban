// Package cache keeps a Project's last loaded Items on disk so the Board can show them at once.
package cache

import (
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/zeotyn/spec-kanban/internal/herdr"
	"github.com/zeotyn/spec-kanban/internal/work"
)

// Version is the cache format version; bump it whenever the stored shape or the reuse rule changes.
const Version = 3

type file struct {
	Version int         `json:"version"`
	Project string      `json:"project"`
	Changes []work.Item `json:"changes"`
}

// Dir returns $HERDR_PLUGIN_STATE_DIR when set, else UserDir.
func Dir() string {
	if d := os.Getenv("HERDR_PLUGIN_STATE_DIR"); d != "" {
		return d
	}
	return UserDir()
}

// UserDir returns os.UserCacheDir()/spec-kanban, "" if it is not available. The Main records and the
// attention markers live here: a worker's pane, where kanban escalate runs, has no $HERDR_PLUGIN_STATE_DIR.
func UserDir() string {
	base, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	return filepath.Join(base, "spec-kanban")
}

// name spreads Projects over files; the project field inside is the authority.
func name(project string) string { return hash(project) + ".json" }

func hash(project string) string {
	sum := sha256.Sum256([]byte(project))
	return hex.EncodeToString(sum[:])[:16]
}

// Read returns the Items cached for project in dir, or nil when there is no valid cache:
// missing/unreadable file, invalid JSON, other version, other project path, or an entry with an empty
// name or a Phase outside Draft…Archived.
func Read(dir, project string) []work.Item {
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
		if c.Name == "" || c.Phase < work.Draft || c.Phase > work.Archived {
			return nil
		}
	}
	if f.Changes == nil {
		return []work.Item{} // valid but empty, unlike "no cache"
	}
	return f.Changes
}

// Write stores changes for project in dir atomically (temp file in dir + os.Rename), creating dir if needed.
func Write(dir, project string, changes []work.Item) error {
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

// mains is where the Main pane of each of project's Changes is recorded, one file per Change.
func mains(dir, project string) string { return filepath.Join(dir, "mains", hash(project)) }

// ReadMain returns the Main pane recorded for change of project in dir, "" when it was unknown, and
// whether there is a readable record at all.
func ReadMain(dir, project, change string) (string, bool) {
	if dir == "" {
		return "", false
	}
	data, err := os.ReadFile(filepath.Join(mains(dir, project), change))
	return string(data), err == nil
}

// WriteMain records pane as the Main pane of change of project in dir, replacing the previous
// record; "" records that there is none.
func WriteMain(dir, project, change, pane string) error {
	if dir == "" {
		return errors.New("cache: no directory")
	}
	return writeAtomic(mains(dir, project), change, []byte(pane))
}

// marks is where the attention markers of project's Changes live, one file per Change.
func marks(dir, project string) string { return filepath.Join(dir, "attention", hash(project)) }

// reasons are why a Change waits for the user.
var reasons = map[string]bool{"escalation": true, "stall": true, "handoff": true}

type mark struct {
	Why string `json:"why"`
}

// WriteMark records that change of project waits for the user because of why: "escalation",
// "stall" or "handoff". It replaces an earlier marker of change.
func WriteMark(dir, project, change, why string) error {
	if dir == "" {
		return errors.New("cache: no directory")
	}
	if !reasons[why] {
		return fmt.Errorf("cache: unknown attention reason %q", why)
	}
	data, err := json.Marshal(mark{why})
	if err != nil {
		return err
	}
	return writeAtomic(marks(dir, project), change, data)
}

// ClearMark removes the attention marker of change of project, if any.
func ClearMark(dir, project, change string) {
	if dir != "" {
		_ = os.Remove(filepath.Join(marks(dir, project), change))
	}
}

// ReadMarks returns the attention markers of project's Changes, Change → reason, skipping
// unreadable files and unknown reasons; nil when there are none.
func ReadMarks(dir, project string) map[string]string {
	if dir == "" {
		return nil
	}
	entries, err := os.ReadDir(marks(dir, project))
	if err != nil {
		return nil
	}
	var got map[string]string
	for _, e := range entries {
		var m mark
		data, err := os.ReadFile(filepath.Join(marks(dir, project), e.Name()))
		if err != nil || json.Unmarshal(data, &m) != nil || !reasons[m.Why] {
			continue
		}
		if got == nil {
			got = map[string]string{}
		}
		got[e.Name()] = m.Why
	}
	return got
}

// Wave is the record of a wave: a group of a Project's Changes merged on branch wave-<Number> by a
// coordinator agent, see the waves spec.
type Wave struct {
	Project     string   `json:"project"`
	Number      int      `json:"number"`
	Branch      string   `json:"branch"`
	Path        string   `json:"path"` // the wave's worktree
	Base        string   `json:"base"`
	Coordinator string   `json:"coordinator"` // the coordinator's pane
	Main        string   `json:"main"`        // the pane the Board was opened from, "" when unknown
	State       string   `json:"state"`       // the plugin's state dir, where members' sessions are recorded
	Members     []string `json:"members"`
	Dropped     []string `json:"dropped,omitempty"`
	Finished    bool     `json:"finished,omitempty"`
}

// waves is where the wave records of project live, one file per wave, so concurrent subcommands
// never write the same file.
func waves(dir, project string) string { return filepath.Join(dir, "waves", hash(project)) }

// WriteWave records w in dir atomically, replacing the earlier record of its number.
func WriteWave(dir string, w Wave) error {
	if dir == "" {
		return errors.New("cache: no directory")
	}
	data, err := json.Marshal(w)
	if err != nil {
		return err
	}
	return writeAtomic(waves(dir, w.Project), strconv.Itoa(w.Number)+".json", data)
}

// RemoveWave removes the record of w from dir, if any.
func RemoveWave(dir string, w Wave) {
	if dir != "" {
		_ = os.Remove(filepath.Join(waves(dir, w.Project), strconv.Itoa(w.Number)+".json"))
	}
}

// ReadWaves returns the waves recorded for project in dir by number, skipping unreadable records;
// nil when there are none.
func ReadWaves(dir, project string) []Wave {
	if dir == "" {
		return nil
	}
	return readWaves(waves(dir, project), project)
}

// readWaves reads the records in d, those of project only unless it is "".
func readWaves(d, project string) []Wave {
	entries, err := os.ReadDir(d)
	if err != nil {
		return nil
	}
	var got []Wave
	for _, e := range entries {
		var w Wave
		data, err := os.ReadFile(filepath.Join(d, e.Name()))
		if err != nil || json.Unmarshal(data, &w) != nil || w.Number < 1 || project != "" && w.Project != project {
			continue
		}
		got = append(got, w)
	}
	slices.SortFunc(got, func(a, b Wave) int { return a.Number - b.Number })
	return got
}

// FindWave returns the wave numbered n whose worktree holds dir, else the one whose Project holds
// dir, else the only one numbered n: the kanban wave subcommands run in agent panes and know only
// their directory, which may be a member's worktree.
func FindWave(dir string, n int, cwd string) (Wave, bool) {
	var all []Wave
	if dir != "" {
		projects, _ := os.ReadDir(filepath.Join(dir, "waves"))
		for _, p := range projects {
			for _, w := range readWaves(filepath.Join(dir, "waves", p.Name()), "") {
				if w.Number == n {
					all = append(all, w)
				}
			}
		}
	}
	for _, in := range []func(Wave) string{func(w Wave) string { return w.Path }, func(w Wave) string { return w.Project }} {
		for _, w := range all {
			if d := filepath.Clean(in(w)); in(w) != "" && (filepath.Clean(cwd) == d || strings.HasPrefix(filepath.Clean(cwd), d+string(filepath.Separator))) {
				return w, true
			}
		}
	}
	if len(all) == 1 {
		return all[0], true
	}
	return Wave{}, false
}

// Badges returns Change → wave number for the members of waves that are neither dropped nor in a
// finished wave.
func Badges(waves []Wave) map[string]int {
	got := map[string]int{}
	for _, w := range waves {
		for _, m := range w.Members {
			if !w.Finished && !slices.Contains(w.Dropped, m) {
				got[m] = w.Number
			}
		}
	}
	return got
}

// NextWave is the number of a Project's next wave: one more than the highest among its recorded
// waves and its branches named wave-<number>, 1 for the first.
func NextWave(waves []Wave, branches []string) int {
	n := 0
	for _, w := range waves {
		n = max(n, w.Number)
	}
	for _, b := range branches {
		if v, err := strconv.Atoi(strings.TrimPrefix(b, "wave-")); strings.HasPrefix(b, "wave-") && err == nil && v > 0 {
			n = max(n, v)
		}
	}
	return n + 1
}

// Failure is one record of the failure log: a detached step (start, run, resume or wave) of a
// Project's Change that failed and showed a herdr notification. A wave's Change is its Changes,
// separated by spaces.
type Failure struct {
	Time    time.Time `json:"time"`
	Project string    `json:"project"`
	Change  string    `json:"change"`
	Step    string    `json:"step"`
	Error   string    `json:"error"`
}

// failures is the failure log, one JSON record per line, and keep how many records it keeps.
const (
	failures = "failures.log"
	keep     = 100
)

// AppendFailure appends f to the failure log in dir with one write, so concurrent writers keep
// their records, then rewrites the newest 100 valid records atomically when it holds more; a record
// appended during that rewrite can be lost.
func AppendFailure(dir string, f Failure) error {
	if dir == "" {
		return errors.New("cache: no directory")
	}
	line, err := json.Marshal(f)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(filepath.Join(dir, failures), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	_, err = file.Write(append(line, '\n'))
	if err := cmp.Or(err, file.Close()); err != nil {
		return err
	}
	all := ReadFailures(dir)
	if len(all) <= keep {
		return nil
	}
	var data []byte
	for _, f := range all[len(all)-keep:] {
		line, _ := json.Marshal(f) // it was unmarshalled from JSON
		data = append(append(data, line...), '\n')
	}
	return writeAtomic(dir, failures, data)
}

// ReadFailures returns the valid records of the failure log in dir, oldest first, skipping lines
// that are not one; nil when there are none or the log cannot be read.
func ReadFailures(dir string) []Failure {
	if dir == "" {
		return nil
	}
	data, err := os.ReadFile(filepath.Join(dir, failures))
	if err != nil {
		return nil
	}
	var got []Failure
	for line := range strings.Lines(string(data)) {
		var f Failure
		if json.Unmarshal([]byte(line), &f) != nil || f.Time.IsZero() {
			continue
		}
		got = append(got, f)
	}
	return got
}

// opened is the record of the Board's last opening for project.
func opened(dir, project string) string { return filepath.Join(dir, "opened", hash(project)) }

// ReadOpened returns when the Board of project last opened, as recorded in dir; an error wrapping
// fs.ErrNotExist when it never did.
func ReadOpened(dir, project string) (time.Time, error) {
	if dir == "" {
		return time.Time{}, errors.New("cache: no directory")
	}
	data, err := os.ReadFile(opened(dir, project))
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339Nano, string(data))
}

// WriteOpened records t as the Board's last opening for project in dir atomically.
func WriteOpened(dir, project string, t time.Time) error {
	if dir == "" {
		return errors.New("cache: no directory")
	}
	return writeAtomic(filepath.Dir(opened(dir, project)), hash(project), []byte(t.Format(time.RFC3339Nano)))
}
