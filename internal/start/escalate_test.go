package start

import (
	"errors"
	"reflect"
	"testing"
)

func TestTell(t *testing.T) {
	const text = "help"
	prompt := "agent prompt " + mainPane + " " + text
	tests := []struct {
		name, main string
		answers    map[string][]any
		want       []string
		took       bool
	}{
		{"Main agent idle", mainPane, map[string][]any{mget: {"agent-get-main.json"}}, []string{mget, prompt}, true},
		{"no Main pane", "", nil, nil, false},
		{"Main pane is the Change's own pane", "w5:p1", nil, nil, false},
		{"no agent in the Main pane", mainPane, map[string][]any{mget: {notFound}}, []string{mget}, false},
		{"Main agent blocked", mainPane, map[string][]any{mget: {"agent-get-main-blocked.json"}}, []string{mget}, false},
		{"prompt rejected", mainPane, map[string][]any{mget: {"agent-get-main.json"}, "agent prompt": {errors.New("agent_blocked")}}, []string{mget, prompt}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
