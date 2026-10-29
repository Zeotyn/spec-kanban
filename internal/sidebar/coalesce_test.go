package sidebar

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestCoalesce(t *testing.T) {
	never := func() error { t.Error("a request during a pass must not run its own pass"); return nil }
	tests := []struct {
		name   string
		during func(dir string) // runs inside the first pass
		passes int
	}{
		{"free lock runs one pass", func(string) {}, 1},
		{"pending during a pass runs one more", func(dir string) {
			if err := os.WriteFile(filepath.Join(dir, "sync.pending"), nil, 0o644); err != nil {
				t.Fatal(err)
			}
		}, 2},
		{"several requests during a pass run one more", func(dir string) {
			for range 5 {
				if err := Coalesce(dir, never); err != nil {
					t.Error(err)
				}
			}
		}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "state") // created on demand
			passes := 0
			err := Coalesce(dir, func() error {
				passes++
				if passes == 1 {
					tt.during(dir)
				}
				return nil
			})
			if err != nil || passes != tt.passes {
				t.Errorf("got %d passes, %v; want %d", passes, err, tt.passes)
			}
			if _, err := os.Stat(filepath.Join(dir, "sync.pending")); err == nil {
				t.Error("pending must be consumed")
			}
		})
	}
}

func TestCoalesceLockHeld(t *testing.T) {
	dir := t.TempDir()
	f, err := os.OpenFile(filepath.Join(dir, "sync.lock"), os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Fatal(err)
	}
	if err := Coalesce(dir, func() error { t.Error("no pass while the lock is held"); return nil }); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "sync.pending")); err != nil {
		t.Errorf("want sync.pending touched: %v", err)
	}
}

func TestCoalescePassError(t *testing.T) {
	boom := errors.New("boom")
	if err := Coalesce(t.TempDir(), func() error { return boom }); !errors.Is(err, boom) {
		t.Errorf("want the pass error, got %v", err)
	}
}
