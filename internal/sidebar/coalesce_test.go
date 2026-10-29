package sidebar

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
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

func TestCoalesceUnusableStateDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(dir, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	boom := errors.New("boom")
	passes := 0
	if err := Coalesce(dir, func() error { passes++; return boom }); !errors.Is(err, boom) || passes != 1 {
		t.Errorf("a broken state dir must not stop the pass: %d passes, %v", passes, err)
	}
}

// TestCoalesceConcurrent races callers against each other: passes never overlap, never outnumber
// the callers, and the last request is always served by a pass that starts after it.
func TestCoalesceConcurrent(t *testing.T) {
	for round := range 200 {
		dir := t.TempDir()
		const callers = 50
		var inFlight, runs atomic.Int32
		var mu sync.Mutex
		var lastStart, lastCall time.Time
		pass := func() error {
			if inFlight.Add(1) > 1 {
				t.Error("passes overlap")
			}
			runs.Add(1)
			mu.Lock()
			lastStart = time.Now()
			mu.Unlock()
			time.Sleep(50 * time.Microsecond)
			inFlight.Add(-1)
			return nil
		}
		var wg sync.WaitGroup
		for i := range callers {
			wg.Add(1)
			go func() {
				defer wg.Done()
				time.Sleep(time.Duration(i%10) * 20 * time.Microsecond)
				mu.Lock()
				if now := time.Now(); now.After(lastCall) {
					lastCall = now
				}
				mu.Unlock()
				if err := Coalesce(dir, pass); err != nil {
					t.Error(err)
				}
			}()
		}
		wg.Wait()
		if n := runs.Load(); n < 1 || n > callers {
			t.Fatalf("round %d: %d passes for %d callers", round, n, callers)
		}
		if !lastStart.After(lastCall) {
			t.Fatalf("round %d: the last request (%v) was not served by a later pass (last start %v)", round, lastCall, lastStart)
		}
	}
}
