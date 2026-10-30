package start

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
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
			b, err := os.ReadFile(filepath.Join("testdata", "herdr", file))
			if errors.Is(err, os.ErrNotExist) {
				b, err = os.ReadFile(filepath.Join("..", "..", "testdata", "herdr", file))
			}
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

// off is the config with worktrees off; noGit fails like git outside a repository.
var off = Worktree{}

func noGit(...string) ([]byte, error) { return nil, errors.New("fatal: not a git repository") }

// fakeGit answers each git call found in out exactly, and fails every other call.
type fakeGit struct {
	out   map[string]string
	calls []string
}

func (g *fakeGit) run(args ...string) ([]byte, error) {
	call := strings.Join(args, " ")
	g.calls = append(g.calls, call)
	if out, ok := g.out[call]; ok {
		return []byte(out), nil
	}
	return nil, errors.New("exit status 1")
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
			if err := Start(proj, change, "claude", "/opsx:apply "+change, off, h.run, fakeList("other", change), noGit); err != nil {
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
		{"node-based agent running: focus only", map[string]string{"agent get": "agent-get.json", "pane process-info": "process-info-node.json"}, nil, []string{get, info, focus}, "", ""},
		{"other program in the foreground: focus only", map[string]string{"agent get": "agent-get.json", "pane process-info": "process-info-other.json"}, nil, []string{get, info, focus}, "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &fakeHerdr{t: t, out: tt.out, fail: tt.fail}
			pane, tab, err := Prepare(proj, change, off, h.run, noGit)
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

func TestLaunchPromptWithLeadingDash(t *testing.T) { // herdr reads the text positionally and rejects --
	var last []string
	run := func(args ...string) ([]byte, error) { last = args; return nil, nil }
	if err := Launch(change, "claude", "-x "+change, "w2:p1J", "w2:tG", run); err != nil {
		t.Fatal(err)
	}
	if want := []string{"agent", "prompt", change, "-x " + change}; !reflect.DeepEqual(last, want) {
		t.Errorf("got %q, want %q", last, want)
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
	err := Start(proj+"2", change, "claude", "/opsx:apply "+change, off, h.run, fakeList(change), noGit)
	if err == nil || !strings.Contains(err.Error(), change) || !strings.Contains(err.Error(), "already runs in "+proj) {
		t.Fatalf("want already-runs error naming the change and %s, got %v", proj, err)
	}
	if !reflect.DeepEqual(h.calls, []string{get}) {
		t.Errorf("want no focus, restart or tab, got %q", h.calls)
	}
}

func TestStartAgentWithoutCwd(t *testing.T) {
	h := &fakeHerdr{t: t, out: map[string]string{"agent get": "agent-get-no-cwd.json", "pane process-info": "process-info-running.json"}}
	if err := Start(proj, change, "claude", "/opsx:apply "+change, off, h.run, fakeList(change), noGit); err != nil {
		t.Fatalf("an unknown cwd must not refuse the agent: %v", err)
	}
	if !reflect.DeepEqual(h.calls, []string{get, info, focus}) {
		t.Errorf("want focus only, got %q", h.calls)
	}
}

func TestStartRejectsChangesThatAreNotOpen(t *testing.T) {
	h := &fakeHerdr{t: t}
	err := Start(proj, "m1-bord-view", "claude", "/opsx:apply m1-bord-view", off, h.run, fakeList(change), noGit)
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
	if err := Start(proj, change, "claude", "/opsx:apply "+change, off, h.run, broken, noGit); err == nil || len(h.calls) != 0 {
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
			err := Start(proj, change, "claude", "/opsx:apply "+change, off, h.run, fakeList(change), noGit)
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
		{"process-info fails", map[string]string{"agent get": "agent-get.json"}, map[string]error{"pane process-info": errors.New("pane_not_found")}, "pane_not_found"},
		{"process-info without shell pid", map[string]string{"agent get": "agent-get.json"}, nil, "process-info"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &fakeHerdr{t: t, out: tt.out, fail: tt.fail}
			err := Start(proj, change, "claude", "/opsx:apply "+change, off, h.run, fakeList(change), noGit)
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
			got, err := Running(tt.project, run, noGit)
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

// wt is the linked worktree of the Change, where herdr's fixtures put it.
const wt = "/Users/U758014/.herdr/worktrees/openspec-kanban/" + change

// gitWith answers the calls worktree.List and worktree.Base make with porcelain listing the
// checkouts as path, branch pairs, origin/HEAD at origin/main and a local main, plus extra.
func gitWith(extra map[string]string, checkouts ...string) *fakeGit {
	list := ""
	for i := 0; i < len(checkouts); i += 2 {
		list += "worktree " + checkouts[i] + "\nHEAD 1234\nbranch refs/heads/" + checkouts[i+1] + "\n\n"
	}
	out := map[string]string{
		"rev-parse --show-prefix":                       "\n",
		"worktree list --porcelain":                     list,
		"symbolic-ref --short refs/remotes/origin/HEAD": "origin/main\n",
		"rev-parse --verify --quiet refs/heads/main":    "1234\n",
		"cat-file -e main:openspec/changes/" + change:   "",
	}
	for k, v := range extra {
		out[k] = v
	}
	return &fakeGit{out: out}
}

func TestPrepareWorktree(t *testing.T) {
	on := Worktree{On: true}
	var (
		wtCreate = "worktree create --cwd " + proj + " --branch " + change + " --base main --label " + change + " --no-focus"
		wtOpen   = "worktree open --cwd " + proj + " --path " + wt + " --label " + change + " --no-focus"
	)
	tests := []struct {
		name            string
		wt              Worktree
		git             *fakeGit
		out             map[string]string
		want            []string
		wantPane, wantT string
		wantGit         string // a git call that must have been made
	}{
		{"new worktree from the detected base", on, gitWith(nil, proj, "main"),
			map[string]string{"worktree create": "worktree-create.json"}, []string{get, wtCreate}, "w3:p1", "w3:t1",
			"cat-file -e main:openspec/changes/" + change},
		{"configured base", Worktree{On: true, Base: "release/2.x"},
			gitWith(map[string]string{"cat-file -e release/2.x:openspec/changes/" + change: ""}, proj, "main"),
			map[string]string{"worktree create": "worktree-create.json"},
			[]string{get, "worktree create --cwd " + proj + " --branch " + change + " --base release/2.x --label " + change + " --no-focus"}, "w3:p1", "w3:t1",
			"cat-file -e release/2.x:openspec/changes/" + change},
		{"existing branch without a checkout", on,
			gitWith(map[string]string{"rev-parse --verify --quiet refs/heads/" + change: "1234\n", "cat-file -e " + change + ":openspec/changes/" + change: ""}, proj, "main"),
			map[string]string{"worktree create": "worktree-create.json"},
			[]string{get, "worktree create --cwd " + proj + " --branch " + change + " --label " + change + " --no-focus"}, "w3:p1", "w3:t1",
			"cat-file -e " + change + ":openspec/changes/" + change},
		{"closed linked checkout: opened", on, gitWith(nil, proj, "main", wt, change),
			map[string]string{"worktree open": "worktree-open.json"}, []string{get, wtOpen}, "w3:p1", "w3:t1", ""},
		{"open linked checkout: new tab in it", on, gitWith(nil, proj, "main", wt, change),
			map[string]string{"worktree open": "worktree-open-already.json", "tab create": "tab-create.json"},
			[]string{get, wtOpen, "tab create --workspace w3 --cwd " + wt + " --label " + change + " --no-focus"}, "w2:p1J", "w2:tG", ""},
		{"Project's own checkout on the branch", on, gitWith(nil, proj, change),
			map[string]string{"tab create": "tab-create.json"}, []string{get, create}, "w2:p1J", "w2:tG", ""},
		{"no checkouts", on, &fakeGit{},
			map[string]string{"tab create": "tab-create.json"}, []string{get, create}, "w2:p1J", "w2:tG", ""},
		{"worktree off", off, gitWith(nil, proj, "main"),
			map[string]string{"tab create": "tab-create.json"}, []string{get, create}, "w2:p1J", "w2:tG", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &fakeHerdr{t: t, out: tt.out, fail: map[string]error{"agent get": notFound}}
			pane, tab, err := Prepare(proj, change, tt.wt, h.run, tt.git.run)
			if err != nil {
				t.Fatal(err)
			}
			if pane != tt.wantPane || tab != tt.wantT {
				t.Errorf("got pane %q tab %q, want %q %q", pane, tab, tt.wantPane, tt.wantT)
			}
			if !reflect.DeepEqual(h.calls, tt.want) {
				t.Errorf("herdr calls\n got %q\nwant %q", h.calls, tt.want)
			}
			if tt.wantGit != "" && !slices.Contains(tt.git.calls, tt.wantGit) {
				t.Errorf("git calls %q lack %q", tt.git.calls, tt.wantGit)
			}
			if tt.wt == off && len(tt.git.calls) != 0 {
				t.Errorf("worktree off must not call git, got %q", tt.git.calls)
			}
			for _, c := range h.calls {
				if strings.Contains(c, "--trust-repository") || strings.HasPrefix(c, "worktree create") && strings.Contains(c, "--path") {
					t.Errorf("call %q must not trust the repository or set the worktree path", c)
				}
			}
		})
	}
}

func TestPrepareChangeNotOnBase(t *testing.T) {
	git := gitWith(nil, proj, "main")
	delete(git.out, "cat-file -e main:openspec/changes/"+change)
	h := &fakeHerdr{t: t, fail: map[string]error{"agent get": notFound}}
	_, _, err := Prepare(proj, change, Worktree{On: true}, h.run, git.run)
	if err == nil || !strings.Contains(err.Error(), change+" is not on main") {
		t.Fatalf("want error naming %s and main, got %v", change, err)
	}
	if !reflect.DeepEqual(h.calls, []string{get}) {
		t.Errorf("nothing may be created, herdr calls %q", h.calls)
	}
}

func TestPrepareAgentInCheckout(t *testing.T) {
	const project = "/work/openspec-kanban" // agent-get.json's cwd is proj, a linked checkout here
	tests := []struct {
		name     string
		info     string
		want     []string
		wantPane string
	}{
		{"running: focused", "process-info-running.json", []string{get, info, focus}, ""},
		{"idle: its tab reused", "process-info-shell.json", []string{get, info}, "w2:p1J"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &fakeHerdr{t: t, out: map[string]string{"agent get": "agent-get.json", "pane process-info": tt.info}}
			pane, _, err := Prepare(project, change, Worktree{On: true}, h.run, gitWith(nil, project, "main", proj, change).run)
			if err != nil || pane != tt.wantPane || !reflect.DeepEqual(h.calls, tt.want) {
				t.Errorf("got pane %q, %v, calls %q; want pane %q, calls %q", pane, err, h.calls, tt.wantPane, tt.want)
			}
		})
	}
	t.Run("worktree off: still focused", func(t *testing.T) {
		h := &fakeHerdr{t: t, out: map[string]string{"agent get": "agent-get.json", "pane process-info": "process-info-running.json"}}
		if _, _, err := Prepare(project, change, off, h.run, gitWith(nil, project, "main", proj, change).run); err != nil || !reflect.DeepEqual(h.calls, []string{get, info, focus}) {
			t.Errorf("got %v, calls %q", err, h.calls)
		}
	})
	t.Run("outside the Project and its checkouts: refused", func(t *testing.T) {
		h := &fakeHerdr{t: t, out: map[string]string{"agent get": "agent-get.json"}}
		_, _, err := Prepare(project, change, Worktree{On: true}, h.run, gitWith(nil, project, "main", wt, change).run)
		if err == nil || !strings.Contains(err.Error(), "already runs in "+proj) || !reflect.DeepEqual(h.calls, []string{get}) {
			t.Errorf("want already-runs error and only agent get, got %v, %q", err, h.calls)
		}
	})
}

func TestRunningInCheckouts(t *testing.T) {
	list, err := os.ReadFile(filepath.Join("testdata", "herdr", "agent-list.json"))
	if err != nil {
		t.Fatal(err)
	}
	run := func(...string) ([]byte, error) { return list, nil }
	got, err := Running("/work/pro", run, gitWith(nil, "/work/pro", "main", "/work/proj", "m1-board-view").run) // /work/proj stands in for a linked checkout
	if want := map[string]bool{"m1-board-view": true, "m2-card-detail": true, "m3-sidebar-summary": true}; err != nil || !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, %v, want %v", got, err, want)
	}
}

func TestPrepareWorktreeFails(t *testing.T) {
	boom := errors.New("boom")
	tests := []struct {
		name, fail string
		git        *fakeGit
	}{
		{"worktree create", "worktree create", gitWith(nil, proj, "main")},
		{"worktree open", "worktree open", gitWith(nil, proj, "main", wt, change)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &fakeHerdr{t: t, fail: map[string]error{"agent get": notFound, tt.fail: boom}}
			if _, _, err := Prepare(proj, change, Worktree{On: true}, h.run, tt.git.run); err == nil || !strings.Contains(err.Error(), tt.fail+": boom") {
				t.Errorf("want %s error, got %v", tt.fail, err)
			}
		})
	}
}
