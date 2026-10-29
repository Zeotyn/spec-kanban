package sidebar

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
)

// Coalesce runs pass under <dir>/sync.lock. When another process holds the lock, it touches
// <dir>/sync.pending and returns at once; the holder then runs one more pass, however many requests
// arrived. It returns the error of the last pass it ran.
func Coalesce(dir string, pass func() error) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(dir, "sync.lock"), os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close() // releases the lock
	pending := filepath.Join(dir, "sync.pending")
	requested := func() bool { _, err := os.Stat(pending); return err == nil }
	touched := false
	var passErr error
	for {
		if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); errors.Is(err, syscall.EWOULDBLOCK) {
			if touched {
				return passErr
			}
			if err := os.WriteFile(pending, nil, 0o644); err != nil {
				return err
			}
			// Try once more: the holder may have looked for pending just before we touched it and
			// released the lock since. If it still holds the lock, it will see pending after unlocking.
			touched = true
			continue
		} else if err != nil {
			return err
		}
		for {
			os.Remove(pending)
			passErr = pass()
			if !requested() {
				break
			}
		}
		syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		if !requested() { // a request that came in between our last look and the unlock
			return passErr
		}
		touched = false
	}
}
