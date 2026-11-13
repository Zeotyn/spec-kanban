package sidebar

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/zeotyn/spec-kanban/internal/source/openspec"
	"github.com/zeotyn/spec-kanban/internal/start"
	"github.com/zeotyn/spec-kanban/internal/work"
)

// projects creates Projects with the given names in a temp dir and returns their paths by name.
func projects(t *testing.T, names ...string) map[string]string {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	paths := map[string]string{"": root}
	for _, n := range names {
		paths[n] = filepath.Join(root, n)
		if err := os.MkdirAll(filepath.Join(paths[n], "openspec", "changes"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return paths
}

type pane struct {
	ws, cwd, fg string
	focused     bool
}

// fakeHerdr answers workspace and pane list and records every other call.
type fakeHerdr struct {
	workspaces []string
	panes      []pane
	fail       map[string]error
	agents     string // agent list output
	agentLists int
	reports    []string
}

func (f *fakeHerdr) run(args ...string) ([]byte, error) {
	call := strings.Join(args, " ")
	if err := f.fail[call]; err != nil {
		return nil, err
	}
	switch call {
	case "workspace list":
		var ws []string
		for _, id := range f.workspaces {
			ws = append(ws, fmt.Sprintf(`{"workspace_id":%q,"label":"x"}`, id))
		}
		return []byte(`{"id":"cli:workspace:list","result":{"type":"workspace_list","workspaces":[` + strings.Join(ws, ",") + `]}}`), nil
	case "pane list":
		var ps []string
		for _, p := range f.panes {
			fg := ""
			if p.fg != "" {
				fg = fmt.Sprintf(`"foreground_cwd":%q,`, p.fg)
			}
			ps = append(ps, fmt.Sprintf(`{"workspace_id":%q,"cwd":%q,%s"focused":%t}`, p.ws, p.cwd, fg, p.focused))
		}
		return []byte(`{"id":"cli:pane:list","result":{"type":"pane_list","panes":[` + strings.Join(ps, ",") + `]}}`), nil
	case "agent list":
		f.agentLists++
		return []byte(f.agents), nil
	}
	f.reports = append(f.reports, call)
	return []byte(`{}`), nil
}

const report = "workspace report-metadata "

// none is a waiting func for which no Change waits.
func none(string, []start.Agent, []work.Item) map[string]bool { return nil }

func TestSync(t *testing.T) {
	p := projects(t, "kanban", "magentic", "empty", "broken")
	open := []work.Item{{Name: "m1-board-view", Done: 3, Total: 6, Phase: work.InProgress}}
	var mu sync.Mutex // Sync loads Projects in parallel
	loads := map[string]int{}
	load := func(proj string) ([]work.Item, error) {
		mu.Lock()
		loads[proj]++
		mu.Unlock()
		switch proj {
		case p["kanban"], p["magentic"]:
			return open, nil
		case p["empty"]:
			return []work.Item{{Name: "old", Phase: work.Archived}}, nil
		}
		return nil, errors.New("openspec list: boom")
	}
	if err := os.Mkdir(filepath.Join(p["kanban"], "internal"), 0o755); err != nil {
		t.Fatal(err)
	}
	tokens := " --source spec-kanban --token kanban_change=▰▰▰▱▱▱ 3/6 m1-board-view --token kanban=1 wip"
	h := &fakeHerdr{
		workspaces: []string{"w1", "w2", "w3", "w4", "w5", "w6"},
		panes: []pane{
			{ws: "w1", cwd: p[""], fg: filepath.Join(p["kanban"], "internal")}, // foreground dir wins over cwd
			{ws: "w2", cwd: p["magentic"]},                                     // no foreground dir: cwd
			{ws: "w2", cwd: p["kanban"], focused: true},                        // focused pane wins
			{ws: "w3", cwd: p[""]},                                             // no Project
			{ws: "w4", cwd: p["empty"]},
			{ws: "w5", cwd: p["broken"]},
			{ws: "w6", cwd: p["kanban"]}, // shares w1's Project
		},
	}
	err := Sync(h.run, openspec.New(nil).Detect, load, none)
	if err == nil || !strings.Contains(err.Error(), "boom") || !strings.Contains(err.Error(), p["broken"]) {
		t.Errorf("want the failed load's error naming its Project, got %v", err)
	}
	want := []string{
		report + "w1" + tokens,
		report + "w2" + tokens,
		report + "w3 --source spec-kanban --clear-token kanban_change --clear-token kanban",
		report + "w4 --source spec-kanban --clear-token kanban_change --token kanban=no open changes",
		report + "w6" + tokens,
	}
	if !reflect.DeepEqual(h.reports, want) {
		t.Errorf("reports\n got %q\nwant %q", h.reports, want)
	}
	if want := map[string]int{p["kanban"]: 1, p["empty"]: 1, p["broken"]: 1}; !reflect.DeepEqual(loads, want) {
		t.Errorf("each shown Project must be loaded once, got %v", loads)
	}
}

func TestSyncWaiting(t *testing.T) {
	p := projects(t, "a", "b")
	open := []work.Item{{Name: "m1-board-view", Done: 3, Total: 6, Phase: work.InProgress}}
	load := func(string) ([]work.Item, error) { return open, nil }
	tests := []struct {
		name, agents string
		fail         error
		want         []start.Agent
	}{
		{"one agent list for all Projects", `{"result":{"agents":[{"name":"m1-board-view","cwd":"` + p["a"] + `","agent_status":"idle"}]}}`, nil,
			[]start.Agent{{Name: "m1-board-view", Cwd: p["a"], Status: "idle"}}},
		{"agent list fails: tokens still reported", "", errors.New("herdr: agent list failed"), nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &fakeHerdr{workspaces: []string{"w1", "w2"}, panes: []pane{{ws: "w1", cwd: p["a"]}, {ws: "w2", cwd: p["b"]}}, agents: tt.agents}
			if tt.fail != nil {
				h.fail = map[string]error{"agent list": tt.fail}
			}
			var mu sync.Mutex
			asked := map[string][]start.Agent{}
			waiting := func(proj string, agents []start.Agent, changes []work.Item) map[string]bool {
				mu.Lock()
				defer mu.Unlock()
				asked[proj] = agents
				return map[string]bool{"m1-board-view": proj == p["a"]}
			}
			if err := Sync(h.run, openspec.New(nil).Detect, load, waiting); err != nil {
				t.Fatal(err)
			}
			if tt.fail == nil && h.agentLists != 1 {
				t.Errorf("agent list ran %d times, want once per pass", h.agentLists)
			}
			if want := map[string][]start.Agent{p["a"]: tt.want, p["b"]: tt.want}; !reflect.DeepEqual(asked, want) {
				t.Errorf("waiting asked with %v, want %v", asked, want)
			}
			change := " --source spec-kanban --token kanban_change=▰▰▰▱▱▱ 3/6 m1-board-view --token kanban="
			if want := []string{report + "w1" + change + "⚑1 1 wip", report + "w2" + change + "1 wip"}; !reflect.DeepEqual(h.reports, want) {
				t.Errorf("reports\n got %q\nwant %q", h.reports, want)
			}
		})
	}
}

func TestSyncHerdrError(t *testing.T) {
	p := projects(t, "kanban")
	for _, call := range []string{"workspace list", "pane list"} {
		t.Run(call, func(t *testing.T) {
			h := &fakeHerdr{
				workspaces: []string{"w1"},
				panes:      []pane{{ws: "w1", cwd: p["kanban"]}},
				fail:       map[string]error{call: errors.New("herdr: connection refused")},
			}
			load := func(string) ([]work.Item, error) { return nil, nil }
			if err := Sync(h.run, openspec.New(nil).Detect, load, none); err == nil || !strings.Contains(err.Error(), "connection refused") {
				t.Errorf("want herdr's error, got %v", err)
			}
			if len(h.reports) > 0 {
				t.Errorf("nothing may be reported, got %q", h.reports)
			}
		})
	}
}

func TestSyncLoadsAtMostFourAtOnce(t *testing.T) {
	names := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	p := projects(t, names...)
	h := &fakeHerdr{}
	for i, n := range names {
		ws := fmt.Sprintf("w%d", i)
		h.workspaces = append(h.workspaces, ws)
		h.panes = append(h.panes, pane{ws: ws, cwd: p[n]})
	}
	var running, peak atomic.Int32
	var mu sync.Mutex
	load := func(string) ([]work.Item, error) {
		n := running.Add(1)
		mu.Lock()
		peak.Store(max(peak.Load(), n))
		mu.Unlock()
		time.Sleep(20 * time.Millisecond)
		running.Add(-1)
		return nil, nil
	}
	if err := Sync(h.run, openspec.New(nil).Detect, load, none); err != nil {
		t.Fatal(err)
	}
	if got := peak.Load(); got != 4 {
		t.Errorf("want 4 loads at once at most (and in parallel), got %d", got)
	}
}

func TestSyncFocusedPanePerWorkspace(t *testing.T) {
	p := projects(t, "a", "b")
	h := &fakeHerdr{
		workspaces: []string{"w1", "w2", "w3"},
		panes: []pane{
			{ws: "w1", cwd: p["a"]},
			{ws: "w1", cwd: p["b"], focused: true},
			{ws: "w2", cwd: p["b"]},
			{ws: "w2", cwd: p["a"], focused: true},
			{ws: "w3", cwd: p["b"]}, // no focused pane: the first listed decides
			{ws: "w3", cwd: p["a"]},
		},
	}
	load := func(proj string) ([]work.Item, error) {
		return []work.Item{{Name: filepath.Base(proj)}}, nil
	}
	if err := Sync(h.run, openspec.New(nil).Detect, load, none); err != nil {
		t.Fatal(err)
	}
	tokens := func(name string) string {
		return " --source spec-kanban --token kanban_change=no tasks " + name + " --token kanban=1 draft"
	}
	want := []string{report + "w1" + tokens("b"), report + "w2" + tokens("a"), report + "w3" + tokens("b")}
	if !reflect.DeepEqual(h.reports, want) {
		t.Errorf("reports\n got %q\nwant %q", h.reports, want)
	}
}
