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
	create = "tab create --cwd /p --label " + change + " --focus"
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
			want: []string{get, create, launch, prompt},
		},
		{
			name: "agent running: focus only",
			out:  map[string]string{"agent get": "agent-get.json", "pane process-info": "process-info-running.json"},
			want: []string{get, info, focus},
		},
		{
			name: "agent exited: restart in its tab",
			out:  map[string]string{"agent get": "agent-get.json", "pane process-info": "process-info-shell.json"},
			want: []string{get, info, launch, prompt, focus},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &fakeHerdr{t: t, out: tt.out, fail: tt.fail}
			if err := Start("/p", change, h.run, fakeList("other", change)); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(h.calls, tt.want) {
				t.Errorf("herdr calls\n got %q\nwant %q", h.calls, tt.want)
			}
		})
	}
}

func TestStartRejectsChangesThatAreNotOpen(t *testing.T) {
	h := &fakeHerdr{t: t}
	err := Start("/p", "m1-bord-view", h.run, fakeList(change))
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
	if err := Start("/p", change, h.run, broken); err == nil || len(h.calls) != 0 {
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
			err := Start("/p", change, h.run, fakeList(change))
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
			err := Start("/p", change, h.run, fakeList(change))
			if err == nil || !strings.Contains(err.Error(), tt.wants) {
				t.Fatalf("want error with %q, got %v", tt.wants, err)
			}
			if got := h.calls[len(h.calls)-1]; got != info {
				t.Errorf("must stop after process-info, last call %q", got)
			}
		})
	}
}
