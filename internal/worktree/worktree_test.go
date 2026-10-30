package worktree

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
)

// fakeGit answers each call whose joined args start with a key; unknown calls fail.
type fakeGit struct {
	calls []string
	out   map[string]string
}

func (f *fakeGit) run(args ...string) ([]byte, error) {
	call := strings.Join(args, " ")
	f.calls = append(f.calls, call)
	if out, ok := f.out[call]; ok {
		return []byte(out), nil
	}
	return nil, errors.New("git " + call + ": exit status 1")
}

func fixture(t *testing.T) string {
	b, err := os.ReadFile("testdata/worktree-list.txt")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestList(t *testing.T) {
	const prefix, list = "rev-parse --show-prefix", "worktree list --porcelain"
	tests := []struct {
		name  string
		out   map[string]string
		want  []Checkout
		calls []string
	}{
		{"top level", map[string]string{prefix: "\n", list: fixture(t)},
			[]Checkout{{"/work/proj", "main"}, {"/work/wt/proj/m1-board-view", "m1-board-view"}},
			[]string{prefix, list}},
		{"subdirectory", map[string]string{prefix: "internal/\n", list: fixture(t)}, nil, []string{prefix}},
		{"outside git", nil, nil, []string{prefix}},
		{"list fails", map[string]string{prefix: "\n"}, nil, []string{prefix, list}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &fakeGit{out: tt.out}
			got := List(g.run)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(g.calls, tt.calls) {
				t.Errorf("calls %v, want %v", g.calls, tt.calls)
			}
		})
	}
}

func TestBase(t *testing.T) {
	const head = "symbolic-ref --short refs/remotes/origin/HEAD"
	ref := func(n string) string { return "rev-parse --verify --quiet refs/heads/" + n }
	cur := "branch --show-current"
	tests := []struct {
		name       string
		configured string
		out        map[string]string
		want       string
		calls      []string
	}{
		{"configured", "release", nil, "release", nil},
		{"origin head with local", "", map[string]string{head: "origin/develop\n", ref("develop"): "x\n"},
			"develop", []string{head, ref("develop")}},
		{"origin head without local", "", map[string]string{head: "origin/develop\n"},
			"origin/develop", []string{head, ref("develop")}},
		{"no origin head", "", map[string]string{ref("master"): "x\n", ref("develop"): "x\n"},
			"master", []string{head, ref("main"), ref("master")}},
		{"current branch", "", map[string]string{cur: "feature\n"},
			"feature", []string{head, ref("main"), ref("master"), ref("develop"), cur}},
		{"detached", "", map[string]string{cur: "\n"},
			"HEAD", []string{head, ref("main"), ref("master"), ref("develop"), cur}},
		{"git failing", "", nil,
			"HEAD", []string{head, ref("main"), ref("master"), ref("develop"), cur}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &fakeGit{out: tt.out}
			if got := Base(tt.configured, g.run); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
			if !reflect.DeepEqual(g.calls, tt.calls) {
				t.Errorf("calls %v, want %v", g.calls, tt.calls)
			}
		})
	}
}
