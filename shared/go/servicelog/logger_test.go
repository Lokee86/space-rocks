package servicelog

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestOpenAttachesServiceContext(t *testing.T) {
	var output bytes.Buffer
	clock := &fakeClock{current: time.Unix(100, 0)}
	runtime, err := openWithDependencies(Config{
		Identity: ServiceIdentity{
			Name:        "game-server",
			InstanceID:  "instance-1",
			Environment: "test",
			Version:     "dev",
		},
		ConsoleEnabled: true,
	}, fakeFilesystem{}.dependencies(&output, clock))
	if err != nil {
		t.Fatalf("openWithDependencies() error = %v", err)
	}
	runtime.Logger().Info("started")

	text := output.String()
	for _, field := range []string{"service=game-server", "service_instance_id=instance-1", "environment=test", "build_version=dev"} {
		if !strings.Contains(text, field) {
			t.Errorf("log output missing %s: %s", field, text)
		}
	}
}

func TestFanoutForwardsRecordsAndConfiguration(t *testing.T) {
	var first, second bytes.Buffer
	fanout := newFanoutHandler(
		slog.NewJSONHandler(&first, nil),
		slog.NewJSONHandler(&second, nil),
	)
	logger := slog.New(fanout).WithGroup("event").With(slog.String("kind", "startup"))
	logger.Info("ready")

	if first.Len() == 0 || second.Len() == 0 {
		t.Fatalf("fanout outputs = %d and %d bytes, want both non-empty", first.Len(), second.Len())
	}
	for name, output := range map[string]string{"first": first.String(), "second": second.String()} {
		if !strings.Contains(output, `"event":{"kind":"startup"}`) {
			t.Errorf("%s output missing grouped attribute: %s", name, output)
		}
	}
	if !fanout.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("fanout Enabled() = false, want true")
	}
}

func TestOpenWritesJSONLFileOutput(t *testing.T) {
	dir := t.TempDir()
	runtime, err := openWithDependencies(Config{
		Identity: ServiceIdentity{
			Name:        "game-server",
			InstanceID:  "instance-1",
			Environment: "test",
			Version:     "dev",
		},
		File: FilePolicy{
			Directory:         dir,
			Prefix:            "game-server",
			SegmentMaxBytes:   1,
			SegmentMaxAge:     time.Hour,
			RetentionMaxAge:   time.Hour,
			RetentionMaxBytes: 1,
		},
		FileEnabled: true,
	}, defaultRuntimeDependencies())
	if err != nil {
		t.Fatalf("openWithDependencies() error = %v", err)
	}

	runtime.Logger().Info("started")
	if runtime.Status().Closed {
		t.Fatal("runtime starts closed")
	}
	if err := runtime.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "active", "game-server.jsonl.open"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(data)
	for _, field := range []string{`"service":"game-server"`, `"msg":"started"`, `"build_version":"dev"`} {
		if !strings.Contains(text, field) {
			t.Fatalf("file output missing %s: %s", field, text)
		}
	}
}

func TestOpenFansOutToConsoleAndFile(t *testing.T) {
	dir := t.TempDir()
	var console bytes.Buffer
	deps := defaultRuntimeDependencies()
	deps.consoleWriter = &console

	runtime, err := openWithDependencies(Config{
		Identity: ServiceIdentity{Name: "game-server", Environment: "test"},
		File: FilePolicy{
			Directory:         dir,
			Prefix:            "game-server",
			SegmentMaxBytes:   1,
			SegmentMaxAge:     time.Hour,
			RetentionMaxAge:   time.Hour,
			RetentionMaxBytes: 1,
		},
		ConsoleEnabled: true,
		FileEnabled:    true,
	}, deps)
	if err != nil {
		t.Fatalf("openWithDependencies() error = %v", err)
	}
	defer func() {
		if err := runtime.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}()

	runtime.Logger().Info("started")
	if err := runtime.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	if !strings.Contains(console.String(), "service=game-server") {
		t.Fatalf("console output missing service identity: %s", console.String())
	}
	if !strings.Contains(console.String(), "msg=started") {
		t.Fatalf("console output missing message: %s", console.String())
	}

	data, err := os.ReadFile(filepath.Join(dir, "active", "game-server.jsonl.open"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(data)
	for _, field := range []string{`"msg":"started"`, `"service":"game-server"`, `"environment":"test"`} {
		if !strings.Contains(text, field) {
			t.Fatalf("file output missing %s: %s", field, text)
		}
	}
}

func TestRuntimeCloseFlushesFileOutputAndIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	runtime, err := openWithDependencies(Config{
		Identity: ServiceIdentity{Name: "game-server", Version: "dev"},
		File: FilePolicy{
			Directory:         dir,
			Prefix:            "game-server",
			SegmentMaxBytes:   1,
			SegmentMaxAge:     time.Hour,
			RetentionMaxAge:   time.Hour,
			RetentionMaxBytes: 1,
		},
		FileEnabled: true,
	}, defaultRuntimeDependencies())
	if err != nil {
		t.Fatalf("openWithDependencies() error = %v", err)
	}

	runtime.Logger().Info("closing")
	if err := runtime.Close(); err != nil {
		t.Fatalf("first Close() error = %v", err)
	}
	if err := runtime.Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
	if !runtime.Status().Closed {
		t.Fatal("runtime remains open after Close()")
	}

	data, err := os.ReadFile(filepath.Join(dir, "active", "game-server.jsonl.open"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(data)
	for _, field := range []string{`"msg":"closing"`, `"build_version":"dev"`} {
		if !strings.Contains(text, field) {
			t.Fatalf("file output missing %s: %s", field, text)
		}
	}
}

func TestOpenReportsActiveStatusAndRunsRetention(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, time.July, 14, 15, 4, 5, 0, time.UTC)
	stale := archiveSegmentPath(dir, "game-server", now.Add(-2*time.Hour), now.Add(-2*time.Hour), 1)
	writeArchiveSegmentFile(t, stale, []byte("old\n"), now.Add(-2*time.Hour))

	clock := &fakeClock{current: now}
	deps := defaultRuntimeDependencies()
	deps.now = clock.now
	runtime, err := openWithDependencies(Config{
		Identity: ServiceIdentity{Name: "game-server"},
		File: FilePolicy{
			Directory:         dir,
			Prefix:            "game-server",
			SegmentMaxBytes:   1,
			SegmentMaxAge:     time.Hour,
			RetentionMaxAge:   time.Hour,
			RetentionMaxBytes: 1024,
		},
		FileEnabled: true,
	}, deps)
	if err != nil {
		t.Fatalf("openWithDependencies() error = %v", err)
	}
	defer runtime.Close()

	status := runtime.Status()
	if !status.FileEnabled || status.FileDegraded || status.FileFailureCount != 0 || status.Closed {
		t.Fatalf("status = %#v, want healthy enabled runtime", status)
	}
	if status.ActivePath != filepath.Join(dir, "active", "game-server.jsonl.open") {
		t.Fatalf("ActivePath = %q", status.ActivePath)
	}
	if status.ActiveBytes != 0 {
		t.Fatalf("ActiveBytes = %d, want 0", status.ActiveBytes)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("stale archive still exists: %v", err)
	}
}

func TestOpenLeavesRuntimeDegradedWhenRetentionCleanupFails(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, time.July, 14, 15, 4, 5, 0, time.UTC)
	stale := archiveSegmentPath(dir, "game-server", now.Add(-2*time.Hour), now.Add(-2*time.Hour), 1)
	writeArchiveSegmentFile(t, stale, []byte("old\n"), now.Add(-2*time.Hour))

	clock := &fakeClock{current: now}
	deps := defaultRuntimeDependencies()
	deps.now = clock.now
	deps.remove = func(path string) error {
		if path == stale {
			return errors.New("retention cleanup failed")
		}
		return os.Remove(path)
	}

	runtime, err := openWithDependencies(Config{
		Identity: ServiceIdentity{Name: "game-server"},
		File: FilePolicy{
			Directory:         dir,
			Prefix:            "game-server",
			SegmentMaxBytes:   1,
			SegmentMaxAge:     time.Hour,
			RetentionMaxAge:   time.Hour,
			RetentionMaxBytes: 1024,
		},
		FileEnabled: true,
	}, deps)
	if err != nil {
		t.Fatalf("openWithDependencies() error = %v", err)
	}
	defer runtime.Close()

	status := runtime.Status()
	if !status.FileEnabled || !status.FileDegraded || status.FileFailureCount != 1 || status.Closed {
		t.Fatalf("status = %#v, want degraded enabled runtime", status)
	}
	if _, err := os.Stat(stale); err != nil {
		t.Fatalf("stale archive missing after failed retention cleanup: %v", err)
	}
}

func TestRuntimeStatusRecordsRotationFailures(t *testing.T) {
	dir := t.TempDir()
	clock := &fakeClock{current: time.Date(2026, time.July, 14, 15, 4, 5, 0, time.UTC)}
	deps := defaultRuntimeDependencies()
	deps.now = clock.now
	deps.rename = func(oldPath, newPath string) error {
		if filepath.Dir(newPath) == filepath.Join(dir, "archive", "2026", "07", "14") {
			return errors.New("rotation rename failed")
		}
		return os.Rename(oldPath, newPath)
	}

	runtime, err := openWithDependencies(Config{
		Identity: ServiceIdentity{Name: "game-server"},
		File: FilePolicy{
			Directory:         dir,
			Prefix:            "game-server",
			SegmentMaxBytes:   1,
			SegmentMaxAge:     time.Hour,
			RetentionMaxAge:   time.Hour,
			RetentionMaxBytes: 1024,
		},
		FileEnabled: true,
	}, deps)
	if err != nil {
		t.Fatalf("openWithDependencies() error = %v", err)
	}
	defer runtime.Close()

	if _, err := runtime.fileWriter.Write([]byte("first\n")); err != nil {
		t.Fatalf("first Write() error = %v", err)
	}
	before := runtime.Status()
	if before.ActiveBytes == 0 || before.FileFailureCount != 0 || before.FileDegraded {
		t.Fatalf("status before failure = %#v, want active file only", before)
	}

	if _, err := runtime.fileWriter.Write([]byte("second\n")); err == nil {
		t.Fatal("second Write() error = nil, want rotation failure")
	}
	after := runtime.Status()
	if !after.FileDegraded || after.FileFailureCount != 1 || after.ActiveBytes != before.ActiveBytes {
		t.Fatalf("status after failure = %#v, want one recorded failure", after)
	}
}

func TestRuntimeMaintenanceLoopRotatesAgeExpiredSegmentAndRunsRetention(t *testing.T) {
	dir := t.TempDir()
	startedAt := time.Date(2026, time.July, 14, 15, 4, 5, 0, time.UTC)
	tickAt := startedAt.Add(2 * time.Hour)
	staleModTime := startedAt.Add(-3 * time.Hour)
	stalePath := archiveSegmentPath(dir, "game-server", staleModTime, staleModTime, 1)
	writeArchiveSegmentFile(t, stalePath, []byte("stale\n"), staleModTime)

	ticker := newManualMaintenanceTicker()
	deps := defaultRuntimeDependencies()
	deps.now = func() time.Time { return startedAt }
	deps.newTicker = func(time.Duration) maintenanceTicker { return ticker }

	runtime, err := openWithDependencies(Config{
		Identity: ServiceIdentity{Name: "game-server"},
		File: FilePolicy{
			Directory:         dir,
			Prefix:            "game-server",
			SegmentMaxBytes:   1024,
			SegmentMaxAge:     time.Hour,
			RetentionMaxAge:   time.Hour,
			RetentionMaxBytes: 1024,
		},
		Flush:      FlushPolicy{Interval: time.Second},
		FileEnabled: true,
	}, deps)
	if err != nil {
		t.Fatalf("openWithDependencies() error = %v", err)
	}
	defer runtime.Close()

	if _, err := runtime.fileWriter.Write([]byte("line one\n")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	tickSent := make(chan struct{})
	go func() {
		ticker.Tick(tickAt)
		close(tickSent)
	}()
	<-tickSent

	if err := runtime.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	archivePath := archiveSegmentPath(dir, "game-server", startedAt, tickAt, 1)
	archiveData, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("ReadFile(archive) error = %v", err)
	}
	if !strings.Contains(string(archiveData), "line one") {
		t.Fatalf("rotated archive contents = %q, want line one", string(archiveData))
	}
	if _, err := os.Stat(stalePath); !os.IsNotExist(err) {
		t.Fatalf("stale archive still exists: %v", err)
	}
	status := runtime.Status()
	if !status.Closed || !status.FileEnabled || status.FileDegraded {
		t.Fatalf("status after maintenance close = %#v", status)
	}
}

func TestRuntimeMaintenanceFailureUpdatesStatus(t *testing.T) {
	dir := t.TempDir()
	startedAt := time.Date(2026, time.July, 14, 15, 4, 5, 0, time.UTC)
	tickAt := startedAt.Add(2 * time.Hour)
	ticker := newManualMaintenanceTicker()
	renameCalled := make(chan struct{})
	var renameOnce sync.Once

	deps := defaultRuntimeDependencies()
	deps.now = func() time.Time { return startedAt }
	deps.newTicker = func(time.Duration) maintenanceTicker { return ticker }
	deps.rename = func(oldPath, newPath string) error {
		renameOnce.Do(func() { close(renameCalled) })
		return errors.New("maintenance rotation failed")
	}

	runtime, err := openWithDependencies(Config{
		Identity: ServiceIdentity{Name: "game-server"},
		File: FilePolicy{
			Directory:         dir,
			Prefix:            "game-server",
			SegmentMaxBytes:   1024,
			SegmentMaxAge:     time.Hour,
			RetentionMaxAge:   time.Hour,
			RetentionMaxBytes: 1024,
		},
		Flush:      FlushPolicy{Interval: time.Second},
		FileEnabled: true,
	}, deps)
	if err != nil {
		t.Fatalf("openWithDependencies() error = %v", err)
	}
	defer runtime.Close()

	runtime.Logger().Info("maintenance")
	go ticker.Tick(tickAt)
	<-renameCalled

	if err := runtime.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	status := runtime.Status()
	if !status.FileDegraded || status.FileFailureCount != 1 {
		t.Fatalf("status = %#v, want one degraded maintenance failure", status)
	}
}

func TestRuntimeCloseWaitsForMaintenanceLoop(t *testing.T) {
	dir := t.TempDir()
	startedAt := time.Date(2026, time.July, 14, 15, 4, 5, 0, time.UTC)
	tickAt := startedAt.Add(2 * time.Hour)
	ticker := newManualMaintenanceTicker()
	renameEntered := make(chan struct{})
	releaseRename := make(chan struct{})
	var renameOnce sync.Once

	deps := defaultRuntimeDependencies()
	deps.now = func() time.Time { return startedAt }
	deps.newTicker = func(time.Duration) maintenanceTicker { return ticker }
	deps.rename = func(oldPath, newPath string) error {
		renameOnce.Do(func() { close(renameEntered) })
		<-releaseRename
		return nil
	}

	runtime, err := openWithDependencies(Config{
		Identity: ServiceIdentity{Name: "game-server"},
		File: FilePolicy{
			Directory:         dir,
			Prefix:            "game-server",
			SegmentMaxBytes:   1024,
			SegmentMaxAge:     time.Hour,
			RetentionMaxAge:   time.Hour,
			RetentionMaxBytes: 1024,
		},
		Flush:      FlushPolicy{Interval: time.Second},
		FileEnabled: true,
	}, deps)
	if err != nil {
		t.Fatalf("openWithDependencies() error = %v", err)
	}

	runtime.Logger().Info("maintenance")
	go ticker.Tick(tickAt)
	<-renameEntered

	closeReturned := make(chan error, 1)
	go func() {
		closeReturned <- runtime.Close()
	}()

	select {
	case err := <-closeReturned:
		t.Fatalf("Close() returned early: %v", err)
	default:
	}

	close(releaseRename)
	if err := <-closeReturned; err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

type manualMaintenanceTicker struct {
	ch       chan time.Time
	stopOnce sync.Once
	stopped  chan struct{}
}

func newManualMaintenanceTicker() *manualMaintenanceTicker {
	return &manualMaintenanceTicker{ch: make(chan time.Time), stopped: make(chan struct{})}
}

func (t *manualMaintenanceTicker) C() <-chan time.Time { return t.ch }

func (t *manualMaintenanceTicker) Stop() {
	t.stopOnce.Do(func() { close(t.stopped) })
}

func (t *manualMaintenanceTicker) Tick(ts time.Time) {
	t.ch <- ts
}
