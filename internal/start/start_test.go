package start

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const change = "m1-board-view"

// proj is the Project the tests start in: the cwd of the agent in agent-get.json.
const proj = "/Users/U758014/code/openspec-kanban"

// notFound is what herdr.Run returns for an unknown agent: the real stderr folded into the error.
var notFound = func() error {
	b, err := os.ReadFile(filepath.Join("..", "..", "testdata", "herdr", "agent-get-missing.json"))
	if err != nil {
		panic(err)
	}
	return errors.New("herdr agent get: exit status 1: " + string(b))
}()

// fakeHerdr answers each call whose joined args start with a key: a fixture file name or an error.
type fakeHerdr struct {
	t     *testing.T
	calls []string
	out   map[string]string
	fail  map[string]error
}

func (f *fakeHerdr) run(args ...string) ([]byte, error) {
	call := strings.Join(args, " ")
	f.calls = append(f.calls, call)
	for prefix, err := range f.fail {
		if strings.HasPrefix(call, prefix) {
			return nil, err
		}
	}
	for prefix, file := range f.out {
		if strings.HasPrefix(call, prefix) {
			b, err := os.ReadFile(filepath.Join("..", "..", "testdata", "herdr", file))
			if err != nil {
				f.t.Fatal(err)
			}
			return b, nil
		}
	}
	return []byte(`{}`), nil
}

func fakeList(names ...string) func(...string) ([]byte, error) {
	return func(args ...string) ([]byte, error) {
		out := `{"changes":[`
		for i, n := range names {
			if i > 0 {
				out += ","
			}
			out += `{"name":"` + n + `"}`
		}
		return []byte(out + `]}`), nil
	}
}

var (
	create = "tab create --cwd " + proj + " --label " + change + " --no-focus"
	tfocus = "tab focus w2:tG"
	launch = "agent start " + change + " --kind claude --pane w2:p1J --timeout 60000"
	prompt = "agent prompt " + change + " /opsx:apply " + change
	focus  = "agent focus " + change
	info   = "pane process-info --pane w2:p1J"
	get    = "agent get " + change
)

func TestStart(t *testing.T) {
	tests := []struct {
		name string
		out  map[string]string
		fail map[string]error
		want []string
	}{
		{
			name: "new tab",
			out:  map[string]string{"tab create": "tab-create.json"},
			fail: map[string]error{"agent get": notFound},
			want: []string{get, create, tfocus, launch, prompt},
		},
		{
			name: "agent running: focus only",
			out:  map[string]string{"agent get": "agent-get.json", "pane process-info": "process-info-running.json"},
			want: []string{get, info, focus},
		},
		{
			name: "agent exited: restart in its tab",
			out:  map[string]string{"agent get": "agent-get.json", "pane process-info": "process-info-shell.json"},
			want: []string{get, info, tfocus, launch, prompt},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &fakeHerdr{t: t, out: tt.out, fail: tt.fail}
			if err := Start(proj, change, "claude", "/opsx:apply "+change, h.run, fakeList("other", change)); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(h.calls, tt.want) {
				t.Errorf("herdr calls\n got %q\nwant %q", h.calls, tt.want)
			}
		})
	}
}

func TestPrepare(t *testing.T) {
	tests := []struct {
		name     string
		out      map[string]string
		fail     map[string]error
		want     []string
		wantPane string
		wantTab  string
	}{
		{"new tab, not focused", map[string]string{"tab create": "tab-create.json"}, map[string]error{"agent get": notFound}, []string{get, create}, "w2:p1J", "w2:tG"},
		{"agent running: focus only", map[string]string{"agent get": "agent-get.json", "pane process-info": "process-info-running.json"}, nil, []string{get, info, focus}, "", ""},
		{"agent exited: its tab, not focused", map[string]string{"agent get": "agent-get.json", "pane process-info": "process-info-shell.json"}, nil, []string{get, info}, "w2:p1J", "w2:tG"},
		{"agent of its reported kind running: focus only", map[string]string{"agent get": "agent-get-codex.json", "pane process-info": "process-info-codex.json"}, nil, []string{get, info, focus}, "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &fakeHerdr{t: t, out: tt.out, fail: tt.fail}
			pane, tab, err := Prepare(proj, change, "claude", h.run) // the reported kind wins over the one passed
			if err != nil {
				t.Fatal(err)
			}
			if pane != tt.wantPane || tab != tt.wantTab {
				t.Errorf("got pane %q tab %q, want %q %q", pane, tab, tt.wantPane, tt.wantTab)
			}
			if !reflect.DeepEqual(h.calls, tt.want) {
				t.Errorf("herdr calls\n got %q\nwant %q", h.calls, tt.want)
			}
		})
	}
}

func TestPrepareWithoutReportedKind(t *testing.T) {
	h := &fakeHerdr{t: t, out: map[string]string{"agent get": "agent-get-no-kind.json", "pane process-info": "process-info-codex.json"}}
	pane, _, err := Prepare(proj, change, "codex", h.run)
	if err != nil || pane != "" || !reflect.DeepEqual(h.calls, []string{get, info, focus}) {
		t.Fatalf("want the passed kind compared and the agent focused, got pane %q, %v, %q", pane, err, h.calls)
	}
	h = &fakeHerdr{t: t, out: map[string]string{"agent get": "agent-get-no-kind.json", "pane process-info": "process-info-codex.json"}}
	if _, _, err := Prepare(proj, change, "claude", h.run); err == nil || !strings.Contains(err.Error(), "runs codex, not claude") {
		t.Fatalf("want an error naming the program and the kind, got %v", err)
	}
}

func TestPrepareForeignProgramNamesKind(t *testing.T) {
	h := &fakeHerdr{t: t, out: map[string]string{"agent get": "agent-get-codex.json", "pane process-info": "process-info-other.json"}}
	if _, _, err := Prepare(proj, change, "claude", h.run); err == nil || !strings.Contains(err.Error(), "pane w2:p1J runs vim, not codex") {
		t.Fatalf("want an error naming vim and codex, got %v", err)
	}
}

func TestKind(t *testing.T) {
	codex, err := os.ReadFile(filepath.Join("..", "..", "testdata", "herdr", "agent-get-codex.json"))
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name, configured, pane, out string
		err                         error
		want                        string
		wantCalls                   []string
	}{
		{"configured kind wins without herdr", "gemini", "w1:p1", string(codex), nil, "gemini", nil},
		{"no pane", "", "", string(codex), nil, "claude", nil},
		{"agent in the pane", "", "w1:p1", string(codex), nil, "codex", []string{"agent get w1:p1"}},
		{"no agent in the pane", "", "w1:p1", "", notFound, "claude", []string{"agent get w1:p1"}},
		{"herdr error", "", "w1:p1", "", errors.New("herdr: not found"), "claude", []string{"agent get w1:p1"}},
		{"unparsable output", "", "w1:p1", "not json", nil, "claude", []string{"agent get w1:p1"}},
		{"kind herdr cannot start", "", "w1:p1", `{"result":{"agent":{"agent":"aider"}}}`, nil, "claude", []string{"agent get w1:p1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []string
			run := func(args ...string) ([]byte, error) {
				calls = append(calls, strings.Join(args, " "))
				return []byte(tt.out), tt.err
			}
			if got := Kind(tt.configured, tt.pane, run); got != tt.want {
				t.Errorf("Kind = %q, want %q", got, tt.want)
			}
			if !reflect.DeepEqual(calls, tt.wantCalls) {
				t.Errorf("herdr calls %q, want %q", calls, tt.wantCalls)
			}
		})
	}
}

func TestLaunchKindAndPrompt(t *testing.T) {
	h := &fakeHerdr{t: t}
	if err := Launch(change, "codex", "/opsx-apply "+change+" with TDD", "w2:p1J", "w2:tG", h.run); err != nil {
		t.Fatal(err)
	}
	want := []string{tfocus, "agent start " + change + " --kind codex --pane w2:p1J --timeout 60000", "agent prompt " + change + " /opsx-apply " + change + " with TDD"}
	if !reflect.DeepEqual(h.calls, want) {
		t.Errorf("herdr calls\n got %q\nwant %q", h.calls, want)
	}
}

func TestLaunch(t *testing.T) {
	boom := errors.New("boom")
	tests := []struct {
		name  string
		fail  map[string]error
		want  []string
		wants string // substring of the error, "" for none
	}{
		{"focuses the tab, starts claude and sends the prompt", nil, []string{tfocus, launch, prompt}, ""},
		{"tab focus", map[string]error{"tab focus": boom}, []string{tfocus}, "tab focus: boom"},
		{"claude not ready", map[string]error{"agent start": boom}, []string{tfocus, launch}, "did not become ready in tab"},
		{"prompt", map[string]error{"agent prompt": boom}, []string{tfocus, launch, prompt}, "but the prompt was not sent"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &fakeHerdr{t: t, fail: tt.fail}
			err := Launch(change, "claude", "/opsx:apply "+change, "w2:p1J", "w2:tG", h.run)
			if tt.wants == "" && err != nil || tt.wants != "" && (err == nil || !strings.Contains(err.Error(), tt.wants)) {
				t.Fatalf("want error with %q, got %v", tt.wants, err)
			}
			if !reflect.DeepEqual(h.calls, tt.want) {
				t.Errorf("herdr calls\n got %q\nwant %q", h.calls, tt.want)
			}
		})
	}
}

func TestStartRefusesAnotherProjectsAgent(t *testing.T) {
	h := &fakeHerdr{t: t, out: map[string]string{"agent get": "agent-get.json"}}
	err := Start(proj+"2", change, "claude", "/opsx:apply "+change, h.run, fakeList(change))
	if err == nil || !strings.Contains(err.Error(), change) || !strings.Contains(err.Error(), "already runs in "+proj) {
		t.Fatalf("want already-runs error naming the change and %s, got %v", proj, err)
	}
	if !reflect.DeepEqual(h.calls, []string{get}) {
		t.Errorf("want no focus, restart or tab, got %q", h.calls)
	}
}

func TestStartAgentWithoutCwd(t *testing.T) {
	h := &fakeHerdr{t: t, out: map[string]string{"agent get": "agent-get-no-cwd.json", "pane process-info": "process-info-running.json"}}
	if err := Start(proj, change, "claude", "/opsx:apply "+change, h.run, fakeList(change)); err != nil {
		t.Fatalf("an unknown cwd must not refuse the agent: %v", err)
	}
	if !reflect.DeepEqual(h.calls, []string{get, info, focus}) {
		t.Errorf("want focus only, got %q", h.calls)
	}
}

func TestStartRejectsChangesThatAreNotOpen(t *testing.T) {
	h := &fakeHerdr{t: t}
	err := Start(proj, "m1-bord-view", "claude", "/opsx:apply m1-bord-view", h.run, fakeList(change))
	if err == nil || !strings.Contains(err.Error(), `"m1-bord-view" is not an open change`) {
		t.Fatalf("want not-open error, got %v", err)
	}
	if len(h.calls) != 0 {
		t.Errorf("herdr must not be called, got %q", h.calls)
	}
}

func TestStartOpenspecFails(t *testing.T) {
	h := &fakeHerdr{t: t}
	broken := func(...string) ([]byte, error) { return nil, errors.New("openspec: not found") }
	if err := Start(proj, change, "claude", "/opsx:apply "+change, h.run, broken); err == nil || len(h.calls) != 0 {
		t.Fatalf("want error and no herdr calls, got %v, %q", err, h.calls)
	}
}

func TestStartReportsFailingStep(t *testing.T) {
	boom := errors.New("boom")
	tests := []struct {
		name  string
		fail  map[string]error
		wants []string // substrings of the error
		last  string   // last herdr call made
	}{
		{"herdr unreachable", map[string]error{"agent get": boom}, []string{"agent get", "boom"}, get},
		{"tab create", map[string]error{"agent get": notFound, "tab create": boom}, []string{"tab create", "boom"}, create},
		{"claude not ready", map[string]error{"agent get": notFound, "agent start": boom}, []string{"prompt was not sent", change, "boom"}, launch},
		{"prompt", map[string]error{"agent get": notFound, "agent prompt": boom}, []string{"prompt was not sent", change, "boom"}, prompt},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &fakeHerdr{t: t, out: map[string]string{"tab create": "tab-create.json"}, fail: tt.fail}
			err := Start(proj, change, "claude", "/opsx:apply "+change, h.run, fakeList(change))
			if err == nil {
				t.Fatal("want error")
			}
			for _, w := range tt.wants {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("error %q lacks %q", err, w)
				}
			}
			if got := h.calls[len(h.calls)-1]; got != tt.last {
				t.Errorf("last call %q, want %q", got, tt.last)
			}
		})
	}
}

func TestStartWithUnexpectedPaneState(t *testing.T) {
	tests := []struct {
		name  string
		out   map[string]string
		fail  map[string]error
		wants string
	}{
		{"other program in the foreground", map[string]string{"agent get": "agent-get.json", "pane process-info": "process-info-other.json"}, nil, "vim"},
		{"process-info fails", map[string]string{"agent get": "agent-get.json"}, map[string]error{"pane process-info": errors.New("pane_not_found")}, "pane_not_found"},
		{"process-info without shell pid", map[string]string{"agent get": "agent-get.json"}, nil, "process-info"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &fakeHerdr{t: t, out: tt.out, fail: tt.fail}
			err := Start(proj, change, "claude", "/opsx:apply "+change, h.run, fakeList(change))
			if err == nil || !strings.Contains(err.Error(), tt.wants) {
				t.Fatalf("want error with %q, got %v", tt.wants, err)
			}
			if got := h.calls[len(h.calls)-1]; got != info {
				t.Errorf("must stop after process-info, last call %q", got)
			}
		})
	}
}

func TestRunning(t *testing.T) {
	list, err := os.ReadFile(filepath.Join("testdata", "herdr", "agent-list.json"))
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		project string
		out     string
		err     error
		want    map[string]bool
		wantErr bool
	}{
		{"named agents in the project and below it, any status", "/work/proj", string(list), nil,
			map[string]bool{"m1-board-view": true, "m2-card-detail": true, "m3-sidebar-summary": true}, false},
		{"other projects and unnamed agents ignored", "/work/other", string(list), nil, map[string]bool{"c1-rename": true}, false},
		{"name prefix is not a subdirectory", "/work/pro", string(list), nil, map[string]bool{}, false},
		{"trailing slash", "/work/project2/", string(list), nil, map[string]bool{"f1-cache": true}, false},
		{"empty list", "/work/proj", `{"id":"cli:agent:list","result":{"agents":[]}}`, nil, map[string]bool{}, false},
		{"herdr error", "/work/proj", "", errors.New("herdr: not found"), nil, true},
		{"unparsable output", "/work/proj", "not json", nil, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []string
			run := func(args ...string) ([]byte, error) {
				calls = append(calls, strings.Join(args, " "))
				return []byte(tt.out), tt.err
			}
			got, err := Running(tt.project, run)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(calls, []string{"agent list"}) {
				t.Errorf("herdr calls %q, want one agent list", calls)
			}
		})
	}
}
