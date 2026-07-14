package servicelog

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
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

func TestOpenRejectsFileOutput(t *testing.T) {
	runtime, err := Open(Config{
		Identity: ServiceIdentity{Name: "game-server"},
		File: FilePolicy{
			Directory:         "logs",
			Prefix:            "game-server",
			SegmentMaxBytes:   1,
			SegmentMaxAge:     time.Hour,
			RetentionMaxAge:   time.Hour,
			RetentionMaxBytes: 1,
		},
		FileEnabled: true,
	})
	if !errors.Is(err, ErrFileOutputUnsupported) {
		t.Fatalf("Open() error = %v, want ErrFileOutputUnsupported", err)
	}
	if runtime != nil {
		t.Fatal("Open() returned runtime for unsupported file output")
	}
}

func TestRuntimeCloseIsIdempotent(t *testing.T) {
	runtime, err := Open(Config{
		Identity:       ServiceIdentity{Name: "game-server"},
		ConsoleEnabled: true,
	})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if runtime.Status().Closed {
		t.Fatal("runtime starts closed")
	}
	if err := runtime.Close(); err != nil {
		t.Fatalf("first Close() error = %v", err)
	}
	if err := runtime.Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
	if !runtime.Status().Closed {
		t.Fatal("runtime remains open after Close()")
	}
}
