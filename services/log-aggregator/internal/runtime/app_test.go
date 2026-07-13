package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/config"
	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/storage"
	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/storage/storagefake"
)

type countingStore struct {
	storage.EventStore
	mu     sync.Mutex
	closes int
}

func (s *countingStore) Close() error {
	s.mu.Lock()
	s.closes++
	s.mu.Unlock()
	return s.EventStore.Close()
}
func (s *countingStore) closeCount() int { s.mu.Lock(); defer s.mu.Unlock(); return s.closes }

func lifecycleLogger(output *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(output, nil))
}
func lifecycleEvents(output *bytes.Buffer) []map[string]any {
	var events []map[string]any
	for _, line := range bytes.Split(bytes.TrimSpace(output.Bytes()), []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		var event map[string]any
		_ = json.Unmarshal(line, &event)
		events = append(events, event)
	}
	return events
}

func appConfig() config.Config {
	return config.Config{ListenAddress: "127.0.0.1:0", ServiceInstanceID: "i", BuildVersion: "v", Environment: "test", ReadHeaderTimeout: time.Second, ReadTimeout: time.Second, WriteTimeout: time.Second, IdleTimeout: time.Second, ShutdownTimeout: time.Second}
}
func appDependencies() (*countingStore, Dependencies) {
	store := &countingStore{EventStore: storagefake.New()}
	d := dependencies()
	d.Store = store
	return store, d
}

func TestAppBindFailureDoesNotBecomeReadyAndClosesStore(t *testing.T) {
	store, deps := appDependencies()
	var output bytes.Buffer
	app, err := NewApp(appConfig(), deps, func(string, string) (net.Listener, error) { return nil, errors.New("address in use") }, lifecycleLogger(&output))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.Run(context.Background()); err == nil || !strings.Contains(err.Error(), "bind") {
		t.Fatalf("error = %v", err)
	}
	if deps.Health.Snapshot().Ready || store.closeCount() != 1 {
		t.Fatalf("unexpected lifecycle: %+v closes=%d", deps.Health.Snapshot(), store.closeCount())
	}
	events := lifecycleEvents(&output)
	if len(events) != 2 || events[0]["event"] != "service_starting" || events[1]["event"] != "service_startup_failed" || events[1]["error_code"] != "bind_failed" {
		t.Fatalf("events = %+v", events)
	}
}

func TestAppReadinessOnlyAfterListenerCreation(t *testing.T) {
	_, deps := appDependencies()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	created := make(chan struct{})
	app, err := NewApp(appConfig(), deps, func(network, address string) (net.Listener, error) {
		listener, listenErr := net.Listen(network, address)
		if listenErr == nil {
			close(created)
		}
		return listener, listenErr
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	finished := make(chan error, 1)
	go func() { finished <- app.Run(ctx) }()
	select {
	case <-created:
	case <-time.After(time.Second):
		t.Fatal("listener was not created")
	}
	if !deps.Health.Snapshot().Ready {
		t.Fatal("listener creation did not mark ready")
	}
	cancel()
	if err := <-finished; err != nil {
		t.Fatal(err)
	}
}

func TestAppCancellationShutsDownAndClosesOnce(t *testing.T) {
	store, deps := appDependencies()
	var output bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	app, err := NewApp(appConfig(), deps, nil, lifecycleLogger(&output))
	if err != nil {
		t.Fatal(err)
	}
	finished := make(chan error, 1)
	go func() { finished <- app.Run(ctx) }()
	time.Sleep(20 * time.Millisecond)
	cancel()
	if err := <-finished; err != nil {
		t.Fatal(err)
	}
	if snapshot := deps.Health.Snapshot(); snapshot.Ready || !snapshot.Stopping {
		t.Fatalf("unexpected state: %+v", snapshot)
	}
	if store.closeCount() != 1 {
		t.Fatalf("store closed %d times", store.closeCount())
	}
	events := lifecycleEvents(&output)
	if len(events) != 4 || events[0]["event"] != "service_starting" || events[1]["event"] != "service_ready" || events[2]["event"] != "service_stopping" || events[3]["event"] != "service_stopped" {
		t.Fatalf("events = %+v", events)
	}
}
func TestAppRejectsSecondRun(t *testing.T) {
	_, deps := appDependencies()
	app, err := NewApp(appConfig(), deps, func(string, string) (net.Listener, error) { return nil, errors.New("bind failed") }, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := app.Run(context.Background()); err == nil {
		t.Fatal("first run unexpectedly succeeded")
	}
	if err := app.Run(context.Background()); err == nil || !strings.Contains(err.Error(), "only be run once") {
		t.Fatalf("second run error = %v", err)
	}
}

type failingListener struct{}

func (failingListener) Accept() (net.Conn, error) { return nil, errors.New("serve failed") }
func (failingListener) Close() error              { return nil }
func (failingListener) Addr() net.Addr            { return &net.TCPAddr{} }

func TestAppUnexpectedServeFailureIsReturned(t *testing.T) {
	store, deps := appDependencies()
	var output bytes.Buffer
	app, err := NewApp(appConfig(), deps, func(string, string) (net.Listener, error) { return failingListener{}, nil }, lifecycleLogger(&output))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.Run(context.Background()); err == nil || !strings.Contains(err.Error(), "serve") {
		t.Fatalf("error = %v", err)
	}
	if snapshot := deps.Health.Snapshot(); snapshot.Ready || !snapshot.Stopping {
		t.Fatalf("unexpected health state: %+v", snapshot)
	}
	if store.closeCount() != 1 {
		t.Fatalf("store closed %d times", store.closeCount())
	}
	events := lifecycleEvents(&output)
	if len(events) != 3 || events[0]["event"] != "service_starting" || events[1]["event"] != "service_ready" || events[2]["event"] != "service_stopped" || events[2]["level"] != "ERROR" || events[2]["error_code"] != "serve_failed" {
		t.Fatalf("events = %+v", events)
	}
}
