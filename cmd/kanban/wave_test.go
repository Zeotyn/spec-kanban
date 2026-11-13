package main

import (
	"errors"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/zeotyn/spec-kanban/internal/cache"
	"github.com/zeotyn/spec-kanban/internal/herdr"
	"github.com/zeotyn/spec-kanban/internal/source/openspec"
	"github.com/zeotyn/spec-kanban/internal/work"
)

// waveFixture records wave 1 of a Project with members f5-bar and c4-foo in the user cache dir and
// returns the record and a Source whose open Changes are open.
func waveFixture(t *testing.T, open ...string) (cache.Wave, func(string) work.Source) {
	t.Helper()
	dir := userDir(t)
	root := t.TempDir()
	w := cache.Wave{Project: filepath.Join(root, "proj"), Number: 1, Branch: "wave-1", Path: filepath.Join(root, "wt", "wave-1"), Base: "main",
		Coordinator: "w3:p1", Main: "w1:p9", Members: []string{"f5-bar", "c4-foo"}}
	if err := cache.WriteWave(dir, w); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HERDR_PLUGIN_CONFIG_DIR", t.TempDir())
	t.Setenv("HERDR_PANE_ID", "w3:p1")
	list := `{"changes":[`
	for i, c := range open {
		if i > 0 {
			list += ","
		}
		list += `{"name":"` + c + `","completedTasks":7,"totalTasks":7}`
	}
	cli := func(string) herdr.Runner {
		return func(...string) ([]byte, error) { return []byte(list + `]}`), nil }
	}
	return w, func(string) work.Source { return openspec.New(cli) }
}

// recorder returns a herdr runner that records its calls and an idle Main agent for agent get.
func recorder(calls *[]string, main bool) herdr.Runner {
	return func(args ...string) ([]byte, error) {
		*calls = append(*calls, strings.Join(args, " "))
		if args[0] == "agent" && args[1] == "get" {
			if !main {
				return nil, errors.New("agent_not_found")
			}
			return []byte(`{"result":{"agent":{"agent":"claude","agent_status":"idle"}}}`), nil
		}
		return nil, nil
	}
}

func TestWaveRefusals(t *testing.T) {
	w, sourceFor := waveFixture(t, "f5-bar", "c4-foo")
	dropped := w
	dropped.Dropped = []string{"c4-foo"}
	finished := w
	finished.Finished = true
	tests := []struct {
		name   string
		record *cache.Wave
		args   []string
		want   string
	}{
		{"unknown wave", nil, []string{"start", "2", "f5-bar"}, "unknown wave 2"},
		{"not a number", nil, []string{"drop", "x", "f5-bar"}, `wave number "x"`},
		{"not a member", nil, []string{"start", "1", "f9-other"}, "f9-other is not a member of wave 1"},
		{"dropped member", &dropped, []string{"start", "1", "c4-foo"}, "c4-foo was dropped from wave 1"},
		{"finished wave: start", &finished, []string{"start", "1", "f5-bar"}, "wave 1 is finished"},
		{"finished wave: drop", &finished, []string{"drop", "1", "f5-bar"}, "wave 1 is finished"},
		{"finished wave: escalate", &finished, []string{"escalate", "1", "help"}, "wave 1 is finished"},
		{"finished wave: finish", &finished, []string{"finish", "1"}, "wave 1 is finished"},
		{"drop a non-member", nil, []string{"drop", "1", "f9-other"}, "f9-other is not a member of wave 1"},
		{"unknown subcommand", nil, []string{"stop", "1"}, `unknown wave command "stop"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := w
			if tt.record != nil {
				rec = *tt.record
			}
			if err := cache.WriteWave(cache.UserDir(), rec); err != nil {
				t.Fatal(err)
			}
			var calls []string
			var out strings.Builder
			err := wave(tt.args, filepath.Join(w.Path, "openspec"), &out, recorder(&calls, true), sourceFor, noGitFor)
			if err == nil || !strings.Contains(err.Error(), tt.want) || len(calls) != 0 || out.Len() != 0 {
				t.Errorf("want an error with %q and nothing sent, got %v, %q, %q", tt.want, err, calls, out.String())
			}
		})
	}
}

func noGitFor(string) herdr.Runner { return noGit }

func TestWaveDrop(t *testing.T) {
	w, sourceFor := waveFixture(t, "f5-bar", "c4-foo")
	var calls []string
	var out strings.Builder
	if err := wave([]string{"drop", "1", "c4-foo"}, w.Project, &out, recorder(&calls, true), sourceFor, noGitFor); err != nil {
		t.Fatal(err)
	}
	if got, _ := cache.FindWave(cache.UserDir(), 1, w.Path); !reflect.DeepEqual(got.Dropped, []string{"c4-foo"}) || out.String() != "dropped c4-foo\n" || len(calls) != 0 {
		t.Errorf("dropped %q, printed %q, herdr calls %q", got.Dropped, out.String(), calls)
	}
	if err := wave([]string{"drop", "1", "c4-foo"}, w.Project, &out, recorder(&calls, true), sourceFor, noGitFor); err != nil {
		t.Fatal(err)
	}
	if got, _ := cache.FindWave(cache.UserDir(), 1, w.Path); len(got.Dropped) != 1 {
		t.Errorf("dropping twice records it once, got %q", got.Dropped)
	}
}

func TestWaveEscalate(t *testing.T) {
	w, sourceFor := waveFixture(t)
	text := "Wave 1 needs a human: c4-foo and f5-bar change the same API in opposite ways. Its coordinator wave-1 waits in tab wave-1, worktree " + w.Path +
		` on branch wave-1. Answer it with herdr agent prompt wave-1 "<answer>", or tell the user exactly what to do.`
	args := []string{"escalate", "1", "c4-foo and f5-bar", "change the same API in opposite ways"}
	t.Run("Main agent takes it", func(t *testing.T) {
		var calls []string
		var out strings.Builder
		if err := wave(args, w.Path, &out, recorder(&calls, true), sourceFor, noGitFor); err != nil {
			t.Fatal(err)
		}
		if want := []string{"agent get w1:p9", "agent prompt w1:p9 " + text}; !reflect.DeepEqual(calls, want) {
			t.Errorf("herdr calls\n got %q\nwant %q", calls, want)
		}
		if out.String() != "Sent to the Main agent. End your turn.\n" {
			t.Errorf("printed %q", out.String())
		}
	})
	t.Run("no Main agent", func(t *testing.T) {
		var calls []string
		var out strings.Builder
		if err := wave(args, w.Path, &out, recorder(&calls, false), sourceFor, noGitFor); err != nil {
			t.Fatal(err)
		}
		if want := []string{"agent get w1:p9", "notification show spec-kanban: wave 1 needs you --body " + text}; !reflect.DeepEqual(calls, want) {
			t.Errorf("herdr calls\n got %q\nwant %q", calls, want)
		}
		if out.String() != "No Main agent took it; herdr shows a notification. End your turn.\n" {
			t.Errorf("printed %q", out.String())
		}
	})
}

func TestWaveFinish(t *testing.T) {
	t.Run("a member is still open", func(t *testing.T) {
		w, sourceFor := waveFixture(t, "c4-foo", "f9-other")
		var calls []string
		var out strings.Builder
		err := wave([]string{"finish", "1"}, w.Path, &out, recorder(&calls, true), sourceFor, noGitFor)
		if err == nil || err.Error() != "wave 1 is not done: c4-foo is neither archived on wave-1 nor dropped" || len(calls) != 0 {
			t.Errorf("got %v, calls %q", err, calls)
		}
		if got, _ := cache.FindWave(cache.UserDir(), 1, w.Path); got.Finished {
			t.Error("the wave must stay active")
		}
	})
	for _, main := range []bool{true, false} {
		t.Run("all through", func(t *testing.T) {
			w, sourceFor := waveFixture(t, "c4-foo")
			if err := wave([]string{"drop", "1", "c4-foo"}, w.Path, &strings.Builder{}, recorder(new([]string), true), sourceFor, noGitFor); err != nil {
				t.Fatal(err)
			}
			var calls []string
			var out strings.Builder
			if err := wave([]string{"finish", "1"}, w.Path, &out, recorder(&calls, main), sourceFor, noGitFor); err != nil {
				t.Fatal(err)
			}
			text := "Wave 1 is finished: archived on wave-1: f5-bar; dropped: c4-foo. 1. Review branch wave-1 in its worktree " + w.Path +
				". 2. Merge it into main. 3. Remove the worktree with git worktree remove " + w.Path + ", never with --force, and keep the branch." +
				" 4. If the merge or the removal does not go through cleanly, stop and tell the user."
			want, printed := []string{"agent get w1:p9", "agent prompt w1:p9 " + text}, "Sent to the Main agent.\n"
			if !main {
				want, printed = []string{"agent get w1:p9", "notification show spec-kanban: wave 1 finished --body " + text}, "No Main agent took it; herdr shows a notification.\n"
			}
			if !reflect.DeepEqual(calls, want) || out.String() != printed {
				t.Errorf("herdr calls\n got %q\nwant %q\nprinted %q", calls, want, out.String())
			}
			if got, _ := cache.FindWave(cache.UserDir(), 1, w.Path); !got.Finished {
				t.Error("the wave must be finished")
			}
		})
	}
}

func TestWaveStart(t *testing.T) {
	for _, existing := range []bool{false, true} {
		t.Run(map[bool]string{false: "new worktree from the wave", true: "existing checkout kept"}[existing], func(t *testing.T) { waveStart(t, existing) })
	}
}

func waveStart(t *testing.T, existing bool) {
	w, sourceFor := waveFixture(t, "f5-bar", "c4-foo")
	var spawned []string
	old := spawn
	spawn = func(args ...string) error { spawned = args; return nil }
	t.Cleanup(func() { spawn = old })
	var calls []string
	run := func(args ...string) ([]byte, error) {
		calls = append(calls, strings.Join(args, " "))
		switch strings.Join(args[:2], " ") {
		case "agent get":
			if args[2] == "w3:p1" { // the coordinator: its kind is the members' kind
				return []byte(`{"result":{"agent":{"agent":"codex","agent_status":"working"}}}`), nil
			}
			return nil, errors.New("agent_not_found")
		case "worktree create", "worktree open":
			return []byte(`{"result":{"root_pane":{"pane_id":"w4:p1","tab_id":"w4:t1"},"tab":{"tab_id":"w4:t1"},"worktree":{"path":"/wt/c4-foo"}}}`), nil
		}
		return nil, nil
	}
	git := func(string) herdr.Runner {
		return func(args ...string) ([]byte, error) {
			switch strings.Join(args, " ") {
			case "cat-file -e wave-1:openspec/changes/c4-foo":
				return nil, nil
			case "rev-parse --show-prefix":
				return []byte("\n"), nil
			case "worktree list --porcelain":
				list := "worktree " + w.Project + "\nbranch refs/heads/main\n\nworktree " + w.Path + "\nbranch refs/heads/wave-1\n"
				if existing {
					list += "\nworktree /wt/c4-foo\nbranch refs/heads/c4-foo\n"
				}
				return []byte(list), nil
			}
			return nil, errors.New("exit status 1")
		}
	}
	var out strings.Builder
	if err := wave([]string{"start", "1", "c4-foo"}, w.Path, &out, run, sourceFor, git); err != nil {
		t.Fatal(err)
	}
	want := "worktree create --cwd " + w.Project + " --branch c4-foo --base wave-1 --label c4-foo --no-focus"
	if existing {
		want = "worktree open --cwd " + w.Project + " --path /wt/c4-foo --label c4-foo --no-focus"
	}
	if !slices.Contains(calls, want) {
		t.Errorf("want %q, herdr calls %q", want, calls)
	}
	for _, c := range calls {
		if strings.Contains(c, "focus ") && !strings.Contains(c, "--no-focus") {
			t.Errorf("a member must not take focus: %q", c)
		}
	}
	if len(spawned) != 6 || spawned[0] != "wave-member" || spawned[1] != "1" || spawned[2] != "c4-foo" || spawned[3] != "w4:p1" || spawned[4] != "codex" {
		t.Fatalf("spawned %q", spawned)
	}
	if p := spawned[5]; !strings.HasPrefix(p, "/opsx:apply c4-foo If you cannot continue without a human, run ") ||
		!strings.HasSuffix(p, " escalate c4-foo \"<what you need>\" and end your turn. Do not archive c4-foo: the coordinator of wave 1 merges and archives it.") {
		t.Errorf("member prompt %q", p)
	}
	if out.String() != "started c4-foo\n" {
		t.Errorf("printed %q", out.String())
	}
}

func TestWaveMember(t *testing.T) {
	w, _ := waveFixture(t)
	t.Setenv("HERDR_PLUGIN_STATE_DIR", t.TempDir())
	var calls []string
	run := func(args ...string) ([]byte, error) {
		calls = append(calls, strings.Join(args, " "))
		if args[0] == "agent" && args[1] == "get" {
			return nil, errors.New("agent_not_found") // the agent went away: watching ends
		}
		return nil, nil
	}
	if err := waveMember(1, "c4-foo", "w4:p1", "claude", "/opsx:apply c4-foo", w.Path, noPane(run)); err != nil {
		t.Fatal(err)
	}
	if want := []string{"agent start c4-foo --kind claude --pane w4:p1 --timeout 60000", "agent prompt c4-foo /opsx:apply c4-foo", "agent get c4-foo"}; !reflect.DeepEqual(calls, want) {
		t.Errorf("want no focus, then the watcher\n got %q\nwant %q", calls, want)
	}
	if got, _ := cache.ReadMain(cache.UserDir(), w.Project, "c4-foo"); got != w.Coordinator {
		t.Errorf("Main pane record %q, want the coordinator %q", got, w.Coordinator)
	}
	calls = nil
	if err := waveMember(2, "c4-foo", "w4:p1", "claude", "/opsx:apply c4-foo", w.Path, noPane(run)); err == nil || err.Error() != "unknown wave 2" ||
		calls[len(calls)-1] != "notification show spec-kanban: start failed --body unknown wave 2" {
		t.Errorf("got %v, calls %q", err, calls)
	}
}

func TestWaveLaunchNotifiesOnFailure(t *testing.T) {
	userDir(t)
	t.Setenv("HERDR_PLUGIN_CONFIG_DIR", t.TempDir())
	t.Setenv("KANBAN_PROJECT", t.TempDir())
	broken := func(string) work.Source {
		return openspec.New(func(string) herdr.Runner {
			return func(...string) ([]byte, error) { return nil, errors.New("openspec: not found") }
		})
	}
	var calls []string
	err := waveLaunch([]string{"f5-bar", "c4-foo"}, recorder(&calls, true), broken, noGitFor)
	if err == nil || len(calls) != 1 || !strings.HasPrefix(calls[0], "notification show spec-kanban: wave failed --body ") {
		t.Errorf("got %v, calls %q", err, calls)
	}
}
