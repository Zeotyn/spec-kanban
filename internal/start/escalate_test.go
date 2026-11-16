package start

import (
	"errors"
	"os"
	"reflect"
	"testing"
	"time"
)

// mread is the herdr call that reads the Main pane's screen.
const mread = "pane read " + mainPane + " --source visible --format ansi"

// fakeTime makes sleep advance now instead of waiting, until the test ends.
func fakeTime(t *testing.T) {
	at := time.Date(2026, 11, 16, 9, 0, 0, 0, time.UTC)
	oldSleep, oldNow := sleep, now
	sleep = func(d time.Duration) { at = at.Add(d) }
	now = func() time.Time { return at }
	t.Cleanup(func() { sleep, now = oldSleep, oldNow })
}

func TestClear(t *testing.T) {
	tests := []struct {
		name, kind string
		answers    map[string][]any
		want       []string
		clear      bool
	}{
		{"claude, empty box", "claude", map[string][]any{mread: {"pane-read-claude-empty.txt"}, mget: {"agent-get-main-focused.json"}}, []string{mread}, true},
		{"claude, placeholder", "claude", map[string][]any{mread: {"pane-read-claude-placeholder.txt"}}, []string{mread}, true},
		{"claude, text in the box", "claude", map[string][]any{mread: {"pane-read-claude-one-line.txt"}, mget: {"agent-get-main.json"}}, []string{mread}, false},
		{"claude, no box, focused", "claude", map[string][]any{mread: {"pane-read-claude-dialog.txt"}, mget: {"agent-get-main-focused.json"}}, []string{mread, mget}, false},
		{"claude, no box, not focused", "claude", map[string][]any{mread: {"pane-read-claude-dialog.txt"}, mget: {"agent-get-main.json"}}, []string{mread, mget}, true},
		{"claude, read error, focused", "claude", map[string][]any{mread: {notFound}, mget: {"agent-get-main-focused.json"}}, []string{mread, mget}, false},
		{"codex, focused", "codex", map[string][]any{mget: {"agent-get-main-codex-focused.json"}}, []string{mget}, false},
		{"codex, not focused", "codex", map[string][]any{mget: {"agent-get-main-codex.json"}}, []string{mget}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &script{t: t, answers: tt.answers}
			if got := clear(h.run, mainPane, tt.kind); got != tt.clear {
				t.Errorf("clear = %v, want %v", got, tt.clear)
			}
			if !reflect.DeepEqual(h.calls, tt.want) {
				t.Errorf("herdr calls\n got %q\nwant %q", h.calls, tt.want)
			}
		})
	}
}

func TestTell(t *testing.T) {
	const text = "help"
	prompt := "agent prompt " + mainPane + " " + text
	tests := []struct {
		name, main string
		answers    map[string][]any
		want       []string
		took       bool
	}{
		{"Main agent idle", mainPane, map[string][]any{mget: {"agent-get-main.json"}, mread: {"pane-read-claude-empty.txt"}}, []string{mget, mread, prompt}, true},
		{"box empties after two polls", mainPane, map[string][]any{mget: {"agent-get-main.json"}, mread: {"pane-read-claude-one-line.txt", "pane-read-claude-lines.txt", "pane-read-claude-placeholder.txt"}},
			[]string{mget, mread, mread, mread, prompt}, true},
		{"focused codex agent leaves the pane", mainPane, map[string][]any{mget: {"agent-get-main-codex-focused.json", "agent-get-main-codex-focused.json", "agent-get-main-codex.json"}},
			[]string{mget, mget, mget, prompt}, true},
		{"no Main pane", "", nil, nil, false},
		{"Main pane is the Change's own pane", "w5:p1", nil, nil, false},
		{"no agent in the Main pane", mainPane, map[string][]any{mget: {notFound}}, []string{mget}, false},
		{"Main agent blocked", mainPane, map[string][]any{mget: {"agent-get-main-blocked.json"}}, []string{mget}, false},
		{"prompt rejected", mainPane, map[string][]any{mget: {"agent-get-main.json"}, mread: {"pane-read-claude-empty.txt"}, "agent prompt": {errors.New("agent_blocked")}}, []string{mget, mread, prompt}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeTime(t)
			h := &script{t: t, answers: tt.answers}
			if took := Tell(h.run, tt.main, "w5:p1", text); took != tt.took {
				t.Errorf("Tell = %v, want %v", took, tt.took)
			}
			if !reflect.DeepEqual(h.calls, tt.want) {
				t.Errorf("herdr calls\n got %q\nwant %q", h.calls, tt.want)
			}
		})
	}
}

func TestTellGivesUpAfter10Minutes(t *testing.T) {
	fakeTime(t)
	start, reads := now(), 0
	run := func(args ...string) ([]byte, error) {
		switch args[0] + " " + args[1] {
		case "agent get":
			return os.ReadFile("testdata/herdr/agent-get-main.json")
		case "pane read":
			reads++
			return os.ReadFile("testdata/herdr/pane-read-claude-one-line.txt")
		}
		t.Fatalf("unexpected herdr call %q", args)
		return nil, nil
	}
	if Tell(run, mainPane, "w5:p1", "help") {
		t.Error("Tell = true, want false")
	}
	if waited := now().Sub(start); waited != 10*time.Minute {
		t.Errorf("waited %v, want 10m", waited)
	}
	if reads != 301 { // every 2 seconds, both ends included
		t.Errorf("%d reads, want 301", reads)
	}
}

func TestTellNow(t *testing.T) {
	const text = "help"
	prompt := "agent prompt " + mainPane + " " + text
	tests := []struct {
		name, main     string
		answers        map[string][]any
		want           []string
		taken, waiting bool
	}{
		{"empty box", mainPane, map[string][]any{mget: {"agent-get-main.json"}, mread: {"pane-read-claude-empty.txt"}}, []string{mget, mread, prompt}, true, false},
		{"text in the box", mainPane, map[string][]any{mget: {"agent-get-main.json"}, mread: {"pane-read-claude-one-line.txt"}}, []string{mget, mread}, false, true},
		{"no Main pane", "", nil, nil, false, false},
		{"no agent in the Main pane", mainPane, map[string][]any{mget: {notFound}}, []string{mget}, false, false},
		{"Main agent blocked", mainPane, map[string][]any{mget: {"agent-get-main-blocked.json"}}, []string{mget}, false, false},
		{"prompt rejected", mainPane, map[string][]any{mget: {"agent-get-main.json"}, mread: {"pane-read-claude-empty.txt"}, "agent prompt": {errors.New("agent_blocked")}}, []string{mget, mread, prompt}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &script{t: t, answers: tt.answers}
			taken, waiting := TellNow(h.run, tt.main, "w5:p1", text)
			if taken != tt.taken || waiting != tt.waiting {
				t.Errorf("TellNow = %v, %v, want %v, %v", taken, waiting, tt.taken, tt.waiting)
			}
			if !reflect.DeepEqual(h.calls, tt.want) {
				t.Errorf("herdr calls\n got %q\nwant %q", h.calls, tt.want)
			}
		})
	}
}

func TestSentence(t *testing.T) {
	tail := " escalate " + change + ` "<what you need>" and end your turn.`
	tests := []struct{ self, want string }{
		{"/opt/herdr/plugins/spec-kanban/kanban", "/opt/herdr/plugins/spec-kanban/kanban"},
		{"/Users/me/Library/Application Support/herdr/kanban", "'/Users/me/Library/Application Support/herdr/kanban'"},
		{"/Users/o'neil/kanban", `'/Users/o'\''neil/kanban'`},
	}
	for _, tt := range tests {
		if got, want := Sentence(tt.self, change), "If you cannot continue without a human, run "+tt.want+tail; got != want {
			t.Errorf("Sentence(%q)\n got %q\nwant %q", tt.self, got, want)
		}
	}
}
