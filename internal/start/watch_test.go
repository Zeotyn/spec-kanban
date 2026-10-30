package start

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/zeotyn/openspec-kanban/internal/herdr"
	"github.com/zeotyn/openspec-kanban/internal/openspec"
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
func loads(results ...any) func() ([]openspec.Change, error) {
	i := 0
	return func() ([]openspec.Change, error) {
		r := results[min(i, len(results)-1)]
		i++
		if err, ok := r.(error); ok {
			return nil, err
		}
		return r.([]openspec.Change), nil
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
	wclose    = "pane close w5:p1"
	mget      = "agent get " + mainPane
	all7      = []openspec.Change{{Name: "other"}, {Name: change, Done: 7, Total: 7}}
	resumeTip = " To look into its session press s on its Card in the Board, or run in " + proj + ": claude --resume " + claudeID
	plainText = change + " is finished (7/7 tasks). Please review it and archive it with /opsx:archive " + change + "." + resumeTip
	cleanEnd  = []string{wclose, mget, "agent prompt " + mainPane + " " + plainText}
)

func watched(load func() ([]openspec.Change, error), rec *[]herdr.Session) Watched {
	return Watched{Project: proj, Change: change, Main: mainPane, On: true, Load: load,
		Record: func(s herdr.Session) { *rec = append(*rec, s) }}
}

func notify(title, body string) string {
	return "notification show openspec-kanban: " + change + " " + title + " --body " + body
}

func TestWatch(t *testing.T) {
	boom := errors.New("boom")
	tests := []struct {
		name    string
		answers map[string][]any
		load    func() ([]openspec.Change, error)
		want    []string
	}{
		{"clean idle settle: handoff", map[string][]any{wget: {"agent-get-watched.json"}, mget: {"agent-get-main.json"}},
			loads(all7), append([]string{wget, wready, wsettle, wget}, cleanEnd...)},
		{"done counts as clean", map[string][]any{wget: {"agent-get-watched.json", "agent-get-watched-done.json"}, mget: {"agent-get-main.json"}},
			loads(all7), append([]string{wget, wready, wsettle, wget}, cleanEnd...)},
		{"wait for working fails: still evaluated", map[string][]any{wget: {"agent-get-watched.json"}, wready: {notFound}, mget: {"agent-get-main.json"}},
			loads(all7), append([]string{wget, wready, wsettle, wget}, cleanEnd...)},
		{"tasks left: notification, then handoff at the next settle", map[string][]any{wget: {"agent-get-watched.json"}, mget: {"agent-get-main.json"}},
			loads([]openspec.Change{{Name: change, Done: 3, Total: 7}}, all7),
			append([]string{wget, wready, wsettle, wget, notify("stopped", change+" stopped at 3/7 tasks"), wget, wworking, wsettle, wget}, cleanEnd...)},
		{"blocked with all tasks done: waits for input", map[string][]any{wget: {"agent-get-watched.json", "agent-get-watched-blocked.json"}, wready: {}, wworking: {notFound}},
			loads(all7), []string{wget, wready, wsettle, wget, notify("stopped", change+" waits for input at 7/7 tasks"), wget, wworking}},
		{"load error: stopped", map[string][]any{wget: {"agent-get-watched.json"}, wready: {}, wworking: {notFound}},
			loads(boom), []string{wget, wready, wsettle, wget, notify("stopped", "boom"), wget, wworking}},
		{"Change neither open nor archived: not clean", map[string][]any{wget: {"agent-get-watched.json"}, wready: {}, wworking: {notFound}},
			loads([]openspec.Change{{Name: "other"}}), []string{wget, wready, wsettle, wget, notify("stopped", change+" is neither open nor archived"), wget, wworking}},
		{"no tasks: not clean", map[string][]any{wget: {"agent-get-watched.json"}, wready: {}, wworking: {notFound}},
			loads([]openspec.Change{{Name: change}}), []string{wget, wready, wsettle, wget, notify("stopped", change+" stopped at 0/0 tasks"), wget, wworking}},
		{"blocked with a load error: waits for input", map[string][]any{wget: {"agent-get-watched.json", "agent-get-watched-blocked.json"}, wready: {}, wworking: {notFound}},
			loads(boom), []string{wget, wready, wsettle, wget, notify("stopped", change+" waits for input"), wget, wworking}},
		{"turn ended while evaluating: evaluated again at once", map[string][]any{wget: {"agent-get-watched.json", "agent-get-watched-blocked.json", "agent-get-watched-later.json"}, mget: {"agent-get-main.json"}},
			loads(all7), append([]string{wget, wready, wsettle, wget, notify("stopped", change+" waits for input at 7/7 tasks"), wget, wsettle, wget}, cleanEnd...)},
		{"working again when read: no handoff, no notification", map[string][]any{wget: {"agent-get-watched.json", "agent-get-watched-working.json", "agent-get-watched.json"}, mget: {"agent-get-main.json"}},
			loads(all7), append([]string{wget, wready, wsettle, wget, wsettle, wget}, cleanEnd...)},
		{"agent gone while settling", map[string][]any{wget: {"agent-get-watched.json"}, wsettle: {notFound}},
			loads(all7), []string{wget, wready, wsettle}},
		{"agent gone after settling", map[string][]any{wget: {"agent-get-watched.json", notFound}},
			loads(all7), []string{wget, wready, wsettle, wget}},
		{"agent gone at once", map[string][]any{wget: {notFound}}, loads(all7), []string{wget}},
		{"archived in the Project", map[string][]any{wget: {"agent-get-watched.json"}, mget: {"agent-get-main.json"}},
			loads([]openspec.Change{{Name: change, Archived: true, ArchivedOn: "2026-10-30"}}),
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
	Watch(watched(loads([]openspec.Change{{Name: change, Done: 3, Total: 7}}, all7), &rec), h.run, noGit)
	if len(rec) != 4 {
		t.Errorf("want 4 records (ready, two settles, after the notification), got %v", rec)
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
	fromWorktree := []openspec.Change{{Name: change, Done: 7, Total: 7, Dir: dir}}
	tests := []struct {
		name, base string
		load       []openspec.Change
		archive    bool
		want       string
	}{
		{"Progress from the checkout, detected base", "", fromWorktree, false, open("main")},
		{"configured base", "release/2.x", fromWorktree, false, open("release/2.x")},
		{"archived inside the checkout", "", []openspec.Change{{Name: change, Total: 7}}, true,
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
	finished := notify("finished", text)
	tests := []struct {
		name    string
		main    string
		answers map[string][]any
		want    []string // calls after closing the pane
	}{
		{"Main agent of another kind", mainPane, map[string][]any{mget: {"agent-get-main-codex.json"}}, []string{mget, "agent prompt " + mainPane + " " + text}},
		{"no Main pane", "", nil, []string{finished}},
		{"Main pane is the agent's pane", "w5:p1", nil, []string{finished}},
		{"agent get fails", mainPane, map[string][]any{mget: {boom}}, []string{mget, finished}},
		{"no agent in the Main pane", mainPane, map[string][]any{mget: {notFound}}, []string{mget, finished}},
		{"Main agent blocked", mainPane, map[string][]any{mget: {"agent-get-main-blocked.json"}}, []string{mget, finished}},
		{"prompt rejected", mainPane, map[string][]any{mget: {"agent-get-main.json"}, "agent prompt": {errors.New("agent_blocked")}},
			[]string{mget, "agent prompt " + mainPane + " " + text, finished}},
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
			Watch(w, h.run, noGit)
			want := append([]string{wget, wready, wsettle, wget, wclose}, tt.want...)
			if !reflect.DeepEqual(h.calls, want) {
				t.Errorf("herdr calls\n got %q\nwant %q", h.calls, want)
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
			Watch(watched(loads([]openspec.Change{{Name: change, Done: 7, Total: 7, Dir: dir}}), &rec), h.run, gitWith(nil, proj, "main", dir, change).run)
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
	want := []string{tfocus, "agent start " + change + " --kind copilot --pane w2:p1J --timeout 60000 -- --resume=abc"}
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
	h = &fakeHerdr{t: t, fail: map[string]error{"agent start": errors.New("boom")}}
	if err := Resume(change, herdr.Session{Agent: "claude", Kind: "id", Value: "abc"}, "w2:p1J", "w2:tG", h.run); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Errorf("want agent start error, got %v", err)
	}
}
