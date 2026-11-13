package start

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/zeotyn/spec-kanban/internal/herdr"
	"github.com/zeotyn/spec-kanban/internal/source/speckit"
	"github.com/zeotyn/spec-kanban/internal/work"
)

// script answers each herdr call whose joined args start with a key with its answers in turn, the
// last one repeating: a fixture file name or an error. Other calls get {}. The longest key wins.
type script struct {
	t       *testing.T
	calls   []string
	answers map[string][]any
	used    map[string]int
}

func (s *script) run(args ...string) ([]byte, error) {
	call := strings.Join(args, " ")
	s.calls = append(s.calls, call)
	if len(s.calls) > 50 { // Watch loops: end it like a closed tab
		s.t.Errorf("Watch does not end: %q", s.calls[len(s.calls)-5:])
		return nil, notFound
	}
	best := ""
	for prefix := range s.answers {
		if strings.HasPrefix(call, prefix) && len(prefix) > len(best) {
			best = prefix
		}
	}
	answers := s.answers[best]
	if len(answers) == 0 {
		return []byte(`{}`), nil
	}
	if s.used == nil {
		s.used = map[string]int{}
	}
	a := answers[min(s.used[best], len(answers)-1)]
	s.used[best]++
	if err, ok := a.(error); ok {
		return nil, err
	}
	b, err := os.ReadFile(filepath.Join("testdata", "herdr", a.(string)))
	if err != nil {
		s.t.Fatal(err)
	}
	return b, nil
}

// loads returns the Change lists (or errors) in turn, the last one repeating.
func loads(results ...any) func() ([]work.Item, error) {
	i := 0
	return func() ([]work.Item, error) {
		r := results[min(i, len(results)-1)]
		i++
		if err, ok := r.(error); ok {
			return nil, err
		}
		return r.([]work.Item), nil
	}
}

const (
	mainPane = "w1:p9"
	claudeID = "00000009-1111-4222-8333-444455556666"
)

var (
	wget      = "agent get " + change
	wready    = "agent wait " + change + " --until working --timeout 60000"
	wsettle   = "agent wait " + change
	wworking  = "agent wait " + change + " --until working"
	wstall    = "agent wait " + change + " --until working --timeout 600000"
	wclose    = "pane close w5:p1"
	mget      = "agent get " + mainPane
	all7      = []work.Item{{Name: "other"}, {Name: change, Done: 7, Total: 7}}
	resumeTip = " To look into its session press s on its Card in the Board, or run in " + proj + ": claude --resume " + claudeID
	plainText = change + " is finished (7/7 tasks). Please review it and archive it with /opsx:archive " + change + "." + resumeTip
	cleanEnd  = []string{wclose, mget, "agent prompt " + mainPane + " " + plainText}
)

func watched(load func() ([]work.Item, error), rec *[]herdr.Session) Watched {
	return Watched{Project: proj, Change: change, Main: mainPane, On: true, Load: load, Source: src,
		Record: func(s herdr.Session) { *rec = append(*rec, s) }, Escalated: func() bool { return false }, Mark: func(string) {}, Clear: func() {}}
}

// stallText is the Stall prompt for change at progress, "" when unknown.
func stallText(progress string) string {
	at := ""
	if progress != "" {
		at = " at " + progress
	}
	return change + " stopped" + at + " 10 minutes ago without asking for help. Read its output with: herdr agent read " + change + " --lines 40." +
		" If it waits for something that finishes by itself, do nothing. If it needs a human, tell the user what it needs." +
		" To let it continue, run: herdr agent prompt " + change + ` "<what to do>". Do not close its tab.`
}

func TestWatch(t *testing.T) {
	boom := errors.New("boom")
	tests := []struct {
		name    string
		answers map[string][]any
		load    func() ([]work.Item, error)
		want    []string
	}{
		{"clean idle settle: handoff", map[string][]any{wget: {"agent-get-watched.json"}, mget: {"agent-get-main.json"}},
			loads(all7), append([]string{wget, wready, wsettle, wget}, cleanEnd...)},
		{"done counts as clean", map[string][]any{wget: {"agent-get-watched.json", "agent-get-watched-done.json"}, mget: {"agent-get-main.json"}},
			loads(all7), append([]string{wget, wready, wsettle, wget}, cleanEnd...)},
		{"wait for working fails: still evaluated", map[string][]any{wget: {"agent-get-watched.json"}, wready: {notFound}, mget: {"agent-get-main.json"}},
			loads(all7), append([]string{wget, wready, wsettle, wget}, cleanEnd...)},
		{"tasks left: no notification, handoff at the next settle", map[string][]any{wget: {"agent-get-watched.json"}, mget: {"agent-get-main.json"}},
			loads([]work.Item{{Name: change, Done: 3, Total: 7}}, all7),
			append([]string{wget, wready, wsettle, wget, wget, wstall, wsettle, wget}, cleanEnd...)},
		{"blocked with all tasks done: no notification, no Stall", map[string][]any{wget: {"agent-get-watched.json", "agent-get-watched-blocked.json"}, wready: {}, wworking: {notFound}},
			loads(all7), []string{wget, wready, wsettle, wget, wget, wworking}},
		{"load error: unfinished", map[string][]any{wget: {"agent-get-watched.json", "agent-get-watched.json", "agent-get-watched.json", notFound}, wready: {}, wstall: {notFound}},
			loads(boom), []string{wget, wready, wsettle, wget, wget, wstall, wget}},
		{"Change neither open nor archived: unfinished", map[string][]any{wget: {"agent-get-watched.json", "agent-get-watched.json", "agent-get-watched.json", notFound}, wready: {}, wstall: {notFound}},
			loads([]work.Item{{Name: "other"}}), []string{wget, wready, wsettle, wget, wget, wstall, wget}},
		{"no tasks: unfinished", map[string][]any{wget: {"agent-get-watched.json", "agent-get-watched.json", "agent-get-watched.json", notFound}, wready: {}, wstall: {notFound}},
			loads([]work.Item{{Name: change}}), []string{wget, wready, wsettle, wget, wget, wstall, wget}},
		{"blocked with a load error: no notification", map[string][]any{wget: {"agent-get-watched.json", "agent-get-watched-blocked.json"}, wready: {}, wworking: {notFound}},
			loads(boom), []string{wget, wready, wsettle, wget, wget, wworking}},
		{"turn ended while evaluating: evaluated again at once", map[string][]any{wget: {"agent-get-watched.json", "agent-get-watched-blocked.json", "agent-get-watched-later.json"}, mget: {"agent-get-main.json"}},
			loads(all7), append([]string{wget, wready, wsettle, wget, wget, wsettle, wget}, cleanEnd...)},
		{"working again when read: no handoff, no notification", map[string][]any{wget: {"agent-get-watched.json", "agent-get-watched-working.json", "agent-get-watched.json"}, mget: {"agent-get-main.json"}},
			loads(all7), append([]string{wget, wready, wsettle, wget, wsettle, wget}, cleanEnd...)},
		{"agent gone while settling", map[string][]any{wget: {"agent-get-watched.json"}, wsettle: {notFound}},
			loads(all7), []string{wget, wready, wsettle}},
		{"agent gone after settling", map[string][]any{wget: {"agent-get-watched.json", notFound}},
			loads(all7), []string{wget, wready, wsettle, wget}},
		{"agent gone at once", map[string][]any{wget: {notFound}}, loads(all7), []string{wget}},
		{"archived in the Project", map[string][]any{wget: {"agent-get-watched.json"}, mget: {"agent-get-main.json"}},
			loads([]work.Item{{Name: change, Phase: work.Archived, ArchivedOn: "2026-10-30"}}),
			[]string{wget, wready, wsettle, wget, wclose, mget, "agent prompt " + mainPane + " " + change + " is finished and archived. Please review it." + resumeTip}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &script{t: t, answers: tt.answers}
			var rec []herdr.Session
			Watch(watched(tt.load, &rec), h.run, noGit)
			if !reflect.DeepEqual(h.calls, tt.want) {
				t.Errorf("herdr calls\n got %q\nwant %q", h.calls, tt.want)
			}
			for _, s := range rec {
				if s != (herdr.Session{Agent: "claude", Kind: "id", Value: claudeID}) {
					t.Errorf("recorded %v", s)
				}
			}
		})
	}
}

func TestWatchRecordsWhenReadyAndAtEverySettle(t *testing.T) {
	h := &script{t: t, answers: map[string][]any{wget: {"agent-get-watched.json"}, mget: {"agent-get-main.json"}}}
	var rec []herdr.Session
	Watch(watched(loads([]work.Item{{Name: change, Done: 3, Total: 7}}, all7), &rec), h.run, noGit)
	if len(rec) != 4 {
		t.Errorf("want 4 records (ready, two settles, after the first check), got %v", rec)
	}
}

func TestWatchCheckout(t *testing.T) {
	dir := t.TempDir()
	tail := " Never use --force and keep the branch; if the merge or the removal does not go through cleanly, stop and tell the user." +
		" To look into its session press s on its Card in the Board, or run in " + dir + ": claude --resume " + claudeID
	open := func(into string) string {
		return change + " is finished (7/7 tasks) on branch " + change + " in " + dir + ". Please review it, merge branch " + change + " into " + into +
			", archive it with /opsx:archive " + change + ", then remove its worktree with: git worktree remove " + dir + "." + tail
	}
	fromWorktree := []work.Item{{Name: change, Done: 7, Total: 7, Dir: dir}}
	tests := []struct {
		name, base string
		load       []work.Item
		archive    bool
		want       string
	}{
		{"Progress from the checkout, detected base", "", fromWorktree, false, open("main")},
		{"configured base", "release/2.x", fromWorktree, false, open("release/2.x")},
		{"archived inside the checkout", "", []work.Item{{Name: change, Total: 7}}, true,
			change + " is finished and archived on branch " + change + " in " + dir + ". Please review it, merge branch " + change +
				" into main, then remove its worktree with: git worktree remove " + dir + "." + tail},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.archive {
				if err := os.MkdirAll(filepath.Join(dir, "openspec", "changes", "archive", "2026-10-30-"+change), 0o755); err != nil {
					t.Fatal(err)
				}
				defer os.RemoveAll(filepath.Join(dir, "openspec"))
			}
			h := &script{t: t, answers: map[string][]any{wget: {"agent-get-watched.json"}, mget: {"agent-get-main.json"}}}
			git := gitWith(nil, proj, "main", dir, change)
			var rec []herdr.Session
			w := watched(loads(tt.load), &rec)
			w.Base = tt.base
			Watch(w, h.run, git.run)
			want := []string{wget, wready, wsettle, wget, wclose, mget, "agent prompt " + mainPane + " " + tt.want}
			if !reflect.DeepEqual(h.calls, want) {
				t.Errorf("herdr calls\n got %q\nwant %q", h.calls, want)
			}
			for _, c := range append(h.calls, git.calls...) {
				for _, never := range []string{"worktree remove", "workspace close", "merge", "branch -d", "branch -D"} {
					if strings.HasPrefix(c, never) {
						t.Errorf("call %q must never be made", c)
					}
				}
			}
		})
	}
}

func TestWatchWorktreeOffIgnoresCheckout(t *testing.T) {
	h := &script{t: t, answers: map[string][]any{wget: {"agent-get-watched.json"}, mget: {"agent-get-main.json"}}}
	var rec []herdr.Session
	w := watched(loads(all7), &rec)
	w.On = false
	Watch(w, h.run, gitWith(nil, proj, "main", t.TempDir(), change).run)
	if want := append([]string{wget, wready, wsettle, wget}, cleanEnd...); !reflect.DeepEqual(h.calls, want) {
		t.Errorf("herdr calls\n got %q\nwant %q", h.calls, want)
	}
}

func TestWatchMainAgent(t *testing.T) {
	boom := errors.New("boom")
	text := plainText
	handoff := []string{"handoff"}
	tests := []struct {
		name    string
		main    string
		answers map[string][]any
		want    []string // calls after closing the pane
		marks   []string // nil: the Main agent took it, which clears the marker
	}{
		{"Handoff clears an older marker", mainPane, map[string][]any{mget: {"agent-get-main.json"}}, []string{mget, "agent prompt " + mainPane + " " + text}, nil},
		{"Main agent of another kind", mainPane, map[string][]any{mget: {"agent-get-main-codex.json"}}, []string{mget, "agent prompt " + mainPane + " " + text}, nil},
		{"Handoff without a Main agent", "", nil, nil, handoff},
		{"Main pane is the agent's pane", "w5:p1", nil, nil, handoff},
		{"agent get fails", mainPane, map[string][]any{mget: {boom}}, []string{mget}, handoff},
		{"Board opened from a shell pane", mainPane, map[string][]any{mget: {notFound}}, []string{mget}, handoff},
		{"Main agent blocked", mainPane, map[string][]any{mget: {"agent-get-main-blocked.json"}}, []string{mget}, handoff},
		{"prompt rejected", mainPane, map[string][]any{mget: {"agent-get-main.json"}, "agent prompt": {errors.New("agent_blocked")}},
			[]string{mget, "agent prompt " + mainPane + " " + text}, handoff},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			answers := map[string][]any{wget: {"agent-get-watched.json"}}
			for k, v := range tt.answers {
				answers[k] = v
			}
			h := &script{t: t, answers: answers}
			var rec []herdr.Session
			w := watched(loads(all7), &rec)
			w.Main = tt.main
			cleared := 0
			var marks []string
			w.Mark = func(why string) { marks = append(marks, why) }
			w.Clear = func() { cleared++ }
			Watch(w, h.run, noGit)
			want := append([]string{wget, wready, wsettle, wget, wclose}, tt.want...)
			if !reflect.DeepEqual(h.calls, want) {
				t.Errorf("herdr calls\n got %q\nwant %q", h.calls, want)
			}
			wantCleared := 0
			if tt.marks == nil {
				wantCleared = 1
			}
			if cleared != wantCleared || !reflect.DeepEqual(marks, tt.marks) {
				t.Errorf("marker cleared %d times and set %q, want %d and %q", cleared, marks, wantCleared, tt.marks)
			}
		})
	}
}

func TestWatchKinds(t *testing.T) {
	dir := t.TempDir()
	tests := []struct{ name, fixture, tip string }{
		{"codex: resume hint in the checkout", "agent-get-watched-codex.json", " or run in " + dir + ": codex resume abc"},
		{"pi: no resume hint", "agent-get-watched-pi.json", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &script{t: t, answers: map[string][]any{wget: {tt.fixture}, mget: {"agent-get-main.json"}}}
			var rec []herdr.Session
			Watch(watched(loads([]work.Item{{Name: change, Done: 7, Total: 7, Dir: dir}}), &rec), h.run, gitWith(nil, proj, "main", dir, change).run)
			last := h.calls[len(h.calls)-1]
			if !strings.HasPrefix(last, "agent prompt "+mainPane+" ") {
				t.Fatalf("want a prompt last, got %q", h.calls)
			}
			if tt.tip == "" && strings.Contains(last, "session") || tt.tip != "" && !strings.HasSuffix(last, tt.tip) {
				t.Errorf("prompt %q, want resume hint %q", last, tt.tip)
			}
			if tt.tip == "" && len(rec) != 0 {
				t.Errorf("an agent without agent_session records nothing, got %v", rec)
			}
		})
	}
}

func TestResumeArgs(t *testing.T) {
	id := func(agent string) herdr.Session { return herdr.Session{Agent: agent, Kind: "id", Value: "abc"} }
	tests := []struct {
		s    herdr.Session
		want string
	}{
		{id("claude"), "claude --resume abc"},
		{id("copilot"), "copilot --resume=abc"},
		{id("codex"), "codex resume abc"},
		{id("opencode"), "opencode --session abc"},
		{id("cursor"), "cursor-agent --resume abc"},
		{id("gemini"), "gemini --resume abc"},
		{herdr.Session{Agent: "claude", Kind: "path", Value: "/x/abc.jsonl"}, ""},
		{id("pi"), ""},
		{id("aider"), ""},
		{herdr.Session{Agent: "claude", Kind: "id"}, ""},
		{herdr.Session{}, ""},
	}
	for _, tt := range tests {
		if got := ResumeCommand(tt.s); got != tt.want {
			t.Errorf("ResumeCommand(%v) = %q, want %q", tt.s, got, tt.want)
		}
		var want []string
		if tt.want != "" {
			want = strings.Fields(tt.want)[1:]
		}
		if got := ResumeArgs(tt.s); !reflect.DeepEqual(got, want) {
			t.Errorf("ResumeArgs(%v) = %q, want %q", tt.s, got, want)
		}
	}
}

func TestResume(t *testing.T) {
	h := &fakeHerdr{t: t}
	if err := Resume(change, herdr.Session{Agent: "copilot", Kind: "id", Value: "abc"}, "w2:p1J", "w2:tG", h.run); err != nil {
		t.Fatal(err)
	}
	want := []string{tfocus, probe, waits, "agent start " + change + " --kind copilot --pane w2:p1J --timeout 60000 -- --resume=abc"}
	if !reflect.DeepEqual(h.calls, want) {
		t.Errorf("herdr calls\n got %q\nwant %q", h.calls, want)
	}
}

func TestResumeFails(t *testing.T) {
	h := &fakeHerdr{t: t}
	if err := Resume(change, herdr.Session{}, "w2:p1J", "w2:tG", h.run); err == nil || err.Error() != "no reopenable session for "+change || len(h.calls) != 0 {
		t.Errorf("no record: want error without herdr calls, got %v, %q", err, h.calls)
	}
	if err := Resume(change, herdr.Session{Agent: "pi", Kind: "id", Value: "abc"}, "w2:p1J", "w2:tG", h.run); err == nil || !strings.Contains(err.Error(), "resume not supported for pi") || len(h.calls) != 0 {
		t.Errorf("want error without herdr calls, got %v, %q", err, h.calls)
	}
	h = &fakeHerdr{t: t, out: map[string]string{"pane read": "pane-read.txt"}, fail: map[string]error{"agent start": errors.New("boom")}}
	if err := Resume(change, herdr.Session{Agent: "claude", Kind: "id", Value: "abc"}, "w2:p1J", "w2:tG", h.run); err == nil || !strings.Contains(err.Error(), "boom; pane shows: "+shown) {
		t.Errorf("want agent start error with the pane's lines, got %v", err)
	}
	h = &fakeHerdr{t: t, out: map[string]string{"pane read": "pane-read.txt"}, fail: map[string]error{"pane wait-output": errors.New("timeout")}}
	if err := Resume(change, herdr.Session{Agent: "claude", Kind: "id", Value: "abc"}, "w2:p1J", "w2:tG", h.run); err == nil || !strings.Contains(err.Error(), `shell did not reach its prompt in tab "m1-board-view": timeout; pane shows: `) || slices.ContainsFunc(h.calls, func(c string) bool { return strings.HasPrefix(c, "agent start") }) {
		t.Errorf("want no agent start without a prompt, got %v, %q", err, h.calls)
	}
}

// noStep is a Source without a finishing step.
type noStep struct{ work.Source }

func (noStep) Finish(string, string) string { return "" }

func TestWatchWithoutFinishingStep(t *testing.T) {
	dir := t.TempDir()
	all9 := []work.Item{{Name: change, Done: 9, Total: 9}}
	tests := []struct {
		name string
		on   bool
		want string
	}{
		{"no checkout", false, change + " is finished (9/9 tasks). Please review it." + resumeTip},
		{"in a worktree", true, change + " is finished (9/9 tasks) on branch " + change + " in " + dir + ". Please review it, merge branch " + change +
			" into main, then remove its worktree with: git worktree remove " + dir + "." +
			" Never use --force and keep the branch; if the merge or the removal does not go through cleanly, stop and tell the user." +
			" To look into its session press s on its Card in the Board, or run in " + dir + ": claude --resume " + claudeID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &script{t: t, answers: map[string][]any{wget: {"agent-get-watched.json"}, mget: {"agent-get-main.json"}}}
			var rec []herdr.Session
			w := watched(loads(all9), &rec)
			w.On, w.Source = tt.on, noStep{src}
			Watch(w, h.run, gitWith(nil, proj, "main", dir, change).run)
			if last := h.calls[len(h.calls)-1]; last != "agent prompt "+mainPane+" "+tt.want {
				t.Errorf("prompt\n got %q\nwant %q", last, "agent prompt "+mainPane+" "+tt.want)
			}
		})
	}
}

// TestWatchSpecKit checks the Handoff of a finished Spec Kit Feature: no archive step, and in a
// worktree the request to delete the feature.json Start wrote.
func TestWatchSpecKit(t *testing.T) {
	dir := t.TempDir()
	all9 := []work.Item{{Name: change, Done: 9, Total: 9}}
	tests := []struct {
		name string
		on   bool
		want string
	}{
		{"Finish without a worktree", false, change + " is finished (9/9 tasks). Please review it." + resumeTip},
		{"Finish in a worktree", true, change + " is finished (9/9 tasks) on branch " + change + " in " + dir + ". Please review it, merge branch " + change +
			" into main, delete " + dir + "/.specify/feature.json unless git tracks it, then remove its worktree with: git worktree remove " + dir + "." +
			" Never use --force and keep the branch; if the merge or the removal does not go through cleanly, stop and tell the user." +
			" To look into its session press s on its Card in the Board, or run in " + dir + ": claude --resume " + claudeID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &script{t: t, answers: map[string][]any{wget: {"agent-get-watched.json"}, mget: {"agent-get-main.json"}}}
			var rec []herdr.Session
			w := watched(loads(all9), &rec)
			w.On, w.Source = tt.on, speckit.New(nil, "")
			Watch(w, h.run, gitWith(nil, proj, "main", dir, change).run)
			if last := h.calls[len(h.calls)-1]; last != "agent prompt "+mainPane+" "+tt.want {
				t.Errorf("prompt\n got %q\nwant %q", last, "agent prompt "+mainPane+" "+tt.want)
			}
		})
	}
}

func TestWatchAndResumeFeatureName(t *testing.T) {
	const feature, agentID = "001-a", "f-001-a"
	ready, working := "agent wait "+agentID+" --until working --timeout 60000", "agent wait "+agentID+" --until working"
	h := &script{t: t, answers: map[string][]any{"agent get " + agentID: {"agent-get-watched.json"}, ready: {}, working: {notFound}, working + " --timeout 600000": {notFound}}}
	w := Watched{Project: proj, Change: feature, Load: loads([]work.Item{{Name: feature, Done: 3, Total: 7}}), Source: src, Record: func(herdr.Session) {},
		Escalated: func() bool { return false }, Mark: func(string) {}, Clear: func() {}}
	Watch(w, h.run, noGit)
	get := "agent get " + agentID
	if want := []string{get, ready, "agent wait " + agentID, get, get, working + " --timeout 600000", get, working}; !reflect.DeepEqual(h.calls, want) {
		t.Errorf("Watch herdr calls %q, want %q", h.calls, want)
	}
	f := &fakeHerdr{t: t}
	if err := Resume(feature, herdr.Session{Agent: "claude", Kind: "id", Value: "abc"}, "w2:p1J", "w2:tG", f.run); err != nil {
		t.Fatal(err)
	}
	if want := "agent start " + agentID + " --kind claude --pane w2:p1J --timeout 60000 -- --resume abc"; f.calls[3] != want || f.calls[1] != probe {
		t.Errorf("Resume calls %q, want the probe, then %q", f.calls, want)
	}
}

func TestWatchStall(t *testing.T) {
	left := loads([]work.Item{{Name: change, Done: 3, Total: 7}})
	stall := "agent prompt " + mainPane + " " + stallText("3/7 tasks")
	again := []any{"agent-get-watched.json", notFound} // works once more, then goes away
	tests := []struct {
		name      string
		main      string
		escalated bool
		answers   map[string][]any
		load      func() ([]work.Item, error)
		want      []string
		cleared   int
		marks     []string
	}{
		{"Worker stops silently", mainPane, false, map[string][]any{wready: {}, wstall: {notFound}, wworking: {notFound}},
			left, []string{wget, wready, wsettle, wget, wget, wstall, wget, mget, stall, wworking}, 0, []string{"stall"}},
		{"Worker goes on by itself", mainPane, false, map[string][]any{wready: {}, wstall: {}, wsettle: again},
			left, []string{wget, wready, wsettle, wget, wget, wstall, wsettle}, 0, nil},
		{"User answers in time", mainPane, false, map[string][]any{wready: {}, wstall: {}, wsettle: again},
			left, []string{wget, wready, wsettle, wget, wget, wstall, wsettle}, 0, nil},
		{"Second silent stop", mainPane, false, map[string][]any{wready: {}, wstall: {notFound}, wworking: again},
			left, []string{wget, wready, wsettle, wget, wget, wstall, wget, mget, stall, wworking, wsettle, wget, wget, wstall, wget, mget, stall, wworking}, 1, []string{"stall", "stall"}},
		{"agent gone after the timed wait", mainPane, false, map[string][]any{wget: {"agent-get-watched.json", "agent-get-watched.json", "agent-get-watched.json", notFound}, wstall: {notFound}},
			left, []string{wget, wready, wsettle, wget, wget, wstall, wget}, 0, nil},
		{"Progress unknown", mainPane, false, map[string][]any{wready: {}, wstall: {notFound}, wworking: {notFound}},
			loads(errors.New("boom")), []string{wget, wready, wsettle, wget, wget, wstall, wget, mget, "agent prompt " + mainPane + " " + stallText(""), wworking}, 0, []string{"stall"}},
		{"Stall without a Main agent", "", false, map[string][]any{wready: {}, wstall: {notFound}, wworking: {notFound}},
			left, []string{wget, wready, wsettle, wget, wget, wstall, wget, wworking}, 0, []string{"stall"}},
		{"timed wait fails while it moved on: no Stall", mainPane, false, map[string][]any{wget: {"agent-get-watched.json", "agent-get-watched.json", "agent-get-watched.json", "agent-get-watched-later.json", notFound},
			wready: {}, wstall: {notFound}, wsettle: {"agent-get-watched.json"}},
			left, []string{wget, wready, wsettle, wget, wget, wstall, wget, wsettle, wget}, 0, nil},
		{"blocked settle without escalation clears nothing", mainPane, false, map[string][]any{wget: {"agent-get-watched-blocked.json"}, wready: {}, wworking: again},
			left, []string{wget, wready, wsettle, wget, wget, wworking, wsettle, wget, wget, wworking}, 0, nil},
		{"Escalation is followed by no Stall", mainPane, true, map[string][]any{wready: {}, wworking: {}, wsettle: again},
			left, []string{wget, wready, wsettle, wget, wget, wworking, wsettle}, 1, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			answers := map[string][]any{wget: {"agent-get-watched.json"}, mget: {"agent-get-main.json"}}
			for k, v := range tt.answers {
				answers[k] = v
			}
			h := &script{t: t, answers: answers}
			var rec []herdr.Session
			w := watched(tt.load, &rec)
			w.Main = tt.main
			cleared := 0
			var marks []string
			w.Escalated = func() bool { return tt.escalated }
			w.Mark = func(why string) { marks = append(marks, why) }
			w.Clear = func() { cleared++ }
			Watch(w, h.run, noGit)
			if !reflect.DeepEqual(h.calls, tt.want) {
				t.Errorf("herdr calls\n got %q\nwant %q", h.calls, tt.want)
			}
			if cleared != tt.cleared || !reflect.DeepEqual(marks, tt.marks) {
				t.Errorf("marker cleared %d times and set %q, want %d and %q", cleared, marks, tt.cleared, tt.marks)
			}
		})
	}
}

func TestWatchResumed(t *testing.T) {
	tests := []struct {
		name    string
		answers map[string][]any
		want    []string
	}{
		{"Resumed session of a finished Change", map[string][]any{wworking: {notFound}}, []string{wget, wworking}},
		{"Resumed session finishes the Change", map[string][]any{wworking: {}}, append([]string{wget, wworking, wsettle, wget}, cleanEnd...)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			answers := map[string][]any{wget: {"agent-get-watched.json"}, mget: {"agent-get-main.json"}}
			for k, v := range tt.answers {
				answers[k] = v
			}
			h := &script{t: t, answers: answers}
			var rec []herdr.Session
			w := watched(loads(all7), &rec)
			w.Resumed = true
			Watch(w, h.run, noGit)
			if !reflect.DeepEqual(h.calls, tt.want) {
				t.Errorf("herdr calls\n got %q\nwant %q", h.calls, tt.want)
			}
		})
	}
}

// TestWatchClearsOnlyTheSettlesEscalation: an escalation written after a blocked settle, when the
// user approved the escalate command, survives that settle and spares the next one its Stall.
func TestWatchClearsOnlyTheSettlesEscalation(t *testing.T) {
	h := &script{t: t, answers: map[string][]any{wget: {"agent-get-watched-blocked.json", "agent-get-watched-blocked.json", "agent-get-watched-blocked.json",
		"agent-get-watched.json", "agent-get-watched.json"}, wready: {}, wworking: {"agent-get-watched.json", notFound}}}
	var rec []herdr.Session
	w := watched(loads([]work.Item{{Name: change, Done: 3, Total: 7}}), &rec)
	reads, cleared := 0, 0
	w.Escalated = func() bool { reads++; return reads > 1 } // written once the user approved it
	w.Clear = func() { cleared++ }
	Watch(w, h.run, noGit)
	if want := []string{wget, wready, wsettle, wget, wget, wworking, wsettle, wget, wget, wworking}; !reflect.DeepEqual(h.calls, want) {
		t.Errorf("herdr calls\n got %q\nwant %q", h.calls, want)
	}
	if cleared != 0 {
		t.Errorf("cleared %d times: the record of the next settle must survive until it works again", cleared)
	}
}
