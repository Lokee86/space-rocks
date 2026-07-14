package servicelog

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func flushTestDependencies(clock *fakeClock, sink *trackedWriteCloser, ticker runtimeTicker) runtimeDependencies {
	deps := testRuntimeDependencies(clock)
	deps.openFile = func(string, int, os.FileMode) (io.WriteCloser, error) {
		return sink, nil
	}
	if ticker != nil {
		deps.newTicker = func(time.Duration) runtimeTicker {
			return ticker
		}
	}
	return deps
}

func waitForCondition(t *testing.T, label string, condition func() bool) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for %s", label)
}

func TestRuntimeFlushesPeriodically(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "logs")
	clock := &fakeClock{current: time.Date(2026, 7, 14, 13, 30, 0, 0, time.UTC)}
	sink := &trackedWriteCloser{}
	ticker := &fakeTicker{ch: make(chan time.Time, 1)}

	runtime, err := openWithDependencies(Config{
		Identity:       ServiceIdentity{Name: "game-server"},
		File:           validFileConfig(directory),
		Flush:          FlushPolicy{Interval: time.Second},
		FileEnabled:    true,
		ConsoleEnabled: false,
	}, flushTestDependencies(clock, sink, ticker))
	if err != nil {
		t.Fatalf("openWithDependencies() error = %v", err)
	}

	ticker.ch <- clock.now()
	waitForCondition(t, "periodic sync", func() bool { return sink.syncCount == 1 })

	if err := runtime.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if sink.syncCount != 2 {
		t.Fatalf("sync count after Close = %d, want 2", sink.syncCount)
	}
	if sink.closeCount != 1 {
		t.Fatalf("close count after Close = %d, want 1", sink.closeCount)
	}
	if ticker.stopCount != 1 {
		t.Fatalf("ticker stop count = %d, want 1", ticker.stopCount)
	}

	ticker.ch <- clock.now()
	time.Sleep(25 * time.Millisecond)
	if sink.syncCount != 2 {
		t.Fatalf("sync count after shutdown tick = %d, want 2", sink.syncCount)
	}
	if err := runtime.Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
	if sink.closeCount != 1 {
		t.Fatalf("close count after second Close = %d, want 1", sink.closeCount)
	}
}

func TestRuntimeCloseFlushesAndStopsOnce(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "logs")
	clock := &fakeClock{current: time.Date(2026, 7, 14, 13, 30, 0, 0, time.UTC)}
	sink := &trackedWriteCloser{}
	ticker := &fakeTicker{ch: make(chan time.Time, 1)}

	runtime, err := openWithDependencies(Config{
		Identity:       ServiceIdentity{Name: "game-server"},
		File:           validFileConfig(directory),
		Flush:          FlushPolicy{Interval: time.Second},
		FileEnabled:    true,
		ConsoleEnabled: false,
	}, flushTestDependencies(clock, sink, ticker))
	if err != nil {
		t.Fatalf("openWithDependencies() error = %v", err)
	}

	if err := runtime.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if sink.syncCount != 1 {
		t.Fatalf("sync count after Close = %d, want 1", sink.syncCount)
	}
	if sink.closeCount != 1 {
		t.Fatalf("close count after Close = %d, want 1", sink.closeCount)
	}
	if ticker.stopCount != 1 {
		t.Fatalf("ticker stop count = %d, want 1", ticker.stopCount)
	}

	ticker.ch <- clock.now()
	time.Sleep(25 * time.Millisecond)
	if sink.syncCount != 1 {
		t.Fatalf("sync count after shutdown tick = %d, want 1", sink.syncCount)
	}
	if err := runtime.Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
	if sink.syncCount != 1 || sink.closeCount != 1 || ticker.stopCount != 1 {
		t.Fatalf("idempotent close changed counts: sync=%d close=%d stop=%d", sink.syncCount, sink.closeCount, ticker.stopCount)
	}
}

func TestRuntimeFlushFailureIsRecordedOnce(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "logs")
	clock := &fakeClock{current: time.Date(2026, 7, 14, 13, 30, 0, 0, time.UTC)}
	flushErr := errors.New("flush failed")
	sink := &trackedWriteCloser{syncErr: flushErr}
	ticker := &fakeTicker{ch: make(chan time.Time, 1)}

	runtime, err := openWithDependencies(Config{
		Identity:       ServiceIdentity{Name: "game-server"},
		File:           validFileConfig(directory),
		Flush:          FlushPolicy{Interval: time.Second},
		FileEnabled:    true,
		ConsoleEnabled: false,
	}, flushTestDependencies(clock, sink, ticker))
	if err != nil {
		t.Fatalf("openWithDependencies() error = %v", err)
	}

	ticker.ch <- clock.now()
	waitForCondition(t, "periodic sync failure", func() bool { return runtime.Status().FailureCount == 1 })

	status := runtime.Status()
	if !status.Degraded {
		t.Fatal("runtime is not degraded after flush failure")
	}
	if status.FailureCount != 1 {
		t.Fatalf("failure count after flush = %d, want 1", status.FailureCount)
	}
	if status.LastError != flushErr.Error() {
		t.Fatalf("last error after flush = %q, want %q", status.LastError, flushErr.Error())
	}
	if err := runtime.Close(); !errors.Is(err, flushErr) {
		t.Fatalf("Close() error = %v, want %v", err, flushErr)
	}
	status = runtime.Status()
	if status.FailureCount != 2 {
		t.Fatalf("failure count after Close = %d, want 2", status.FailureCount)
	}
	if status.LastError != flushErr.Error() {
		t.Fatalf("last error after Close = %q, want %q", status.LastError, flushErr.Error())
	}
}

func TestZeroFlushIntervalDisablesPeriodicLoop(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "logs")
	clock := &fakeClock{current: time.Date(2026, 7, 14, 13, 30, 0, 0, time.UTC)}
	sink := &trackedWriteCloser{}
	called := false

	deps := testRuntimeDependencies(clock)
	deps.openFile = func(string, int, os.FileMode) (io.WriteCloser, error) {
		return sink, nil
	}
	deps.newTicker = func(time.Duration) runtimeTicker {
		called = true
		return &fakeTicker{ch: make(chan time.Time, 1)}
	}

	runtime, err := openWithDependencies(Config{
		Identity:       ServiceIdentity{Name: "game-server"},
		File:           validFileConfig(directory),
		Flush:          FlushPolicy{Interval: 0},
		FileEnabled:    true,
		ConsoleEnabled: false,
	}, deps)
	if err != nil {
		t.Fatalf("openWithDependencies() error = %v", err)
	}

	if called {
		t.Fatal("newTicker called for zero flush interval")
	}
	if err := runtime.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if sink.syncCount != 1 {
		t.Fatalf("sync count after Close = %d, want 1", sink.syncCount)
	}
	if sink.closeCount != 1 {
		t.Fatalf("close count after Close = %d, want 1", sink.closeCount)
	}
}
